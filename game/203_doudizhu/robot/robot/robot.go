/*
 * @Author: yhlyl
 * @Date: 2019-11-27 14:48:15
 * @LastEditTime: 2019-11-27 15:02:47
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/101_longhudou/robot/robot/robot.go
 * @https://github.com/android-coco
 */
package robot

import (
	"sync"
	"time"
	gameLogic "xj_game_server/game/203_doudizhu/game/logic"
	"xj_game_server/game/203_doudizhu/global"
	"xj_game_server/game/203_doudizhu/msg"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/util"
)

// RobotList 机器人列表 uid --> Robot
//var List = make(map[int32]*Robot)
var List sync.Map

var Count int

// Robot 机器人
type Robot struct {
	*robot.Agent                       // 机器人连接代理
	batchID          int32             // 批次id
	userID           int32             // 用户id
	userChairID      int32             // 座位号
	tableStatus      bool              //桌子状态
	userStatus       bool              //用户状态 playing
	currentChairID   int32             //当前出牌玩家
	bankerChairID    int32             //地主椅子号
	dizPokers        []int32           //地主牌
	gold             float32           //用户金币
	diamond          float32           //用户余额
	rounds           int32             //当前轮数
	pokers           []int32           //牌
	currentMultiple  int32             //当前倍数
	withProbability  int32             //推出概率
	nearestChairID   int32             //最近出牌椅子号
	nearestPokers    []int32           //最近出牌扑克
	nearestPokerType int32             //最近出牌扑克类型
	gameStatus       int               //游戏状态 1.叫分状态,2出牌状态
	cronTimer        *time.Timer       // 定时器
	HandPokers       map[int32][]int32 // 手上牌逻辑值//数量 牌
	handAlarm        map[int32]int32   // 报警剩单或者剩双 座位 数量
	outCards         map[int32][]int32 // 出过的牌 座位 牌
	multipleJF       int32             // 叫分
}

// OnInit 初始化
func (r *Robot) OnInit(userID int32, batchID int32, gate *gate.Gate, userCallBack func(args []interface{})) {
	r.Agent = new(robot.Agent)
	r.Agent.OnInit(gate, userCallBack)
	r.batchID = batchID
	r.userID = userID
	r.userChairID = -1
	r.currentChairID = -1
	r.currentMultiple = 0
	r.bankerChairID = -1
	r.dizPokers = make([]int32, 0)
	r.nearestChairID = -1
	r.gameStatus = 0
	r.HandPokers = make(map[int32][]int32, 0)
	r.handAlarm = make(map[int32]int32, 0)
	r.outCards = make(map[int32][]int32, 0)
	r.cronTimer = &time.Timer{}
}

// 获取地主牌
func (r *Robot) GetDizPoker() []int32 {
	return r.dizPokers
}

// 获取地主座位号
func (r *Robot) GetDizChairID() int32 {
	return r.bankerChairID
}

// 设置地主牌
func (r *Robot) SetDizPokerAndChairId(chairId int32, dizPoker []int32) {
	var dizCard []int32
	for _, v := range dizPoker {
		dizCard = append(dizCard, v)
	}
	r.bankerChairID = chairId
	r.dizPokers = dizCard

}

//GetUserID 获取用户id
func (r *Robot) GetUserID() int32 {
	return r.userID
}

//GetBatchID 获取批次号
func (r *Robot) GetBatchID() int32 {
	return r.batchID
}

//GetGold 获取金币
func (r *Robot) GetGold() float32 {
	return r.gold
}

//GetDiamond 获取余额
func (r *Robot) GetDiamond() float32 {
	return r.diamond
}

// Login 登录
func (r *Robot) Login() {
	r.WriteMsg(&msg.Game_C_RobotLogin{
		UserID:  r.userID,
		BatchID: r.batchID,
	})
}

// Assignment 给机器人给金币
func (r *Robot) Assignment() {
	batchId := r.GetBatchID()
	maxCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxCoin
	minCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMinCoin

	maxDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxDiamond
	minDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMinDiamond
	if minDiamond < store.GameControl.GetGameInfo().MinEnterScore {
		minDiamond = store.GameControl.GetGameInfo().MinEnterScore
	}
	coin := float32(util.RandInterval(int32(minCoin), int32(maxCoin)-1))
	diamond := float32(util.RandInterval(int32(minDiamond), int32(maxDiamond)-1))

	value, ok := user.List.Load(r.userID)
	if ok {
		value.(*user.Item).UserGold = coin
		value.(*user.Item).UserDiamond = diamond
	}
	r.gold = coin
	r.diamond = diamond
}

// 发起准备
func (r *Robot) Prepare() {
	r.WriteMsg(&msg.Game_C_UserPrepare{})
}

//修改准备状态 坐下
func (r *Robot) SetTableStatus(tableStatus bool) {
	r.tableStatus = tableStatus
}

//获取准备状态
func (r *Robot) GetTableStatus() bool {
	return r.tableStatus
}

//修改准备状态 坐下
func (r *Robot) SetUserStatus(userStatus bool) {
	r.userStatus = userStatus
}

// 设置游戏状态
func (r *Robot) SetGameStatus(status int) {
	r.gameStatus = status
}

// 获取游戏状态
func (r *Robot) GetGameStatus() int {
	return r.gameStatus
}

// 设置最近出牌
func (r *Robot) SetCurrentPokers(chairId, pokerType int32, pokers []int32) {
	r.nearestChairID = chairId
	r.nearestPokers = pokers
	r.nearestPokerType = pokerType
}

// 设置当前操作座位号
func (r *Robot) SetCurrentChairID(chairId int32) {
	r.currentChairID = chairId
}

// 处理叫分事件和出牌事件  随机任务时间
func (r *Robot) PokersCpORJF() {
	if r.gameStatus == global.GameStatusJF {
		r.cronTimer.Reset(time.Millisecond * time.Duration(util.RandInterval(1000, 4000)))
	} else {
		r.cronTimer.Reset(time.Millisecond * time.Duration(util.RandInterval(500, 1500))) //  出牌时间随机
	}
}

// 定时处理出牌和叫分
func (r *Robot) onEventGameTimer() {
	for r.gameStatus != global.GameStatusFree {
		select {
		case <-r.cronTimer.C:
			if r.userChairID == r.currentChairID {
				switch r.gameStatus {
				case global.GameStatusJF:
					if r.currentMultiple == r.multipleJF {
						r.multipleJF = 0
					}
					r.StartJF(r.multipleJF)
				case global.GameStatusPlay:
					//if r.nearestChairID == r.userChairID {
					//	r.StartFirstCP()
					//} else {
					//	r.StartCP()
					//}
					gameLogic.Client.TernaryOperator(r.nearestChairID == r.userChairID, r.StartFirstCP, r.StartCP)
				}
			}
		}
	}
}

// 获取当前的操作座位号
func (r *Robot) GetCurrentChairID() int32 {
	return r.currentChairID
}

//获取准备状态
func (r *Robot) GetUserStatus() bool {
	return r.userStatus
}

//修改过牌状态
func (r *Robot) SetRounds(rounds int32) {
	r.rounds = rounds
}

//获取过牌状态
func (r *Robot) GetRounds() int32 {
	return r.rounds
}

//修改用户座位号
func (r *Robot) SetUserChairID(userChairID int32) {
	r.userChairID = userChairID
}

//获取用户座位号
func (r *Robot) GetUserChairID() int32 {
	return r.userChairID
}

//获取当前分数
func (r *Robot) GetChip() int32 {
	return r.currentMultiple
}

//设置当前分数
func (r *Robot) SetChip(currentMultiple int32) {
	r.currentMultiple = currentMultiple
}
func (r *Robot) GetPokers() []int32 {
	return r.pokers
}

func (r *Robot) SetPokers(pokers []int32) {
	r.pokers = pokers
}

func (r *Robot) IsPokersEmpty() bool {
	return len(r.pokers) == 0
}

//SitDown 坐下
func (r *Robot) SitDown() {
	r.WriteMsg(&msg.Game_C_UserSitDown{
		ChairID: -1,
	})
}

//StandUp 起立
func (r *Robot) StandUp() {
	r.WriteMsg(&msg.Game_C_UserStandUp{})
}

//CheckBatchTimeOut 检查批次是否过期
func (r *Robot) CheckBatchTimeOut() bool {
	// 批次号是否过期，过期必退
	_, ok := robot.RobotConfigItem.GetConfig()[r.batchID]
	return ok
}

// 开始游戏
func (r *Robot) StartGame(pokers []int32) {
	cards := make([]int32, 0)
	for _, card := range pokers {
		cards = append(cards, card)
	}

	r.multipleJF = gameLogic.Client.GetJFMultiple(cards)
	r.tableStatus = true
	r.userStatus = true
	r.pokers = cards
	r.cronTimer = time.NewTimer(time.Second * time.Duration(util.RandInterval(1, 4)))
	go r.onEventGameTimer()
	r.conversionPokerType()
}

// 游戏结束，清空数据
func (r *Robot) GameEnd() {
	r.tableStatus = false
	r.userStatus = false
	r.gameStatus = 0
	r.rounds = 0
	r.currentMultiple = 0
	r.currentChairID = -1
	r.pokers = make([]int32, 0)
	r.HandPokers = make(map[int32][]int32, 0)
	r.withProbability += 5
	r.currentMultiple = 0
	r.bankerChairID = -1
	r.multipleJF = 0
	r.handAlarm = make(map[int32]int32, 0)
	r.outCards = make(map[int32][]int32, 0)
	r.dizPokers = make([]int32, 0)
	r.RandStandUp()
}

// 随机推出
func (r *Robot) RandStandUp() {
	//	var isStandUp = false
	t := time.NewTimer(time.Millisecond * time.Duration(util.RandInterval(500, 1000)))
	go func() {
		for {
			select {
			case <-t.C:
				//if !isStandUp {
				//	r.StandUp()
				//	isStandUp = true
				//	t.Reset(time.Second * 2)
				//} else {
				//	r.SitDown()
				//	return
				//}
				r.SitDown()
				return
			}
		}
	}()
}

// 叫分
func (r *Robot) StartJF(Multiple int32) {
	r.WriteMsg(&msg.Game_C_UserGrabLandlord{Multiple: Multiple})

}

// 出牌
func (r *Robot) StartCP() {
	var outCards = make([]int32, 0)
	var isPass = false
	var bomb []int32
	randN := util.RandInterval(0, 8)
	if r.nearestPokerType == global.CardTypeKINGBOMB {
		r.WriteMsg(&msg.Game_C_UserPass{})
		return
	}
	if len(r.handAlarm) > 0 {
		outCards = gameLogic.Client.GetAutoOutPokersByType(r.nearestPokerType, r.nearestPokers, r.pokers)
		for chairID, num := range r.handAlarm {
			if chairID != r.bankerChairID && r.userChairID != r.bankerChairID {
				if chairID == r.userChairID && num == r.nearestPokerType {
					isPass = true
				}
				if randN <= 7 && r.nearestChairID == chairID {
					isPass = true
				}
			}
			if chairID == r.bankerChairID && r.nearestChairID != r.bankerChairID && r.userChairID != r.bankerChairID {
				isPass = true
			}
		}
		if isPass && len(outCards) != len(r.pokers) {
			r.WriteMsg(&msg.Game_C_UserPass{})
			return
		}
		if len(outCards) > 0 {
			r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
			return
		} else if len(outCards) == 0 && r.nearestPokerType < global.CardTypeBOMB && (len(r.HandPokers[4]) > 0 || len(r.HandPokers[5]) == 2) {
			if len(r.HandPokers[4]) > 0 {
				bomb = []int32{r.HandPokers[4][0]}
			} else if len(r.HandPokers[5]) == 2 {
				bomb = r.HandPokers[5]
			}
			bomb = r.getOriginData(bomb, false)
			r.WriteMsg(&msg.Game_C_UserCP{Pokers: bomb})
			return
		}
		if r.nearestPokerType == global.CardTypeSINGLE {
			for i := len(r.pokers) - 1; i >= 0; i-- {
				if gameLogic.Client.GetLogicValue(r.pokers[i]) > gameLogic.Client.GetLogicValue(r.nearestPokers[0]) {
					outCards = []int32{r.pokers[i]}
					break
				}
			}
			if len(outCards) > 0 {
				r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
				return
			}
		}
		r.WriteMsg(&msg.Game_C_UserPass{})
		return
	}
	outCards = gameLogic.Client.GetAutoOutPokersByType(r.nearestPokerType, r.nearestPokers, r.pokers)

	// 最近出牌的是队友，并且小于顺子则不出
	if r.userChairID != r.bankerChairID && r.nearestChairID != r.bankerChairID && len(outCards) > 0 {
		if len(outCards) != len(r.pokers) && r.nearestPokerType >= global.CardTypeSINGLEALONE {
			outCards = make([]int32, 0)
		}
		if r.nearestPokerType == global.CardTypeSINGLE || r.nearestPokerType == global.CardTypeDOUBLE {
			_, maxCard := gameLogic.Client.GetPokerType(r.nearestPokers)
			if gameLogic.Client.GetLogicValue(maxCard) >= 0x0F || gameLogic.Client.GetLogicValue(outCards[0]) >= 0x0F {
				outCards = make([]int32, 0)
			}
		}
	} else if r.userChairID != r.bankerChairID && r.nearestPokerType == global.CardTypeSINGLE &&
		r.nearestChairID == r.bankerChairID && len(outCards) == 0 && randN <= 5 {
		for i := len(r.pokers) - 1; i > 0; i-- {
			if gameLogic.Client.GetLogicValue(r.pokers[i]) == 0x10 || gameLogic.Client.GetLogicValue(r.pokers[i]) == 0x11 {
				continue
			}
			if gameLogic.Client.GetLogicValue(r.pokers[i]) > gameLogic.Client.GetLogicValue(r.nearestPokers[0]) {
				outCards = []int32{r.pokers[i]}
				break
			}
		}
	}
	// 自己是地主的话要随机接牌
	if r.userChairID == r.bankerChairID && r.nearestPokerType == global.CardTypeSINGLE && len(outCards) == 0 && randN <= 5 {
		for i := len(r.pokers) - 1; i > 0; i-- {
			if gameLogic.Client.GetLogicValue(r.pokers[i]) == 0x10 || gameLogic.Client.GetLogicValue(r.pokers[i]) == 0x11 {
				continue
			}
			if gameLogic.Client.GetLogicValue(r.pokers[i]) > gameLogic.Client.GetLogicValue(r.nearestPokers[0]) {
				outCards = []int32{r.pokers[i]}
				break
			}
		}
	}
	if len(outCards) > 0 {
		r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
	} else {
		outCards = gameLogic.Client.GetBomb(r.nearestPokerType, r.pokers)
		if len(outCards) > 0 {
			r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
		} else {
			r.WriteMsg(&msg.Game_C_UserPass{})
		}
	}
}

// 其它座位预警的情况 首次出牌
func (r *Robot) getPokersType(num int32, isEqualType bool) (cards []int32) {
	var count int32
	var pokers []int32
	cards = make([]int32, 0)
	for count, pokers = range r.HandPokers {
		if count == num {
			if len(pokers) > 0 && isEqualType {
				cards = []int32{pokers[0]}
				break
			}
			continue
		}
		if len(pokers) > 0 && count != 5 {
			cards = []int32{pokers[0]}
			break
		}
	}
	if count == 3 {
		if len(r.HandPokers[1]) > 0 {
			cards = append(cards, r.HandPokers[1][0])
		} else if len(r.HandPokers[2]) > 0 {
			cards = append(cards, r.HandPokers[2][0])
		}
	} else if count == 4 {
		if len(r.HandPokers[1]) > 1 {
			cards = append(cards, r.HandPokers[1][0], r.HandPokers[1][1])
		} else if len(r.HandPokers[2]) >= 1 {
			cards = append(cards, r.HandPokers[2][0])
			if len(r.HandPokers[2]) >= 2 {
				cards = append(cards, r.HandPokers[2][1])
			}
		}
	}
	if len(cards) == 0 && !isEqualType {
		if len(r.HandPokers[5]) != 2 && num == 1 {
			cards = []int32{r.pokers[len(r.pokers)-1]}
		} else if len(r.HandPokers[num]) > 0 {
			cards = []int32{r.HandPokers[num][len(r.HandPokers[num])-1]}
		}
	}
	cards = r.getOriginData(cards, false)
	return
}

// 首次出牌策略,出掉最多的,留手上最少 类型1，单牌 2对子，3，顺子，4连对，5三带n(n<=2)
func (r *Robot) StartFirstCP() {
	var outCards []int32
	var determine = false
	var pokerNum int32 // 牌的张数
	if len(r.handAlarm) > 0 {
		for chairID, num := range r.handAlarm {
			if chairID == r.bankerChairID && r.userChairID != r.bankerChairID {
				outCards = r.getPokersType(num, false)
			} else if chairID != r.bankerChairID && r.userChairID != r.bankerChairID {
				if chairID == r.currentChairID+1%3 { // 下一次出牌的是不是队友
					if num == 1 {
						for _, card := range r.pokers {
							if gameLogic.Client.GetLogicValue(card) < 0x0A { // 出牌10以下的
								outCards = []int32{card}
								break
							}
						}
					}
					if len(outCards) == 0 {
						outCards = r.getPokersType(num, true)
					}
				} else if len(outCards) == 0 {
					outCards = r.getPokersType(num, true)
				}
			} else if chairID != r.bankerChairID && r.userChairID == r.bankerChairID {
				outCards = r.getPokersType(num, false)
			}
		}
		if len(outCards) == 0 {
			if len(r.pokers) == 2 && len(r.HandPokers[5]) == 2 {
				outCards = r.HandPokers[5]
			} else {
				outCards = []int32{r.pokers[0]}
			}
		}
		r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
		return
	}
	poker := r.pokers[0]
	for cardNum, cards := range r.HandPokers {
		for _, card := range cards {
			if gameLogic.Client.GetLogicValue(poker) == gameLogic.Client.GetLogicValue(card) {
				determine = true
				break
			}
		}
		if determine {
			pokerNum = cardNum
			break
		}

	}
	if gameLogic.Client.GetLogicValue(poker) < 0x07 { // 小于7出顺子
		if pokerNum <= 2 {
			outCards = gameLogic.Client.GetContinuityByCardType(global.CardTypeSINGLEALONE, r.pokers)
			outCards = r.getOriginData(outCards, true)
		}
		if pokerNum == 2 && len(outCards) == 0 {
			outCards = gameLogic.Client.GetContinuityByCardType(global.CardTypeDOUBLEALONE, r.pokers)
			outCards = r.getOriginData(outCards, false)
		}
		if len(outCards) > 0 {
			r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
			return
		}
	}
	outCards = []int32{poker}
	switch pokerNum {
	case 1:
		if len(r.pokers) == 4 && len(r.HandPokers[3]) == 1 {
			outCards = append(outCards, r.HandPokers[3]...)
		}
		outCards = r.getOriginData(outCards, false)
	case 2: // 优先出连对
		outPokers := gameLogic.Client.GetContinuityByCardType(global.CardTypeDOUBLEALONE, r.pokers)
		if len(outPokers) > 0 {
			outCards = outPokers
		}
		outCards = r.getOriginData(outCards, false)
	case 3:
		k := len(outCards) - 1
		for i := range r.HandPokers[3] {
			if gameLogic.Client.GetLogicValue(outCards[k]) >= 0x0E {
				break
			}
			if gameLogic.Client.GetLogicValue(outCards[k])+1 == gameLogic.Client.GetLogicValue(r.HandPokers[3][i]) {
				outCards = append(outCards, r.HandPokers[3][i])
				k = len(outCards) - 1
			}
		}
		if len(r.HandPokers[1]) >= len(outCards) {
			outCards = append(outCards, r.HandPokers[1][0:len(outCards)]...)
		} else if len(r.HandPokers[2]) >= len(outCards) {
			outCards = append(outCards, r.HandPokers[2][0:len(outCards)]...)
		}
		outCards = r.getOriginData(outCards, false)
	case 4:
		if len(r.HandPokers[1]) > 1 {
			outCards = append(outCards, r.HandPokers[1][0], r.HandPokers[1][1])
		} else if len(r.HandPokers[2]) > 1 {
			outCards = append(outCards, r.HandPokers[2][0], r.HandPokers[2][1])
		} else if len(r.HandPokers[2]) == 1 {
			outCards = append(outCards, r.HandPokers[2][0])
		}
		outCards = r.getOriginData(outCards, false)
	case 5:
		if len(r.HandPokers[5]) == 2 {
			for _, card := range r.HandPokers[5] {
				if poker != card {
					outCards = append(outCards, card)
				}
			}
		}
		outCards = r.getOriginData(outCards, false)
	}
	r.WriteMsg(&msg.Game_C_UserCP{Pokers: outCards})
}

// 重新设置手牌
func (r *Robot) ResetPoker() {
	for i := 0; i < len(r.nearestPokers); i++ {
		for j := 0; j < len(r.pokers); j++ {
			if r.nearestPokers[i] == r.pokers[j] {
				r.pokers = append(r.pokers[:j], r.pokers[j+1:]...)
				break
			}
		}
	}
	r.conversionPokerType()
}

// 添加出过的牌
func (r *Robot) AddOutCards(outCards []int32, chairID int32) {
	for _, card := range outCards {
		r.outCards[chairID] = append(r.outCards[chairID], card)
	}
	var num int32 = 17 // 非地主的牌数
	if chairID == r.bankerChairID {
		num = 20
	}
	num = num - int32(len(r.outCards[chairID]))
	if num == 1 || num == 2 {
		r.handAlarm[chairID] = num
	}
}

// 手牌添加地主牌
func (r *Robot) AppendDzPokers() {
	r.pokers = append(r.pokers, r.dizPokers...)
	r.conversionPokerType()
}

// 逻辑值转换为源数据
func (r *Robot) getOriginData(card []int32, isSingle bool) []int32 {
	pokerVal := make([]int32, 0)
	var pokers = r.pokers
	size := len(pokers)
	for j := 0; j < len(card); j++ {
		for i := 0; i < size; i++ {
			if gameLogic.Client.GetLogicValue(pokers[i]) == gameLogic.Client.GetLogicValue(card[j]) {
				if isSingle {
					pokerVal = append(pokerVal, pokers[i])
					break
				} else {
					pokerVal = append(pokerVal, pokers[i])
				}
			}
		}
	}
	return pokerVal
}

// 数量-牌,是王  key 表示5 value表示大王或者小王
func (r *Robot) conversionPokerType() {
	var mold = make(map[int32]int32, 0) //牌 数量

	for _, card := range r.pokers {
		if card == 0x4F || card == 0x5F {
			mold[card] = 0
		} else {
			mold[card&0x0F]++
		}
	}

	var moldNum = make(map[int32][]int32, 0) //数量 牌
	for card, num := range mold {
		var cards = moldNum[num]
		if card == 0x4F || card == 0x5F {
			moldNum[5] = append(moldNum[5], card)
		} else {
			moldNum[num] = append(cards, card)
		}
	}

	//排序扑克
	for k, v := range moldNum {
		for i := 0; i < len(v)-1; i++ {
			for j := i + 1; j < len(v); j++ {
				if gameLogic.Client.GetLogicValue(v[i]) > gameLogic.Client.GetLogicValue(v[j]) {
					moldNum[k][i], moldNum[k][j] = moldNum[k][j], moldNum[k][i]
				}
			}
		}
	}
	r.HandPokers = moldNum

	for i := 0; i < len(r.pokers)-1; i++ {
		for j := i + 1; j < len(r.pokers); j++ {
			if gameLogic.Client.GetLogicValue(r.pokers[i]) > gameLogic.Client.GetLogicValue(r.pokers[j]) {
				r.pokers[i], r.pokers[j] = r.pokers[j], r.pokers[i]
			}
		}
	}
}
