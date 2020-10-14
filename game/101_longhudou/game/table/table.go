package table

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
	"xj_game_server/game/101_longhudou/conf"
	"xj_game_server/game/101_longhudou/game/logic"
	"xj_game_server/game/101_longhudou/global"
	"xj_game_server/game/101_longhudou/model"
	"xj_game_server/game/101_longhudou/msg"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/segment"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/public"
	"xj_game_server/util/leaf/log"

	rand "xj_game_server/util/leaf/util"

	"time"
)

/*
龙虎斗
*/
var List = make([]*Item, 0)

//初始化桌子
func OnInit() {
	//初始化桌子
	for i := int32(0); i < store.GameControl.GetGameInfo().TableCount; i++ {
		temp := &Item{
			tableID:            i,
			chair:              &model.Chair{},
			gameStatus:         0,
			sceneStartTime:     0,
			userList:           sync.Map{},
			userCount:          0,
			androidCount:       0,
			lotteryRecord:      make([]int32, 0),
			userListJetton:     sync.Map{},
			userListJettons:    []sync.Map{},
			userListAreaJetton: sync.Map{},
			userListWinRecord:  sync.Map{},
			lotteryPoker:       make([]int32, 0),
			winArea:            make([]bool, 0),
			userListLoss:       make(map[int32]float32, 0),
			userTax:            make(map[int32]float32, 0),
			systemScore:        0,
			specialUserList:    make([]*msg.Game_S_User, 0),
			areaJettions:       make([]*msg.Game_C_AreaJetton, 0),
			cronTimer:          &time.Timer{},
			updateTimer:        &time.Timer{},
		}
		temp.chair.OnInit(store.GameControl.GetGameInfo().ChairCount)
		List = append(List, temp)
		go List[i].sendUserAreaLottery()
		time.Sleep(1 * time.Second)
		go List[i].onEventGameStart()
		go List[i].sendRoomRedis()
	}
}

type Item struct {
	tableID            int32                    //桌子号
	chair              *model.Chair             //椅子
	gameStatus         int32                    //游戏状态
	sceneStartTime     int64                    //场景开始时间
	userList           sync.Map                 //玩家列表 座位号-->uid map[int32]int32
	userCount          int32                    //玩家数量
	androidCount       int32                    //机器人数量
	lotteryRecord      []int32                  //开奖记录 global.LotteryCount
	userListJetton     sync.Map                 //玩家总下注 1局 座位号 map[int32]float32
	userListJettons    []sync.Map               //玩家总下注 20局 [20]userListJetton
	userListAreaJetton sync.Map                 //玩家每个区域下注信息 座位号 map[int32][global.AreaCount]float32
	userListWinRecord  sync.Map                 //玩家输赢记录 --->座位号map[int32][global.WinRecordCount]bool
	lotteryPoker       []int32                  //开奖扑克 global.PokerCount
	winArea            []bool                   //输赢区域 global.AreaCount
	userListLoss       map[int32]float32        //用户盈亏 座位号
	userTax            map[int32]float32        //用户税收
	systemScore        float32                  //系统盈亏
	specialUserList    []*msg.Game_S_User       //富豪榜/神算子
	areaJettions       []*msg.Game_C_AreaJetton //每个区域的下注数
	cronTimer          *time.Timer              //定时器下注
	updateTimer        *time.Timer              //定时器redis更新数据
	drawID             string                   // 游戏记录id
	roundOrder         string                   `json:"round_order"` // 局号
}

// 定时发送发送每个区域每个人的总下注
func (it *Item) sendUserAreaLottery() {
	for {
		it.cronTimer = time.NewTimer(time.Second * 1)
		select {
		case <-it.cronTimer.C:
			if it.gameStatus == global.GameStatusJetton {
				var areaJettions = make([]*msg.Game_C_AreaJetton, 0)
				for i := 0; i < global.AreaCount; i++ {
					var tempArearJettion msg.Game_C_AreaJetton
					var tempJetton float32
					it.userListAreaJetton.Range(func(chairID, value interface{}) bool {
						tempJetton += value.([global.AreaCount]float32)[i]
						return true
					})
					tempArearJettion = msg.Game_C_AreaJetton{
						Area:   int32(i),
						Jetton: tempJetton,
					}
					areaJettions = append(areaJettions, &tempArearJettion)
				}
				it.areaJettions = areaJettions
				it.sendAllUser(&msg.Game_S_AreaJetton{
					TableID:        it.tableID,
					UserArraJetton: areaJettions,
					UserCount:      it.userCount,
				})
			}
		}
	}
}

func (it *Item) sendRoomRedis() {
	for {
		// 发送每个区域每个人的总下注
		it.updateTimer = time.NewTimer(time.Second * 1)
		select {
		case <-it.updateTimer.C:
			it.onUpdateRedisMsg()

		}
	}
}

func (it *Item) GetTableID() int32 {
	return it.tableID
}

func (it *Item) GetLotteryRecord() []int32 {
	return it.lotteryRecord
}

func (it *Item) GetGameStatus() int32 {
	return it.gameStatus
}

//是否下注
func (it *Item) IsUserBet(user *user.Item) bool {
	load, ok := it.userListJetton.Load(user.ChairID)
	if ok && it.gameStatus == global.GameStatusJetton {
		if load.(float32) > 0 {
			return true
		}
	}
	return false
}

// 剩余时间
func (it *Item) GetSceneStartTime() int32 {
	var sceneTime int32
	switch it.gameStatus {
	case global.GameStatusJetton: // 下注场景
		sceneTime = conf.GetServer().GameJettonTime - int32(time.Now().Unix()-it.sceneStartTime)
		if sceneTime < 0 {
			sceneTime = 0
		}
	case global.GameStatusLottery: //开奖场景
		sceneTime = conf.GetServer().GameLotteryTime - int32(time.Now().Unix()-it.sceneStartTime)
		if sceneTime < 0 {
			sceneTime = 0
		}
	}
	return sceneTime
}
func (it *Item) GetUserCount() int32 {
	return it.userCount
}

// 游戏大厅场景发送 登录或者游戏结束发送
func (it *Item) SendLotteryRecord(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem != nil {
		userItem.WriteMsg(&msg.Game_S_Hall{
			TableID:       it.tableID,
			LotteryRecord: it.lotteryRecord[len(it.lotteryRecord)-1:][0],
			UserCount:     it.userCount,
		})
	}
}

func (it *Item) onUpdateRedisMsg() {
	segment := &segment.GameSegmentReq{}
	segment.TableId = int(it.tableID)
	segment.UserCount = int(it.userCount)
	segment.JettonTime = int(conf.GetServer().GameJettonTime)
	segment.LotteryTime = int(conf.GetServer().GameLotteryTime)
	segment.RoomStatus = int(it.gameStatus)

	segment.JettonList = conf.GetServer().JettonList

	segment.Astrict = store.GameControl.GetGameInfo().MinEnterScore

	segment.ResidueTime = int(it.GetSceneStartTime())

	for _, record := range it.lotteryRecord {
		segment.LotteryRecord = append(segment.LotteryRecord, record)
	}

	lKey := fmt.Sprintf("%s%d:%d:", public.RedisKeyTableServerList, conf.GetServer().KindID, conf.GetServer().GameID)

	key := fmt.Sprintf("%s%d:%d:%d:", public.RedisKeyTableServer, conf.GetServer().KindID, conf.GetServer().GameID, it.tableID)
	value, _ := json.Marshal(segment)

	redis.GameClient.Client.LRem(lKey, 0, key)
	err := redis.GameClient.Client.LPush(lKey, key).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterGame err %v", err)
	}

	redis.GameClient.Client.Set(key, string(value), time.Duration(conf.GetServer().GameJettonTime+conf.GetServer().GameLotteryTime)*time.Second)
}

//开始游戏
func (it *Item) onEventGameStart() {
	//清除上局的数据
	it.userListAreaJetton = sync.Map{}
	it.userListJetton = sync.Map{}
	//设置游戏状态
	it.gameStatus = global.GameStatusJetton
	// 发送每个区域每个人的总下注
	it.cronTimer.Reset(1 * time.Second)
	it.sceneStartTime = time.Now().Unix()
	// 计算神算子或富豪榜
	it.specialUser()
	//发送开始游戏给所有用户
	randNum := rand.Krand(6, 3)
	//data, _ := json.Marshal(it.specialUserList)
	it.roundOrder = fmt.Sprintf("%v%v%s", conf.GetServer().GameID, time.Now().Unix(), randNum)
	it.sendAllUser(&msg.Game_S_GameStart{
		Data: it.specialUserList,
	})

	// 记录总开局数 1天
	key := fmt.Sprintf("kind-%d:game-%d:table-%d:", conf.GetServer().KindID, conf.GetServer().GameID, it.tableID)
	err := redis.GameClient.Client.Incr(key).Err()
	//设置 超时时间 每天凌晨到期
	day := rand.EndOfDay(time.Now()).Sub(time.Now())
	err = redis.GameClient.Client.Expire(key, day).Err()
	if err != nil {
		_ = log.Logger.Errorf("write total game err %v", err)
	}
	fmt.Printf("[龙虎斗]桌子%d开始游戏,游戏状态%d,游戏人数:%d,神算子富豪榜%d\n", it.tableID, it.gameStatus, it.userCount, len(it.specialUserList))

	//下注定时器
	t := time.NewTimer(time.Second * time.Duration(conf.GetServer().GameJettonTime))
	select {
	case <-t.C:
		it.onEventGameConclude()
	}
}

//神算子和富豪榜
func (it *Item) specialUser() {
	var richUser model.Special         //富豪
	var operator model.SpecialOperator //神算子
	it.specialUserList = make([]*msg.Game_S_User, 0)
	it.userListWinRecord.Range(func(key, value interface{}) bool {
		var temp *model.Operator
		uid, ok := it.userList.Load(key.(int32))
		if !ok {
			//_ = log.Logger.Errorf("specialUser it.userList err %d",key.(int32))
			return true
		}
		userItem, ok := user.List.Load(uid.(int32))
		if !ok {
			_ = log.Logger.Errorf("specialUser user.List err %d", uid.(int32))
			return true
		}
		for _, v := range value.([global.WinRecordCount]bool) {
			var i = 0
			if v {
				i++
				temp = &model.Operator{
					User: &msg.Game_S_User{
						UserID:       userItem.(*user.Item).Info.UserID,
						NikeName:     userItem.(*user.Item).Info.NikeName,
						UserGold:     userItem.(*user.Item).Info.UserGold,
						UserDiamond:  userItem.(*user.Item).Info.UserDiamond,
						MemberOrder:  userItem.(*user.Item).Info.MemberOrder,
						HeadImageUrl: userItem.(*user.Item).Info.HeadImageUrl,
						FaceID:       userItem.(*user.Item).Info.FaceID,
						RoleID:       userItem.(*user.Item).Info.RoleID,
						SuitID:       userItem.(*user.Item).Info.SuitID,
						PhotoFrameID: userItem.(*user.Item).Info.PhotoFrameID,
						TableID:      userItem.(*user.Item).Info.TableID,
						ChairID:      userItem.(*user.Item).Info.ChairID,
						Status:       userItem.(*user.Item).Info.Status,
						Gender:       userItem.(*user.Item).Info.Gender,
					},
					Count: i,
				}
			}
		}
		if temp != nil {
			operator = append(operator, temp)
		}
		//富豪榜
		richUser = append(richUser, &msg.Game_S_User{
			UserID:       userItem.(*user.Item).Info.UserID,
			NikeName:     userItem.(*user.Item).Info.NikeName,
			UserGold:     userItem.(*user.Item).Info.UserGold,
			UserDiamond:  userItem.(*user.Item).Info.UserDiamond,
			MemberOrder:  userItem.(*user.Item).Info.MemberOrder,
			HeadImageUrl: userItem.(*user.Item).Info.HeadImageUrl,
			FaceID:       userItem.(*user.Item).Info.FaceID,
			RoleID:       userItem.(*user.Item).Info.RoleID,
			SuitID:       userItem.(*user.Item).Info.SuitID,
			PhotoFrameID: userItem.(*user.Item).Info.PhotoFrameID,
			TableID:      userItem.(*user.Item).Info.TableID,
			ChairID:      userItem.(*user.Item).Info.ChairID,
			Status:       userItem.(*user.Item).Info.Status,
			Gender:       userItem.(*user.Item).Info.Gender,
		})
		return true
	})
	// 前2名
	sort.Sort(richUser)
	// 前2名
	sort.Sort(operator)
	if len(richUser) >= global.ListSize {
		it.specialUserList = append(it.specialUserList, richUser[0:2]...)
	}
	if len(operator) > global.ListSize {
		it.specialUserList = append(it.specialUserList, operator[0].User)
		it.specialUserList = append(it.specialUserList, operator[1].User)
	}

}

//结束游戏
func (it *Item) onEventGameConclude() {
	it.cronTimer.Reset(time.Second*time.Duration(conf.GetServer().GameLotteryTime) + 5)
	it.gameStatus = global.GameStatusLottery
	it.sceneStartTime = time.Now().Unix()
	//randNumber := rand.RandInterval(0, 101)
	//var min float32
	//var count int32
	//fmt.Printf("[龙虎斗]==测试死循环==桌子%d结束游戏,随机值%d 数据库值:%f\n", it.tableID, randNumber, store.GameControl.GetUserWinRate())
	////开奖
	//for {
	//	var userWinSum float32
	//	// 发牌
	//	it.lotteryPoker = logic.Client.DispatchTableCard()
	//	// 获取赢的区域
	//	it.winArea = logic.Client.GetWinArea(it.lotteryPoker, it.userListAreaJetton)
	//	// 系统盈亏和用户盈亏
	//	it.systemScore, it.userListLoss, it.userTax, userWinSum = logic.Client.GetSystemLoss(it.winArea, it.userListAreaJetton, it.userList)
	//
	//	if float32(randNumber) < store.GameControl.GetUserWinRate() { //用户赢
	//		// 系统库存不够的时候用户输
	//		if store.GameControl.GetStore() < 0 {
	//			break
	//		}
	//		if min > userWinSum {
	//			min = userWinSum
	//			count++
	//		}
	//		if count == 5 {
	//			break
	//		}
	//		if userWinSum >= 0 {
	//			break
	//		}
	//	} else { //用户输
	//		if userWinSum <= 0 {
	//			break
	//		}
	//	}
	//}

	var winAreas = make([]store.WinAreas, 0)

	for i := 0; i < 5; i++ {
		var w = store.WinAreas{}
		// 发牌
		w.LotteryPoker = logic.Client.DispatchTableCard()
		// 获取赢的区域
		w.WinArea = logic.Client.GetWinArea(w.LotteryPoker, it.userListAreaJetton)
		// 系统盈亏和用户盈亏
		w.SystemScore, w.UserListLoss, w.UserTax, _ = logic.Client.GetSystemLoss(w.WinArea, it.userListAreaJetton, it.userList)

		w.Stores = store.GameControl.GetStore1() + w.SystemScore

		winAreas = append(winAreas, w)
	}

	sort.Sort(store.IntSlice(winAreas))

	it.lotteryPoker = winAreas[0].LotteryPoker
	it.winArea = winAreas[0].WinArea
	it.systemScore = winAreas[0].SystemScore
	it.userListLoss = winAreas[0].UserListLoss
	it.userTax = winAreas[0].UserTax

	// 更新系统库存
	store.GameControl.ChangeStore(it.systemScore)

	//更新开奖记录
	it.onUpdateWins()
	//更新用户输赢记录
	it.onUpdateUsersWins()

	//记录游戏记录
	it.onWriteGameRecord()
	//用户写分
	it.onWriteGameScore()
	// 热更数据
	it.onUpdateAgentData()

	var gameConclude msg.Game_S_GameConclude
	gameConclude.LotteryPoker = it.lotteryPoker
	gameConclude.WinArea = it.winArea

	// 发自己和特殊的四人的记录
	var newListLoss = make(map[int32]float32)
	var new2ListLoss = make(map[int32]float32)
	//神算子/富豪
	for _, v := range it.specialUserList {

		uid, ok := it.userList.Load(v.ChairID)
		if ok {
			value, ok := user.List.Load(uid)
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					newListLoss[v.ChairID] = value.(*user.Item).UserGold
				} else {
					newListLoss[v.ChairID] = value.(*user.Item).UserDiamond
				}

			}
			//新盈亏 加上下注数
			load, o := it.userListJetton.Load(v.ChairID)
			if o {
				new2ListLoss[v.ChairID] = it.userListLoss[v.ChairID] + load.(float32)
			}

		}

	}

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
			_, ok := newListLoss[mySelf.ChairID]
			if !ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					newListLoss[mySelf.ChairID] = value.(*user.Item).UserGold
				} else {
					newListLoss[mySelf.ChairID] = value.(*user.Item).UserDiamond
				}
			}
			//新盈亏 加上下注数
			load, o := it.userListJetton.Load(mySelf.ChairID)
			if o {
				new2ListLoss[mySelf.ChairID] = it.userListLoss[mySelf.ChairID] + load.(float32)
			}
			gameConclude.UserLoss = new2ListLoss[mySelf.ChairID]
			// 重新赋值 减去税收
			gameConclude.UserListMoney = newListLoss
			gameConclude.UserListLoss = new2ListLoss
			gameConclude.ThisMoney = value.(*user.Item).UserDiamond
			value.(*user.Item).WriteMsg(&gameConclude)
			if !ok {
				delete(newListLoss, mySelf.ChairID)
			}
		}
		return true
	})

	//user.List.Range(func(key, value interface{}) bool {
	//	// 空闲状态
	//	//if value.(*user.Item).Status == user.StatusFree {
	//	//	//发送空闲状态数据
	//	//	it.SendLotteryRecord(value)
	//	//	return true
	//	//}
	//	if value.(*user.Item).Agent == nil {
	//		return true
	//	}
	//	if value.(*user.Item).IsRobot() {
	//		return true
	//	}
	//	value.(*user.Item).WriteMsg(&msg.Game_S_Hall{
	//		TableID:       it.tableID,
	//		LotteryRecord: it.lotteryRecord[len(it.lotteryRecord)-1:][0],
	//		UserCount:     it.userCount,
	//	})
	//	//it.SendLotteryRecord(value)
	//	return true
	//})
	//看是否记录了20局
	if len(it.userListJettons) == global.WinRecordCount {
		it.userListJettons = it.userListJettons[1:]
		it.userListJettons = append(it.userListJettons, it.userListJetton)
	} else {
		it.userListJettons = append(it.userListJettons, it.userListJetton)
	}
	fmt.Printf("[龙虎斗]桌子%d结束游戏,系统损耗%f\n,开奖号码% X\n,开奖记录%v\n,中奖区域%v\n", it.tableID, it.systemScore, it.lotteryPoker, it.lotteryRecord, it.winArea)
	//清空上局数据
	it.userListLoss = make(map[int32]float32)
	it.systemScore = 0
	it.drawID = ""
	//it.lotteryPoker = make([]int32, 0)

	//开奖定时器
	t := time.NewTimer(time.Second * time.Duration(conf.GetServer().GameLotteryTime))
	select {
	case <-t.C:
		it.onEventGameStart()
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

//更新开奖记录
func (it *Item) onUpdateWins() {
	//self.winArea
	//
	//for i := 1; i < len(it.lotteryRecord); i++ {
	//	it.lotteryRecord[i], it.lotteryRecord[i-1] = it.lotteryRecord[i-1], it.lotteryRecord[i]
	//}
	if len(it.lotteryRecord) >= global.LotteryCount {
		it.lotteryRecord = it.lotteryRecord[1:]
	}

	var wins int32

	if it.winArea[0] {
		wins = 0
	} else if it.winArea[3] {
		wins = 3
	} else if it.winArea[8] {
		wins = 8
	}

	it.lotteryRecord = append(it.lotteryRecord, wins)
}

//更新用户输赢记录
func (it *Item) onUpdateUsersWins() {
	it.userList.Range(func(key, value interface{}) bool {
		// 取出20句的所有用户输赢记录
		v, _ := it.userListWinRecord.LoadOrStore(key.(int32), [global.WinRecordCount]bool{})
		// 迁移以为
		temp := v.([global.WinRecordCount]bool)
		for i := 1; i < len(temp); i++ {
			temp[i], temp[i-1] = temp[i-1], temp[i]
		}
		// 收过税的人肯定是赢了
		temp[len(temp)-1] = it.userListLoss[key.(int32)] > 0
		it.userListWinRecord.Store(key.(int32), temp)

		return true
	})
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
					var tempScore float32
					v, ok := it.userListJetton.Load(userItem.(*user.Item).ChairID)
					if ok {
						tempScore = v.(float32)
					}
					userItem.(*user.Item).UserDiamond = float32(userDiamond) + tempScore
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
			it.userTax[key], //  税收写分
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

		if !userItem.(*user.Item).IsRobot() && store.GameControl.GetGameInfo().DeductionsType == 1 {
			redis.GameClient.SetDiamond(userItem.(*user.Item).UserID, userItem.(*user.Item).UserDiamond)
			redis.GameClient.RegisterRecharge(userItem.(*user.Item).UserID)
		}

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
		it.androidCount,
		it.systemScore,
		taxSum,
		startTime,
		endTime,
		store.GameControl.GetGameInfo().DeductionsType,
		"",
	)
	if errorCode != 0 {
		_ = log.Logger.Errorf(" mysql GSP_RecordDrawInfo存储过程报错 %n %s ", errorCode, errorMsg)
		return
	}
	it.drawID = drawID
}

//发送场景
func (it *Item) onEventSendGameScene(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem.UserID == 138684 || userItem.UserID == 138677 {
		fmt.Println(it.tableID, "   onEventSendGameScene11111111", "onEventSendGameScene22222", time.Now())
	}
	switch it.gameStatus {
	case global.GameStatusJetton: //下注场景消息
		var jettonScene msg.Game_S_JettonScene

		sceneTime := conf.GetServer().GameJettonTime - int32(time.Now().Unix()-it.sceneStartTime)
		if sceneTime < 0 {
			sceneTime = 0
		}
		jettonScene.SceneStartTime = sceneTime
		jettonScene.UserChairID = userItem.ChairID
		jettonScene.UserList = it.specialUserList
		jettonScene.RecordID = it.roundOrder

		if !it.isOnTable(userItem) {
			//自己
			jettonScene.UserList = append(jettonScene.UserList, &msg.Game_S_User{
				UserID:       userItem.GetUserInfo().UserID,
				NikeName:     userItem.GetUserInfo().NikeName,
				UserGold:     userItem.GetUserInfo().UserGold,
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
			})
		}
		jettonScene.LotteryRecord = it.lotteryRecord
		//jettonScene.LotteryRecord = it.lotteryRecord
		jettonScene.UserArraJetton = it.areaJettions
		//玩家每个区域的下注数
		var userJettonList = make([]*msg.Game_C_AreaJetton, 0)
		load, ok := it.userListAreaJetton.Load(userItem.ChairID)
		if ok {
			for k, v := range load.([global.AreaCount]float32) {
				userJetton := &msg.Game_C_AreaJetton{
					Area:   int32(k),
					Jetton: v,
				}
				userJettonList = append(userJettonList, userJetton)
			}
		} else {
			for i := 0; i < global.AreaCount; i++ {
				userJetton := &msg.Game_C_AreaJetton{
					Area:   int32(i),
					Jetton: 0,
				}
				userJettonList = append(userJettonList, userJetton)
			}
		}
		jettonScene.UserJetton = userJettonList
		userItem.WriteMsg(&jettonScene)
	case global.GameStatusLottery: //开奖场景消息
		var lotteryScene msg.Game_S_LotteryScene

		sceneTime := conf.GetServer().GameLotteryTime - int32(time.Now().Unix()-it.sceneStartTime)

		if sceneTime < 0 {
			sceneTime = 0
		}

		lotteryScene.SceneStartTime = sceneTime
		lotteryScene.UserChairID = userItem.ChairID
		lotteryScene.UserList = it.specialUserList
		if !it.isOnTable(userItem) {
			//自己
			lotteryScene.UserList = append(lotteryScene.UserList, &msg.Game_S_User{
				UserID:       userItem.GetUserInfo().UserID,
				NikeName:     userItem.GetUserInfo().NikeName,
				UserGold:     userItem.GetUserInfo().UserGold,
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
			})
		}
		lotteryScene.LotteryRecord = it.lotteryRecord
		//lotteryScene.LotteryRecord = it.lotteryRecord
		lotteryScene.LotteryPoker = it.lotteryPoker

		lotteryScene.UserArraJetton = it.areaJettions
		//玩家每个区域的下注数
		var userJettonList = make([]*msg.Game_C_AreaJetton, 0)
		load, ok := it.userListAreaJetton.Load(userItem.ChairID)
		if ok {
			for k, v := range load.([global.AreaCount]float32) {
				userJetton := &msg.Game_C_AreaJetton{
					Area:   int32(k),
					Jetton: v,
				}
				userJettonList = append(userJettonList, userJetton)
			}
		} else {
			for i := 0; i < global.AreaCount; i++ {
				userJetton := &msg.Game_C_AreaJetton{
					Area:   int32(i),
					Jetton: 0,
				}
				userJettonList = append(userJettonList, userJetton)
			}
		}
		lotteryScene.UserJetton = userJettonList
		//lotteryScene.WinArea = it.winArea
		//lotteryScene.UserArraJetton = it.areaJettions
		//data, err := json.Marshal(lotteryScene)
		//if err != nil {
		//	_ = log.Logger.Errorf("onEventSendGameScene err %v", err)
		//	userItem.Close()
		//	return
		//}
		if userItem.UserID == 138684 || userItem.UserID == 138677 {
			fmt.Println("开奖场景消息", "开奖场景消息", time.Now())
		}
		userItem.WriteMsg(&lotteryScene)
	}
}

//用户坐下
func (it *Item) OnActionUserSitDown(args ...interface{}) {
	userItem := args[0].(*user.Item)
	// 检查是否锁定
	lock := mysql.GameClient.IsLock(userItem.UserID)
	if lock {
		_ = log.Logger.Errorf("OnActionUserSitDown 坐下失败 err %s", "坐下失败, 上局游戏未结束")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError1,
			ErrorMsg:  "坐下失败, 上局游戏未结束",
		})
		userItem.Close()
		return
	}

	//校验是否满人
	if it.userCount >= store.GameControl.GetGameInfo().ChairCount || it.chair.IsFull() {
		_ = log.Logger.Errorf("OnActionUserSitDown 坐下失败 err %s", "坐下失败, 房间人数已满!")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError2,
			ErrorMsg:  "坐下失败, 房间人数已满!",
		})
		userItem.Close()
		return
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
	// 取出椅子
	chair := it.chair.GetChair()
	if chair < 0 {
		_ = log.Logger.Errorf("OnActionUserSitDown 坐下失败 err %s", "坐下失败, 房间人数已满!")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError2,
			ErrorMsg:  "坐下失败, 房间人数已满!",
		})
		userItem.Close()
		return
	}
	//加入游戏用户列表
	it.userList.Store(chair, userItem.UserID)
	it.userCount++
	if userItem.IsRobot() {
		it.androidCount++
	}
	userItem.SitDown(it.tableID, chair)
	//for i := int32(0); i < store.GameControl.GetGameInfo().ChairCount; i++ {
	//	_, ok := it.userList.Load(i)
	//	if ok {
	//		continue
	//	}
	//	it.userList.Store(i, userItem.UserID)
	//	it.userCount++
	//	userItem.SitDown(it.tableID, i)
	//	break
	//}

	//发送坐下通知
	//it.sendAllUser(&msg.Game_S_SitDownNotify{
	//	UserChairID: chair,
	//	User: &msg.Game_S_User{
	//		UserID:       userItem.GetUserInfo().UserID,
	//		NikeName:     userItem.GetUserInfo().NikeName,
	//		UserGold:     userItem.GetUserInfo().UserGold,
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

	userItem.WriteMsg(&msg.Game_S_SitDownNotify{
		UserChairID: chair,
		User: &msg.Game_S_User{
			UserID:       userItem.GetUserInfo().UserID,
			NikeName:     userItem.GetUserInfo().NikeName,
			UserGold:     userItem.GetUserInfo().UserGold,
			UserDiamond:  userItem.GetUserInfo().UserDiamond,
			MemberOrder:  userItem.GetUserInfo().MemberOrder,
			HeadImageUrl: userItem.GetUserInfo().HeadImageUrl,
			FaceID:       userItem.GetUserInfo().FaceID,
			RoleID:       userItem.GetUserInfo().RoleID,
			SuitID:       userItem.GetUserInfo().SuitID,
			PhotoFrameID: userItem.GetUserInfo().PhotoFrameID,
			TableID:      userItem.GetUserInfo().TableID,
			ChairID:      userItem.GetUserInfo().ChairID,
			Status:       userItem.GetUserInfo().SuitID,
			Gender:       userItem.GetUserInfo().Gender,
		},
	})

	//发送场景消息
	if userItem.UserID == 138684 || userItem.UserID == 138677 {
		fmt.Println(it.tableID, "  坐下", "坐下", time.Now())
	}
	it.onEventSendGameScene(userItem)

}

//用户起立
func (it *Item) OnActionUserStandUp(args ...interface{}) {
	userItem := args[0].(*user.Item)
	flag := args[1].(bool)
	if !flag {
		//检测是否已押注
		v, ok := it.userListAreaJetton.Load(userItem.ChairID)
		if ok && it.gameStatus == global.GameStatusJetton {
			for _, v1 := range v.([global.AreaCount]float32) {
				if v1 != 0 {
					_ = log.Logger.Errorf("OnActionUserStandUp 用户起立 err %s", "押注状态下不允许退出!")
					userItem.WriteMsg(&msg.Game_S_ReqlyFail{
						ErrorCode: global.StandUpError1,
						ErrorMsg:  "押注状态下不允许退出!",
					})
					return
				}
			}
		}

	}
	//移出游戏用户列表
	var oldChairID = userItem.ChairID
	it.userList.Delete(oldChairID)
	it.chair.AddChair(oldChairID)
	it.userCount--
	if userItem.IsRobot() {
		it.androidCount--
	}
	userItem.StandUp()
	//解锁
	if err := mysql.GameClient.UnLock(userItem.UserID); err != nil {
		_ = log.Logger.Errorf("解锁用户失败 err %v", err)
		return
	}

	//删除这个座位的输赢记录
	//delete(it.userListWinRecord, userItem.ChairID)
	it.userListWinRecord.Delete(oldChairID)

	//起立通知桌面上的人通知和自己
	//if it.isOnTable(userItem) {
	//	it.sendAllUser(&msg.Game_S_StandUpNotify{
	//		ChairID: oldChairID,
	//	})
	//	if !userItem.IsRobot() {
	//		userItem.WriteMsg(&msg.Game_S_StandUpNotify{
	//			ChairID: oldChairID,
	//		})
	//	}
	//} else {
	//	if !userItem.IsRobot() {
	//		userItem.WriteMsg(&msg.Game_S_StandUpNotify{
	//			ChairID: oldChairID,
	//		})
	//	}
	//}

	userItem.WriteMsg(&msg.Game_S_StandUpNotify{
		ChairID: oldChairID,
	})
}

//用户断线
func (it *Item) OnActionUserOffLine(args ...interface{}) {
	userItem := args[0].(*user.Item)
	//设置用户状态
	userItem.Status = user.StatusOffline
	//断线通知桌面上的人
	if it.isOnTable(userItem) {
		it.sendOtherUser(userItem.UserID, &msg.Game_S_OffLineNotify{
			ChairID: userItem.ChairID,
		})
	}
}

//用户重入
func (it *Item) OnActionUserReconnect(args ...interface{}) {
	userItem := args[0].(*user.Item)
	if userItem.Status == user.StatusOffline {
		//设置用户状态
		userItem.Status = user.StatusPlaying
		//发送上线通知给其他玩家
		if it.isOnTable(userItem) {
			it.sendOtherUser(userItem.UserID, &msg.Game_S_OnLineNotify{
				ChairID: userItem.ChairID,
			})
		}
	}

	if !userItem.IsRobot() {
		userItem.WriteMsg(&msg.Game_S_SitDownNotify{
			UserChairID: userItem.ChairID,
			User: &msg.Game_S_User{
				UserID:       userItem.GetUserInfo().UserID,
				NikeName:     userItem.GetUserInfo().NikeName,
				UserGold:     userItem.GetUserInfo().UserGold,
				UserDiamond:  userItem.GetUserInfo().UserDiamond,
				MemberOrder:  userItem.GetUserInfo().MemberOrder,
				HeadImageUrl: userItem.GetUserInfo().HeadImageUrl,
				FaceID:       userItem.GetUserInfo().FaceID,
				RoleID:       userItem.GetUserInfo().RoleID,
				SuitID:       userItem.GetUserInfo().SuitID,
				PhotoFrameID: userItem.GetUserInfo().PhotoFrameID,
				TableID:      userItem.GetUserInfo().TableID,
				ChairID:      userItem.GetUserInfo().ChairID,
				Status:       userItem.GetUserInfo().SuitID,
				Gender:       userItem.GetUserInfo().Gender,
			},
		})
	}

	//发送场景消息
	it.onEventSendGameScene(userItem)
}

//下注事件
func (it *Item) OnUserPlaceJetton(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserJetton)
	//检测数据是否异常
	if m.JettonArea < 0 || m.JettonArea >= global.AreaCount || m.JettonScore <= 0 {
		_ = log.Logger.Errorf("OnUserPlaceJetton err %s", "下注失败, 无效的数据")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注失败, 无效的数据",
		})
		userItem.Close()
		return
	}

	if m.JettonArea == 1 || m.JettonArea == 2 {
		return
	}

	//检验是否是下注状态
	if it.gameStatus != global.GameStatusJetton {
		_ = log.Logger.Errorf("OnUserPlaceJetton err %s uid: %d", "下注失败, 非下注状态", userItem.UserID)
		//userItem.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError1,
		//	ErrorMsg:  "下注失败, 非下注状态",
		//})
		//userItem.Close()
		return
	}

	value, ok := it.userList.Load(userItem.ChairID)

	//检验用户是否在用户列表里
	if ok && userItem.UserID != value.(int32) {
		_ = log.Logger.Errorf("OnUserPlaceJetton err %s", "下注失败, 用户不在用户列表里")
		userItem.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注失败, 用户不在用户列表里",
		})
		userItem.Close()
		return
	}

	var tempScore float32
	v, ok := it.userListJetton.Load(userItem.ChairID)
	if ok {
		tempScore = v.(float32)
	}

	if !userItem.IsRobot() {

		if !redis.GameClient.IsExistsDiamond(userItem.UserID) {
			scoreInfo, _ := mysql.GetGameScoreInfoByUserId(mysql.GameClient.GetXJGameDB, userItem.UserID)
			redis.GameClient.SetDiamond(userItem.UserID, scoreInfo.Diamond)
		}

		userDiamond, err := redis.GameClient.GetDiamond(userItem.UserID)
		if err != nil {
			log.Logger.Error("GetDiamond err:", err)
		} else {

			userItem.UserDiamond = float32(userDiamond) + tempScore
		}
	}

	//判断下注积分是否足够
	if store.GameControl.GetGameInfo().DeductionsType == 0 {

		if userItem.UserGold < tempScore+m.JettonScore {
			_ = log.Logger.Errorf("OnUserPlaceJetton err %s", "下注失败, 金币不足!")
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.JettonError1,
				ErrorMsg:  "下注失败, 金币不足!",
			})
			return
		}
	} else {
		if userItem.UserDiamond < tempScore+m.JettonScore {
			_ = log.Logger.Errorf("OnUserPlaceJetton err %s", "下注失败, 余额不足!")
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.JettonError2,
				ErrorMsg:  "下注失败, 余额不足",
			})
			return
		}
	}

	if !userItem.IsRobot() {
		err := redis.GameClient.SetDiamond(userItem.UserID, userItem.UserDiamond-(tempScore+m.JettonScore))
		if err != nil {
			userItem.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.JettonError2,
				ErrorMsg:  "下注失败",
			})
			return
		}

		redis.GameClient.RegisterRecharge(userItem.UserID)
	}

	//记录下注
	userJetton, ok := it.userListAreaJetton.Load(userItem.ChairID)
	if ok {
		temp := userJetton.([global.AreaCount]float32)
		temp[m.JettonArea] += m.JettonScore
		it.userListAreaJetton.Store(userItem.ChairID, temp)
	} else {
		var temp [global.AreaCount]float32
		temp[m.JettonArea] += m.JettonScore
		it.userListAreaJetton.Store(userItem.ChairID, temp)
	}

	if score, ok := it.userListJetton.Load(userItem.ChairID); ok {
		temp := score.(float32)
		temp += m.JettonScore
		it.userListJetton.Store(userItem.ChairID, temp)
	} else {
		it.userListJetton.Store(userItem.ChairID, m.JettonScore)
	}

	// 富豪榜和神算子下注 才通知
	//if it.isOnTable(userItem) {
	//	it.sendAllUser(&msg.Game_S_UserJetton{
	//		ChairID:     userItem.ChairID,
	//		JettonArea:  m.JettonArea,
	//		JettonScore: m.JettonScore,
	//	})
	//} else {
	//	userItem.WriteMsg(&msg.Game_S_UserJetton{
	//		ChairID:     userItem.ChairID,
	//		JettonArea:  m.JettonArea,
	//		JettonScore: m.JettonScore,
	//	})
	//}

	userItem.WriteMsg(&msg.Game_S_UserJetton{
		ChairID:     userItem.ChairID,
		JettonArea:  m.JettonArea,
		JettonScore: m.JettonScore,
		UserScore:   userItem.UserDiamond - (tempScore + m.JettonScore),
	})
}

// 获取用户列表
func (it *Item) GetUserList(args ...interface{}) {
	userItem := args[0].(*user.Item)
	m := args[1].(*msg.Game_C_UserList)
	if m.Page <= 0 {
		m.Page = 1
	}
	// 控制数量
	if m.Size >= 10 {
		m.Size = 10
	}
	var list = make([]*user.Info, 0)
	it.userList.Range(func(key, uid interface{}) bool {
		item, ok := user.List.Load(uid)
		if !ok {
			_ = log.Logger.Errorf("GetUserList err %d", uid)
			return false
		}
		list = append(list, item.(*user.Item).GetUserInfo())
		return true
	})
	start := (m.Page - 1) * m.Size
	end := start + m.Size

	lenSize := int32(len(list))
	if lenSize <= m.Size {
		start = 0
		end = lenSize
	} else {
		if start >= lenSize {
			start = lenSize - 1
			if start <= 0 {
				start = 0
			}
		}
		if end >= lenSize {
			end = lenSize - 1
			if end <= 0 {
				end = start
			}
		}
	}
	//type tempUser struct {
	//	*user.Info
	//	TotalJetton float32 `json:"total_jetton"`
	//	TotalWin    int32   `json:"total_win"`
	//}
	var dataTempUser = make([]*msg.Game_S_TempUser, 0)
	for _, v := range list[start:end] {
		var totalJetton float32
		var totalWin int32
		for _, jetton := range it.userListJettons {
			value, ok := jetton.Load(v.ChairID)
			if ok {
				totalJetton += value.(float32)
			}
		}
		it.userListWinRecord.Range(func(key, value interface{}) bool {
			if key.(int32) == v.ChairID {
				for _, win := range value.([global.WinRecordCount]bool) {
					if win {
						totalWin++
					}
				}
			}
			return true
		})
		dataTempUser = append(dataTempUser, &msg.Game_S_TempUser{
			User: &msg.Game_S_User{
				UserID:       v.UserID,
				NikeName:     v.NikeName,
				UserGold:     v.UserGold,
				UserDiamond:  v.UserDiamond,
				MemberOrder:  v.MemberOrder,
				HeadImageUrl: v.HeadImageUrl,
				FaceID:       v.FaceID,
				RoleID:       v.RoleID,
				SuitID:       v.SuitID,
				PhotoFrameID: v.PhotoFrameID,
				TableID:      v.TableID,
				ChairID:      v.ChairID,
				Status:       v.Status,
				Gender:       v.Gender,
			},
			TotalJetton: totalJetton,
			TotalWin:    totalWin,
		})
	}

	userItem.WriteMsg(&msg.Game_S_UserList{
		Data: dataTempUser,
	})
}

// 发送所有人
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

// 发送其他人
func (it *Item) sendOtherUser(userID int32, data interface{}) {
	it.userList.Range(func(chairID, uid interface{}) bool {
		//过滤userID
		if userID == uid {
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

// 判断当前人是否在桌子显示
func (it *Item) isOnTable(userItem *user.Item) bool {
	isSend := false
	for _, v := range it.specialUserList {
		if v.UserID == userItem.UserID {
			isSend = true
			break
		}
	}
	return isSend
}
