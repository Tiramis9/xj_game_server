package table

import (
	"fmt"
	"math"
	"net"
	"xj_game_server/game/204_redmahjong/conf"
	"xj_game_server/game/204_redmahjong/game/logic"
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/204_redmahjong/msg"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/log"

	"sync"

	rand "xj_game_server/util/leaf/util"

	"time"
)

/*
红中麻将
*/
var List = make([]*Item, 0)

var TablesList sync.Map //桌子列表 人数 桌子号,桌子 key int32  map[int32]*Item

// 等待列表
var UserQueue sync.Map  // 真人等待列表
var RobotQueue sync.Map //机器人等待列表
var RobotCount int32    //机器人个数
var UserCount int32     //真人个数
var cronTimer *time.Timer

//初始化桌子
func OnInit() {
	TablesList = sync.Map{}
	tempTables := make(map[int32]*Item, 0)
	//初始化桌子
	for i := int32(0); i < store.GameControl.GetGameInfo().TableCount; i++ {
		temp := &Item{
			tableID:             i,
			gameStatus:          0,
			sceneStartTime:      0,
			userList:            sync.Map{},
			userCount:           0,
			robotCount:          0,
			userPrepare:         make(map[int32]bool), //准备玩家
			userListLoss:        make(map[int32]float32),
			userTax:             make(map[int32]float32),
			huType:              make([]int32, 0),
			systemScore:         0,
			outMjChairID:        -1,
			currentChairID:      -1,
			outMj:               0,
			sendMj:              0,
			userListAction:      make(map[int32]int32),
			userListOperate:     make(map[int32]int32),
			winChairID:          -1,
			winCount:            0,
			bankerChairID:       -1,
			sice:                make([]int32, 0),
			mahjongHeap:         make([]int32, 0),
			mjHeapHeadCount:     -1,
			mjHeapTailCount:     -1,
			userListMahjong:     make(map[int32]map[int32]int32),
			mahjongNum:          make(map[int32]int32, 0),
			userListDiskMj:      make(map[int32]*msg.DiskMahjongList),
			userListOutCard:     make(map[int32][]int32),
			maCount:             0,
			userListTrusteeship: sync.Map{},
			userListTing:        make(map[int32]bool),
			cronTimer:           &time.Timer{},
			mutex:               sync.Mutex{},
		}
		List = append(List, temp)
		tempTables[i] = temp
	}
	go initMatchTable()
	TablesList.Store(int32(0), tempTables) //初始化桌子列表
}

// 动态添加桌子
func appendTableList() {
	n := len(List)
	size := int32(n * 2)
	for i := int32(n); i < size; i++ {
		temp := &Item{
			tableID:             i,
			gameStatus:          0,
			sceneStartTime:      0,
			userList:            sync.Map{},
			userCount:           0,
			robotCount:          0,
			userPrepare:         make(map[int32]bool), //准备玩家
			userListLoss:        make(map[int32]float32),
			userTax:             make(map[int32]float32),
			huType:              make([]int32, 0),
			systemScore:         0,
			outMjChairID:        -1,
			currentChairID:      -1,
			outMj:               0,
			sendMj:              0,
			userListAction:      make(map[int32]int32),
			userListOperate:     make(map[int32]int32),
			winChairID:          -1,
			winCount:            0,
			bankerChairID:       -1,
			sice:                make([]int32, 0),
			mahjongHeap:         make([]int32, 0),
			mjHeapHeadCount:     -1,
			mjHeapTailCount:     -1,
			userListMahjong:     make(map[int32]map[int32]int32),
			mahjongNum:          make(map[int32]int32, 0),
			userListDiskMj:      make(map[int32]*msg.DiskMahjongList),
			userListOutCard:     make(map[int32][]int32),
			maCount:             0,
			userListTrusteeship: sync.Map{},
			userListTing:        make(map[int32]bool),
			cronTimer:           &time.Timer{},
			mutex:               sync.Mutex{},
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
			var tableId int32
			if UserCount <= 0 {
				continue
			}
			if (RobotCount + UserCount) < global.TablePlayCount {
				global.NoticeRobotOnline <- global.TablePlayCount - UserCount
				continue
			}
			for key, _ := range List {
				tableId = List[key].GetFreeTableID()
				if List[key].GetRobotCount() > 0 {
					continue
				}
				if tableId >= 0 {
					break
				}
			}
			if tableId < 0 {
				appendTableList()
				global.NoticeLoadMath <- 0
				continue
			}
			log.Logger.Debug("tableId:", tableId, "机器人人数:", RobotCount, "匹配队列真人人数:", UserCount)

			//算出差多少机器人
			haveRootCount := global.TablePlayCount - UserCount%global.TablePlayCount
			var userCont int32 = 0
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
			if haveRootCount > 0 && tableId >= 0 {
				var roobotCont int32 = 0
				// 机器人入场
				RobotQueue.Range(func(userID, userItem interface{}) bool {
					if List[tableId].RoomRobotFull() {
						return false
					}
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
	tableID             int32                          //桌子号
	gameStatus          int32                          //游戏状态
	sceneStartTime      int64                          //场景开始时间
	userList            sync.Map                       //玩家列表 座位号-->uid map[int32]int32
	userCount           int32                          //玩家数量
	robotCount          int32                          //机器人数量
	userPrepare         map[int32]bool                 //准备玩家
	userListLoss        map[int32]float32              //用户盈亏 座位号
	newUserListLoss     map[int32]float32              //用户盈亏 座位号->金额 只包含胡牌结算的金额
	huType              []int32                        // 胡牌类型: 1、平胡;2、四红中;3、七对;4、天胡;5、地胡;6、抢杠胡;7、杠开;8、癞子胡
	userTax             map[int32]float32              //用户税收
	systemScore         float32                        //系统盈亏
	outMjChairID        int32                          //出牌用户椅子号
	currentChairID      int32                          //当前用户椅子号
	outMj               int32                          //当前出牌
	sendMj              int32                          //当前摸牌
	userListAction      map[int32]int32                //当前操作
	userListOperate     map[int32]int32                //用户选择的操作
	winChairID          int32                          //胜利椅子号
	winCount            int32                          //连赢次数
	bankerChairID       int32                          //庄家的椅子
	sice                []int32                        //骰子
	mahjongHeap         []int32                        //麻将堆
	mjHeapHeadCount     int32                          //麻将堆头部位置
	mjHeapTailCount     int32                          //麻将堆尾部位置
	userListMahjong     map[int32]map[int32]int32      //玩家麻将数据
	mahjongNum          map[int32]int32                //麻将剩余数据 key 麻将->剩余数
	userListDiskMj      map[int32]*msg.DiskMahjongList //玩家桌面麻将(已操作)
	userListOutCard     map[int32][]int32              //玩家出牌记录
	maCount             int32                          //中码数量
	userListTrusteeship sync.Map                       //用户托管
	userListTing        map[int32]bool                 //用户听牌
	cronTimer           *time.Timer                    //定时任务
	settlementType      int32                          // 是否流局 1为流局 0为正常结算
	drawID              string                         // 游戏记录id
	mutex               sync.Mutex
	roundOrder          string // 局号
}

//定时操作 处理超时操作
func (it *Item) onEventGameTimer() {
	fmt.Println("定时器开始,桌子号:", it.tableID)
	for it.gameStatus == global.GameStatusPlay {
		select {
		case <-it.cronTimer.C:
			log.Logger.Debugf("定时器超时:%v,%v", len(it.userListOperate) == 0, it.currentChairID != it.outMjChairID)
			if len(it.userListAction) == 0 || (len(it.userListOperate) == 0 && it.currentChairID != it.outMjChairID) {
				item := it.getUserItem(it.currentChairID)
				if item == nil {
					return
				}

				var mjData int32
				if it.sendMj != 0 {
					mjData = it.sendMj
				} else {
					for k, v := range it.userListMahjong[it.currentChairID] {
						if v != 0 {
							mjData = k
							break
						}
					}
				}

				if _, ok := it.userListTrusteeship.Load(it.currentChairID); ok {
					if it.userListAction[it.currentChairID]&global.WIK_HU != 0 {
						it.OnUserOperate(item, &msg.Game_C_UserOperate{
							OperateCode: global.WIK_HU,
						})
						continue
					}
				}

				it.OnUserOutCard(item, &msg.Game_C_UserOutCard{
					MjData: mjData,
				})
				continue
			}

			//操作响应
			var chairID, operateCode int32
			for k, v := range it.userListOperate {
				if v == global.WIK_NULL || v == 0 {
					continue
				}

				if v > operateCode {
					operateCode = v
					chairID = k
				}
			}
			//过
			if operateCode == 0 {
				//过抢杠，允许补杠
				if it.userListAction[it.currentChairID]&global.CHR_QIANG_GANG_HU != 0 {
					//扣除补杠费用
					//for k, _ := range it.userListLoss {
					//	if k == chairID {
					//		it.userListLoss[k] += store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
					//	}
					//	it.userListLoss[k] -= store.GameControl.GetGameInfo().CellScore
					//}
					newUserListLoss := make(map[int32]float32)
					it.userList.Range(func(key, value interface{}) bool {
						if key.(int32) == chairID {
							it.userListLoss[key.(int32)] += store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
							newUserListLoss[key.(int32)] = store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						} else {
							it.userListLoss[key.(int32)] -= store.GameControl.GetGameInfo().CellScore
							newUserListLoss[key.(int32)] -= store.GameControl.GetGameInfo().CellScore
						}
						return true
					})
					//操作通知
					it.sendAllUser(&msg.Game_S_UserOperate{
						OperateCode:   operateCode,
						OperateMj:     it.sendMj,
						OperateUser:   chairID,
						ProvideUser:   chairID,
						UserListLoss:  make([]msg.UserLossData, 0),
						UserListMoney: make([]msg.UserListData, 0),
					})

					//摸牌
					item := it.getUserItem(it.currentChairID)
					if item == nil {
						return
					}
					it.onSendMahjong(item, false)
					if it.gameStatus != global.GameStatusPlay {
						break
					}

					if _, ok := it.userListTrusteeship.Load(item.ChairID); ok {
						//托管处理
						it.cronTimer.Reset(500 * time.Millisecond)
					} else if len(it.userListAction) == 0 {
						//重置出牌超时时间
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
					} else {
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
					}

				} else {
					//下家摸牌
					it.currentChairID = (it.currentChairID + 1) % store.GameControl.GetGameInfo().ChairCount
					item := it.getUserItem(it.currentChairID)
					if item == nil {
						return
					}
					it.onSendMahjong(item, true)
					if it.gameStatus != global.GameStatusPlay {
						break
					}
					if _, ok := it.userListTrusteeship.Load(item.ChairID); ok {
						//托管处理
						it.cronTimer.Reset(500 * time.Millisecond)
					} else if len(it.userListAction) == 0 {
						//重置出牌超时时间
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
					} else {
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
					}
				}
			}

			//胡
			if operateCode == global.WIK_HU {
				//胜利用户
				it.winChairID = chairID
				if it.winChairID == chairID {
					it.winCount++
				} else {
					it.winCount = 1
				}
				newUserListLoss := make(map[int32]float32, 0)
				var qiangGangeHu = make(map[int32]bool, 0)
				for k, v := range it.userListOperate {
					if v != global.WIK_HU {
						continue
					}

					var multiple int32
					if it.userListAction[k]&global.CHR_SI_HONG_ZHONG != 0 { //四红中
						multiple += 10
						it.huType = append(it.huType, global.CHR_SI_HONG_ZHONG)
					} else if it.userListAction[k]&global.CHR_QI_DUI != 0 { //七对
						multiple += 4
						it.huType = append(it.huType, global.CHR_QI_DUI)
					} else if it.userListAction[k]&global.CHR_PING_HU != 0 { //平胡
						multiple += 2
						it.huType = append(it.huType, global.CHR_PING_HU)
					}

					//无癞子胡
					if it.userListAction[k]&global.CHR_MAGIC == 0 {
						multiple += 2
					}

					if it.userListAction[k]&global.CHR_QIANG_GANG_HU != 0 { //抢杠胡
						it.userListLoss[it.currentChairID] -= float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						it.userListLoss[k] += float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						newUserListLoss[it.currentChairID] -= float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						newUserListLoss[k] += float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)

						var userLoss = make([]msg.UserLossData, 0)
						var userMoney = make([]msg.UserListData, 0)
						for k, v := range newUserListLoss {
							userLoss = append(userLoss, msg.UserLossData{ChairID: k, UserLoss: v})
						}

						for k, v := range it.GetAllUserListMoney() {
							userMoney = append(userMoney, msg.UserListData{ChairID: k, UserMoney: v})
						}
						qiangGangeHu[k] = true
						////操作通知
						it.sendAllUser(&msg.Game_S_UserOperate{
							OperateCode:   operateCode,
							OperateMj:     it.outMj,
							OperateUser:   chairID,
							ProvideUser:   it.outMjChairID,
							UserListLoss:  userLoss,
							UserListMoney: userMoney,
						})
						continue
					}

					if it.userListAction[k]&global.CHR_GANG_KAI != 0 { //杠开
						multiple += 2
					} else if it.userListAction[k]&global.CHR_TIAN_HU != 0 { //天胡
						multiple += 10
					} else if it.userListAction[k]&global.CHR_DI_HU != 0 { //地胡
						multiple += 10
					}
					log.Logger.Debugf("桌子号:%v,loss=%v,胡牌:%X,倍数:%v,Action:%X", it.tableID, it.userListLoss, it.huType, multiple, it.userListAction[k])

					it.userList.Range(func(key, value interface{}) bool {
						if key.(int32) == chairID {
							it.userListLoss[key.(int32)] += float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
							newUserListLoss[key.(int32)] = float32(multiple) * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						} else {
							it.userListLoss[key.(int32)] -= float32(multiple) * store.GameControl.GetGameInfo().CellScore
							newUserListLoss[key.(int32)] -= float32(multiple) * store.GameControl.GetGameInfo().CellScore
						}
						return true
					})

					var userLoss = make([]msg.UserLossData, 0)
					for k, v := range newUserListLoss {
						userLoss = append(userLoss, msg.UserLossData{ChairID: k, UserLoss: v})
					}
					userMoney := make([]msg.UserListData, 0)
					for k, v := range it.GetAllUserListMoney() {
						userMoney = append(userMoney, msg.UserListData{ChairID: k, UserMoney: v})
					}
					//操作通知
					it.sendAllUser(&msg.Game_S_UserOperate{
						OperateCode:   operateCode,
						OperateMj:     it.sendMj,
						OperateUser:   chairID,
						ProvideUser:   chairID,
						UserListLoss:  userLoss,
						UserListMoney: userMoney,
					})
				}

				//抽码
				for _, mjData := range it.mahjongHeap[:conf.GetServer().MaCount] {
					if mjData&0x0F == 0x01 || mjData&0x0F == 0x05 || mjData&0x0F == 0x09 {
						it.maCount++
					}
				}
				if it.maCount != 0 {
					if len(qiangGangeHu) > 0 {
						for k, _ := range qiangGangeHu {
							it.userListLoss[k] += 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * float32(store.GameControl.GetGameInfo().ChairCount-1) * 3
							newUserListLoss[k] += 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * float32(store.GameControl.GetGameInfo().ChairCount-1) * 3
							it.userListLoss[it.currentChairID] -= 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * 3
							newUserListLoss[it.currentChairID] -= 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * 3
						}
					} else {
						for chairId, _ := range it.userListLoss {
							if chairId == it.winChairID {
								it.userListLoss[it.winChairID] += 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * float32(store.GameControl.GetGameInfo().ChairCount-1)
								newUserListLoss[it.winChairID] += 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount) * float32(store.GameControl.GetGameInfo().ChairCount-1)
								continue
							}
							it.userListLoss[chairId] -= 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount)
							newUserListLoss[chairId] -= 2 * store.GameControl.GetGameInfo().CellScore * float32(it.maCount)
						}
					}
				}
				it.newUserListLoss = newUserListLoss
				it.onEventGameConclude()
				return
			}

			//碰
			if operateCode == global.WIK_PENG {
				//检测是否异常
				if it.userListMahjong[chairID][it.outMj] < 2 {
					_ = log.Logger.Errorf("onEventGameTimer err %s", "代码异常")
					return
				}
				//清除手牌
				it.userListMahjong[chairID][it.outMj] -= 2
				it.mahjongNum[it.outMj] -= 2
				if it.userListMahjong[chairID][it.outMj] == 0 {
					delete(it.userListMahjong[chairID], it.outMj)
				}

				//加入桌面记录
				it.userListDiskMj[chairID].Data = append(it.userListDiskMj[chairID].Data, msg.DiskMahjong{
					Data:    it.outMj,
					Code:    operateCode,
					ChairID: it.currentChairID,
				})

				//操作通知
				it.sendAllUser(&msg.Game_S_UserOperate{
					OperateCode:   operateCode,
					OperateMj:     it.outMj,
					OperateUser:   chairID,
					ProvideUser:   it.outMjChairID,
					UserListMoney: make([]msg.UserListData, 0),
					UserListLoss:  make([]msg.UserLossData, 0),
				})

				// 检查听牌的数据
				tingMjData := logic.Client.GetTingMjData(it.userListMahjong[chairID], it.mahjongNum)
				if len(tingMjData) > 0 {
					var userTing = msg.Game_S_UserTing{
						UserMajData: make([]msg.UserListMjData, 0),
					}
					for k, v := range tingMjData {
						var mjData = make([]msg.MjData, 0)
						for val, num := range v {
							mjData = append(mjData, msg.MjData{MjKey: val, MjValue: num})
						}
						userTing.UserMajData = append(userTing.UserMajData, msg.UserListMjData{OutCard: k, MjData: mjData})
					}
					it.getUserItem(chairID).WriteMsg(&userTing)
					// debug 测试
					if !it.getUserItem(chairID).IsRobot() {
						log.Logger.Debugf("碰牌 桌子号:%v,chairID:%v,听胡:%v", it.tableID, chairID, tingMjData)
					}
				}
				//清除数据
				it.userListAction = make(map[int32]int32, 0)
				it.sendMj = 0
				//设置当前用户
				it.currentChairID = chairID
				//重置出牌超时时间
				it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
			}

			//补杠
			if operateCode == global.WIK_BU_GANG {
				//检测是否异常
				if it.userListMahjong[chairID][it.sendMj] != 1 {
					_ = log.Logger.Errorf("onEventGameTimer err %s", "代码异常")
					return
				}

				//清除手牌
				it.userListMahjong[chairID][it.sendMj] -= 1
				it.mahjongNum[it.sendMj] -= 1
				if it.userListMahjong[chairID][it.sendMj] == 0 {
					delete(it.userListMahjong[chairID], it.sendMj)
				}
				for _, v := range it.userListDiskMj[chairID].Data {
					if v.Code != global.WIK_PENG || v.Data != it.sendMj {
						continue
					}

					//修改桌面记录
					v.Code = global.WIK_BU_GANG
				}

				//判断其他玩家是否有抢杠胡
				it.userListAction = logic.Client.GetBuGangResponse(chairID, it.sendMj, it.userListMahjong)
				if len(it.userListAction) != 0 {
					//操作通知
					it.sendAllUser(&msg.Game_S_UserOperate{
						OperateCode:   operateCode,
						OperateMj:     it.sendMj,
						OperateUser:   chairID,
						ProvideUser:   chairID,
						UserListMoney: make([]msg.UserListData, 0),
						UserListLoss:  make([]msg.UserLossData, 0),
					})
					var Trusteeship = false
					for k, v := range it.userListAction {

						it.userListAction[k] |= global.CHR_QIANG_GANG_HU
						ResponseCode := make([]int32, 0)
						if v == 0 {
							continue
						}
						ResponseCode = append(ResponseCode, global.WIK_HU)
						it.getUserItem(k).WriteMsg(&msg.Game_S_OperateNotify{
							Response: ResponseCode,
						})
						if _, ok := it.userListTrusteeship.Load(k); ok {
							Trusteeship = true
							it.userListOperate[k] = global.WIK_NULL
						}
					}
					if Trusteeship {
						//托管处理
						it.cronTimer.Reset(500 * time.Millisecond)
					} else {
						//重置操作超时时间
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
					}
				} else {
					//扣除补杠费用
					newUserListLoss := make(map[int32]float32)
					it.userList.Range(func(key, value interface{}) bool {
						if key.(int32) == chairID {
							it.userListLoss[key.(int32)] += store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
							newUserListLoss[key.(int32)] = store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						} else {
							it.userListLoss[key.(int32)] -= store.GameControl.GetGameInfo().CellScore
							newUserListLoss[key.(int32)] -= store.GameControl.GetGameInfo().CellScore
						}
						return true
					})
					var userLoss = make([]msg.UserLossData, 0)
					for k, v := range newUserListLoss {
						userLoss = append(userLoss, msg.UserLossData{ChairID: k, UserLoss: v})
					}
					userMoney := make([]msg.UserListData, 0)
					for k, v := range it.GetAllUserListMoney() {
						userMoney = append(userMoney, msg.UserListData{ChairID: k, UserMoney: v})
					}
					//操作通知
					it.sendAllUser(&msg.Game_S_UserOperate{
						OperateCode:   operateCode,
						OperateMj:     it.sendMj,
						OperateUser:   chairID,
						ProvideUser:   chairID,
						UserListLoss:  userLoss,
						UserListMoney: userMoney,
					})

					//摸牌
					item := it.getUserItem(it.currentChairID)
					if item == nil {
						return
					}
					it.onSendMahjong(item, false)
					if it.gameStatus != global.GameStatusPlay {
						break
					}

					if _, ok := it.userListTrusteeship.Load(item.ChairID); ok {
						//托管处理
						it.cronTimer.Reset(500 * time.Millisecond)
					} else if len(it.userListAction) == 0 {
						//重置出牌超时时间
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
					} else {
						it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
					}
				}
			}

			//明杠
			if operateCode == global.WIK_MING_GANG {
				//检测是否异常
				if it.userListMahjong[chairID][it.outMj] != 3 {
					_ = log.Logger.Errorf("onEventGameTimer err %s", "代码异常")
					return
				}
				//清除手牌
				//for k, card := range it.userListOutCard[it.outMjChairID] {
				//	if card == it.outMj {
				//		it.userListOutCard[it.outMjChairID] = append(it.userListOutCard[it.outMjChairID][:k], it.userListOutCard[it.outMjChairID][k+1:]...)
				//		break
				//	}
				//}
				it.userListMahjong[chairID][it.outMj] -= 3
				it.mahjongNum[it.outMj] -= 3
				if it.userListMahjong[chairID][it.outMj] == 0 {
					delete(it.userListMahjong[chairID], it.outMj)
				}
				//加入桌面记录
				it.userListDiskMj[chairID].Data = append(it.userListDiskMj[chairID].Data, msg.DiskMahjong{
					Data:    it.outMj,
					Code:    operateCode,
					ChairID: it.currentChairID,
				})

				//扣除补杠费用
				it.userListLoss[chairID] += store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
				it.userListLoss[it.currentChairID] -= store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)

				var newUserListLoss = map[int32]float32{chairID: store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1),
					it.currentChairID: -store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)}

				var userLoss = make([]msg.UserLossData, 0)
				for k, v := range newUserListLoss {
					userLoss = append(userLoss, msg.UserLossData{ChairID: k, UserLoss: v})
				}
				userMoney := make([]msg.UserListData, 0)
				for k, v := range it.GetAllUserListMoney() {
					userMoney = append(userMoney, msg.UserListData{ChairID: k, UserMoney: v})
				}

				//操作通知
				it.sendAllUser(&msg.Game_S_UserOperate{
					OperateCode:   operateCode,
					OperateMj:     it.outMj,
					OperateUser:   chairID,
					ProvideUser:   it.outMjChairID,
					UserListLoss:  userLoss,
					UserListMoney: userMoney,
				})

				//设置当前用户
				it.currentChairID = chairID

				//摸牌
				item := it.getUserItem(it.currentChairID)
				if item == nil {
					return
				}
				it.onSendMahjong(item, false)
				if it.gameStatus != global.GameStatusPlay {
					break
				}

				if _, ok := it.userListTrusteeship.Load(item.ChairID); ok {
					//托管处理
					it.cronTimer.Reset(500 * time.Millisecond)
				} else if len(it.userListAction) == 0 {
					//重置出牌超时时间
					it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
				} else {
					it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
				}
			}

			//暗杠
			if operateCode == global.WIK_AN_GANG {
				//检测是起手杠
				if it.userListMahjong[chairID][it.sendMj] != 4 && len(it.userListOutCard[chairID]) == 0 {
					var card int32 = 0
					for k, v := range it.userListMahjong[chairID] {
						if v == 4 {
							card = k
						}
					}
					if card != 0 {
						//清除手牌
						it.userListMahjong[chairID][card] -= 4
						it.mahjongNum[card] -= 4
						if it.userListMahjong[chairID][card] == 0 {
							delete(it.userListMahjong[chairID], card)
						}
						//加入桌面记录
						it.userListDiskMj[chairID].Data = append(it.userListDiskMj[chairID].Data, msg.DiskMahjong{
							Data:    card,
							Code:    operateCode,
							ChairID: it.currentChairID,
						})
						if it.sendMj != card {
							it.sendMj = card
						}
					}
				} else if it.userListMahjong[chairID][it.sendMj] == 4 {
					//清除手牌
					it.userListMahjong[chairID][it.sendMj] -= 4
					it.mahjongNum[it.sendMj] -= 4
					if it.userListMahjong[chairID][it.sendMj] == 0 {
						delete(it.userListMahjong[chairID], it.sendMj)
					}
					//加入桌面记录
					it.userListDiskMj[chairID].Data = append(it.userListDiskMj[chairID].Data, msg.DiskMahjong{
						Data:    it.sendMj,
						Code:    operateCode,
						ChairID: it.currentChairID,
					})
				} else {
					break
				}

				//扣除暗杠费用
				newUserListLoss := make(map[int32]float32)
				it.userList.Range(func(key, value interface{}) bool {
					if key.(int32) == chairID {
						it.userListLoss[key.(int32)] += 2 * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
						newUserListLoss[key.(int32)] = 2 * store.GameControl.GetGameInfo().CellScore * float32(store.GameControl.GetGameInfo().ChairCount-1)
					} else {
						it.userListLoss[key.(int32)] -= 2 * store.GameControl.GetGameInfo().CellScore
						newUserListLoss[key.(int32)] -= 2 * store.GameControl.GetGameInfo().CellScore
					}
					return true
				})
				var userLoss = make([]msg.UserLossData, 0)
				for k, v := range newUserListLoss {
					userLoss = append(userLoss, msg.UserLossData{ChairID: k, UserLoss: v})
				}
				userMoney := make([]msg.UserListData, 0)
				for k, v := range it.GetAllUserListMoney() {
					userMoney = append(userMoney, msg.UserListData{ChairID: k, UserMoney: v})
				}
				//操作通知
				it.sendAllUser(&msg.Game_S_UserOperate{
					OperateCode:   operateCode,
					OperateMj:     it.sendMj,
					OperateUser:   chairID,
					ProvideUser:   chairID,
					UserListLoss:  userLoss,
					UserListMoney: userMoney,
				})

				//摸牌
				item := it.getUserItem(it.currentChairID)
				if item == nil {
					return
				}
				it.onSendMahjong(item, false)
				if it.gameStatus != global.GameStatusPlay {
					break
				}

				if _, ok := it.userListTrusteeship.Load(item.ChairID); ok {
					//托管处理
					it.cronTimer.Reset(500 * time.Millisecond)
				} else if len(it.userListAction) == 0 {
					//重置出牌超时时间
					it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
				} else {
					//重置出牌超时时间
					it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
				}

			}
		}
	}
	fmt.Println("定时器结束 桌子号:", it.tableID)
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
				return false
			}
			freeScene.UserList = append(freeScene.UserList, UserInfoToGrameUser(userItemTemp.(*user.Item)))
			return true
		})
		//freeScene.PrepareUserList = it.userPrepare
		freeScene.SceneStartTime = it.sceneStartTime

		userItem.WriteMsg(&freeScene)
	case global.GameStatusPlay:
		var playScene msg.Game_S_PlayScene
		playScene.UserList = make([]msg.Game_S_User, 0)
		it.userList.Range(func(chairID, userID interface{}) bool {
			userItemTemp, ok := user.List.Load(userID.(int32))
			if !ok {
				_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
				return true
			}
			playScene.UserList = append(playScene.UserList, UserInfoToGrameUser(userItemTemp.(*user.Item)))
			return true
		})
		var sceneStartTime int64
		var ResponseCode = make([]int32, 0)
		if len(it.userListAction) != 0 {
			if it.userListAction[userItem.ChairID]&global.WIK_HU != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_SI_HONG_ZHONG != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_QI_DUI != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_PING_HU != 0 {
				ResponseCode = append(ResponseCode, global.WIK_HU)
			}
			if it.userListAction[userItem.ChairID]&global.WIK_MING_GANG != 0 {
				ResponseCode = append(ResponseCode, global.WIK_MING_GANG)
			}
			if it.userListAction[userItem.ChairID]&global.WIK_PENG != 0 {
				ResponseCode = append(ResponseCode, global.WIK_PENG)
			}
			sceneStartTime = int64(conf.GetServer().GameOperateTime) - (time.Now().Unix() - it.sceneStartTime)
		} else {
			sceneStartTime = int64(conf.GetServer().GameOutCardTime) - (time.Now().Unix() - it.sceneStartTime)
		}
		if sceneStartTime <= 0 {
			sceneStartTime = 0
		}

		var mjData = make([]msg.MjData, 0)
		for k, v := range it.userListMahjong[userItem.ChairID] {
			mjData = append(mjData, msg.MjData{MjKey: k, MjValue: v})
		}
		var userTuoGuan = make([]msg.UserTuoGuan, 0)
		var userTiPai = make([]msg.UserTingPai, 0)
		it.userListTrusteeship.Range(func(key, value interface{}) bool {
			userTuoGuan = append(userTuoGuan, msg.UserTuoGuan{
				ChairID: key.(int32),
				IsTing:  value.(bool),
			})
			return true
		})
		for k, v := range it.userListTing {
			userTiPai = append(userTiPai, msg.UserTingPai{ChairID: k, IsTing: v})
		}
		playScene.SceneStartTime = int32(sceneStartTime)
		playScene.UserMahjong = mjData
		playScene.UserListTrusteeship = userTuoGuan
		playScene.UserListTing = userTiPai
		playScene.CurrentChairID = it.currentChairID
		playScene.OutMjChairID = it.outMjChairID
		playScene.OutMj = it.outMj
		playScene.SendMj = it.sendMj
		playScene.UserAction = ResponseCode
		playScene.BankerUser = it.bankerChairID
		playScene.UserListOutCardRecord = make([]msg.Int32Array, 0)
		for k, v := range it.userListOutCard {
			playScene.UserListOutCardRecord = append(playScene.UserListOutCardRecord, msg.Int32Array{
				ChairID: k,
				Data:    v,
			})
		}
		for _, datas := range it.userListDiskMj {
			for i := range datas.Data {
				if datas.Data[i].Code == global.WIK_PENG || datas.Data[i].Code == global.WIK_BU_GANG || datas.Data[i].Code == global.WIK_MING_GANG {
					for charID, values := range playScene.UserListOutCardRecord {
						var isFind = false
						for index := range values.Data {
							if values.Data[index] == datas.Data[i].Data {
								playScene.UserListOutCardRecord[charID].Data = append(playScene.UserListOutCardRecord[charID].Data[:index], playScene.UserListOutCardRecord[charID].Data[index+1:]...)
								isFind = true
								break
							}
						}
						if isFind {
							break
						}
					}
				}
			}
		}
		log.Logger.Debugf("场景消息,桌子id=%v,剩余牌数=%v,已出牌:%v", it.tableID, len(it.mahjongHeap), playScene.UserListOutCardRecord)

		userListDiskMj := make([]msg.DiskMahjongList, 0)
		for k, v := range it.userListDiskMj {
			tempData := make([]msg.DiskMahjong, 0)
			if len(v.Data) > 0 {
				tempData = v.Data
			}
			userListDiskMj = append(userListDiskMj, msg.DiskMahjongList{
				ChairID: k,
				Data:    tempData,
			})
		}
		playScene.UserListDiskMahjong = userListDiskMj
		playScene.DiskMahjongNum = int32(len(it.mahjongHeap))
		userItem.WriteMsg(&playScene)
		// 检查听牌的数据
		if it.currentChairID == userItem.ChairID {
			tingMjData := logic.Client.GetTingMjData(it.userListMahjong[userItem.ChairID], it.mahjongNum)
			if len(tingMjData) > 0 {
				var userTing = msg.Game_S_UserTing{
					UserMajData: make([]msg.UserListMjData, 0),
				}
				for k, v := range tingMjData {
					var mjData = make([]msg.MjData, 0)
					for val, num := range v {
						mjData = append(mjData, msg.MjData{MjKey: val, MjValue: num})
					}
					userTing.UserMajData = append(userTing.UserMajData, msg.UserListMjData{OutCard: k, MjData: mjData})
				}
				userItem.WriteMsg(&userTing) //发送数据给当前用户
			}
			if !userItem.IsRobot() {
				log.Logger.Debugf("重连 桌子号:%v,userID:%v,听胡:%v", it.tableID, userItem.UserID, tingMjData)
			}
		}
	}
}

//开始游戏
func (it *Item) onEventGameStart() {
	//设置游戏状态
	it.gameStatus = global.GameStatusPlay
	it.sceneStartTime = time.Now().Unix()
	randNum := rand.Krand(6, 3)

	it.roundOrder = fmt.Sprintf("%v%v%s", conf.GetServer().GameID, time.Now().Unix(), randNum)
	//定庄
	if it.winChairID != -1 {
		it.bankerChairID = it.winChairID
	} else {
		it.bankerChairID = rand.RandInterval(0, store.GameControl.GetGameInfo().ChairCount-1)
	}
	//初始化骰子
	it.sice = make([]int32, 0)
	for i := 0; i < 2; i++ {
		it.sice = append(it.sice, rand.RandInterval(0, 6))
	}

	//洗牌
	it.mahjongHeap = make([]int32, 0)
	var mjData = make(map[int32][]msg.MjData, 0)

	heap := logic.Client.DispatchTableCard()
	for _, card := range heap {
		it.mahjongHeap = append(it.mahjongHeap, card)
		it.mahjongNum[card] = 4
	}
	it.mjHeapTailCount = 0
	//初始化用户手牌
	it.userListMahjong = make(map[int32]map[int32]int32, 0)
	it.userList.Range(func(chairID, uid interface{}) bool {
		//初始化桌面牌
		it.userListDiskMj[chairID.(int32)] = new(msg.DiskMahjongList)
		//发牌
		it.userListMahjong[chairID.(int32)] = make(map[int32]int32)
		for _, v := range it.mahjongHeap[0:13] {
			it.userListMahjong[chairID.(int32)][v]++
		}
		it.mahjongHeap = it.mahjongHeap[13:]
		it.mjHeapHeadCount = global.MahjongCount - 13

		value, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("user.List err %d", uid)
			return true
		}
		for k, v := range it.userListMahjong[chairID.(int32)] {
			mjData[chairID.(int32)] = append(mjData[chairID.(int32)], msg.MjData{MjKey: k, MjValue: v})
		}
		value.(*user.Item).WriteMsg(&msg.Game_S_GameStart{
			BankerUser: it.bankerChairID,
			SiceData:   it.sice,
			UserMjData: mjData[chairID.(int32)],
		})
		return true
	})
	it.cronTimer = time.NewTimer(15 * time.Second)
	//log.Logger.Debugf("======开始游戏===== 桌子号:%v,用户牌:%v,剩余麻将堆:%v", it.tableID, it.userListMahjong, it.mahjongHeap)
	t := time.NewTimer(time.Second * 5)
	select {
	case <-t.C:
		time.Sleep(time.Second * 5) // 初始化预留时间，客户端跑动画
		it.currentChairID = it.bankerChairID
		//庄家摸牌
		bankerUserID, ok := it.userList.Load(it.bankerChairID)
		if !ok {
			_ = log.Logger.Errorf("onEventGameStart err %d", bankerUserID)
			return
		}
		bankerItem, ok := user.List.Load(bankerUserID)
		if !ok {
			_ = log.Logger.Errorf("onEventGameStart err %d", bankerUserID)
			return
		}
		it.onSendMahjong(bankerItem, true)

		if _, ok := it.userListTrusteeship.Load(bankerItem.(*user.Item).ChairID); ok {
			//托管处理
			it.cronTimer.Reset(500 * time.Millisecond)
		} else if len(it.userListAction) == 0 {
			//重置出牌超时时间
			it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
		} else {
			//重置出牌超时时间
			it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
		}
		//启动定时器
		it.onEventGameTimer()
		break
	}
}

//结束游戏 判断是否钱够不够 直接踢出去
func (it *Item) onEventGameConclude() {
	it.cronTimer.Stop()
	//设置游戏状态
	it.gameStatus = global.GameStatusFree
	it.sceneStartTime = time.Now().Unix()
	//计算系统损耗
	for k, v := range it.userListLoss {
		item := it.getUserItem(k)
		if item == nil {
			return
		}

		//计算税收
		if it.userListLoss[k] > 0 {
			it.userTax[k] = it.userListLoss[k] * store.GameControl.GetGameInfo().RevenueRatio
			it.userListLoss[k] -= it.userTax[k]
		}

		if item.IsRobot() {
			continue
		}
		it.systemScore += v
	}
	//更新库存
	store.GameControl.ChangeStore(it.systemScore)
	//记录游戏记录
	it.onWriteGameRecord()
	//用户写分
	it.onWriteGameScore()
	// 热更数据
	it.onUpdateAgentData()

	log.Logger.Debugf("桌子号:%v,中的码:%v,loss=%v,倍数:%v", it.tableID, it.maCount, it.userListLoss, it.huType)
	newListLoss := make(map[int32]float32)
	oldUserList := make(map[int32]int32)
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
	var flag = false
	//通知用户
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
			oldUserList[chairID.(int32)] = uid.(int32)
			if !mySelf.IsRobot() {
				mySelf.Status = user.StatusFree
			}
			//结束通知
			var gameConclude msg.Game_S_GameConclude
			gameConclude.WinChairID = it.winChairID
			gameConclude.ProvideMj = it.sendMj
			gameConclude.UserListMjData = make([]msg.Int32MapInt32, 0)
			for k, v := range it.userListMahjong {
				var mjData = make([]msg.MjData, 0)

				for val, num := range v {
					mjData = append(mjData, msg.MjData{MjKey: val, MjValue: num})
				}
				gameConclude.UserListMjData = append(gameConclude.UserListMjData, msg.Int32MapInt32{
					ChairID: k,
					Data:    mjData,
				})
			}
			if it.winChairID >= 0 && !flag {
				log.Logger.Debugf("结束游戏,桌子号:%v,Win data:%v，输赢:%v", it.tableID,
					gameConclude.UserListMjData[it.winChairID].Data, it.newUserListLoss)
				flag = true
			}

			var newUserLostLoss = make([]msg.UserLossData, 0)
			for k, v := range it.newUserListLoss {
				newUserLostLoss = append(newUserLostLoss, msg.UserLossData{
					ChairID:  k,
					UserLoss: v,
				})
			}
			var newUserMoney = make([]msg.UserListData, 0)
			for K, V := range newListLoss {
				newUserMoney = append(newUserMoney, msg.UserListData{
					ChairID:   K,
					UserMoney: V,
				})
			}
			gameConclude.MaData = it.mahjongHeap[:conf.GetServer().MaCount]
			gameConclude.UserListLoss = newUserLostLoss
			gameConclude.UserListMoney = newUserMoney
			gameConclude.SettlementType = it.settlementType
			gameConclude.HuType = it.huType
			value.(*user.Item).WriteMsg(&gameConclude)

			//判断下注积分是否足够
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				if value.(*user.Item).UserGold < store.GameControl.GetGameInfo().MinEnterScore {
					_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 金币不足!")
					value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
						ErrorCode: global.SitDownError3,
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
						ErrorCode: global.SitDownError3,
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
		it.OnActionUserStandUp(value.(*user.Item), true)
		return true
	})

	//清除数据
	it.userTax = make(map[int32]float32)
	it.systemScore = 0
	it.userListLoss = make(map[int32]float32)
	it.userListDiskMj = make(map[int32]*msg.DiskMahjongList)
	it.userListOutCard = make(map[int32][]int32)
	it.maCount = 0
	it.userCount = 0
	it.robotCount = 0
	it.userListTrusteeship = sync.Map{}
	it.userListTing = make(map[int32]bool)
	it.drawID = ""
	it.userListMahjong = make(map[int32]map[int32]int32, 0)
	it.outMjChairID = -1
	it.currentChairID = -1
	it.mahjongHeap = make([]int32, 0)
	it.mahjongNum = make(map[int32]int32, 0)
	//初始化准备
	it.userPrepare = make(map[int32]bool)
	it.huType = make([]int32, 0)
	it.settlementType = 0
	it.newUserListLoss = make(map[int32]float32)
	//go func() {
	//	for len(oldUserList) > 0 {
	//		t := time.NewTicker(time.Second * 60)
	//		select {
	//		case <-t.C:
	//			it.OnMoveUserByChairID(oldUserList)
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
		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			userItem.(*user.Item).UserGold += v
		} else {
			if !userItem.(*user.Item).IsRobot() {
				if !redis.GameClient.IsExistsDiamond(userItem.(*user.Item).UserID) {
					scoreInfo, _ := mysql.GetGameScoreInfoByUserId(mysql.GameClient.GetXJGameDB, userItem.(*user.Item).UserID)
					redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, scoreInfo.Diamond)
					userItem.(*user.Item).UserDiamond = scoreInfo.Diamond
				}

				//userDiamond, err := redis.GameClient.GetDiamond(userItem.(*user.Item).UserID)
				//if err != nil {
				//	log.Logger.Error("GetDiamond err:", err)
				//} else {
				//	userItem.(*user.Item).UserDiamond = float32(userDiamond)
				//}

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
			0,
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
		redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userItem.(*user.Item).UserDiamond)

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

	var startTime = time.Unix(it.sceneStartTime, 0).Format("2006-01-02 15:04:05")
	var taxSum float32 = 0
	for _, v := range it.userTax {
		taxSum += v
	}
	errorCode, errorMsg, drawID := mysql.GameClient.WriteGameRecord(
		it.tableID,
		it.userCount,
		it.robotCount,
		it.systemScore,
		taxSum,
		startTime,
		endTime,
		store.GameControl.GetGameInfo().DeductionsType,
		"{}",
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
	if userItem.TableID >= 0 && userItem.ChairID >= 0 {
		List[userItem.TableID].OnMoveUserByChairID(map[int32]int32{userItem.ChairID: userItem.UserID})
	}
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
	//	log.Logger.Debugf(" OnActionUserSitDown do 坐下中 桌子id:%v,桌子人数:%v,uid:%v,前桌子id:%v,前椅子id:%v", it.tableID, it.userCount, userItem.UserID, userItem.TableID, userItem.ChairID)
	// TODO 匹配桌子
	it.mutex.Lock() // 协程互斥锁
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
			it.robotCount++
		}
		it.userPrepare[userItem.ChairID] = false
		break
	}
	it.mutex.Unlock()
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
	// 重新排序桌子
	log.Logger.Debugf("成功坐下,发送场景消息,桌子号:%v 人数:%v,uid:%v,椅子id:%v", it.tableID, it.userCount, userItem.UserID, userItem.ChairID)

	go func() {
		t := time.NewTimer(time.Second * 5)
		for {
			select {
			case <-t.C:
				if it.gameStatus == global.GameStatusFree && it.userCount != global.TablePlayCount {
					log.Logger.Debug("用户重新加入队列", userItem.UserID)
					it.OnActionUserStandUp(userItem, true)
					downLoad()
					//清除数据
					it.userTax = make(map[int32]float32)
					it.systemScore = 0
					it.userListLoss = make(map[int32]float32)
					it.userListDiskMj = make(map[int32]*msg.DiskMahjongList)
					it.userListOutCard = make(map[int32][]int32)
					it.maCount = 0
					it.userCount = 0
					it.robotCount = 0
					it.userListTrusteeship = sync.Map{}
					it.userListTing = make(map[int32]bool)
					it.drawID = ""
					it.userListMahjong = make(map[int32]map[int32]int32, 0)
					it.outMjChairID = -1
					it.currentChairID = -1
					it.mahjongHeap = make([]int32, 0)
					it.mahjongNum = make(map[int32]int32, 0)
					//初始化准备
					it.userPrepare = make(map[int32]bool)
					it.huType = make([]int32, 0)
					it.settlementType = 0
					it.newUserListLoss = make(map[int32]float32)
					if !userItem.IsRobot() {
						global.NoticeLoadMath <- 0
					}
					return
				}
			}
			if it.userCount != global.TablePlayCount {
				time.Sleep(time.Millisecond * 100)
			}
			it.onEventSendGameScene(userItem)
			it.OnUserPrepare(userItem, true)
			break
		}
	}()
}

// 移除座位
func (it *Item) OnMoveUserByChairID(OldUserList map[int32]int32) {
	for chairID, oldUserId := range OldUserList {
		UserID, ok := it.userList.Load(chairID)
		if ok {
			userItem, IsExist := user.List.Load(UserID)
			if oldUserId == UserID.(int32) && IsExist && userItem.(*user.Item).Status == user.StatusFree {
				it.userList.Delete(chairID)
				if it.gameStatus == global.GameStatusFree && it.userCount > 0 {
					it.userCount--
					it.sendAllUser(&msg.Game_S_StandUpNotify{
						ChairID: chairID,
					})
					// 重新排序桌子

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

//用户起立
func (it *Item) OnActionUserStandUp(args ...interface{}) {
	userItem := args[0].(*user.Item)
	flag := args[1].(bool)
	log.Logger.Debugf(" 起立 OnActionUserStandUp- uid:%v", userItem.UserID)
	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if !ok || userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnActionUserStandUp err %s", "起立失败, 用户不在用户列表里")
		userItem.Close()
		return
	}

	if !flag {
		//检测是否游戏中
		if it.gameStatus == global.GameStatusPlay {
			_ = log.Logger.Errorf("OnActionUserStandUp %s", "游戏中不允许退出")
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
	//取消准备
	delete(it.userPrepare, oldChairID)

	//it.userCount--
	//if userItem.IsRobot() {
	//	it.robotCount--
	//}
	userItem.StandUp()
	//解锁
	if err := mysql.GameClient.UnLock(userItem.UserID); err != nil {
		_ = log.Logger.Errorf("解锁用户失败 err %v", err)
		return
	}

	// 重新排序桌子

	//通知其他玩家
	//it.sendAllUser(&msg.Game_S_StandUpNotify{
	//	ChairID: oldChairID,
	//})
}

//用户重入
func (it *Item) OnActionUserReconnect(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem.Status == user.StatusOffline {
		//设置用户状态
		userItem.Status = user.StatusPlaying
		//解除托管
		it.userListTrusteeship.Delete(userItem.ChairID)

		//通知其他玩家
		it.sendOtherUser(userItem.ChairID, &msg.Game_S_OnLineNotify{
			ChairID: userItem.ChairID,
		})
	}

	//发送场景消息
	it.onEventSendGameScene(userItem)
}

//用户断线
func (it *Item) OnActionUserOffLine(args ...interface{}) {
	userItem := args[0].(*user.Item)

	//设置用户状态
	userItem.Status = user.StatusOffline
	//设置用户托管

	it.userListTrusteeship.Store(userItem.ChairID, true)

	log.Logger.Debugf("用户断线,桌子号:%v,uid:%v 椅子号:%v ,托管状态:%v", it.tableID, userItem.UserID, userItem.ChairID, it.userListTrusteeship)
	//通知其他玩家
	it.sendOtherUser(userItem.ChairID, &msg.Game_S_OffLineNotify{
		ChairID: userItem.ChairID,
	})
}

//用户出牌
func (it *Item) OnUserOutCard(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserOutCard)

	//判断游戏状态
	if it.gameStatus != global.GameStatusPlay {
		_ = log.Logger.Error("OnUserOutCard游戏状态异常:", it.gameStatus)
		userItem.Close()
		return
	}

	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserOutCard err %s", "出牌失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OutCardError2,
			ErrorMsg:  "出牌失败, 用户不在用户列表里",
		})
		userItem.Close()
		return
	}

	//判断是否当前操作玩家
	if userItem.ChairID != it.currentChairID || userItem.ChairID == it.outMjChairID {
		_ = log.Logger.Errorf("OnUserOutCard err %s uid=%v", "出牌失败, 非当前用户出牌", userItem.UserID)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.OutCardError2,
		//	ErrorMsg:  "出牌失败, 非当前用户出牌!",
		//})
		// userItem.Close()
		return
	}

	//判断数据是否异常
	if logic.Client.SwitchToIndex(m.MjData) < 0 && logic.Client.SwitchToIndex(m.MjData) >= 31 {
		_ = log.Logger.Errorf("OnUserOutCard err %s uid:%v ,出牌:%X", "出牌失败, 麻将数据异常", userItem.UserID, m.MjData)
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OutCardError2,
			ErrorMsg:  "出牌失败, 麻将数据异常!",
		})
		return
	}

	//判断是否是手牌
	if it.userListMahjong[userItem.ChairID][m.MjData] == 0 {
		_ = log.Logger.Errorf("OnUserOutCard err %s uid:%v 出牌%X ,手牌:%v", "出牌失败, 非手牌数据!", userItem.UserID, m.MjData, it.userListMahjong[userItem.ChairID])
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OutCardError2,
			ErrorMsg:  "出牌失败, 非手牌数据!",
		})
		return
	}
	// debug
	log.Logger.Debugf("OnUserOutCard 桌子号:%v 椅子号:%v,uid:%v,成功出牌:%d=%x", it.tableID, userItem.ChairID, userItem.UserID, m.MjData, m.MjData)
	//记录用户出牌
	it.userListOutCard[userItem.ChairID] = append(it.userListOutCard[userItem.ChairID], m.MjData)
	//移除麻将
	it.userListMahjong[userItem.ChairID][m.MjData]--
	it.mahjongNum[m.MjData]--
	if it.userListMahjong[userItem.ChairID][m.MjData] == 0 {
		delete(it.userListMahjong[userItem.ChairID], m.MjData)
	}

	//更新出牌用户与麻将
	it.outMjChairID = userItem.ChairID
	it.outMj = m.MjData
	//检测其他玩家是否有响应
	it.userListAction = logic.Client.OutMjResponse(userItem.ChairID, m.MjData, it.userListMahjong)
	//重置用户操作状态
	it.userListOperate = make(map[int32]int32, 0)

	//通知用户
	it.userList.Range(func(chairID, uid interface{}) bool {
		item, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("send user.List err %d", uid)
			return true
		}

		item.(*user.Item).WriteMsg(&msg.Game_S_UserOutCard{
			ChairID: userItem.ChairID,
			MjData:  m.MjData,
		})

		return true
	})

	//触发其他用户操作
	if len(it.userListAction) != 0 {
		var Trusteeship = false
		for k, v := range it.userListAction {
			ResponseCode := make([]int32, 0)
			if v == 0 {
				continue
			}
			if v&global.WIK_MING_GANG != 0 {
				ResponseCode = append(ResponseCode, global.WIK_MING_GANG)
			}
			if v&global.WIK_PENG != 0 {
				ResponseCode = append(ResponseCode, global.WIK_PENG)
			}
			it.getUserItem(k).WriteMsg(&msg.Game_S_OperateNotify{
				Response: ResponseCode,
			})

			if _, ok = it.userListTrusteeship.Load(k); ok {
				Trusteeship = true
				it.userListOperate[k] = global.WIK_NULL
			}

		}
		if Trusteeship {
			//托管处理
			it.cronTimer.Reset(500 * time.Millisecond)
		} else {
			it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
		}
		return
	}
	//下家摸牌
	it.currentChairID = (it.currentChairID + 1) % store.GameControl.GetGameInfo().ChairCount
	uid, ok := it.userList.Load(it.currentChairID)
	if !ok {
		_ = log.Logger.Errorf("onEventGameStart err %d", uid)
		return
	}
	item, ok := user.List.Load(uid)
	if !ok {
		_ = log.Logger.Errorf("onEventGameStart err %d", uid)
		return
	}
	it.onSendMahjong(item, true)
	if it.gameStatus != global.GameStatusPlay {
		return
	}
	if _, ok = it.userListTrusteeship.Load(it.currentChairID); ok {
		//托管处理
		it.cronTimer.Reset(500 * time.Millisecond)
	} else if len(it.userListAction) == 0 {
		//重置出牌超时时间
		it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOutCardTime))
	} else {
		//重置出牌超时时间
		it.cronTimer.Reset(time.Second * time.Duration(conf.GetServer().GameOperateTime))
	}
}

//用户操作
func (it *Item) OnUserOperate(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserOperate)
	// debug
	log.Logger.Debugf("OnUserOperate 桌子号:%v 椅子号:%v,uid:%v,操作code:%X", it.tableID, userItem.ChairID, userItem.UserID, m.OperateCode)

	//判断游戏状态
	if it.gameStatus != global.GameStatusPlay {
		_ = log.Logger.Error("OnUserOperate游戏状态异常:", it.gameStatus)
		userItem.Close()
		return
	}

	//检验用户是否在用户列表里
	value, ok := it.userList.Load(userItem.ChairID)
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserOperate err %s", "操作失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OperateError2,
			ErrorMsg:  "操作失败, 用户不在用户列表里",
		})
		userItem.Close()
		return
	}

	//校验数据
	if m.OperateCode < global.WIK_NULL || (m.OperateCode != global.WIK_NULL && it.userListAction[userItem.ChairID]&m.OperateCode == 0) || it.userListAction[userItem.ChairID] == 0 {
		_ = log.Logger.Errorf("OnUserOperate err %s uid:%v,%X", "操作失败, 数据异常", userItem.UserID, m.OperateCode)
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OperateError2,
			ErrorMsg:  "操作失败, 数据异常",
		})
		//	userItem.Close()
		return
	}
	if it.userListOperate[userItem.ChairID] > 0 {
		_ = log.Logger.Errorf("OnUserOperate err %s uid:%v,%X 已经操作:%X", "操作失败, 数据异常", userItem.UserID, m.OperateCode, it.userListOperate[userItem.ChairID])
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OperateError2,
			ErrorMsg:  "操作失败, 不能重复操作",
		})
		//	userItem.Close()
		return
	}

	//记录状态
	it.userListOperate[userItem.ChairID] = m.OperateCode
	//判断是否需要等待其他人操作
	var temp int32
	for k, v := range it.userListAction {
		if it.userListOperate[k] == global.WIK_NULL {
			continue
		}

		if v&0x1FF > temp {
			temp = v & 0x1FF
		}
	}

	wait := false
	for k, v := range it.userListAction {
		if v&0x1FF != temp {
			continue
		}

		if it.userListOperate[k] == 0 {
			wait = true
		}
	}
	if m.OperateCode == global.WIK_NULL && it.currentChairID == userItem.ChairID && (it.userListAction[userItem.ChairID]&global.WIK_HU != 0 ||
		it.userListAction[userItem.ChairID]&global.CHR_SI_HONG_ZHONG != 0 ||
		it.userListAction[userItem.ChairID]&global.CHR_QI_DUI != 0 ||
		it.userListAction[userItem.ChairID]&global.CHR_PING_HU != 0) {
		delete(it.userListOperate, userItem.ChairID)
	}
	if !wait {
		it.cronTimer.Reset(500 * time.Millisecond)
	}
}

//用户摸牌
func (it *Item) onSendMahjong(args ...interface{}) {
	userItem := args[0].(*user.Item)
	flag := args[1].(bool)

	//判断流局
	if int32(len(it.mahjongHeap)) <= conf.GetServer().MaCount {
		it.settlementType = 1
		it.onEventGameConclude()
		return
	}
	it.sceneStartTime = time.Now().Unix()
	//摸牌
	it.userListAction = make(map[int32]int32, 0)
	it.sendMj = it.mahjongHeap[0]
	it.userListMahjong[userItem.ChairID][it.sendMj]++
	it.mahjongHeap = it.mahjongHeap[1:]

	//log.Logger.Debugf("onSendMahjong 用户摸牌 桌子号:%v,椅子号:%v uid:%v  摸牌:%v=%x,麻将堆%v\n", it.tableID, userItem.ChairID, userItem.UserID, it.sendMj, it.sendMj, it.mahjongHeap)

	if flag {
		it.mjHeapHeadCount--
	} else {
		it.mjHeapTailCount--
	}

	//判断是否有摸牌响应
	action := logic.Client.SendMjResponse(it.sendMj, it.userListMahjong[userItem.ChairID], it.userListDiskMj[userItem.ChairID], len(it.userListOutCard[userItem.ChairID]) == 0)
	if action != 0 {
		it.userListAction[userItem.ChairID] = action
		if it.userListOperate[userItem.ChairID] == global.WIK_BU_GANG ||
			it.userListOperate[userItem.ChairID] == global.WIK_MING_GANG ||
			it.userListOperate[userItem.ChairID] == global.WIK_AN_GANG {
			it.userListAction[userItem.ChairID] |= global.CHR_GANG_KAI
		}
		if len(it.userListOutCard[userItem.ChairID]) == 0 &&
			(it.bankerChairID == userItem.ChairID &&
				it.userListAction[userItem.ChairID]&global.WIK_HU != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_SI_HONG_ZHONG != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_QI_DUI != 0 ||
				it.userListAction[userItem.ChairID]&global.CHR_PING_HU != 0) {
			if it.bankerChairID == userItem.ChairID {
				it.userListAction[userItem.ChairID] |= global.CHR_TIAN_HU
			} else {
				it.userListAction[userItem.ChairID] |= global.CHR_DI_HU
			}
		}
	}

	//重置用户操作状态
	it.userListOperate = make(map[int32]int32, 0)

	//发送数据给当前用户
	userItem.WriteMsg(&msg.Game_S_SendMj{
		MjData:         it.sendMj,
		CurrentChairID: userItem.ChairID,
		Tail:           flag,
	})

	//发送给其他人
	it.sendOtherUser(userItem.ChairID, &msg.Game_S_SendMj{
		CurrentChairID: userItem.ChairID,
		Tail:           flag,
	})

	// 检查听牌的数据
	tingMjData := logic.Client.GetTingMjData(it.userListMahjong[userItem.ChairID], it.mahjongNum)
	if len(tingMjData) > 0 {
		var userTing = msg.Game_S_UserTing{
			UserMajData: make([]msg.UserListMjData, 0),
		}
		for k, v := range tingMjData {
			var mjData = make([]msg.MjData, 0)
			for val, num := range v {
				mjData = append(mjData, msg.MjData{
					MjKey:   val,
					MjValue: num,
				})
			}
			userTing.UserMajData = append(userTing.UserMajData, msg.UserListMjData{OutCard: k, MjData: mjData})
		}
		userItem.WriteMsg(&userTing) //发送数据给当前用户
	}
	// debug 测试
	if !userItem.IsRobot() {
		log.Logger.Debugf("桌子号:%v,uid:%v,用户摸牌:%v,手上牌:%v,麻将剩余数量:%v,听胡:%v", it.tableID, userItem.UserID, it.sendMj, it.userListMahjong[userItem.ChairID], it.mahjongNum, tingMjData)
	}
	//发送操作提示
	if it.userListAction[userItem.ChairID] != 0 {
		var ResponseCode = make([]int32, 0)
		if it.userListAction[userItem.ChairID]&global.WIK_HU != 0 ||
			it.userListAction[userItem.ChairID]&global.CHR_SI_HONG_ZHONG != 0 ||
			it.userListAction[userItem.ChairID]&global.CHR_QI_DUI != 0 ||
			it.userListAction[userItem.ChairID]&global.CHR_PING_HU != 0 {
			ResponseCode = append(ResponseCode, global.WIK_HU)
		}
		if it.userListAction[userItem.ChairID]&global.WIK_AN_GANG != 0 {
			ResponseCode = append(ResponseCode, global.WIK_AN_GANG)
		}
		if it.userListAction[userItem.ChairID]&global.WIK_BU_GANG != 0 {
			ResponseCode = append(ResponseCode, global.WIK_BU_GANG)
		}
		log.Logger.Debugf("Game_S_OperateNotify  send 2 uid:%v,code:%v", userItem.UserID, ResponseCode)
		userItem.WriteMsg(&msg.Game_S_OperateNotify{
			Response: ResponseCode,
		})
	}
	//log.Logger.Debugf("onSendMahjong 用户摸牌 桌子号:%v,椅子号:%v uid:%v  摸牌:%v=%x,手上牌%v,碰牌:%v\n", it.tableID, userItem.ChairID, userItem.UserID, it.sendMj, it.sendMj, it.userListMahjong[userItem.ChairID], it.userListDiskMj[userItem.ChairID])
}

//用户准备
func (it *Item) OnUserPrepare(args ...interface{}) {
	userItem := args[0].(*user.Item)

	if it.gameStatus != global.GameStatusFree {
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError1,
		//	ErrorMsg:  "当前不能准备!",
		//})
		return
	}

	userPrepareStatus, ok := it.userPrepare[userItem.ChairID]

	if ok && userPrepareStatus {
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.ZBError2,
		//	ErrorMsg:  "不能重复准备!",
		//})
		return
	}

	it.userPrepare[userItem.ChairID] = true

	//it.sendAllUser(&msg.Game_S_UserPrepare{
	//	ChairID: userItem.ChairID,
	//})

	//检测是否都准备
	var userPrepareCount int
	for _, v := range it.userPrepare {
		if v {
			userPrepareCount++
		}
	}
	if userPrepareCount == global.TablePlayCount {
		go it.onEventGameStart()
	}
}

//用户取消准备
func (it *Item) OnUserUnPrepare(args ...interface{}) {
	userItem := args[0].(*user.Item)

	if it.gameStatus != global.GameStatusFree {
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError1,
			ErrorMsg:  "当前不能取消准备!",
		})
		return
	}

	userPrepareStatus := it.userPrepare[userItem.ChairID]
	if !userPrepareStatus {
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError2,
			ErrorMsg:  "暂未准备!",
		})
		return
	}

	it.userPrepare[userItem.ChairID] = false

	it.sendAllUser(&msg.Game_S_UserUnPrepare{
		ChairID: userItem.ChairID,
	})
}

//用户听牌
func (it *Item) OnUserTing(args ...interface{}) {
	userItem := args[0].(*user.Item)

	//检测状态是否异常
	if it.gameStatus != global.GameStatusPlay {
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "听牌失败, 状态异常",
		})
		return
	}

	//检测是否听牌
	//if logic.Client.CheckTing(it.userListMahjong[userItem.ChairID]) != true {
	//	log.Logger.Errorf("OnUserTing 听牌错误! uid:%v", userItem.UserID)
	//	userItem.WriteMsg(&msg.Game_S_ReqlyFail{
	//		ErrorCode: global.TingError2,
	//		ErrorMsg:  "听牌失败, 数据异常",
	//	})
	//	return
	//}

	//通知
	//it.sendAllUser(&msg.Game_S_UserTing{
	//	ChairID: userItem.ChairID,
	//})
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

// 获取空闲的座位 优先匹配有真人的房间
func getFreeTableID() (int32, int32) {
	var tableID int32 = -1
	var IsBreak = false
	var userCount int32
	for i := int32(3); i >= 0; i-- {
		tempList, ok := TablesList.Load(i)
		if !ok {
			continue
		}
		for k, _ := range tempList.(map[int32]*Item) {
			tableID = k
			IsBreak = true
			userCount = List[k].userCount - List[k].robotCount
			if i > 0 && userCount > 0 {
				break
			} else if i == 0 {
				break
			}
		}
		if IsBreak {
			break
		}
	}
	return tableID, userCount
}

//获取房间没有机器人的桌子,和没有上局挂机的人
func (it *Item) GetRobotCount() int32 {
	if int(it.userCount) != len(it.userPrepare) {
		return 1
	}
	return -1
}

//获取房间最多能装多少机器人 3个
func (it *Item) RoomRobotFull() bool {
	return it.robotCount == global.TablePlayCount-1
}

// 获取用户临时金额
func (it *Item) GetAllUserListMoney() map[int32]float32 {
	var userListMoney = make(map[int32]float32)
	it.userList.Range(func(chairID, value interface{}) bool {
		userItem, ok := user.List.Load(value.(int32))
		if ok {
			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				userListMoney[chairID.(int32)] = userItem.(*user.Item).UserGold + it.userListLoss[chairID.(int32)]
			} else {
				userListMoney[chairID.(int32)] = userItem.(*user.Item).UserDiamond + it.userListLoss[chairID.(int32)]
			}

			redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userListMoney[chairID.(int32)])
		}
		return true
	})

	return userListMoney
}

//判断是否加入(机器人)
func (it *Item) CheckRobotSitDown() bool {
	//判断是否是空闲状态或人满
	if it.gameStatus != global.GameStatusFree || it.userCount == global.TablePlayCount {
		return false
	}

	//判断是否有真人
	var flag bool
	it.userList.Range(func(chairID, userID interface{}) bool {
		item, ok := user.List.Load(userID.(int32))
		if !ok {
			_ = log.Logger.Errorf("onEventSendGameScene global.GameStatusFree err user.List.Load")
			return true
		}

		if !item.(*user.Item).IsRobot() {
			flag = true
			return false
		}

		return true
	})

	return flag
}

//获取用户item
func (it *Item) getUserItem(chairID int32) *user.Item {
	uid, ok := it.userList.Load(chairID)
	if !ok {
		_ = log.Logger.Errorf("onEventGameStart err %d 桌子号:%v,椅子号:%v", uid, it.tableID, chairID)
		return nil
	}
	item, ok := user.List.Load(uid)
	if !ok {
		_ = log.Logger.Errorf("onEventGameStart err %d 桌子号:%v,椅子号:%v", uid, it.tableID, chairID)
		return nil
	}

	return item.(*user.Item)
}

//获取桌子状态
func (it *Item) GetGameStatus() int32 {
	return it.gameStatus
}

// 用户托管
func (it *Item) OnUserAutoManage(args ...interface{}) {
	userItem := args[0].(*user.Item)

	it.userListTrusteeship.Store(userItem.ChairID, true)

	if it.gameStatus != global.GameStatusFree {
		var sceneStartTime int64
		sceneStartTime = int64(conf.GetServer().GameOutCardTime) - (time.Now().Unix() - it.sceneStartTime)
		if sceneStartTime > 0 && it.currentChairID == userItem.ChairID {
			it.cronTimer.Reset(time.Second * 1)
		}
		//	log.Logger.Debugf(" handlerAutoManage end  桌子号:%v,uid:%v,bool=%v,场景时间:%v,托管:%v", it.tableID, userItem.UserID, it.currentChairID == userItem.ChairID, sceneStartTime, it.userListTrusteeship)
	}
	it.sendAllUser(&msg.Game_S_AutoManage{ChairID: userItem.ChairID})
}

// 用户取消托管
func (it *Item) OnUserUnAutoManage(args ...interface{}) {
	userItem := args[0].(*user.Item)
	it.userListTrusteeship.Delete(userItem.ChairID)
	if it.gameStatus != global.GameStatusFree {
		var sceneStartTime int64

		if len(it.userListAction) == 0 {
			sceneStartTime = int64(conf.GetServer().GameOutCardTime) - (time.Now().Unix() - it.sceneStartTime)
		} else {
			sceneStartTime = int64(conf.GetServer().GameOperateTime) - (time.Now().Unix() - it.sceneStartTime)
		}
		if sceneStartTime > 0 && it.currentChairID == userItem.ChairID {
			delete(it.userListOperate, userItem.ChairID)
			it.cronTimer.Reset(time.Second * time.Duration(sceneStartTime))
		}
	}
	it.sendAllUser(&msg.Game_S_UnAutoManage{ChairID: userItem.ChairID})
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
