package table

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"xj_game_server/game/203_doudizhu/conf"
	"xj_game_server/game/203_doudizhu/game/logic"
	"xj_game_server/game/203_doudizhu/global"
	"xj_game_server/game/203_doudizhu/msg"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/log"
	rand "xj_game_server/util/leaf/util"

	"time"
)

/*
斗地主
*/
var List = make([]*Item, 0)

// 等待列表
var UserQueue sync.Map  // 真人等待列表
var RobotQueue sync.Map //机器人等待列表
var RobotCount int32    //机器人个数
var UserCount int32     //真人个数
var cronTimer *time.Timer

//初始化桌子
func OnInit() {
	//初始化桌子
	for i := int32(0); i < store.GameControl.GetGameInfo().TableCount; i++ {
		temp := &Item{
			tableID:             i,                                     //桌子号
			gameStatus:          0,                                     //游戏状态
			sceneStartTime:      0,                                     //场景开始时间
			userList:            sync.Map{},                            //玩家列表 座位号-->uid map[int32]int32
			userCount:           0,                                     //玩家数量
			userListLoss:        make(map[int32]float32),               //用户盈亏 座位号
			userTax:             make(map[int32]float32),               //用户税收
			systemScore:         0,                                     //系统盈亏
			currentChairID:      -1,                                    //当前操作椅子号
			bankerChairID:       -1,                                    //地主椅子号
			userPrepare:         sync.Map{},                            //准备玩家
			userPokers:          make(map[int32]map[int32]interface{}), //玩家牌
			dizPokers:           []int32{},                             //地主牌
			userGrabLandlord:    make(map[int32]int32),                 //用户叫分
			currentMultiple:     0,                                     //当前倍数
			nearestChairID:      -1,                                    //最近出牌椅子号
			nearestPokers:       make([]int32, 0),                      //最近出牌扑克
			nearestCardType:     0,                                     //最近出牌类型
			nearestMaxCard:      0,                                     //最近出牌中最大牌
			dizCPCount:          0,                                     //地主出牌次数
			nongmCPCount:        0,                                     //农民出牌次数
			userListTrusteeship: make(map[int32]bool),                  //用户托管

			cronTimer: &time.Timer{}, //定时任务
		}
		List = append(List, temp)
	}
	go initMatchTable()
}

// 动态添加桌子
func appendTableList() {
	n := len(List)
	size := int32(n * 2)
	for i := int32(n); i < size; i++ {
		temp := &Item{
			tableID:             i,                                     //桌子号
			gameStatus:          0,                                     //游戏状态
			sceneStartTime:      0,                                     //场景开始时间
			userList:            sync.Map{},                            //玩家列表 座位号-->uid map[int32]int32
			userCount:           0,                                     //玩家数量
			userListLoss:        make(map[int32]float32),               //用户盈亏 座位号
			userTax:             make(map[int32]float32),               //用户税收
			systemScore:         0,                                     //系统盈亏
			currentChairID:      -1,                                    //当前操作椅子号
			bankerChairID:       -1,                                    //地主椅子号
			userPrepare:         sync.Map{},                            //准备玩家
			userPokers:          make(map[int32]map[int32]interface{}), //玩家牌
			dizPokers:           []int32{},                             //地主牌
			userGrabLandlord:    make(map[int32]int32),                 //用户叫分
			currentMultiple:     0,                                     //当前倍数
			nearestChairID:      -1,                                    //最近出牌椅子号
			nearestPokers:       make([]int32, 0),                      //最近出牌扑克
			nearestCardType:     0,                                     //最近出牌类型
			nearestMaxCard:      0,                                     //最近出牌中最大牌
			dizCPCount:          0,                                     //地主出牌次数
			nongmCPCount:        0,                                     //农民出牌次数
			userListTrusteeship: make(map[int32]bool),                  //用户托管
			cronTimer:           &time.Timer{},                         //定时任务
		}
		List = append(List, temp)
	}
}

//匹配桌子
func initMatchTable() {
	UserQueue = sync.Map{}
	RobotQueue = sync.Map{}
	//开启匹配桌子 线程
	for {
		cronTimer = time.NewTimer(time.Second * 3)
		select {
		case <-cronTimer.C:

			if UserCount <= 0 {
				continue
			}
			if (RobotCount + UserCount) < global.TablePlayCount {
				global.NoticeRobotOnline <- global.TablePlayCount - UserCount
				continue
			}
			//TODO 查询空闲座位
			var tableId int32
			for key, _ := range List {
				tableId = List[key].GetFreeTableID()
				if List[key].GetRobotCount() > 0 {
					continue
				}
				if tableId >= 0 {
					break
				}
			}
			log.Logger.Debug("真人人数:", UserCount, "机器人人数:", RobotCount, "run match table->id:", tableId)
			if tableId < 0 {
				appendTableList()
				global.NoticeLoadMath <- 0
				continue
				//UserQueue.Range(func(key, value interface{}) bool {
				//	_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 没有空闲的桌子", value.(*user.Item).Status)
				//	value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				//		ErrorCode: global.SitDownError2,
				//		ErrorMsg:  "坐下失败, 没有空闲的桌子",
				//	})
				//	value.(*user.Item).Close()
				//	return true
				//})
				//RobotQueue.Range(func(key, value interface{}) bool {
				//	_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 没有空闲的桌子", value.(*user.Item).Status)
				//	value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				//		ErrorCode: global.SitDownError2,
				//		ErrorMsg:  "坐下失败, 没有空闲的桌子",
				//	})
				//	value.(*user.Item).Close()
				//	return true
				//})
				//continue
			}

			//算出差多少机器人
			haveRootCount := global.TablePlayCount - UserCount%global.TablePlayCount
			var userCont int32 = 0

			log.Logger.Debug("tableId:", tableId, "真人人数:", UserCount, "机器人人数:", RobotCount, "需要多少机器人：", haveRootCount)

			// 真人坐下
			UserQueue.Range(func(userID, userItem interface{}) bool {
				// 离线状态
				if userItem.(*user.Item).Status == user.StatusOffline {
					UserCount--
					UserQueue.Delete(userItem.(*user.Item).UserID)
					return true
				}
				go List[tableId].OnActionUserSitDown(userItem.(*user.Item))
				UserQueue.Delete(userID)
				UserCount--
				userCont++
				return true
			})
			if haveRootCount > 0 && tableId >= 0 { // 必须有真人才开始匹配机器人
				var roobotCont int32 = 0
				// 机器人入场
				RobotQueue.Range(func(userID, userItem interface{}) bool {
					if roobotCont >= haveRootCount {
						return false
					}
					// 离线状态
					if userItem.(*user.Item).Status == user.StatusOffline {
						RobotQueue.Delete(userItem.(*user.Item).UserID)
						RobotCount--
						return true
					}
					go List[tableId].OnActionUserSitDown(userItem.(*user.Item))
					RobotQueue.Delete(userID)
					RobotCount--
					roobotCont++
					return true
				})
			}

		case loadNumber := <-global.NoticeLoadMath:
			for RobotCount < loadNumber {
				time.Sleep(time.Millisecond * 100)
			}
			cronTimer.Reset(0)
		}
	}
}

type Item struct {
	tableID             int32                           //桌子号
	gameStatus          int32                           //游戏状态
	sceneStartTime      int64                           //场景开始时间
	userList            sync.Map                        //玩家列表 座位号-->uid map[int32]int32
	userCount           int32                           //玩家数量
	userListLoss        map[int32]float32               //用户盈亏 座位号
	userTax             map[int32]float32               //用户税收
	systemScore         float32                         //系统盈亏
	currentChairID      int32                           //当前操作椅子号
	bankerChairID       int32                           //地主椅子号
	userPrepare         sync.Map                        //准备玩家
	userPokers          map[int32]map[int32]interface{} //玩家牌
	userListPoker       map[int32][]int32               // 用户牌
	dizPokers           []int32                         //地主牌
	userGrabLandlord    map[int32]int32                 //用户叫分
	currentMultiple     int32                           //当前倍数
	nearestChairID      int32                           //最近出牌椅子号
	nearestPokers       []int32                         //最近出牌扑克
	nearestCardType     int32                           //最近出牌类型
	nearestMaxCard      int32                           //最近出牌中最大牌
	dizCPCount          int32                           //地主出牌次数
	nongmCPCount        int32                           //农民出牌次数
	userListTrusteeship map[int32]bool                  //用户托管
	cronTimer           *time.Timer                     //定时任务
	mutex               sync.Mutex
	robotCount          int32  // 机器人数量
	drawID              string // 游戏记录id
	roundOrder          string // 局号
}

//超时处理
func (it *Item) onEventGameTimer() {
	for it.gameStatus != global.GameStatusFree {
		select {
		case <-it.cronTimer.C:
			switch it.gameStatus {
			case global.GameStatusJF: //叫分超时
				userID, ok := it.userList.Load(it.currentChairID)
				if !ok {
					_ = log.Logger.Errorf("onEventGameTimer err %d", it.currentChairID)
					return
				}
				userItem, ok := user.List.Load(userID.(int32))
				if !ok {
					_ = log.Logger.Errorf("onEventGameTimer err %d", userID.(int32))
					return
				}
				it.OnUserGrabLandlord(userItem, &msg.Game_C_UserGrabLandlord{
					Multiple: 0,
				})
			case global.GameStatusPlay: // 出牌超时或托管
				userID, ok := it.userList.Load(it.currentChairID)
				if !ok {
					_ = log.Logger.Errorf("onEventGameTimer err %d", it.currentChairID)
					return
				}
				userItem, ok := user.List.Load(userID.(int32))
				if !ok {
					_ = log.Logger.Errorf("onEventGameTimer err %d", userID.(int32))
					return
				}
				ok, _ = it.userListTrusteeship[it.currentChairID]
				if ok {
					// 托管
					outPoker := it.GetAutoManagePokers()
					if len(outPoker) > 0 {
						it.OnUserCP(userItem, &msg.Game_C_UserCP{
							Pokers: outPoker,
						})
					} else {
						it.OnUserPass(userItem, &msg.Game_C_UserPass{})
					}
					break
				}
				// 出牌超时 自己先出
				if it.currentChairID == it.nearestChairID {
					min := int32(0x5F)
					var userPoker []int32
					for key, _ := range it.userPokers[it.currentChairID] {
						if logic.Client.GetLogicValue(min) > logic.Client.GetLogicValue(key) {
							min = key
						}
						userPoker = append(userPoker, key)
					}
					//TODO 出最小的牌
					poker := logic.Client.GetAutoOutPokers(min, userPoker)
					it.OnUserCP(userItem, &msg.Game_C_UserCP{
						Pokers: poker,
					})
				} else {
					it.OnUserPass(userItem, &msg.Game_C_UserPass{})
				}
			}
		}
	}
}

//发送场景
func (it *Item) onEventSendGameScene(args ...interface{}) {
	userItem := args[0].(*user.Item)
	switch it.gameStatus {
	case global.GameStatusFree:
		var freeScene msg.Game_S_FreeScene
		freeScene.UserList = make([]msg.Game_S_User, 0)
		it.userList.Range(func(chairID, userID interface{}) bool {
			userItemTemp, ok := user.List.Load(userID.(int32))
			if !ok {
				_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
				return true
			}
			freeScene.UserList = append(freeScene.UserList, UserInfoToGrameUser(userItemTemp.(*user.Item)))
			return true
		})
		var userPrepare = make([]msg.UserListStatus, 0)
		it.userPrepare.Range(func(key, value interface{}) bool {

			userPrepare = append(userPrepare, msg.UserListStatus{
				ChairID: key.(int32),
				Status:  value.(bool),
			})

			return true
		})
		freeScene.PrepareList = userPrepare
		userItem.WriteMsg(&freeScene)
	case global.GameStatusJF:
		var grabLandlordScene msg.Game_S_GrabLandlordScene
		grabLandlordScene.UserList = make([]msg.Game_S_User, 0)
		it.userList.Range(func(chairID, userID interface{}) bool {
			userItemTemp, ok := user.List.Load(userID.(int32))
			if !ok {
				_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
				return true
			}
			grabLandlordScene.UserList = append(grabLandlordScene.UserList, UserInfoToGrameUser(userItemTemp.(*user.Item)))
			return true
		})
		sceneTime := int64(conf.GetServer().GameJFTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime < 0 {
			sceneTime = 0
		}
		grabLandlordScene.SceneStartTime = sceneTime
		for k, _ := range it.userPokers[userItem.ChairID] {
			grabLandlordScene.UserPoker = append(grabLandlordScene.UserPoker, k)
		}
		var userGrabLandlord = make([]msg.UserListGrabLandlord, 0)
		for k, v := range it.userGrabLandlord {
			userGrabLandlord = append(userGrabLandlord, msg.UserListGrabLandlord{Score: v, ChairID: k})
		}

		grabLandlordScene.UserListGrabLandlord = userGrabLandlord
		grabLandlordScene.CurrentChairID = it.currentChairID
		userItem.WriteMsg(&grabLandlordScene)
	case global.GameStatusPlay:
		var playScene msg.Game_S_PlayScene
		var userListStatus = make([]msg.UserListTrusteeship, 0)
		playScene.UserList = make([]msg.Game_S_User, 0)
		var userListPokerCount = make([]msg.UserListPokerCount, 0)
		it.userList.Range(func(chairID, userID interface{}) bool {
			userItemTemp, ok := user.List.Load(userID.(int32))
			if !ok {
				_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
				return true
			}
			playScene.UserList = append(playScene.UserList, UserInfoToGrameUser(userItemTemp.(*user.Item)))
			return true
		})
		sceneTime := int64(conf.GetServer().GameCPTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneTime < 0 {
			sceneTime = 0
		}
		for k, v := range it.userListTrusteeship {
			userListStatus = append(userListStatus, msg.UserListTrusteeship{ChairID: k, Status: v})
		}
		playScene.UserListTrusteeship = userListStatus
		playScene.SceneStartTime = sceneTime
		playScene.CurrentChairID = it.currentChairID
		playScene.BankerChairID = it.bankerChairID
		for k, _ := range it.userPokers[userItem.ChairID] {
			playScene.UserPoker = append(playScene.UserPoker, k)
		}
		for k, v := range it.userPokers {
			userListPokerCount = append(userListPokerCount, msg.UserListPokerCount{ChairID: k, Count: int32(len(v))})
		}
		playScene.UserListPokerCount = userListPokerCount
		playScene.Multiple = it.userGrabLandlord[it.bankerChairID]
		playScene.SumMultiple = it.currentMultiple
		playScene.LandlordPokers = it.dizPokers
		playScene.NearestChairID = it.nearestChairID
		playScene.NearestPokers = it.nearestPokers
		playScene.NearestCardType = it.nearestCardType
		playScene.PokerCount = make([]int32, 0)
		for chairID, v := range it.userPokers {
			if chairID == userItem.ChairID {
				continue
			}
			for data, _ := range v {
				playScene.PokerCount = append(playScene.PokerCount, data)
			}
		}
		userItem.WriteMsg(&playScene)
	}
}

// 测试代码
func getTestPokers() map[int32][]int32 {
	var card = make(map[int32][]int32, 0)
	var pokers = []int32{
		//0方块: 2 3 4 5 6 7 8 9 10 J Q K A
		0x07, 0x08, 0x09, 0x0A, 0x0B, 0x21, 0x38, 0x39, 0x3C, 0x5F, 0x18, 0x19, 0x28, 0x29, 0x36, 0x37, 0x3D,
		//1梅花: 2 3 4 5 6 7 8 9 10 J Q K A
		0x13, 0x14, 0x15, 0x16, 0x17, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1A, 0x1B, 0x1C, 0x2A, 0x2B, 0x3A, 0x3B,
		//2红桃: 2 3 4 5 6 7 8 9 10 J Q K A
		0x23, 0x24, 0x25, 0x26, 0x27, 0x32, 0x33, 0x34, 0x35, 0x2C, 0x2D, 0x1D, 0x0C, 0x11, 0x31, 0x0D, 0x01,
		//3黑桃: 2 3 4 5 6 7 8 9 10 J Q K A
		//4鬼：大鬼 小鬼 十进制 79 95
		0x4F, 0x22, 0x12}
	card[0] = pokers[:17]
	card[1] = pokers[17:34]
	card[2] = pokers[34:51]
	card[3] = pokers[51:]
	return card
}

//开始游戏
func (it *Item) onEventGameStart() {
	it.gameStatus = global.GameStatusJF
	fmt.Println("开始游戏，清空上局数据")
	it.userListPoker = make(map[int32][]int32, 0)
	it.dizPokers = make([]int32, 0)
	//发牌
	pokers := logic.Client.DispatchTableCard()
	/*todo del**************debug**************/
	//var isTest = false
	//var testChairId int32
	//it.userList.Range(func(chairID, uid interface{}) bool {
	//	if uid.(int32) == 138684 || uid.(int32) == 138652 {
	//		isTest = true
	//		testChairId = chairID.(int32)
	//	}
	//	return true
	//})
	//if isTest {
	//	pokers = getTestPokers()
	//}
	/*todo del*************debug**************/
	for i := int32(0); i < 3; i++ {
		it.userPokers[i] = make(map[int32]interface{})
		for _, v := range pokers[i] {
			it.userPokers[i][v] = struct{}{}
			it.userListPoker[i] = append(it.userListPoker[i], v)
		}
	}
	it.userListPoker[3] = append(it.userListPoker[3], pokers[3]...)

	for _, v := range pokers[3] {
		it.dizPokers = append(it.dizPokers, v)
	}
	it.userList.Range(func(chairID, uid interface{}) bool {
		v, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("onEventGameStart err %d", uid)
			return false
		}

		v.(*user.Item).WriteMsg(&msg.Game_S_StartGame{
			UserPoker: pokers[chairID.(int32)],
			RecordID:  it.roundOrder,
		})
		return true
	})
	// 初始化参数
	it.userGrabLandlord = make(map[int32]int32, 0)
	//设置时间
	it.sceneStartTime = time.Now().Unix()
	//随机用户开叫
	it.currentChairID = rand.RandInterval(0, store.GameControl.GetGameInfo().ChairCount-1)
	//发送当前用户
	it.sendAllUser(&msg.Game_S_CurrentUser{
		CurrentChairID: it.currentChairID,
	})

	//启动定时器
	it.cronTimer = time.NewTimer(time.Second * time.Duration(conf.GetServer().GameJFTime))
	go it.onEventGameTimer()
}

//结束游戏 判断是否钱够不够 直接踢出去
func (it *Item) onEventGameConclude() {
	//设置游戏状态
	it.gameStatus = global.GameStatusFree
	it.sceneStartTime = time.Now().Unix()
	it.cronTimer.Stop()
	//判断春天
	if it.dizCPCount == 1 || it.nongmCPCount == 0 {
		it.currentMultiple *= 2
	}
	// 春天和反春
	var SpringType int32
	if it.nongmCPCount == 0 {
		SpringType = 1
	} else if it.dizCPCount == 1 {
		SpringType = 2
	}
	it.systemScore = 0
	it.userListLoss = make(map[int32]float32)

	it.systemScore, it.userListLoss, it.userTax = logic.Client.GetSystemLoss(it.bankerChairID, it.userList, it.currentMultiple, it.nearestChairID == it.bankerChairID)
	log.Logger.Debugf("===游戏结束 结算==桌子号:%v loss=%v 倍数:%v", it.tableID, it.userListLoss, it.currentMultiple)

	//更新库存
	store.GameControl.ChangeStore(it.systemScore)
	//记录游戏记录
	it.onWriteGameRecord()
	//用户写分
	it.onWriteGameScore()
	// 热更数据
	it.onUpdateAgentData()

	//结算通知
	newListLoss := make(map[int32]float32)
	handPoker := make([]msg.Game_S_HandPoker, 0)
	for k, cards := range it.userPokers {
		tempCards := make([]int32, 0)
		for card, _ := range cards {
			tempCards = append(tempCards, card)
		}
		tempHand := msg.Game_S_HandPoker{
			Poker:   tempCards,
			ChairID: k,
		}
		handPoker = append(handPoker, tempHand)
	}
	it.userList.Range(func(chairID, uid interface{}) bool {
		value, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("onEventGameConclude err 用户不存在!")
			return true
		}
		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			newListLoss[chairID.(int32)] = value.(*user.Item).UserGold
		} else {
			newListLoss[chairID.(int32)] = value.(*user.Item).UserDiamond
		}
		return true
	})
	//log.Logger.Debugf("===游戏结束 结算==桌子号:%v loss=%v 倍数:%v", it.tableID, it.userListLoss, it.currentMultiple)
	var resultList = make([]msg.Game_S_GameResult, 0)
	var userListMoney = make([]msg.Game_S_GameResult, 0)
	for k, v := range it.userListLoss {
		resultList = append(resultList, msg.Game_S_GameResult{ChairID: k, Result: v})
	}
	for k, v := range newListLoss {
		userListMoney = append(userListMoney, msg.Game_S_GameResult{ChairID: k, Result: v})
	}
	it.sendAllUser(&msg.Game_S_GameConclude{
		UserListLoss:    resultList,
		UserHandPoker:   handPoker,
		SpringType:      SpringType,
		CurrentMultiple: it.currentMultiple,
		UserListMoney:   userListMoney,
	})
	//判断积分是否足够
	oldUserList := make(map[int32]int32, 0) // 座位号 uid
	it.userList.Range(func(chairID, uid interface{}) bool {

		value, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("onEventGameConclude err 用户不存在!")
			return true
		}
		mySelf := value.(*user.Item)
		// 如果游戏结束的时候用户在离线状态 解锁用户
		if mySelf.Status == user.StatusOffline {
			// 起立 强制退出
			it.OnActionUserStandUp(value.(*user.Item), true)
			// map 中移除
			user.List.Delete(uid.(int32))
			return true
		} else if !mySelf.IsRobot() {
			mySelf.Status = user.StatusFree
		}

		// 取消托管
		it.OnUnAutoManage(mySelf)
		oldUserList[mySelf.ChairID] = mySelf.UserID
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
				user.List.Delete(uid.(int32))
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
				user.List.Delete(uid.(int32))
				return true
			}
		}

		// 起立 结束游戏全部退出
		it.OnActionUserStandUp(value.(*user.Item), true)
		return true
	})
	// 清空数据
	it.currentChairID = it.bankerChairID
	it.userPrepare = sync.Map{}
	it.userPokers = make(map[int32]map[int32]interface{})
	it.dizPokers = []int32{}
	it.mutex = sync.Mutex{}
	it.userList = sync.Map{}
	it.userGrabLandlord = make(map[int32]int32)
	it.nearestChairID = -1
	it.nearestPokers = make([]int32, 0)
	it.nearestCardType = 0
	it.nearestMaxCard = 0
	it.currentMultiple = 0
	it.dizCPCount = 0
	it.userCount = 0
	it.robotCount = 0
	it.nongmCPCount = 0
	it.userListTrusteeship = make(map[int32]bool)
	it.drawID = ""
	//go func() {
	//	for len(oldUserList) > 0 {
	//		t := time.NewTicker(time.Second * 30)
	//		select {
	//		case <-t.C:
	//			it.OnMoveUserByChairID(oldUserList)
	//			return
	//		}
	//	}
	//}()
}

// 移除座位
func (it *Item) OnMoveUserByChairID(OldUserList map[int32]int32) {
	for chairID, oldUserId := range OldUserList {
		UserID, ok := it.userList.Load(chairID)
		if ok {
			_, exist := it.userPrepare.Load(chairID)
			userItem, IsExist := user.List.Load(UserID)
			if !exist && oldUserId == UserID.(int32) && IsExist && userItem.(*user.Item).Status == user.StatusFree {
				it.userList.Delete(chairID)
				if it.gameStatus == global.GameStatusFree && it.userCount > 0 {
					it.userCount--
					it.sendAllUser(&msg.Game_S_StandUpNotify{
						ChairID: chairID,
					})
				}
				if err := mysql.GameClient.UnLock(userItem.(*user.Item).UserID); err != nil {
					_ = log.Logger.Errorf("解锁用户失败 err %v", err)
					return
				}
				log.Logger.Debugf("结束游戏 桌子号%v,游戏状态：%v,桌子人数:%v，主动移除座位=%v,uid:%v", it.tableID, it.gameStatus, it.userCount, chairID, oldUserId)
			}
		}
	}
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
		intLostCount,    //失败盘数
		intDrawCount,    //和局盘数
		intFleeCount,    //逃跑数目
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

		if !redis.GameClient.IsExistsDiamond(userItem.(*user.Item).UserID) && !userItem.(*user.Item).IsRobot() {
			redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userItem.(*user.Item).UserDiamond)
		}
		UserDiamond, _ := redis.GameClient.GetDiamond(userItem.(*user.Item).UserID)
		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			userItem.(*user.Item).UserGold += v
		} else {
			userItem.(*user.Item).UserDiamond = float32(UserDiamond)
			userItem.(*user.Item).UserDiamond += v
		}
		// 过滤机器人
		if userItem.(*user.Item).BatchID != -1 {
			continue
		}
		redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userItem.(*user.Item).UserDiamond)

		host, _, _ := net.SplitHostPort(userItem.(*user.Item).Agent.RemoteAddr().String())
		errorCode, errorMsg := mysql.GameClient.WriteUserScore(uid.(int32),
			v,
			store.GameControl.GetGameInfo().DeductionsType,
			it.userTax[key],
			intWinCount,
			intLostCount,
			intDrawCount,
			intFleeCount,
			conf.GetServer().GameCPTime,
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

		if errorCode != common.StatusOK {
			_ = log.Logger.Errorf(" mysql GSP_WriteGameScore存储过程 %n %s ", errorCode, errorMsg)
			return
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

	var endTime = time.Now().Format("2006-01-02 15:04:05")
	newUserListLoss := make(map[int32]float32, 0)
	newUserTax := make(map[int32]float32, 0)
	newUserMoney := make(map[int32]float32, 0)

	var startTime = time.Unix(it.sceneStartTime, 0).Format("2006-01-02 15:04:05")
	var taxSum float32 = 0
	for k, v := range it.userTax {
		taxSum += v
		newUserTax[k] = v
	}

	var userList = make(map[int32]int32, 0)
	var userHandPoker = make(map[int32][]int32, 0)

	for k, v := range it.userPokers {
		for val, _ := range v {
			userHandPoker[k] = append(userHandPoker[k], val)
		}
	}
	it.userList.Range(func(key, value interface{}) bool {
		userList[key.(int32)] = value.(int32)
		if userItem, ok := user.List.Load(value.(int32)); ok {
			newUserMoney[userItem.(*user.Item).ChairID] = userItem.(*user.Item).UserDiamond
		}
		return true
	})
	// 春天和反春
	var SpringType int32
	if it.nongmCPCount == 0 {
		SpringType = 1
	} else if it.dizCPCount == 1 {
		SpringType = 2
	}

	for k, v := range it.userListLoss {
		newUserListLoss[k] = v
	}
	// 局数详细信息
	Detail := struct {
		UserList          map[int32]int32   `json:"user_list"`            // uid --->椅子号
		UserListPoker     map[int32][]int32 `json:"user_original_poker"`  //玩家初始扑克
		UserListHandPoker map[int32][]int32 `json:"user_list_hand_poker"` //玩家剩余扑克
		BankerChairID     int32             `json:"banker_chair_id"`      //地主椅子ID
		CurrentMultiple   int32             `json:"current_multiple"`     // 当前倍数
		SpringType        int32             `json:"spring_type"`          // 春天类型1,为春天2为反春
		UserListLoss      map[int32]float32 `json:"user_list_loss"`       //用户盈亏 座位号
		UserTax           map[int32]float32 `json:"user_tax"`             //用户税收
		UserMoney         map[int32]float32 `json:"user_pre_money"`       // 更改前的金币
	}{
		UserList:          userList,
		UserListPoker:     it.userListPoker,
		UserTax:           newUserTax,
		UserListLoss:      newUserListLoss,
		BankerChairID:     it.bankerChairID,
		SpringType:        SpringType,
		UserListHandPoker: userHandPoker,
		CurrentMultiple:   it.currentMultiple,
		UserMoney:         newUserMoney,
	}
	detail, _ := json.Marshal(Detail)
	errorCode, errorMsg, drawID := mysql.GameClient.WriteGameRecord(
		it.tableID,
		it.userCount,
		it.robotCount,
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

//用户坐下
func (it *Item) OnActionUserSitDown(args ...interface{}) {
	fmt.Printf("%c[1;40;31m用户坐下=====%c[0m\n", 0x1B, 0x1B)
	userItem := args[0].(*user.Item)
	downLoad := func() {
		it.userList.Range(func(key, value interface{}) bool {
			userInfo, ok := user.List.Load(value.(int32))
			if ok {
				if userInfo.(*user.Item).IsRobot() {
					_, ex := RobotQueue.Load(userInfo.(*user.Item).UserID)
					if !ex {
						RobotQueue.Store(userInfo.(*user.Item).UserID, userInfo.(*user.Item))
						RobotCount++
					}
				} else {
					_, ex := UserQueue.Load(userInfo.(*user.Item).UserID)
					if !ex {
						UserQueue.Store(userInfo.(*user.Item).UserID, userInfo.(*user.Item))
						UserCount++
					}
				}
			}
			return true
		})
	}
	//校验是否满人
	if it.userCount >= store.GameControl.GetGameInfo().ChairCount {
		downLoad()
		if !userItem.IsRobot() {
			global.NoticeLoadMath <- 0
		}
		return
	}
	// 用户在其它桌子上未起立，则移除
	//if userItem.TableID >= 0 && userItem.ChairID >= 0 {
	//	List[userItem.TableID].OnMoveUserByChairID(map[int32]int32{userItem.ChairID: userItem.UserID})
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
	//log.Logger.Debugf("坐下中桌子号:%v,人数%v,坐下uid=%v,座位号%v,坐下前状态%v", it.tableID, it.userCount, userItem.UserID, userItem.ChairID, userItem.Status)

	it.mutex.Lock()
	//等待加入游戏用户列表
	for i := int32(0); i < store.GameControl.GetGameInfo().ChairCount; i++ {
		_, ok := it.userList.Load(i)
		if ok {
			continue
		}
		prepare, exist := it.userPrepare.Load(i)
		if exist && !prepare.(bool) {
			continue
		}
		//加入游戏
		userItem.SitDown(it.tableID, i)
		it.userList.Store(i, userItem.UserID)
		it.userPrepare.Store(i, false)
		it.userCount++
		if userItem.IsRobot() {
			it.robotCount++
		}
		break
	}
	it.mutex.Unlock()

	log.Logger.Debugf("成功坐下--桌子号:%v,座位号%v,人数%v,uid=%v,用户状态%v", it.tableID, userItem.ChairID, it.userCount, userItem.UserID, userItem.Status)

	go func() {
		t := time.NewTimer(time.Second * 5)
		for {
			select {
			case <-t.C:
				if it.gameStatus == global.GameStatusFree && it.userCount != global.TablePlayCount {
					log.Logger.Debug("用户重新加入队列", userItem.UserID)
					it.OnActionUserStandUp(userItem, true)
					downLoad()
					// 清空数据
					it.userPrepare = sync.Map{}
					it.userPokers = make(map[int32]map[int32]interface{})
					it.dizPokers = []int32{}
					it.mutex = sync.Mutex{}
					it.userList = sync.Map{}
					it.userGrabLandlord = make(map[int32]int32)
					it.nearestChairID = -1
					it.nearestPokers = make([]int32, 0)
					it.nearestCardType = 0
					it.nearestMaxCard = 0
					it.currentMultiple = 0
					it.dizCPCount = 0
					it.userCount = 0
					it.robotCount = 0
					it.nongmCPCount = 0
					it.userListTrusteeship = make(map[int32]bool)
					it.drawID = ""
					if !userItem.IsRobot() {
						global.NoticeLoadMath <- 0
					}
					return
				}

			}
			if it.userCount != global.TablePlayCount {
				time.Sleep(time.Microsecond * 500)
			}
			t.Stop()
			it.onEventSendGameScene(userItem)
			it.OnUserPrepare(userItem, &msg.Game_C_UserPrepare{})
			break
		}
	}()

	// 发送场景消息
	//for it.userCount != global.TablePlayCount {
	//	time.Sleep(time.Microsecond * 500)
	//}
	//it.onEventSendGameScene(userItem)
	//it.OnUserPrepare(userItem, &msg.Game_C_UserPrepare{})

	//it.sendAllUser(&msg.Game_S_SitDownNotify{
	//	Data: &msg.Game_S_User{
	//		UserID:       userItem.GetUserInfo().UserID,
	//		NikeName:     userItem.GetUserInfo().NikeName,
	//		UserDiamond:  userItem.GetUserInfo().UserDiamond,
	//		MemberOrder:  userItem.GetUserInfo().MemberOrder,
	//		HeadImageUrl: userItem.GetUserInfo().HeadImageUrl,
	//		FaceID:       userItem.GetUserInfo().FaceID,
	//		RoleID:       userItem.GetUserInfo().RoleID,
	//		SuitID:       userItem.GetUserInfo().SuitID,
	//		PhotoFrameID: userItem.GetUserInfo().PhotoFrameID,
	//		TableID:      userItem.GetUserInfo().TableID,
	//		ChairID:      userItem.GetUserInfo().ChairID,
	//		Status:       userItem.GetUserInfo().SuitID,
	//		Gender:       userItem.GetUserInfo().Gender,
	//	},
	//})
}

//用户起立
func (it *Item) OnActionUserStandUp(args ...interface{}) {
	userItem := args[0].(*user.Item)
	flag := args[1].(bool)
	//log.Logger.Debugf("OnActionUserStandUp 桌子id=%v,椅子id=%v,桌子人数%v,桌子状态%v,uid=%v,用户状态=%v", it.tableID, userItem.ChairID, it.userCount, it.gameStatus, userItem.UserID, userItem.Status)
	if !flag {
		//检测是否游戏中
		if it.gameStatus == global.GameStatusPlay {
			_ = log.Logger.Errorf("OnActionUserStandUp %s", "游戏中不允许退出")
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.StandUpError2,
				ErrorMsg:  "游戏中不允许退出!",
			})
			return
		}
	}
	//取消准备
	it.userPrepare.Delete(userItem.ChairID)

	//移出游戏用户列表
	oldChairID := userItem.ChairID
	it.userList.Delete(oldChairID)
	//if it.userCount > 0 {
	//	it.userCount--
	//}
	//if userItem.IsRobot() {
	//	it.robotCount--
	//}
	userItem.StandUp()
	//解锁
	if err := mysql.GameClient.UnLock(userItem.UserID); err != nil {
		_ = log.Logger.Errorf("解锁用户失败 err %v", err)
		return
	}

	////通知其他玩家
	//it.sendAllUser(&msg.Game_S_StandUpNotify{
	//	ChairID: oldChairID,
	//})
	// 有人准备，重新进入匹配
	//it.onResetEnterMatch()
}

// 起立如果有人准备，则为准备的人重新进入匹配
func (it *Item) onResetEnterMatch() {
	it.userPrepare.Range(func(chairId, Prepare interface{}) bool {
		if Prepare.(bool) {
			userID, ok := it.userList.Load(chairId)
			if ok {
				userItem, exist := user.List.Load(userID.(int32))
				if exist && !userItem.(*user.Item).IsRobot() {
					UserQueue.Store(userItem.(*user.Item).UserID, userItem)
					UserCount++
					it.userCount--
					log.Logger.Debugf("onResetEnterMatch 起立准备 桌子id=%v,椅子id=%v,桌子人数%v,桌子状态%v,uid=%v,用户状态=%v", it.tableID, userItem.(*user.Item).ChairID, it.userCount, it.gameStatus, userItem.(*user.Item).UserID, userItem.(*user.Item).Status)

				}
			}
		}
		return true
	})
}

// 用户主动托管
func (it *Item) OnAutoManage(args ...interface{}) {
	userItem := args[0].(*user.Item)
	it.userListTrusteeship[userItem.ChairID] = true
	if userItem.ChairID == it.currentChairID && it.gameStatus != global.GameStatusFree {
		it.cronTimer.Reset(time.Second * 2)
	}
	it.sendAllUser(&msg.Game_S_AutoManage{
		ChairID: userItem.ChairID,
	})
}

// 用户取消托管
func (it *Item) OnUnAutoManage(args ...interface{}) {
	userItem := args[0].(*user.Item)
	delete(it.userListTrusteeship, userItem.ChairID)
	if userItem.ChairID == it.currentChairID && it.gameStatus != global.GameStatusFree {
		var sceneStartTime int64
		if it.gameStatus == global.GameStatusJF {
			sceneStartTime = int64(conf.GetServer().GameJFTime) - (time.Now().Unix() - it.sceneStartTime)
		} else if it.gameStatus == global.GameStatusPlay {
			sceneStartTime = int64(conf.GetServer().GameCPTime) - (time.Now().Unix() - it.sceneStartTime)
		}
		if sceneStartTime > 0 {
			it.cronTimer.Reset(time.Second * time.Duration(sceneStartTime))
		}
		log.Logger.Debugf("玩家取消托管--桌子号:%v,桌子状态:%v,玩家座位号:%v,xz=%v,重新设置时间=%v", it.tableID, it.gameStatus, userItem.ChairID, it.currentChairID, sceneStartTime)
	}
	it.sendAllUser(&msg.Game_S_UnAutoManage{
		ChairID: userItem.ChairID,
	})
}

func (it *Item) OnActionUserOffLine(args ...interface{}) { //用户断线
	userItem := args[0].(*user.Item)

	//设置用户状态
	userItem.Status = user.StatusOffline
	//设置用户托管
	it.userListTrusteeship[userItem.ChairID] = true

	//通知其他玩家
	it.sendOtherUser(userItem.ChairID, &msg.Game_S_OffLineNotify{
		ChairID: userItem.ChairID,
	})
}

//用户重入
func (it *Item) OnActionUserReconnect(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem.Status == user.StatusOffline {
		//设置用户状态
		if it.gameStatus == global.GameStatusFree {
			userItem.Status = user.StatusFree
		} else {
			userItem.Status = user.StatusPlaying
		}

		//解除托管
		it.userListTrusteeship[userItem.ChairID] = false

		//通知其他玩家
		it.sendOtherUser(userItem.ChairID, &msg.Game_S_OnLineNotify{
			ChairID: userItem.ChairID,
		})
	}

	//发送场景消息
	it.onEventSendGameScene(userItem)
}

//用户准备
func (it *Item) OnUserPrepare(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if it.gameStatus != global.GameStatusFree {
		log.Logger.Error("当前不能准备", userItem.UserID, userItem.Status)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError1,
		//	ErrorMsg:  "当前不能准备!",
		//})
		return
	}
	userPrepareStatus, ok := it.userPrepare.Load(userItem.ChairID)
	if ok && userPrepareStatus.(bool) {
		log.Logger.Error("不能重复准备", userItem.UserID, userItem.Status)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError2,
		//	ErrorMsg:  "不能重复准备!",
		//})
		return
	}
	it.userPrepare.Store(userItem.ChairID, true)

	//it.sendAllUser(&msg.Game_S_UserPrepare{
	//	ChairID: userItem.ChairID,
	//})

	//检测是否都准备
	var userPrepareCount int
	it.userPrepare.Range(func(key, value interface{}) bool {
		if value.(bool) {
			userPrepareCount++
		}
		return true
	})
	if userPrepareCount == global.TablePlayCount {
		randNum := rand.Krand(6, 3)

		it.roundOrder = fmt.Sprintf("%v%v%s", conf.GetServer().GameID, time.Now().Unix(), randNum)
		go it.onEventGameStart()
	}
}

//用户取消准备
func (it *Item) OnUserUnPrepare(args ...interface{}) {
	userItem := args[0].(*user.Item)

	if it.gameStatus != global.GameStatusFree {
		log.Logger.Error("当前不能取消准备", userItem.UserID, userItem.Status)

		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError1,
		//	ErrorMsg:  "当前不能取消准备!",
		//})
		return
	}

	userPrepareStatus, ok := it.userPrepare.Load(userItem.ChairID)
	if ok && !userPrepareStatus.(bool) {
		log.Logger.Error("暂未准备", userItem.UserID, userItem.Status)

		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError2,
		//	ErrorMsg:  "暂未准备!",
		//})
		return
	}

	it.userPrepare.Delete(userItem.ChairID)

	it.sendAllUser(&msg.Game_S_UserUnPrepare{
		ChairID: userItem.ChairID,
	})
}

//玩家叫分
func (it *Item) OnUserGrabLandlord(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserGrabLandlord)

	//判断游戏状态
	if it.gameStatus != global.GameStatusJF {
		_ = log.Logger.Error("OnUserGrabLandlord游戏状态异常:", it.gameStatus)
		//userItem.Close()
		return
	}

	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JFError1,
			ErrorMsg:  "叫分失败, 用户不在用户列表里",
		})
		//userItem.Close()
		return
	}

	//判断是否当前操作玩家
	if userItem.ChairID != it.currentChairID {
		_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 非当前用户出牌")
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JFError1,
		//	ErrorMsg:  "叫分失败, 非当前用户出牌!",
		//})
		//userItem.Close()
		return
	}

	//记录用户叫分
	it.userGrabLandlord[userItem.ChairID] = m.Multiple
	//通知其他用户
	it.sendAllUser(&msg.Game_S_UserGrabLandlord{
		ChairID:  userItem.ChairID,
		Multiple: m.Multiple,
	})

	if m.Multiple == 3 || len(it.userGrabLandlord) == 3 { //定地主
		bankerChairID := it.confirmBanker()
		if bankerChairID == -1 {
			//没人叫分，重新开始
			it.sendAllUser(&msg.Game_S_GameRestart{})
			it.cronTimer.Stop()
			time.Sleep(time.Second * 2) // 重新开始延迟2s，客户端跑动画
			it.onEventGameStart()
			return
		}
		/********打印测试***********/
		var tempPoker = make(map[int32][]int32, 0)
		var tempDiz []int32
		for chairID, v := range it.userPokers {
			for card, _ := range v {
				tempPoker[chairID] = append(tempPoker[chairID], card&0x0F)
			}
		}
		for _, card := range it.dizPokers {
			tempDiz = append(tempDiz, card&0x0F)
		}
		log.Logger.Debugf("确认地主--桌子号:%v,原始牌:%v,地主牌:%v", it.tableID, tempPoker, tempDiz)
		/********打印测试***********/
		//确定地主
		it.gameStatus = global.GameStatusPlay
		it.bankerChairID = bankerChairID
		it.currentChairID = it.bankerChairID
		it.nearestChairID = it.bankerChairID
		it.currentMultiple = it.userGrabLandlord[it.bankerChairID]
		for _, v := range it.dizPokers {
			it.userPokers[it.bankerChairID][v] = struct{}{}
		}
		//通知所有人
		it.sendAllUser(&msg.Game_S_StartCPDetermine{
			CurrentChairID: it.currentChairID,
			Multiple:       it.userGrabLandlord[it.bankerChairID],
			LandlordPokers: it.dizPokers,
		})

		//重置出牌超时时间
		it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameCPTime))
		return
	}

	//设置时间
	it.sceneStartTime = time.Now().Unix()
	//下一个玩家
	it.currentChairID = (it.currentChairID + 1) % store.GameControl.GetGameInfo().ChairCount
	//发送当前用户
	it.sendAllUser(&msg.Game_S_CurrentUser{
		CurrentChairID: it.currentChairID,
	})
	//重置出牌超时时间
	it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameJFTime))
}

//确认地主
func (it *Item) confirmBanker() int32 {
	var multiple int32
	multiple = 0
	var chairID int32
	chairID = -1

	for k, v := range it.userGrabLandlord {
		if v > multiple {
			multiple = v
			chairID = k
		}
	}

	return chairID
}

//玩家出牌
func (it *Item) OnUserCP(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserCP)

	/**********打印************/
	var outCards []int32
	for _, card := range m.Pokers {
		tempCard := card & 0x0F
		outCards = append(outCards, tempCard)
	}
	// log.Logger.Debug("出牌 OnUserCP:", it.currentChairID, userItem.UserID, outCards, it.userListTrusteeship)
	//判断游戏状态
	if it.gameStatus != global.GameStatusPlay {
		_ = log.Logger.Error("OnUserCP游戏状态异常:", it.gameStatus)
		//userItem.Close()
		return
	}

	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.CPError1,
			ErrorMsg:  "出牌失败, 用户不在用户列表里",
		})
		//userItem.Close()
		return
	}

	//判断是否当前操作玩家
	if userItem.ChairID != it.currentChairID {
		_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 非当前用户出牌")
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.CPError3,
		//	ErrorMsg:  "出牌失败, 非当前用户出牌!",
		//})
		//userItem.Close()
		return
	}

	//判断牌型是否合格
	cardLen := len(m.Pokers)
	cardType, maxCard := logic.Client.GetPokerType(m.Pokers)
	if cardType <= 0 || cardType >= 12 {
		_ = log.Logger.Errorf("OnUserOutCard err %s cards=%v", "出牌失败, 出牌数据异常", m.Pokers)
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.CPError3,
			ErrorMsg:  "出牌失败, 牌型错误!",
		})
		//userItem.Close()
		return
	}

	//不是自己先出
	if it.nearestChairID != userItem.ChairID {
		if cardType != it.nearestCardType {
			if cardType < 10 {
				_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 出牌数据异常")
				_ = log.Logger.Errorf("OnUserOutCard err 最近出牌最大牌%d,类型%d=%v 现在出牌最大牌%d,类型%d=%v",
					logic.Client.GetLogicValue(it.nearestMaxCard), it.nearestCardType, it.nearestPokers,
					logic.Client.GetLogicValue(maxCard), cardType, m.Pokers)
				userItem.WriteMsg(&msg.Game_S_ReqlyFail{
					ErrorCode: global.CPError3,
					ErrorMsg:  "出牌失败, 出牌数据异常",
				})
				//userItem.Close()
				return
			}
		} else {
			if cardLen != len(it.nearestPokers) || logic.Client.GetLogicValue(maxCard) <= logic.Client.GetLogicValue(it.nearestMaxCard) {
				_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 出牌数据异常")
				_ = log.Logger.Errorf("OnUserOutCard err 最近出牌最大牌%d,类型%d=%v 现在出牌最大牌%d,类型%d=%v",
					logic.Client.GetLogicValue(it.nearestMaxCard), it.nearestCardType, it.nearestPokers,
					logic.Client.GetLogicValue(maxCard), cardType, m.Pokers)
				userItem.WriteMsg(&msg.Game_S_ReqlyFail{
					ErrorCode: global.CPError3,
					ErrorMsg:  "出牌失败, 出牌数据异常",
				})
				//userItem.Close()
				return
			}
		}
	}
	var tempPokers = make(map[int32]int32)
	//检查牌类型
	for _, v := range m.Pokers {
		_, exist := it.userPokers[userItem.ChairID][v]
		if !exist {
			_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 非手牌数据")
			_ = log.Logger.Errorf("OnUserOutCard err 最近出牌最大牌%d,类型%d=%v 现在出牌最大牌%d,类型%d=%v，服务器原始牌=[%v]，用户id=%d,座位号%d",
				logic.Client.GetLogicValue(it.nearestMaxCard), it.nearestCardType, it.nearestPokers,
				logic.Client.GetLogicValue(maxCard), cardType, m.Pokers, it.userPokers, userItem.UserID, userItem.ChairID)
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.CPError3,
				ErrorMsg:  "出牌失败, 非手牌数据",
			})
			//userItem.Close()
			return
		}
		if _, exist = tempPokers[v]; exist {
			_ = log.Logger.Errorf("OnUserOutCard err 现在出牌最大牌%d,类型%d=%v，用户id=%d,座位号%d",
				logic.Client.GetLogicValue(maxCard), cardType, m.Pokers, userItem.UserID, userItem.ChairID)
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.CPError3,
				ErrorMsg:  "出牌失败, 出牌数据异常",
			})
			return
		} else {
			tempPokers[v] = v
		}
		// delete(it.userPokers[userItem.ChairID], v)
	}
	//去掉手牌
	for _, v := range m.Pokers {
		delete(it.userPokers[userItem.ChairID], v)
	}
	it.nearestChairID = userItem.ChairID
	it.nearestCardType = cardType
	it.nearestMaxCard = maxCard
	it.nearestPokers = m.Pokers

	//记录出牌次数
	if userItem.ChairID == it.bankerChairID {
		it.dizCPCount++
	} else {
		it.nongmCPCount++
	}

	//判断是否是炸弹
	if cardType >= 10 {
		it.currentMultiple *= 2
	}
	pokerNum := int32(len(it.userPokers[userItem.ChairID]))
	log.Logger.Debugf("出牌 OnUserCP后: 桌子号:%v,椅子号:%v,uid:=%v,出牌:%v,托管%v", it.tableID, it.currentChairID, userItem.UserID, outCards, it.userListTrusteeship)
	//出牌通知
	it.sendAllUser(&msg.Game_S_UserCP{
		ChairID:         userItem.ChairID,
		Pokers:          m.Pokers,
		PokerType:       cardType,
		CurrentMultiple: it.currentMultiple,
		PokerCount:      pokerNum,
	})
	//判断是否结束
	if pokerNum == 0 {
		it.onEventGameConclude()
		return
	}

	//设置时间
	it.sceneStartTime = time.Now().Unix()
	//下一个玩家
	it.currentChairID = (it.currentChairID + 1) % store.GameControl.GetGameInfo().ChairCount
	//发送当前用户
	it.sendAllUser(&msg.Game_S_CurrentUser{
		CurrentChairID: it.currentChairID,
	})
	//重置出牌超时时间
	// 玩家是否托管
	trusteeshipStatus, exist := it.userListTrusteeship[it.currentChairID]
	if exist && trusteeshipStatus {
		it.cronTimer.Reset(time.Second * 2)
		return
	}
	it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameCPTime))
}

// 托管出牌
func (it *Item) GetAutoManagePokers() []int32 {
	var outPoker []int32
	min := int32(0x5F)
	var HandPokers []int32
	for key, _ := range it.userPokers[it.currentChairID] {
		HandPokers = append(HandPokers, key)
		if logic.Client.GetLogicValue(min) > logic.Client.GetLogicValue(key) {
			min = key
		}
	}
	if it.nearestChairID == it.currentChairID {
		outPoker = logic.Client.GetAutoOutPokers(min, HandPokers)
	} else {
		outPoker = logic.Client.GetAutoOutPokersByType(it.nearestCardType, it.nearestPokers, HandPokers)
		log.Logger.Debugf("--托管接牌桌子号:%v,椅子id=%v，出牌:%v", it.tableID, it.currentChairID, outPoker)
		if len(outPoker) == 0 {
			outPoker = logic.Client.GetBomb(it.nearestCardType, HandPokers)
		} else if len(outPoker) != len(it.nearestPokers) {
			outPoker = make([]int32, 0)
		}
	}
	return outPoker
}

//玩家过牌
func (it *Item) OnUserPass(args ...interface{}) {
	userItem := args[0].(*user.Item)
	_ = args[1].(*msg.Game_C_UserPass)
	//判断游戏状态
	if it.gameStatus != global.GameStatusPlay {
		_ = log.Logger.Error("OnUserPass游戏状态异常:", it.gameStatus)
		//userItem.Close()
		return
	}

	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserPass err %s", "过牌失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.PassError1,
			ErrorMsg:  "过牌失败, 用户不在用户列表里",
		})
		//userItem.Close()
		return
	}

	//判断是否当前操作玩家
	if userItem.ChairID != it.currentChairID {
		_ = log.Logger.Errorf("OnUserPass err %s", "过牌失败, 非当前用户出牌")
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.PassError1,
		//	ErrorMsg:  "过牌失败, 非当前用户出牌!",
		//})
		//userItem.Close()
		return
	}

	//自己先出
	if it.nearestChairID == userItem.ChairID {
		_ = log.Logger.Errorf("OnUserPass err %s", "过牌失败, 用户不在用户列表里")
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.PassError1,
		//	ErrorMsg:  "过牌失败, 操作异常",
		//})
		//userItem.Close()
		return
	}
	log.Logger.Debugf("过牌 OnUserPass:后: 桌子号:%v,椅子号:%v,uid:=%v,托管%v", it.tableID, it.currentChairID, userItem.UserID, it.userListTrusteeship)
	//通知用户
	it.sendAllUser(&msg.Game_S_UserPass{
		ChairID: userItem.ChairID,
	})

	//设置时间
	it.sceneStartTime = time.Now().Unix()
	//下一个玩家
	it.currentChairID = (it.currentChairID + 1) % store.GameControl.GetGameInfo().ChairCount
	//发送当前用户
	it.sendAllUser(&msg.Game_S_CurrentUser{
		CurrentChairID: it.currentChairID,
	})

	//重置出牌超时时间
	trusteeshipStatus, exist := it.userListTrusteeship[it.currentChairID]
	if exist && trusteeshipStatus {
		it.cronTimer.Reset(time.Second * 2)
		return
	}
	it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameCPTime))

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

//获取空闲桌子 TODO 还可以优化 找到空闲最近满员的房间
func (it *Item) GetFreeTableID() int32 {
	if it.userCount < store.GameControl.GetGameInfo().ChairCount {
		//poor := store.GameControl.GetGameInfo().ChairCount - it.userCount
		return it.tableID
	}
	return -1
}

func (it *Item) GetRobotCount() int32 {
	var userPrepareNum int32
	it.userPrepare.Range(func(key, value interface{}) bool {
		userPrepareNum++
		return true
	})
	if userPrepareNum != it.userCount {
		return 1
	}
	return it.robotCount
}

//获取桌子状态
func (it *Item) GetGameStatus() int32 {
	return it.gameStatus
}

func UserInfoToGrameUser(userItem *user.Item) msg.Game_S_User {
	return msg.Game_S_User{
		UserID:       userItem.GetUserInfo().UserID,
		NikeName:     userItem.GetUserInfo().NikeName,
		UserDiamond:  userItem.GetUserInfo().UserDiamond,
		MemberOrder:  userItem.GetUserInfo().MemberOrder,
		HeadImageUrl: userItem.GetUserInfo().HeadImageUrl,
		FaceID:       userItem.GetUserInfo().FaceID,
		RoleID:       userItem.GetUserInfo().RoleID,
		SuitID:       userItem.GetUserInfo().SuitID,
		PhotoFrameID: userItem.GetUserInfo().PhotoFrameID,
		TableID:      userItem.GetUserInfo().TableID,
		ChairID:      userItem.GetUserInfo().ChairID,
		Status:       userItem.GetUserInfo().Status,
		Gender:       userItem.GetUserInfo().Gender,
	}
}
