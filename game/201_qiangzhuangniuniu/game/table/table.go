package table

import (
	"encoding/json"
	"fmt"
	"math"
	"xj_game_server/game/201_qiangzhuangniuniu/game/logic"
	"xj_game_server/game/public/redis"

	"net"
	"sync"
	"xj_game_server/game/201_qiangzhuangniuniu/conf"
	"xj_game_server/game/201_qiangzhuangniuniu/global"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/log"
	rand "xj_game_server/util/leaf/util"

	"time"
)

/*
抢庄牛牛
*/
var List = make([]*Item, 0)

// 匹配池
var queuePool *Queue

// 匹配列表
type Queue struct {
	UserQueue  sync.Map // 真人等待列表
	RobotQueue sync.Map //机器人等待列表
	RobotCount int32    //机器人个数
	UserCount  int32    //真人个数
	cronTimer  *time.Timer
	//TablesList sync.Map //桌子列表 人数 桌子号,桌子 key int32  map[int32]*Item
}

func init() {
	queuePool = &Queue{
		UserCount:  0,
		RobotCount: 0,
		UserQueue:  sync.Map{},
		RobotQueue: sync.Map{},
		//	TablesList: sync.Map{},
	}
}

type Item struct {
	tableID        int32                              //桌子号
	drawID         string                             //游戏记录ID
	gameStatus     int32                              //游戏状态
	sceneStartTime int64                              //场景开始时间
	userList       sync.Map                           //玩家列表 座位号-->uid map[int32]int32
	userCount      int32                              //玩家数量
	androidCount   int32                              //机器人数量
	userPlaying    map[int32]bool                     //用户是否游戏中	座位号map[int32]bool
	userListQZ     sync.Map                           //用户抢庄	座位号	map[int32]int32
	userListJetton sync.Map                           //玩家下注 座位号 map[int32]int32
	userListTP     sync.Map                           //玩家摊牌	座位号 map[int32]bool
	userListPoker  map[int32]*msg.Game_S_LotteryPoker //玩家扑克	每个用户5张扑克
	userListLoss   map[int32]float32                  //用户盈亏 座位号
	userTax        map[int32]float32                  //用户税收
	systemScore    float32                            //系统盈亏
	bankerChairID  int32                              //庄家椅子号
	bankerMultiple int32                              //庄家抢庄倍数
	cronTimer      *time.Timer                        //定时任务
	tablePlayCount int32                              // 桌子人数
	queueToSitDown chan *user.Item                    //用户信息 坐下队列
	mutex          sync.Mutex                         // 队列锁
	roundOrder     string                             `json:"round_order"` // 局号
}

//初始化桌子
func OnInit() {
	//初始化桌子 	for i := int32(0); i < store.GameControl.GetGameInfo().TableCount; i++
	for i := int32(0); i < 4; i++ {
		temp := &Item{
			tableID:        i,
			drawID:         "",
			gameStatus:     0,
			sceneStartTime: 0,
			userList:       sync.Map{},
			userCount:      0,
			androidCount:   0,
			userPlaying:    map[int32]bool{},
			userListQZ:     sync.Map{},
			userListJetton: sync.Map{},
			userListTP:     sync.Map{},
			userListPoker:  make(map[int32]*msg.Game_S_LotteryPoker),
			userListLoss:   make(map[int32]float32),
			userTax:        make(map[int32]float32),
			systemScore:    0,
			bankerChairID:  -1,
			bankerMultiple: 0,
			cronTimer:      &time.Timer{},
			queueToSitDown: make(chan *user.Item, store.GameControl.GetGameInfo().ChairCount),
		}
		List = append(List, temp)
	}
	//	queuePool.TablesList.Store(int32(0), tempTables)
	//go initMatchTable()
	go actionMatchTable()

}

// 动态添加桌子
func appendTableList() {
	n := len(List)
	size := int32(n * 2)
	for i := int32(n); i < size; i++ {
		temp := &Item{
			tableID:        i,
			drawID:         "",
			gameStatus:     0,
			sceneStartTime: 0,
			userList:       sync.Map{},
			userCount:      0,
			androidCount:   0,
			userPlaying:    map[int32]bool{},
			userListQZ:     sync.Map{},
			userListJetton: sync.Map{},
			userListTP:     sync.Map{},
			userListPoker:  make(map[int32]*msg.Game_S_LotteryPoker),
			userListLoss:   make(map[int32]float32),
			userTax:        make(map[int32]float32),
			systemScore:    0,
			bankerChairID:  -1,
			bankerMultiple: 0,
			cronTimer:      &time.Timer{},
			queueToSitDown: make(chan *user.Item, store.GameControl.GetGameInfo().ChairCount),
		}
		List = append(List, temp)
	}
}

// 删除一半没有使用的桌子
func delTableList() {
	if queuePool.UserCount != 0 || len(List) <= 4 {
		return
	}
	nlen := len(List) / 2
	var isFree = true
	for i := nlen; i < len(List); i++ {
		if List[i].gameStatus != global.GameStatusFree {
			isFree = false
			break
		}
	}
	if isFree {
		List = List[:nlen]
	}
}

func GetQueuePool() *Queue {
	return queuePool
}

// 加入匹配队列
func ADDQueueInfo(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if !userItem.IsRobot() {
		log.Logger.Debugf("ADDQueueInfo--用户坐下,加入匹配队列 uid:%v,chairID:%v,queuePoolCount:%v,robotCount:%v,桌子数:%v", userItem.UserID, userItem.ChairID, queuePool.UserCount, queuePool.RobotCount, len(List))
		time.Sleep(time.Millisecond * time.Duration(rand.RandInterval(1000, 3000)))
		if userItem.Status != user.StatusFree {
			return
		}
		if _, ok := queuePool.UserQueue.Load(userItem.UserID); !ok {
			queuePool.UserCount++
		}
		queuePool.UserQueue.Store(userItem.UserID, userItem)
		for len(noticeSignal) >= 1 || queuePool.UserCount > 1 {
			time.Sleep(time.Millisecond * 500)
		}
		noticeSignal <- struct{}{}
	} else {
		queuePool.RobotQueue.Store(userItem.UserID, userItem)
		queuePool.RobotCount++
	}
}

// 获取空闲的座位 优先匹配有真人的房间
/*
func getFreeTableID() (int32, int32) {
	var tableID int32 = -1
	var IsBreak = false
	var userCount int32
	for i := store.GameControl.GetGameInfo().ChairCount - 1; i >= 0; i-- {
		tempList, ok := queuePool.TablesList.Load(i)
		if !ok {
			continue
		}
		for k, _ := range tempList.(map[int32]*Item) {
			userCount = List[k].userCount - List[k].androidCount
			if i > 0 && userCount > 0 && List[k].gameStatus == global.GameStatusFree {
				tableID = k
				IsBreak = true
				break
			} else if i == 0 {
				IsBreak = true
				tableID = k
				break
			}
		}
		if IsBreak {
			break
		}
	}
	return tableID, userCount
}

*/
// 用户进入房间通知消息
var (
	// 通知匹配
	noticeSignal = make(chan struct{}, 10000)
	// 通知机器人上线 globao.Notice
	//noticeRobot = make(chan struct{}, 10000)
)

// 玩家进入房间匹配桌子
func actionMatchTable() {
	for {
		select {
		case <-noticeSignal:
			if queuePool.UserCount == 0 {
				continue
			}
			var tableId int32 = -1
			for key, _ := range List {
				tableId = List[key].GetFreeTableID()
				if tableId < 0 {
					continue
				} else {
					go List[key].ConsumePlayer2Table()
					break
				}
			}
			if tableId < 0 {
				appendTableList()

				noticeSignal <- struct{}{}

				continue
			}
			// 随机2-5人开局  机器人至少2-5
			playCount := List[tableId].tablePlayCount
			if playCount == 0 {
				playCount = rand.RandInterval(global.TablePlayCount+1, store.GameControl.GetGameInfo().ChairCount)
				List[tableId].tablePlayCount = playCount
			}
			haveRootCount := playCount - queuePool.UserCount
			log.Logger.Debugf("tableId:%v,机器人人数:%v,匹配队列真人人数:%v,,需要匹配机器人:%v,rand:%v", tableId, queuePool.RobotCount, queuePool.UserCount, haveRootCount, playCount)

			if queuePool.RobotCount < haveRootCount {
				global.NoticeRobotOnline <- haveRootCount - queuePool.RobotCount

				continue
			}
			// 真人坐下
			var userCont int32
			queuePool.UserQueue.Range(func(userID, userItem interface{}) bool {

				_, ok := user.List.Load(userItem.(*user.Item).UserID)
				// 离线状态  或者玩家退出游戏
				if userItem.(*user.Item).Status == user.StatusOffline || !ok {
					queuePool.UserCount--
					queuePool.UserQueue.Delete(userItem.(*user.Item).UserID)
					return true
				}
				if List[tableId].tablePlayCount-userCont <= 0 {
					return false
				}
				List[tableId].AppendToQueue(userItem.(*user.Item))
				userCont++
				queuePool.UserQueue.Delete(userItem.(*user.Item).UserID)
				queuePool.UserCount--
				return true
			})
			if haveRootCount > 0 && tableId >= 0 {
				var roobotCont int32 = 0
				// 机器人入场
				queuePool.RobotQueue.Range(func(userID, userItem interface{}) bool {
					if List[tableId].RoomRobotFull() {
						return false
					}
					if roobotCont >= haveRootCount {
						return false
					}

					_, ok := user.List.Load(userItem.(*user.Item).UserID)
					// 离线状态
					if userItem.(*user.Item).Status == user.StatusOffline || !ok {
						queuePool.RobotQueue.Delete(userItem.(*user.Item).UserID)
						queuePool.RobotCount--
						return true
					}
					List[tableId].AppendToQueue(userItem.(*user.Item))
					roobotCont++
					queuePool.RobotQueue.Delete(userItem.(*user.Item).UserID)
					queuePool.RobotCount--
					return true
				})
			}

		case loadNum := <-global.NoticeLoadMath:
			for queuePool.RobotCount < loadNum || queuePool.UserCount > 1 {
				time.Sleep(time.Millisecond * 100)
			}
			noticeSignal <- struct{}{}
		}
	}
}

/*
//匹配桌子
func initMatchTable() {

	//开启匹配桌子 线程
	for {
		queuePool.cronTimer = time.NewTimer(time.Second * 3)
		select {
		case <-queuePool.cronTimer.C:
			var tableId int32
			var roomUserCount int32 // 房间里的真人数量
			tableId, roomUserCount = getFreeTableID()

			if (tableId >= 0 && List[tableId].userCount+(queuePool.RobotCount+queuePool.UserCount) < global.TablePlayCount) || queuePool.UserCount == 0 {
				continue
			}
			if tableId < 0 {
				queuePool.UserQueue.Range(func(key, value interface{}) bool {
					_ = log.Logger.Errorf("handlerUserSitDown %s---uid:%v status:%d ", "坐下失败, 没有空闲的桌子", value.(*user.Item).UserID, value.(*user.Item).Status)
					return false
				})
				continue
			}

			// 随机2-5人开局  机器人至少2-5
			//playCount := rand.RandInterval(global.TablePlayCount, store.GameControl.GetGameInfo().ChairCount)
			playCount := rand.RandInterval(queuePool.UserCount, store.GameControl.GetGameInfo().ChairCount)
			List[tableId].tablePlayCount = playCount

			haveRootCount := playCount - queuePool.UserCount //+ roomUserCount%playCount
			log.Logger.Debugf("tableId:%v,机器人人数:%v,匹配队列真人人数:%v,房间待匹配真人:%v,需要匹配机器人:%v,rand:%v", tableId, queuePool.RobotCount, queuePool.UserCount, roomUserCount, haveRootCount, playCount)

			//算出差多少机器人

			var userCont int32 = 0
			// 真人坐下
			queuePool.UserQueue.Range(func(userID, userItem interface{}) bool {
				// 离线状态
				if userItem.(*user.Item).Status == user.StatusOffline {
					queuePool.UserCount--
					queuePool.UserQueue.Delete(userItem.(*user.Item).UserID)
					return true
				}
				if userCont%global.TablePlayCount == 0 {
					tableId, _ = getFreeTableID()
					if tableId < 0 {
						queuePool.UserQueue.Range(func(key, value interface{}) bool {
							_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 没有空闲的桌子", value.(*user.Item).Status)
							return true
						})
						return false
					}
				}
				//queuePool.UserQueue.Delete(userID)
				//queuePool.UserCount--
				//userCont++
				go List[tableId].OnActionUserSitDown(userItem.(*user.Item))
				if List[tableId].gameStatus == global.GameStatusFree {
					userCont++
				}
				return true
			})
			if haveRootCount > 0 && tableId >= 0 {
				var roobotCont int32 = 0
				// 机器人入场
				queuePool.RobotQueue.Range(func(userID, userItem interface{}) bool {
					if List[tableId].RoomRobotFull() {
						return false
					}
					if roobotCont >= haveRootCount {
						return false
					}
					// 离线状态
					if userItem.(*user.Item).Status == user.StatusOffline {
						queuePool.RobotQueue.Delete(userItem.(*user.Item).UserID)
						queuePool.RobotCount--
						return true
					}
					go List[tableId].OnActionUserSitDown(userItem.(*user.Item))
					//queuePool.RobotQueue.Delete(userID)
					//queuePool.RobotCount--
					//roobotCont++
					if List[tableId].gameStatus == global.GameStatusFree {
						roobotCont++
					}
					return true
				})
			}
		}
	}
}
*/
// 更新桌子列表
/*
func (it *Item) ResetTablesList() {
	it.deleteListByTableID()
	it.lock.Lock()
	defer it.lock.Unlock()
	mapTables := make(map[int32]*Item, 0)
	mapTables[it.tableID] = it
	queuePool.TablesList.Store(it.userCount, mapTables)

}

//删除桌子列表
func (it *Item) deleteListByTableID() {
	it.lock.Lock()
	defer it.lock.Unlock()
	queuePool.TablesList.Range(func(key, value interface{}) bool {
		mapTables := value.(map[int32]*Item)
		if _, ok := mapTables[it.tableID]; ok {
			delete(mapTables, it.tableID)
		}
		return true
	})
}

*/
// 添加队列
func (it *Item) AppendToQueue(userInfo *user.Item) { // == userInfo
	// todo 限制chan 数
	it.queueToSitDown <- userInfo
}

// 处理队列
func (it *Item) ConsumePlayer2Table() {
	ConsumerFunc := func() {
		for it.gameStatus == global.GameStatusFree {
			userInfo := <-it.queueToSitDown
			it.OnActionUserSitDown(userInfo) //坐下 it.OnActionUserSitDown
		}
	}
	ConsumerFunc()
}

//定时操作 处理超时操作
func (it *Item) onEventGameTimer() {
	switch it.gameStatus {
	case global.GameStatusStart:
		it.changeGameStatus(global.GameStatusQZ, conf.GetServer().GameQZTime)
	case global.GameStatusQZ:
		// 处理超时抢庄 0倍
		for chairId, isIn := range it.userPlaying {
			if isIn {
				_, ok := it.userListQZ.Load(chairId)
				//没有抢庄
				if !ok {
					uid, okUser := it.userList.Load(chairId)
					if !okUser {
						continue
					}
					userItem, okUserItem := user.List.Load(uid)
					if !okUserItem {
						continue
					}
					it.OnUserQZ(userItem, &msg.Game_C_UserQZ{
						Multiple: int32(0),
					})
				}
			}
		}

		//定庄
		it.userListQZ.Range(func(chairID, multiple interface{}) bool {
			if it.bankerMultiple < multiple.(int32) {
				it.bankerChairID = chairID.(int32)
				it.bankerMultiple = multiple.(int32)
			} else if it.bankerMultiple == multiple.(int32) {
				userID1, okUser1 := it.userList.Load(chairID)
				if !okUser1 {
					_ = log.Logger.Errorf("")
					return true
				}
				userItem1, okUser2 := user.List.Load(userID1.(int32))
				if !okUser2 {
					//_ = log.Logger.Errorf("")
					return true
				}

				userID2, okUser3 := it.userList.Load(it.bankerChairID)
				if !okUser3 {
					//_ = log.Logger.Errorf("")
					return true
				}
				userItem2, okUser4 := user.List.Load(userID2.(int32))
				if !okUser4 {
					//_ = log.Logger.Errorf("")
					return true
				}

				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					if userItem1.(*user.Item).UserGold > userItem2.(*user.Item).UserGold {
						it.bankerChairID = chairID.(int32)
						it.bankerMultiple = multiple.(int32)
					}
				} else {
					if userItem1.(*user.Item).UserDiamond > userItem2.(*user.Item).UserDiamond {
						it.bankerChairID = chairID.(int32)
						it.bankerMultiple = multiple.(int32)
					}
				}
			}
			return true
		})
		//随机定庄
		if it.bankerChairID == -1 {
			for {
				it.bankerChairID = rand.RandInterval(0, store.GameControl.GetGameInfo().ChairCount-1)
				_, ok := it.userPlaying[it.bankerChairID]
				if ok {
					it.bankerMultiple = 1
					break
				}
			}
		}
		//定庄通知
		it.sendAllUser(&msg.Game_S_GameDZ{
			ChairID:  it.bankerChairID,
			Multiple: it.bankerMultiple,
		})
		//庄家不用下注
		it.userListJetton.Store(it.bankerChairID, int32(0))
		//休眠3秒到下个场景
		time.Sleep(2 * time.Second)
		//进入下个场景
		it.changeGameStatus(global.GameStatusJetton, conf.GetServer().GameJettonTime)

	case global.GameStatusJetton:
		// 处理下注 最低倍数
		for chairId, isIn := range it.userPlaying {
			if isIn {
				_, ok := it.userListJetton.Load(chairId)
				//没有下注
				if !ok {
					uid, okUser := it.userList.Load(chairId)
					if !okUser {
						continue
					}
					userItem, okUserItem := user.List.Load(uid)
					if !okUserItem {
						continue
					}
					it.OnUserPlaceJetton(userItem, &msg.Game_C_UserJetton{
						Multiple: conf.GetServer().JettonList[0],
					})
				}
			}
		}
		//it.userListPoker = logic.Client.DispatchTableCard(it.userPlaying)
		//it.systemScore, it.userListLoss, it.userTax, _ = logic.Client.GetSystemLoss(it.bankerChairID, it.bankerMultiple, it.userListJetton, it.userListPoker, it.userList)
		it.systemScore, it.userListLoss, it.userTax, it.userListPoker = logic.Client.GetUserSystemLoss(it.bankerChairID, it.bankerMultiple, it.userListJetton, logic.Client.DispatchTableCard(it.userPlaying), it.userList)
		//randNumber := rand.RandInterval(0, 101)
		//for {
		//	it.userListPoker = logic.Client.DispatchTableCard(it.userPlaying)
		//	it.systemScore, it.userListLoss, it.userTax = logic.Client.GetSystemLoss(it.bankerChairID, it.bankerMultiple, it.userListJetton, it.userListPoker, it.userList)
		//	// 系统库存不够的时候用户输
		//	if store.GameControl.GetStore()+it.systemScore < 0 {
		//		randNumber = int32(store.GameControl.GetUserWinRate() + 1)
		//	}
		//	if float32(randNumber) < store.GameControl.GetUserWinRate() { //用户赢
		//		if it.systemScore <= 0 && store.GameControl.GetStore()+it.systemScore >= 0 {
		//			break
		//		}
		//	} else { //用户输
		//		if it.systemScore >= 0 {
		//			break
		//		}
		//	}
		//}

		//// 提前发送自己的牌
		//for key, v := range it.userPlaying {
		//	if v {
		//		uid, okUid := it.userList.Load(key)
		//		if !okUid {
		//			continue
		//		}
		//		userItem, ok := user.List.Load(uid)
		//		if ok {
		//			var userTP msg.Game_S_UserTP
		//			userTP.ChairID = key
		//			var userTp msg.Game_S_LotteryPoker
		//			userTp.PokerType = it.userListPoker[key].PokerType & 0xf00 / 0x100
		//			userTp.LotteryPoker = it.userListPoker[key].LotteryPoker
		//			userTP.Poker = &userTp
		//			userItem.(*user.Item).WriteMsg(&userTP)
		//		}
		//	}
		//}
		//休眠3秒到下个场景
		time.Sleep(2 * time.Second)
		it.changeGameStatus(global.GameStatusTP, conf.GetServer().GameTPTime)

	case global.GameStatusTP:

		// 处理摊牌
		for chairId, isIn := range it.userPlaying {
			if isIn {
				_, ok := it.userListTP.Load(chairId)
				if !ok {
					uid, okUser := it.userList.Load(chairId)
					if !okUser {
						continue
					}
					userItem, okUserItem := user.List.Load(uid)
					if !okUserItem {
						continue
					}
					it.OnUserTP(userItem, &msg.Game_C_UserTP{})
				}
			}
		}
		//休眠3秒到下个场景
		time.Sleep(2 * time.Second)
		it.onEventGameConclude()
	default:
		fmt.Println("未知：", it.gameStatus)
		_ = log.Logger.Errorf("未知：%d", it.gameStatus)
	}
}

//获取房间最多能装多少机器人 3个
func (it *Item) RoomRobotFull() bool {
	return it.androidCount == store.GameControl.GetGameInfo().ChairCount-1
}

//开始游戏
func (it *Item) onEventGameStart() {
	//人数不足
	if it.userCount < global.TablePlayCount {
		it.gameStatus = global.GameStatusFree
		return
	}
	if len(it.userPlaying) > 0 {
		if int32(len(it.userPlaying)) != it.userCount {
			it.userList.Range(func(chairID, userID interface{}) bool {
				it.userPlaying[chairID.(int32)] = true
				return true
			})
		}
		return
	}
	//设定游戏玩家
	it.userList.Range(func(chairID, userID interface{}) bool {
		it.userPlaying[chairID.(int32)] = true
		return true
	})
	// 设定开始游戏  todo 待优化发送坐下消息
	userListMap := make(map[int32]*msg.Game_S_User, 0)
	userListArr := make([]msg.Game_S_User, 0)
	var tempPrintInt32 []int32       // 临时打印
	var tempPrintInt32Chaird []int32 // 临时打印 椅子号
	var tempPrintCount int32         // 临时打印
	it.userList.Range(func(chairID, userID interface{}) bool {
		userItemTemp, ok := user.List.Load(userID.(int32))
		if ok {
			userTemp := &msg.Game_S_User{
				UserID:       userItemTemp.(*user.Item).UserID,
				NikeName:     userItemTemp.(*user.Item).NikeName,
				UserGold:     userItemTemp.(*user.Item).UserGold,
				UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
				MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
				HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
				FaceID:       userItemTemp.(*user.Item).FaceID,
				RoleID:       userItemTemp.(*user.Item).RoleID,
				PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
				TableID:      userItemTemp.(*user.Item).TableID,
				ChairID:      userItemTemp.(*user.Item).ChairID,
			}
			userListMap[userItemTemp.(*user.Item).ChairID] = userTemp
			tempPrintInt32 = append(tempPrintInt32, userItemTemp.(*user.Item).UserID)
			tempPrintInt32Chaird = append(tempPrintInt32Chaird, userItemTemp.(*user.Item).ChairID)
			tempPrintCount++
		}

		return true
	})
	for _, v := range userListMap {
		userListArr = append(userListArr, *v)
	}
	log.Logger.Debugf("Game_S_SitDownNotify 桌子号:%v,坐下通知列表：%v,椅子号:%v,房间人数:%v,实际人数:%v", it.tableID, tempPrintInt32, tempPrintInt32Chaird, it.userCount, tempPrintCount)

	randNum := rand.Krand(6, 3)

	it.roundOrder = fmt.Sprintf("%v%v%s", conf.GetServer().GameID, time.Now().Unix(), randNum)
	it.sendAllUser(&msg.Game_S_SitDownNotify{
		Data:    userListArr,
		TableID: it.tableID,
	})
	go func() {
		t := time.NewTimer(time.Second * time.Duration(conf.GetServer().GameShuffleTime)) // --8s  洗牌环节
		it.sceneStartTime = time.Now().Unix()
		for {
			select {
			case <-t.C:
				it.changeGameStatus(global.GameStatusStart, conf.GetServer().GameStartTime)
				return
			}
		}
	}()
}

//是否游戏中
func (it *Item) IsInGame(user *user.Item) bool {
	//判断用户是否正在游戏中
	_, ok := it.userPlaying[user.ChairID]
	//if it.gameStatus == global.GameStatusEnd {
	//	return false
	//}
	if it.gameStatus == global.GameStatusFree {
		return false
	}
	return ok
}

//结束游戏  判断是否钱够不够 直接踢出去
func (it *Item) onEventGameConclude() {
	it.gameStatus = global.GameStatusEnd
	it.cronTimer.Stop()
	//更新库存
	store.GameControl.ChangeStore(it.systemScore)

	//记录游戏记录
	it.onWriteGameRecord()
	//用户写分
	it.onWriteGameScore()
	// 热更数据
	it.onUpdateAgentData()

	var gameConclude = msg.Game_S_GameConclude{
		ResultList: make([]msg.Game_S_GameResult, 0),
	}
	var listLoss = make(map[int32]float32)

	for k, v := range it.userListLoss {
		temp := msg.Game_S_GameResult{
			ChairID: k,
			Result:  v,
		}
		gameConclude.ResultList = append(gameConclude.ResultList, temp)
	}
	it.userList.Range(func(chairID, uid interface{}) bool {
		value, ok := user.List.Load(uid)
		_, okPlay := it.userPlaying[chairID.(int32)]
		if ok && okPlay {
			//判断下注积分是否足够
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				listLoss[chairID.(int32)] = value.(*user.Item).UserGold
			} else {
				listLoss[chairID.(int32)] = value.(*user.Item).UserDiamond
			}

		}
		return true
	})
	//gameConclude.UserListLoss = it.userListLoss
	//gameConclude.UserListMoney = listLoss
	it.userList.Range(func(chairID, uid interface{}) bool {
		value, ok := user.List.Load(uid)
		if ok {
			mySelf := value.(*user.Item)
			// 如果游戏结束的时候用户在离线状态 解锁用户
			if mySelf.Status == user.StatusOffline {
				// 起立 强制退出
				it.OnActionUserStandUp(value.(*user.Item), true)
				// map 中移除
				user.List.Delete(uid)
				return true
			}
			// todo 变动金币通知心跳
			if !value.(*user.Item).IsRobot() {
				redis.GameClient.RegisterRecharge(value.(*user.Item).UserID)
				for k, v := range gameConclude.ResultList {
					if value.(*user.Item).ChairID == v.ChairID {
						gameConclude.ResultList[k], gameConclude.ResultList[0] = gameConclude.ResultList[0], gameConclude.ResultList[k]
						break
					}
				}
			}

			//结束通知
			value.(*user.Item).WriteMsg(&gameConclude)

			//判断下注积分是否足够
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				if value.(*user.Item).UserGold < store.GameControl.GetGameInfo().MinEnterScore {
					_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 金币不足!")
					value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
						ErrorCode: global.JCError1,
						ErrorMsg:  "退出房间, 金币不足!",
					})
					// 起立 强制退出
					it.OnActionUserStandUp(value.(*user.Item), true)

					value.(*user.Item).Close()
					// map 中移除
					user.List.Delete(uid)
					return true
				}
			} else {
				if value.(*user.Item).UserDiamond < store.GameControl.GetGameInfo().MinEnterScore {
					_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 金币不足!")
					value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
						ErrorCode: global.JCError1,
						ErrorMsg:  "退出房间, 余额不足",
					})
					// 起立 强制退出
					it.OnActionUserStandUp(value.(*user.Item), true)

					value.(*user.Item).Close()
					// map 中移除
					user.List.Delete(uid)
					return true
				}
			}
		}
		return true
	})

	//清空上局数据
	it.systemScore = 0
	it.bankerChairID = -1
	it.bankerMultiple = 0
	it.userListLoss = make(map[int32]float32)
	it.userPlaying = make(map[int32]bool)
	it.userListJetton = sync.Map{}
	it.userListTP = sync.Map{}
	it.userListQZ = sync.Map{}
	it.userCount = 0
	it.androidCount = 0
	it.gameStatus = global.GameStatusFree
	it.tablePlayCount = 0
	// 结束游戏全部起立
	var oldUserList []int32
	it.userList.Range(func(chairID, uid interface{}) bool {
		value, ok := user.List.Load(uid)
		if ok {
			oldUserList = append(oldUserList, value.(*user.Item).UserID)
			it.OnActionUserStandUp(value.(*user.Item), true)
		}

		return true
	})
	it.userList = sync.Map{}
	// 删除空闲桌子
	delTableList()
	//开奖定时器
	//t := time.NewTimer(time.Second * time.Duration(conf.GetServer().GameStartTime))
	//select {
	//case <-t.C:
	//	it.onEventGameStart()
	//}
	return
	// 超时不操作断开TCP
	//go func() {
	//	for len(oldUserList) > 0 {
	//		t := time.NewTicker(time.Second * 60 * 5)
	//		select {
	//		case <-t.C:
	//
	//			return
	//		}
	//	}
	//}()
}

// 热更数据
func (it *Item) onUpdateAgentData() {
	var decPercentValue float32
	var intDataType = 2 //数据类型：1 注册，2 游戏输赢，3 返佣，4 充值，5 兑换，6 领取佣金
	for key, v := range it.userListLoss {
		decAmount := v // 输赢金额
		uid, ok := it.userList.Load(key)
		if !ok {
			_ = log.Logger.Errorf("onUpdateAgentData it.userList err: %d", key)
			return
		}
		userItem, ok := user.List.Load(uid.(int32))
		// 过滤机器人
		if userItem.(*user.Item).BatchID != -1 {
			continue
		}
		mysql.GameClient.WriteAgentData(uid.(int32),
			intDataType,
			decAmount,
			decPercentValue,
		)
	}
}

//用户写分
func (it *Item) onWriteGameScore() {
	//GSP_WriteGameScore
	var endTime = time.Now().Format("2006-01-02 15:04:05")
	for key, v := range it.userListLoss {
		var intWinCount, //胜利盘数
			intLostCount, //失败盘数
			intDrawCount, //和局盘数
			intFleeCount, //逃跑数目
			tintTaskForward int32 // 任务跟进
		uid, ok := it.userList.Load(key)
		// 过滤机器人
		if !ok {
			_ = log.Logger.Errorf("onWriteGameScore it.userList err: %d", key)
			return
		}
		userItem, ok := user.List.Load(uid.(int32))
		if !ok {
			_ = log.Logger.Errorf("onWriteGameScore user.List err:%d", uid.(int32))
			return
		}
		if v > 0 {
			intWinCount = 1
		}
		if v == 0 {
			intDrawCount = 1
		}
		if v < 0 {
			intDrawCount = 1
		}
		if userItem.(*user.Item).Status == user.StatusOffline {
			intFleeCount = 1
		}
		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			userItem.(*user.Item).UserGold += v
		} else {
			if !userItem.(*user.Item).IsRobot() {
				if !redis.GameClient.IsExistsDiamond(userItem.(*user.Item).UserID) {
					scoreInfo, _ := mysql.GetGameScoreInfoByUserId(mysql.GameClient.GetXJGameDB, userItem.(*user.Item).UserID)
					redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, scoreInfo.Diamond)
				}

				userDiamond, err := redis.GameClient.GetDiamond(userItem.(*user.Item).UserID)
				if err != nil {
					log.Logger.Error("GetDiamond err:", err)
				} else {
					userItem.(*user.Item).UserDiamond = float32(userDiamond)
				}

				var dia = v + it.userTax[key]

				userItem.(*user.Item).Jackpot += dia + float32(math.Abs(float64(store.GameControl.GetGameInfo().UmRevenueRatio*dia)))
			}
			userItem.(*user.Item).UserDiamond += v

		}
		// 过滤机器人
		if userItem.(*user.Item).BatchID != -1 {
			continue
		}
		host, _, _ := net.SplitHostPort(userItem.(*user.Item).Agent.RemoteAddr().String())
		errorCode, errorMsg := mysql.GameClient.WriteUserScore(uid.(int32),
			v,
			store.GameControl.GetGameInfo().DeductionsType,
			it.userTax[key],
			intWinCount,
			intLostCount,
			intDrawCount,
			intFleeCount,
			conf.GetServer().GameJettonTime,
			tintTaskForward,
			store.GameControl.GetGameInfo().KindID,
			store.GameControl.GetGameInfo().GameID,
			v,
			host,
			time.Unix(it.sceneStartTime, 0).Format("2006-01-02 15:04:05"),
			endTime,
			it.drawID,
			userItem.(*user.Item).Jackpot,
			userItem.(*user.Item).UserDiamond,
			it.roundOrder,
		)

		if userItem.(*user.Item).UserID == 1000514 {
			fmt.Printf("UserDiamond: %f Jackpot:%f   userTax %f userloss%f \n", userItem.(*user.Item).UserDiamond, userItem.(*user.Item).Jackpot, it.userTax[key], v)
		}

		if errorCode != common.StatusOK {
			_ = log.Logger.Errorf(" mysql GSP_WriteGameScore存储过程 %v %s ", errorCode, errorMsg)
			return
		}
		if !userItem.(*user.Item).IsRobot() && store.GameControl.GetGameInfo().DeductionsType == 1 {
			redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userItem.(*user.Item).UserDiamond)
		}
	}
}

//游戏记录
func (it *Item) onWriteGameRecord() {
	//GSP_RecordDrawInfo
	//-- intTableID：桌子ID
	//-- intUserCount：用户数量
	//-- intAndroidCount：机器人数量
	//-- decWasteCount：损耗数目
	//-- decResveueCount：税收数目
	//-- timeEnterTime：游戏开始时间
	//-- timeLeaveTime：游戏结束时间
	//-- tintScoreType：金币类型
	// --gameRoundInfo: 游戏局数详细信息

	var endTime = time.Now().Format("2006-01-02 15:04:05")

	var startTime = time.Unix(it.sceneStartTime, 0).Format("2006-01-02 15:04:05")
	var taxSum float32 = 0
	for _, v := range it.userTax {
		taxSum += v
	}

	// 局数详细信息
	roundDetail := struct {
		UserList       map[int32]int32                    `json:"user_list"`        // uid --->椅子号
		UserListQZ     map[int32]int32                    `json:"user_list_qz"`     //用户抢庄	座位号	map[int32]int32
		UserListJetton map[int32]int32                    `json:"user_list_jetton"` //玩家下注 座位号 map[int32]int32
		UserListPoker  map[int32]*msg.Game_S_LotteryPoker `json:"user_list_poker"`  //玩家扑克	每个用户5张扑克
		UserListLoss   map[int32]float32                  `json:"user_list_loss"`   //用户盈亏 座位号
		UserTax        map[int32]float32                  `json:"user_tax"`         //用户税收
	}{}
	roundDetail.UserTax = make(map[int32]float32, 0)
	roundDetail.UserListLoss = make(map[int32]float32, 0)
	roundDetail.UserList = make(map[int32]int32, 0)
	roundDetail.UserListPoker = make(map[int32]*msg.Game_S_LotteryPoker, 0)
	roundDetail.UserListJetton = make(map[int32]int32, 0)
	roundDetail.UserListQZ = make(map[int32]int32, 0)

	for k, v := range it.userTax {
		roundDetail.UserTax[k] = v
	}
	for k, v := range it.userListLoss {
		roundDetail.UserListLoss[k] = v
	}
	it.userList.Range(func(key, value interface{}) bool {
		roundDetail.UserList[key.(int32)] = value.(int32)
		return true
	})

	for k, v := range it.userListPoker {
		poker := new(msg.Game_S_LotteryPoker)
		for _, v1 := range v.LotteryPoker {
			poker.LotteryPoker = append(poker.LotteryPoker, v1)
		}
		poker.PokerType = v.PokerType
		roundDetail.UserListPoker[k] = poker

	}

	it.userListJetton.Range(func(key, value interface{}) bool {
		roundDetail.UserListJetton[key.(int32)] = value.(int32)
		return true
	})
	it.userListQZ.Range(func(key, value interface{}) bool {
		roundDetail.UserListQZ[key.(int32)] = value.(int32)
		return true
	})

	detail, err := json.Marshal(roundDetail)
	if err != nil {
		log.Logger.Debugf("onWriteGameRecord err:%s", err.Error())
	}
	errorCode, errorMsg, drawID := mysql.GameClient.WriteGameRecord(
		it.tableID,
		it.userCount,
		it.androidCount,
		it.systemScore,
		taxSum,
		startTime,
		endTime,
		store.GameControl.GetGameInfo().DeductionsType,
		string(detail),
	)
	if errorCode != common.StatusOK {
		_ = log.Logger.Errorf(" mysql GSP_RecordDrawInfo存储过程报错 %n %s ", errorCode, errorMsg)
		return
	}
	it.drawID = drawID
}

//发送场景
func (it *Item) onEventSendGameScene(args ...interface{}) {
	userItem := args[0].(*user.Item)
	switch it.gameStatus {
	case global.GameStatusStart:
		fallthrough
	case global.GameStatusEnd:
		fallthrough
	case global.GameStatusFree:
		sceneTime := int64(conf.GetServer().GameStartTime+conf.GetServer().GameShuffleTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime > 1 {
			sceneTime -= 1
			if sceneTime >= int64(conf.GetServer().GameStartTime) {
				//设定游戏玩家
				it.userList.Range(func(chairID, userID interface{}) bool {
					it.userPlaying[chairID.(int32)] = true
					return true
				})
				// 设定开始游戏  todo 待优化发送坐下消息
				userListMap := make(map[int32]*msg.Game_S_User, 0)
				userListArr := make([]msg.Game_S_User, 0)
				it.userList.Range(func(chairID, userID interface{}) bool {
					userItemTemp, ok := user.List.Load(userID.(int32))
					if ok {
						userTemp := &msg.Game_S_User{
							UserID:       userItemTemp.(*user.Item).UserID,
							NikeName:     userItemTemp.(*user.Item).NikeName,
							UserGold:     userItemTemp.(*user.Item).UserGold,
							UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
							MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
							HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
							FaceID:       userItemTemp.(*user.Item).FaceID,
							RoleID:       userItemTemp.(*user.Item).RoleID,
							PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
							TableID:      userItemTemp.(*user.Item).TableID,
							ChairID:      userItemTemp.(*user.Item).ChairID,
						}
						userListMap[userItemTemp.(*user.Item).ChairID] = userTemp
						if !userItemTemp.(*user.Item).IsRobot() {
							log.Logger.Debugf("Game_S_SitDownNotify 桌子号:%v,坐下通知：%v,人数:%v", it.tableID, userItemTemp.(*user.Item).UserID, it.userCount)
						}
					}

					return true
				})
				for _, v := range userListMap {
					userListArr = append(userListArr, *v)
				}
				it.sendAllUser(&msg.Game_S_SitDownNotify{
					Data:    userListArr,
					TableID: it.tableID,
				})
			} else {
				var freeScene msg.Game_S_FreeScene
				freeScene.UserList = make([]msg.Game_S_User, 0)
				it.userList.Range(func(chairID, userID interface{}) bool {
					userItemTemp, ok := user.List.Load(userID.(int32))
					if !ok {
						_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
						return true
					}
					userTemp := msg.Game_S_User{
						UserID:       userItemTemp.(*user.Item).UserID,
						NikeName:     userItemTemp.(*user.Item).NikeName,
						UserGold:     userItemTemp.(*user.Item).UserGold,
						UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
						MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
						HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
						FaceID:       userItemTemp.(*user.Item).FaceID,
						RoleID:       userItemTemp.(*user.Item).RoleID,
						PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
						TableID:      userItemTemp.(*user.Item).TableID,
						ChairID:      userItemTemp.(*user.Item).ChairID,
					}
					freeScene.UserList = append(freeScene.UserList, userTemp)
					//freeScene.UserList[chairID.(int32)] = UserInfoToGrameUser(userItemTemp.(*user.Item))
					return true
				})
				freeScene.UserChairID = userItem.ChairID
				freeScene.TableID = it.tableID
				freeScene.SceneStartTime = sceneTime
				userItem.WriteMsg(&freeScene)
			}
			break
		} else if sceneTime == 1 {
			time.Sleep(time.Second * time.Duration(sceneTime))
			it.onEventSendGameScene(userItem)
			break
		}
		fallthrough
	case global.GameStatusQZ:
		sceneTime := int64(conf.GetServer().GameQZTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime > 1 {
			sceneTime -= 1 // 减掉安全时间1s
			var qzScene msg.Game_S_QZScene
			qzScene.UserList = make([]msg.Game_S_User, 0)
			it.userList.Range(func(chairID, userID interface{}) bool {
				userItemTemp, ok := user.List.Load(userID.(int32))
				if !ok {
					_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusQZ err user.List.Load")
					return true
				}
				userTemp := msg.Game_S_User{
					UserID:       userItemTemp.(*user.Item).UserID,
					NikeName:     userItemTemp.(*user.Item).NikeName,
					UserGold:     userItemTemp.(*user.Item).UserGold,
					UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
					MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
					HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
					FaceID:       userItemTemp.(*user.Item).FaceID,
					RoleID:       userItemTemp.(*user.Item).RoleID,
					PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
					TableID:      userItemTemp.(*user.Item).TableID,
					ChairID:      userItemTemp.(*user.Item).ChairID,
				}
				qzScene.UserList = append(qzScene.UserList, userTemp)
				//qzScene.UserList[chairID.(int32)] = UserInfoToGrameUser(userItemTemp.(*user.Item))
				return true
			})
			userMultiple := make([]int32, 0)
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				for _, v := range conf.GetServer().MultipleList {
					if userItem.UserGold > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.userCount)) {
						userMultiple = append(userMultiple, v)
					}
				}
			} else {
				for _, v := range conf.GetServer().MultipleList {
					if userItem.UserDiamond > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.userCount)) {
						userMultiple = append(userMultiple, v)
					}
				}
			}
			if len(userMultiple) == 0 {
				userMultiple = append(userMultiple, 0)
			}
			qzScene.Multiple = userMultiple
			qzScene.UserChairID = userItem.ChairID
			qzScene.TableID = it.tableID
			qzScene.RecordID = it.roundOrder
			qzScene.SceneStartTime = sceneTime
			qzScene.UserListQZ = make([]msg.Game_S_UserQZ, 0)
			it.userListQZ.Range(func(chairID, multiple interface{}) bool {
				temp := msg.Game_S_UserQZ{
					ChairID:  chairID.(int32),
					Multiple: multiple.(int32),
				}
				qzScene.UserListQZ = append(qzScene.UserListQZ, temp)
				return true
			})

			userItem.WriteMsg(&qzScene)
			break
		} else if sceneTime == 1 {
			time.Sleep(time.Second * time.Duration(sceneTime))
			it.onEventSendGameScene(userItem)
			break
		}
		fallthrough
	case global.GameStatusJetton:
		sceneTime := int64(conf.GetServer().GameJettonTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime < 0 && it.gameStatus == global.GameStatusQZ {
			sceneTime = int64(conf.GetServer().GameJettonTime)
		}
		if sceneTime > 1 {
			sceneTime -= 1
			var jettonScene msg.Game_S_JettonScene
			jettonScene.UserList = make([]msg.Game_S_User, 0)
			it.userList.Range(func(chairID, userID interface{}) bool {
				userItemTemp, ok := user.List.Load(userID.(int32))
				if !ok {
					_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusQZ err user.List.Load")
					return true
				}
				userTemp := msg.Game_S_User{
					UserID:       userItemTemp.(*user.Item).UserID,
					NikeName:     userItemTemp.(*user.Item).NikeName,
					UserGold:     userItemTemp.(*user.Item).UserGold,
					UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
					MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
					HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
					FaceID:       userItemTemp.(*user.Item).FaceID,
					RoleID:       userItemTemp.(*user.Item).RoleID,
					PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
					TableID:      userItemTemp.(*user.Item).TableID,
					ChairID:      userItemTemp.(*user.Item).ChairID,
				}
				jettonScene.UserList = append(jettonScene.UserList, userTemp)
				return true
			})
			jettonScene.UserChairID = userItem.ChairID

			jettonScene.SceneStartTime = sceneTime
			jettonScene.RecordID = it.roundOrder
			jettonScene.UserListJetton = make([]msg.Game_S_UserJetton, 0)
			it.userListJetton.Range(func(chairID, multiple interface{}) bool {
				if chairID == it.bankerChairID {
					return true
				}
				temp := msg.Game_S_UserJetton{
					ChairID:  chairID.(int32),
					Multiple: multiple.(int32),
				}
				jettonScene.UserListJetton = append(jettonScene.UserListJetton, temp)
				return true
			})

			var userMultiple = make([]int32, 0)
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				for _, v := range conf.GetServer().JettonList {
					if userItem.UserGold > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.bankerMultiple)) {
						userMultiple = append(userMultiple, v)
					}
				}
			} else {
				for _, v := range conf.GetServer().JettonList {
					if userItem.UserDiamond > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.bankerMultiple)) {
						userMultiple = append(userMultiple, v)
					}
				}
			}
			if len(userMultiple) == 0 {
				userMultiple = append(userMultiple, 0)
			}
			jettonScene.Multiple = userMultiple
			jettonScene.BankerChairID = it.bankerChairID
			jettonScene.BankerMultiple = it.bankerMultiple
			userItem.WriteMsg(&jettonScene)
			break
		} else if sceneTime == 1 {
			time.Sleep(time.Second * time.Duration(sceneTime))
			it.onEventSendGameScene(userItem)
			break
		}
		fallthrough
	case global.GameStatusTP:
		sceneTime := int64(conf.GetServer().GameTPTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime >= 1 {
			sceneTime -= 1
		}
		// 摊牌场景
		var tpScene msg.Game_S_TPScene
		tpScene.UserList = make([]msg.Game_S_User, 0)
		it.userList.Range(func(chairID, userID interface{}) bool {
			userItemTemp, ok := user.List.Load(userID.(int32))
			if !ok {
				_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusQZ err user.List.Load")
				return true
			}
			userTemp := msg.Game_S_User{
				UserID:       userItemTemp.(*user.Item).UserID,
				NikeName:     userItemTemp.(*user.Item).NikeName,
				UserGold:     userItemTemp.(*user.Item).UserGold,
				UserDiamond:  userItemTemp.(*user.Item).UserDiamond,
				MemberOrder:  userItemTemp.(*user.Item).MemberOrder,
				HeadImageUrl: userItemTemp.(*user.Item).HeadImageUrl,
				FaceID:       userItemTemp.(*user.Item).FaceID,
				RoleID:       userItemTemp.(*user.Item).RoleID,
				PhotoFrameID: userItemTemp.(*user.Item).PhotoFrameID,
				TableID:      userItemTemp.(*user.Item).TableID,
				ChairID:      userItemTemp.(*user.Item).ChairID,
			}
			tpScene.UserList = append(tpScene.UserList, userTemp)
			//tpScene.UserList[chairID.(int32)] = UserInfoToGrameUser(userItemTemp.(*user.Item))
			return true
		})
		tpScene.UserChairID = userItem.ChairID

		tpScene.SceneStartTime = sceneTime
		tpScene.RecordID = it.roundOrder
		var userListTp = make([]msg.Game_S_UserTP, 0)
		it.userListTP.Range(func(chairID, status interface{}) bool {

			if !status.(bool) {
				return true
			}

			poker := it.userListPoker[chairID.(int32)]
			d := msg.Game_S_UserTP{
				ChairID:      chairID.(int32),
				PokerType:    poker.PokerType & 0xf00 / 0x100,
				LotteryPoker: poker.LotteryPoker,
			}
			userListTp = append(userListTp, d)
			return true
		})
		tpScene.UserListJetton = make([]msg.Game_S_UserJetton, 0)
		it.userListJetton.Range(func(chairID, multiple interface{}) bool {
			if chairID == it.bankerChairID {
				return true
			}
			temp := msg.Game_S_UserJetton{
				ChairID:  chairID.(int32),
				Multiple: multiple.(int32),
			}
			tpScene.UserListJetton = append(tpScene.UserListJetton, temp)
			return true
		})
		tpScene.UserListTP = userListTp
		tpScene.BankerChairID = it.bankerChairID
		tpScene.BankerMultiple = it.bankerMultiple
		//tpScene.UserPlaying = it.userPlaying
		userItem.WriteMsg(&tpScene)
	default:
		log.Logger.Info("default error ", it.gameStatus)
	}

}

//用户坐下
func (it *Item) OnActionUserSitDown(args ...interface{}) {
	//fmt.Printf("%c[1;40;31m用户坐下=====%c[0m\n", 0x1B, 0x1B)
	userItem := args[0].(*user.Item)
	// 检查是否锁定
	lock := mysql.GameClient.IsLock(userItem.UserID)
	if lock {
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError1,
			ErrorMsg:  "坐下失败, 上局游戏未结束",
		})
		userItem.Close()
		return
	}

	//校验是否满人
	if it.userCount >= store.GameControl.GetGameInfo().ChairCount {
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError2,
			ErrorMsg:  "坐下失败, 房间人数已满!",
		})
		userItem.Close()
		return
	}
	if it.gameStatus != global.GameStatusFree {
		log.Logger.Infof("OnActionUserSitDown 重新进入匹配池 uid:%v", userItem.UserID)
		return
	}
	//if userItem.ChairID>=0{
	//	log.Logger.Infof("OnActionUserSitDown 已经坐下 uid:%v chairid:%v", userItem.UserID,userItem.ChairID)
	//	return
	//}
	host, _, _ := net.SplitHostPort(userItem.Agent.RemoteAddr().String())
	//strings.Split(userItem.Agent.RemoteAddr().String(), ":")[0]
	// 锁定
	if err := mysql.GameClient.Lock(userItem.UserID, host); err != nil {
		_ = log.Logger.Errorf("锁定用户失败 err %v", err)
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError3,
			ErrorMsg:  "坐下失败, 服务器繁忙!",
		})
		userItem.Close()
		return
	}
	it.mutex.Lock()
	//等待加入游戏用户列表
	for i := int32(0); i < store.GameControl.GetGameInfo().ChairCount; i++ {
		_, ok := it.userList.Load(i)
		if ok {
			continue
		}
		//加入游戏
		userItem.SitDown(it.tableID, i)
		it.userList.Store(i, userItem.UserID)
		it.userCount++
		if userItem.IsRobot() {
			it.androidCount++
		}
		break
	}
	it.mutex.Unlock()
	//it.ResetTablesList()
	//发送坐下通知给其他玩家
	//data, err := json.Marshal(userItem.GetUserInfo())
	//if err != nil {
	//	_ = log.Logger.Errorf("OnActionUserSitDown %s", err.Error())
	//}
	// 坐下成功不发送场景消息
	// it.onEventSendGameScene(userItem)
	log.Logger.Debugf("OnActionUserSitDown tableID:%v 人数:%v,gameStatus:%v,uid:%v,chairID:%v, rand 人数:%v", it.tableID, it.userCount, it.gameStatus, userItem.UserID, userItem.ChairID, it.tablePlayCount)
	//runtime.Gosched()

	if it.userCount >= it.tablePlayCount && it.gameStatus == global.GameStatusFree {
		it.gameStatus = global.GameStatusStart
		go it.onEventGameStart()
	}
	var isEnter = false
	go func() {
		timer := time.NewTimer(time.Second * time.Duration(10+it.userCount))
		for {
			select {
			case <-timer.C:
				if it.userCount >= global.TablePlayCount && it.gameStatus == global.GameStatusFree {
					go it.onEventGameStart()
				}
				isEnter = true
				break
			}
			if isEnter {
				break
			}
		}
	}()
}

//用户起立
func (it *Item) OnActionUserStandUp(args ...interface{}) {
	userItem := args[0].(*user.Item)
	flag := args[1].(bool)
	if !flag {
		//检测是否游戏中
		v, _ := it.userPlaying[userItem.ChairID]
		if v {
			log.Logger.Errorf("OnActionUserStandUp %s", "游戏中不允许退出")
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.StandUpError1,
				ErrorMsg:  "游戏中不允许退出!",
			})
			return
		}
	}

	//移出游戏用户列表
	oldChairID := userItem.ChairID
	it.userList.Delete(oldChairID)
	//it.userCount--
	//if userItem.IsRobot() {
	//	it.androidCount--
	//}
	userItem.StandUp()
	//解锁
	if err := mysql.GameClient.UnLock(userItem.UserID); err != nil {
		_ = log.Logger.Errorf("解锁用户失败 err %v", err)
		return
	}
	//通知其他玩家 不通知消息 7.18需求
	//it.sendAllUser(&msg.Game_S_StandUpNotify{
	//	ChairID: oldChairID,
	//})
}

//用户断线
func (it *Item) OnActionUserOffLine(args ...interface{}) {
	userItem := args[0].(*user.Item)

	//设置用户状态
	userItem.Status = user.StatusOffline

	//通知其他玩家
	it.sendOtherUser(userItem.ChairID, &msg.Game_S_OffLineNotify{
		ChairID: userItem.ChairID,
	})
}

//删除准备中的用户
func (it *Item) OnActionUserClose(args ...interface{}) {
	userItem := args[0].(*user.Item)
	it.userList.Delete(userItem.ChairID)
}

//用户重入
func (it *Item) OnActionUserReconnect(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem.Status == user.StatusOffline {
		//设置用户状态
		userItem.Status = user.StatusPlaying
		//通知其他玩家
		it.sendOtherUser(userItem.ChairID, &msg.Game_S_OnLineNotify{
			ChairID: userItem.ChairID,
		})
	}

	//发送场景消息
	it.onEventSendGameScene(userItem)
}

//用户抢庄
func (it *Item) OnUserQZ(args ...interface{}) {
	//fmt.Printf("%c[1;40;31m用户抢庄=====%c[0m\n", 0x1B, 0x1B)
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserQZ)

	//判断用户是否正在游戏中
	_, ok := it.userPlaying[userItem.ChairID]
	if !ok {
		//_ = log.Logger.Errorf("")
		_ = log.Logger.Error("OnUserQZ用户状态异常:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.QZError1,
		//	ErrorMsg:  "抢庄异常,用户状态异常!",
		//})
		//userItem.Close()
		return
	}

	//判断游戏状态
	if it.gameStatus != global.GameStatusQZ {
		_ = log.Logger.Error("OnUserQZ游戏状态异常:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.QZError1,
		//	ErrorMsg:  "抢庄异常,现在不能抢庄!",
		//})
		//userItem.Close()
		return
	}

	//判断是否已操作
	_, ok = it.userListQZ.Load(userItem.ChairID)
	if ok {
		_ = log.Logger.Error("OnUserQZ游戏状态异常:", "抢庄异常,请勿重复抢庄!")

		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.QZError1,
		//	ErrorMsg:  "抢庄异常,请勿重复抢庄!",
		//})
		//userItem.Close()
		return
	}

	//判断数据是否异常
	if m.Multiple < 0 {
		_ = log.Logger.Error("OnUserQZ游戏状态异常:", "抢庄异常,倍数异常!")

		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.QZError1,
			ErrorMsg:  "抢庄异常,倍数异常!",
		})
		//userItem.Close()
		return
	}

	//记录用户抢庄
	it.userListQZ.Store(userItem.ChairID, m.Multiple)

	it.sendAllUser(&msg.Game_S_UserQZ{
		ChairID:  userItem.ChairID,
		Multiple: m.Multiple,
	})

	//判断是否可以进入下个状态
	ok = it.checkEnterNextStatus(it.userListQZ)
	if ok {
		//提前完成倒计时
		it.cronTimer.Reset(0)

	}

}

//用户下注
func (it *Item) OnUserPlaceJetton(args ...interface{}) {
	//fmt.Printf("%c[1;40;31m用户下注=====%c[0m\n", 0x1B, 0x1B)
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserJetton)

	//fmt.Printf("%c[1;43;31m用户下注=====%c[0m %v %v \n", 0x1B, 0x1B,userItem,m.Multiple)

	//判断用户是否正在游戏中
	_, ok := it.userPlaying[userItem.ChairID]
	if !ok {
		_ = log.Logger.Error("OnUserPlaceJettony用户状态异常:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError1,
		//	ErrorMsg:  "下注异常,用户状态异常!",
		//})
		//userItem.Close()
		return
	}

	//判断游戏状态
	if it.gameStatus != global.GameStatusJetton {
		_ = log.Logger.Error("OnUserPlaceJetton现在不能叫倍数:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError1,
		//	ErrorMsg:  "下注异常,现在不能叫倍数!",
		//})
		//userItem.Close()
		return
	}

	//判断是否已操作
	_, ok = it.userListJetton.Load(userItem.ChairID)
	if ok {
		if userItem.IsRobot() {
			return
		}
		_ = log.Logger.Errorf("OnUserPlaceJetton: 已经叫过倍数! uid:  %v", userItem.UserID)

		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注异常,已经叫过倍数!",
		})
		//userItem.Close()
		return
	}

	//判断是否是庄家
	if userItem.ChairID == it.bankerChairID {
		_ = log.Logger.Error("OnUserPlaceJetton", "下注异常,庄家不能下注")

		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError1,
		//	ErrorMsg:  "下注异常,庄家不能下注",
		//})
		//userItem.Close()
		return
	}

	//判断数据是否异常
	if m.Multiple <= 0 {
		_ = log.Logger.Error("OnUserPlaceJetton", "下注异常,倍数异常!")

		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注异常,倍数异常!",
		})
		//userItem.Close()
		return
	}

	//记录用户下注
	it.userListJetton.Store(userItem.ChairID, m.Multiple)

	//通知其他人
	it.sendAllUser(&msg.Game_S_UserJetton{
		ChairID:  userItem.ChairID,
		Multiple: m.Multiple,
	})

	//判断是否可以进入下个状态
	ok = it.checkEnterNextStatus(it.userListJetton)
	if ok {
		//提前完成倒计时
		it.cronTimer.Reset(0)
	}

}

//用户摊牌
func (it *Item) OnUserTP(args ...interface{}) {
	//fmt.Printf("%c[1;40;31m用户摊牌=====%c[0m\n", 0x1B, 0x1B)
	userItem := args[0].(*user.Item)
	_ = args[1].(*msg.Game_C_UserTP)

	//判断用户是否正在游戏中
	_, ok := it.userPlaying[userItem.ChairID]
	if !ok {
		_ = log.Logger.Error("OnUserTP用户状态异常:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.TPError1,
		//	ErrorMsg:  "摊牌异常,用户状态异常!",
		//})
		//userItem.Close()
		return
	}

	//判断游戏状态
	if it.gameStatus != global.GameStatusTP {
		_ = log.Logger.Error("OnUserTP游戏状态异常:", it.gameStatus)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.TPError1,
		//	ErrorMsg:  "摊牌异常,用户状态异常!",
		//})
		//userItem.Close()
		return
	}

	//判断是否已操作
	_, ok = it.userListTP.Load(userItem.ChairID)
	if ok {
		_ = log.Logger.Error("OnUserTP不能重复摊牌:", "不能重复摊牌")
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.TPError1,
		//	ErrorMsg:  "摊牌异常,不能重复摊牌!",
		//})
		//userItem.Close()
		return
	}

	//记录用户摊牌
	it.userListTP.Store(userItem.ChairID, true)

	//通知其他人
	var userTP msg.Game_S_UserTP
	userTP.ChairID = userItem.ChairID

	//var userTp msg.Game_S_LotteryPoker

	//userTp.PokerType = it.userListPoker[userItem.ChairID].PokerType & 0xf00 / 0x100
	//userTp.LotteryPoker = it.userListPoker[userItem.ChairID].LotteryPoker

	//userTP.Poker = &userTp
	userTP.PokerType = it.userListPoker[userItem.ChairID].PokerType & 0xf00 / 0x100
	userTP.LotteryPoker = it.userListPoker[userItem.ChairID].LotteryPoker
	it.sendAllUser(&userTP)

	//判断是否可以进入下个状态
	ok = it.checkEnterNextStatus(it.userListTP)
	if ok {
		//提前完成倒计时
		it.cronTimer.Reset(0)
	}

}

//发送所有人
func (it *Item) sendAllUser(data interface{}) {
	it.userList.Range(func(chairID, uid interface{}) bool {
		value, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("sendAll user.List err %d", uid)
			return true
		}
		value.(*user.Item).WriteMsg(data)
		return true
	})
}

//发送其他人
func (it *Item) sendOtherUser(userChairID int32, data interface{}) {
	it.userList.Range(func(chairID, uid interface{}) bool {
		//过滤userChairID
		if userChairID == chairID {
			return true
		}

		value, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("sendAll user.List err %d", uid)
			return true
		}
		value.(*user.Item).WriteMsg(data)
		return true
	})
}

//判断是否可以进入下个状态
func (it *Item) checkEnterNextStatus(value sync.Map) bool {
	var temp int32
	value.Range(func(v1, v2 interface{}) bool {
		temp++
		return true
	})
	// 都已经抢庄
	if temp == int32(len(it.userPlaying)) {
		return true
	}

	return false
}

//改变状态
func (it *Item) changeGameStatus(gameStatus int32, statusTime int32) {
	//设置游戏状态
	it.gameStatus = gameStatus
	it.sceneStartTime = time.Now().Unix()
	switch it.gameStatus {
	case global.GameStatusFree:
		fmt.Println("[抢庄牛牛]桌子", it.tableID, " 空闲倒计时开始：", it.userPlaying)
	case global.GameStatusStart:
		//发送发牌通知Game_S_CardRound
		it.sendAllUser(&msg.Game_S_CardRound{
			RecordID: it.roundOrder,
		})
		// 开始游戏清空数据
		it.userListJetton = sync.Map{}
		it.userListTP = sync.Map{}
		it.userListQZ = sync.Map{}
		fmt.Println("[抢庄牛牛]桌子", it.tableID, " 开始游戏倒计时开始：", it.userPlaying)
	case global.GameStatusQZ: // todo--done 动态返回可抢庄倍数 公式例：用户金币>= 10（底注）*4(牌型最大倍数)*3（最大下注倍数）*3（玩家匹配个数）=360
		it.userList.Range(func(chairID, userID interface{}) bool {
			var userMultiple = make([]int32, 0)
			userItemTemp, ok := user.List.Load(userID.(int32))
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					for _, v := range conf.GetServer().MultipleList {
						if userItemTemp.(*user.Item).UserGold > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.userCount)) {
							userMultiple = append(userMultiple, v)
						}
					}
				} else {
					for _, v := range conf.GetServer().MultipleList {
						if userItemTemp.(*user.Item).UserDiamond > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.userCount)) {
							userMultiple = append(userMultiple, v)
						}
					}
				}
				if len(userMultiple) == 0 {
					userMultiple = append(userMultiple, 0)
				}
				userItemTemp.(*user.Item).WriteMsg(&msg.Game_S_CallRound{
					Multiple: userMultiple,
				})
			}
			return true
		})
		fmt.Println("[抢庄牛牛]桌子", it.tableID, " 抢庄倒计时开始：", it.userPlaying)
	case global.GameStatusJetton: // todo-done 动态返回可下倍数 用户金币>= 10（底注）*4 (最大牌型倍数)*3(庄家抢庄倍数)*15(最高倍)=1800
		it.userList.Range(func(chairID, userID interface{}) bool {
			var userMultiple = make([]int32, 0)
			userItemTemp, ok := user.List.Load(userID.(int32))
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					for _, v := range conf.GetServer().JettonList {
						if userItemTemp.(*user.Item).UserGold > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.bankerMultiple)) {
							userMultiple = append(userMultiple, v)
						}
					}
				} else {
					for _, v := range conf.GetServer().JettonList {
						if userItemTemp.(*user.Item).UserDiamond > (float32(v) * store.GameControl.GetGameInfo().CellScore * float32(conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) * float32(it.bankerMultiple)) {
							userMultiple = append(userMultiple, v)
						}
					}
				}
				if len(userMultiple) == 0 {
					userMultiple = append(userMultiple, conf.GetServer().JettonList[0])
				}
				userItemTemp.(*user.Item).WriteMsg(&msg.Game_S_BetRound{
					Multiple: userMultiple,
				})
			}
			return true
		})
		fmt.Println("[抢庄牛牛]桌子", it.tableID, " 下注倒计时开始：", it.userPlaying)
	case global.GameStatusTP:
		it.sendAllUser(&msg.Game_S_ShowRound{})
		fmt.Println("[抢庄牛牛]桌子", it.tableID, " 摊牌倒计时开始：", it.userPlaying)
		//for key, v := range it.userPlaying {
		//	if v {
		//		uid, okUid := it.userList.Load(key)
		//		if !okUid {
		//			continue
		//		}
		//userItem, ok := user.List.Load(uid)
		//if ok {
		//var userTP msg.Game_S_UserTP
		//userTP.ChairID = key
		//var userTp msg.Game_S_LotteryPoker
		//userTp.PokerType = it.userListPoker[key].PokerType & 0xf00 / 0x100
		//userTp.LotteryPoker = it.userListPoker[key].LotteryPoker
		//userTP.Poker = &userTp
		//userItem.(*user.Item).WriteMsg(&userTP)
		//var userTP msg.Game_S_GameCard
		//userTP.ChairID = key
		//userTP.PokerType = it.userListPoker[key].PokerType & 0xf00 / 0x100
		//userTP.LotteryPoker = it.userListPoker[key].LotteryPoker
		//userItem.(*user.Item).WriteMsg(&userTP)
		//	}
		//	}
		//}
	}

	//定时器
	it.cronTimer = time.NewTimer(time.Second * time.Duration(statusTime))
	select {
	case <-it.cronTimer.C:
		switch it.gameStatus {
		case global.GameStatusFree:
			fmt.Println("[抢庄牛牛]桌子", it.tableID, " 空闲倒计时结束：", it.userPlaying)
		case global.GameStatusStart:
			fmt.Println("[抢庄牛牛]桌子", it.tableID, " 开始游戏倒计时结束：", it.userPlaying)
		case global.GameStatusQZ:
			fmt.Println("[抢庄牛牛]桌子", it.tableID, " 抢庄倒计时结束：", it.userPlaying)
		case global.GameStatusJetton:
			fmt.Println("[抢庄牛牛]桌子", it.tableID, " 下注倒计时结束：", it.userPlaying)
		case global.GameStatusTP:
			fmt.Println("[抢庄牛牛]桌子", it.tableID, " 摊牌倒计时结束：", it.userPlaying)
		}
		it.onEventGameTimer()
	}
}

//获取空闲桌子 TODO 还可以优化 找到空闲最近满员的房间
func (it *Item) GetFreeTableID() int32 {
	//if it.userCount < store.GameControl.GetGameInfo().ChairCount {
	//	//poor := store.GameControl.GetGameInfo().ChairCount - it.userCount
	//	return it.tableID
	//}
	if it.userCount < store.GameControl.GetGameInfo().ChairCount && it.gameStatus == global.GameStatusFree {
		//poor := store.GameControl.GetGameInfo().ChairCount - it.userCount
		return it.tableID
	}

	return -1
}

//获取桌子状态
func (it *Item) GetGameStatus() int32 {
	return it.gameStatus
}

func UserInfoToGrameUser(userItem *user.Item) *msg.Game_S_User {
	return &msg.Game_S_User{
		UserID:       userItem.GetUserInfo().UserID,
		NikeName:     userItem.GetUserInfo().NikeName,
		UserGold:     userItem.GetUserInfo().UserGold,
		UserDiamond:  userItem.GetUserInfo().UserDiamond,
		MemberOrder:  userItem.GetUserInfo().MemberOrder,
		HeadImageUrl: userItem.GetUserInfo().HeadImageUrl,
		FaceID:       userItem.GetUserInfo().FaceID,
		RoleID:       userItem.GetUserInfo().RoleID,
		//SuitID:       userItem.GetUserInfo().SuitID,
		PhotoFrameID: userItem.GetUserInfo().PhotoFrameID,
		TableID:      userItem.GetUserInfo().TableID,
		ChairID:      userItem.GetUserInfo().ChairID,
		//Status:       userItem.GetUserInfo().Status,
		//Gender:       userItem.GetUserInfo().Gender,
	}
}
