package logic

import (
	"sync"
	"xj_game_server/game/103_bairenniuniu/global"
	"xj_game_server/game/103_bairenniuniu/msg"
	publicLogic "xj_game_server/game/public/logic"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
)

var Client = new(Logic)

type Logic struct {
}

//发牌
func (logic *Logic) DispatchTableCard() map[int32]*msg.Game_S_LotteryPoker {

	var lotteryPoker = make(map[int32]*msg.Game_S_LotteryPoker)
	//0x01:第一位代表花色 第二位代表牌大小
	pokers := publicLogic.Poker{
		Count:        1,     //1副牌
		IsNeedGhost:  false, //不需要大小鬼
		ShuffleCount: 10,    //洗牌次数
	}.GetPokers()

	//赋值
	for i := int32(0); i < 5; i++ {

		var d msg.Game_S_LotteryPoker
		d.LotteryPoker = pokers[(i * 5):((i + 1) * 5)]
		d.PokerType = logic.getPokerType(d.LotteryPoker)
		lotteryPoker[i] = &d
	}

	//var d msg.Game_S_LotteryPoker
	//d.LotteryPoker = []int32{0x02, 0x03, 0x04, 0x05, 0x0D}
	//d.PokerType = logic.getPokerType(d.LotteryPoker)
	//lotteryPoker[2] = &d
	//
	//var d1 msg.Game_S_LotteryPoker
	//d1.LotteryPoker = []int32{0x12, 0x13, 0x14, 0x15, 0x1D}
	//d1.PokerType = logic.getPokerType(d1.LotteryPoker)
	//lotteryPoker[0] = &d1
	//
	//var d2 msg.Game_S_LotteryPoker
	//d2.LotteryPoker = []int32{0x22, 0x23, 0x24, 0x25, 0x2D}
	//d2.PokerType = logic.getPokerType(d2.LotteryPoker)
	//lotteryPoker[3] = &d2
	//
	//var d3 msg.Game_S_LotteryPoker
	//d3.LotteryPoker = []int32{0x32, 0x33, 0x34, 0x35, 0x3D}
	//d3.PokerType = logic.getPokerType(d3.LotteryPoker)
	//lotteryPoker[1] = &d3
	//
	//var d4 msg.Game_S_LotteryPoker
	//d4.LotteryPoker = []int32{0x09, 0x38, 0x18, 0x08, 0x0C}
	//d4.PokerType = logic.getPokerType(d4.LotteryPoker)
	//lotteryPoker[4] = &d4

	return lotteryPoker
}

//比牌 反赔率
func (logic *Logic) compareType(poker1 int32, poker2 int32) float32 {

	var area0, area1, area2, area3, area4, area5 int32

	area0 = poker1 & 0xf00 / 0x100
	area1 = poker2 & 0xf00 / 0x100
	area2 = poker1 & 0xf0 / 0x10
	area3 = poker2 & 0xf0 / 0x10
	area4 = poker1 & 0xf
	area5 = poker2 & 0xf

	if area0 > area1 {
		return -global.AreaMultiple[area0]
	} else if area0 < area1 {
		return global.AreaMultiple[area1]
	} else {

		if area4 > area5 {
			return -global.AreaMultiple[area0]
		} else if area4 < area5 {
			return global.AreaMultiple[area1]
		} else {
			if area2 > area3 {
				return -global.AreaMultiple[area0]
			} else {
				return global.AreaMultiple[area1]
			}
		}

	}

}

//获取获胜区域
//lotteryPoker 开奖扑克
func (logic *Logic) GetWinArea(lotteryPoker map[int32]*msg.Game_S_LotteryPoker) []float32 {
	var winArea = make([]float32, 0)

	var wins [5]int32
	wins = logic.getWins(lotteryPoker)
	for i := 1; i < len(wins); i++ {

		winArea = append(winArea, logic.compareType(wins[0], wins[i]))
	}

	return winArea
}

func (logic *Logic) getWins(lotteryPoker map[int32]*msg.Game_S_LotteryPoker) [5]int32 {

	var wins [5]int32

	for k, v := range lotteryPoker {
		//wins[k] = logic.getPokerType(v.LotteryPoker)
		wins[k] = v.PokerType
	}

	return wins

}

//返回系统损耗
func (*Logic) GetSystemLoss(winArea []float32, userListAreaJetton sync.Map, userList sync.Map) (float32, map[int32]float32, map[int32]float32, float32) {

	userListLoss := make(map[int32]float32)
	userTax := make(map[int32]float32)
	var systemScore, userNoRootTax float32

	userListAreaJetton.Range(func(k, v interface{}) bool {
		//[global.AreaCount]float32
		var userItem interface{}
		if value, ok := userList.Load(k); ok {
			userItem, ok = user.List.Load(value)
			if !ok {
				return true
			}
		} else {
			return true
		}

		for i := 0; i < global.AreaCount; i++ {
			if v.([global.AreaCount]float32)[i] == 0 {
				continue
			}
			//机器人批次号为 > 0
			if !userItem.(*user.Item).IsRobot() {
				userNoRootTax += userListLoss[k.(int32)]
			}
			userWins := v.([global.AreaCount]float32)[i] * winArea[i]
			if winArea[i] > 0 {

				//机器人批次号为 > 0
				if !userItem.(*user.Item).IsRobot() {
					userNoRootTax += userWins
				}
				// 判断是否>0 赢了 要扣税
				userListLoss[k.(int32)] += userWins - userWins*store.GameControl.GetGameInfo().RevenueRatio
				// 记录税收
				userTax[k.(int32)] += userWins * store.GameControl.GetGameInfo().RevenueRatio
			} else {
				userListLoss[k.(int32)] += userWins
			}

		}

		//机器人批次号为 > 0
		if userItem.(*user.Item).BatchID != -1 {
			return true
		}

		systemScore -= userListLoss[k.(int32)]
		return true
	})

	return systemScore, userListLoss, userTax, userNoRootTax
}

// 没牛
// 牛1 2 3 4 5 6
// 牛7 8
// 牛9
// 牛牛
// 五花牛
// 炸弹
// 五小牛
func (logic *Logic) getPokerType(lotteryPoker []int32) int32 {

	var pokerType int32

	num := publicLogic.GetCombined(lotteryPoker...)

	var sum int32
	var cardMax int32
	var smallCowNum int32
	var bigCowNum int32
	var friedNum int32
	var friedNum1 int32

	for i, v := range lotteryPoker {

		cardNum := v & 0x0F
		sum += cardNum

		cardMax = publicLogic.ContrastSize(cardMax, lotteryPoker[i])

		if cardNum > 10 {
			bigCowNum++
		}

		if i != 0 {
			if cardNum == (lotteryPoker[0] & 0x0F) {
				friedNum++
			}

			if i != 1 {
				if cardNum == (lotteryPoker[1] & 0x0F) {
					friedNum1++
				}
			}

		}

		if cardNum <= 5 {
			smallCowNum++
		}
	}

	if smallCowNum == 5 && sum <= 10 {

		pokerType = 0xd00 + int32(cardMax)

		return pokerType
	}

	if friedNum >= 4 || friedNum1 >= 4 {
		pokerType = 0xc00 + int32(cardMax)

		return pokerType
	}

	if bigCowNum == 5 {
		pokerType = 0xb00 + int32(cardMax)

		return pokerType
	}

	if bigCowNum == 3 || bigCowNum == 4 {

		if num == 0 {
			pokerType = 0xa00 + int32(cardMax)
			return pokerType
		} else if num == 9 {
			pokerType = 0x900 + int32(cardMax)
			return pokerType
		} else if num == 8 {
			pokerType = 0x800 + int32(cardMax)
			return pokerType
		} else if num == 7 {
			pokerType = 0x700 + int32(cardMax)
			return pokerType
		} else {
			pokerType = (0x100 * int32(num)) + int32(cardMax)
			return pokerType
		}
	}

	judge := logic.judge(num, lotteryPoker)

	if !judge {
		return int32(cardMax)
	} else {
		if num == 0 {
			pokerType = 0xa00 + int32(cardMax)
			return pokerType
		} else {
			pokerType = (0x100 * int32(num)) + int32(cardMax)
			return pokerType
		}

	}
}

// 求合 算牛
func (logic *Logic) judge(sum int32, lotteryPoker []int32) bool {

	if len(lotteryPoker) == 0 {
		return false
	}

	for i := 1; i < len(lotteryPoker); i++ {

		if publicLogic.GetCombined(lotteryPoker[0], lotteryPoker[i]) == sum {
			return true
		}

	}

	return logic.judge(sum, lotteryPoker[1:])

}
