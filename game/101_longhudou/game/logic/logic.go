package logic

import (
	"xj_game_server/game/101_longhudou/global"
	publicLogic "xj_game_server/game/public/logic"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"sync"
)

var Client = new(Logic)

type Logic struct {
}

//发牌[global.PokerCount]uint8
func (*Logic) DispatchTableCard() []int32 {
	//0x01:第一位代表花色 第二位代表牌大小
	pokers := publicLogic.Poker{
		Count:        3,     //3副牌
		IsNeedGhost:  false, //不需要大小鬼
		ShuffleCount: 100,   //洗牌次数
	}.GetPokers()
	//赋值
	var lotteryPoker = make([]int32, 0)
	for i := 0; i < global.PokerCount; i++ {
		lotteryPoker = append(lotteryPoker, int32(pokers[i]))
	}
	return lotteryPoker
}

//比牌 1 龙赢 -1 虎赢 0 和
func (logic *Logic) compare(poker1 int32, poker2 int32) int {
	poker1Value := publicLogic.GetValue(poker1)
	poker2Value := publicLogic.GetValue(poker2)
	//poker1Color := publicLogic.GetColor(poker1)
	//poker2Color := publicLogic.GetColor(poker2)
	// 获取牌大小 去掉花色
	if poker1Value > poker2Value {
		return 1
	} else if poker1Value < poker2Value {
		return -1
	}
	////获取花色 去掉牌大小
	//if poker1Color > poker2Color {
	//	return 1
	//} else if poker1Color < poker2Color {
	//	return -1
	//}
	return 0
}

//获取获胜区域
//lotteryPoker 开奖扑克
//areaJetton 区域下注筹码
func (logic *Logic) GetWinArea(lotteryPoker []int32, areaJetton sync.Map) []bool {
	var winArea = make([]bool, global.AreaCount)
	switch logic.compare(lotteryPoker[0], lotteryPoker[1]) {
	case 1: //龙
		winArea[0] = true
		//判断庄输赢
		//押庄赢：开奖结果为龙或者虎,且龙或虎中获胜一方下注金额小于失败一方
		//押庄输：开奖结果为龙或者虎,且龙或虎中获胜一方下注金额大于失败一方
		var winJetton, lossJetton float32
		areaJetton.Range(func(key, value interface{}) bool {
			winJetton += value.([global.AreaCount]float32)[0]
			lossJetton += value.([global.AreaCount]float32)[3]
			return true
		})
		if winJetton < lossJetton {
			winArea[1] = true //押庄赢
		}
		if winJetton > lossJetton {
			winArea[2] = true //押庄输
		}
	case -1: //虎
		winArea[3] = true
		//判断庄输赢
		//押庄赢：开奖结果为龙或者虎,且龙或虎中获胜一方下注金额小于失败一方
		//押庄输：开奖结果为龙或者虎,且龙或虎中获胜一方下注金额大于失败一方
		var winJetton, lossJetton float32
		//for _, v := range areaJetton {
		//	winJetton += v[3]
		//	lossJetton += v[0]
		//}
		areaJetton.Range(func(key, value interface{}) bool {
			winJetton += value.([global.AreaCount]float32)[3]
			lossJetton += value.([global.AreaCount]float32)[0]
			return true
		})
		if winJetton < lossJetton {
			winArea[1] = true //押庄赢
		}
		if winJetton > lossJetton {
			winArea[2] = true //押庄输
		}
	case 0: // 和
		winArea[8] = true
	}

	switch publicLogic.GetColor(lotteryPoker[0]) {
	case 0:
		winArea[7] = true
	case 1:
		winArea[6] = true
	case 2:
		winArea[5] = true
	case 3:
		winArea[4] = true
	}

	switch publicLogic.GetColor(lotteryPoker[1]) {
	case 0:
		winArea[9] = true
	case 1:
		winArea[10] = true
	case 2:
		winArea[11] = true
	case 3:
		winArea[12] = true
	}

	return winArea
}

//返回系统损耗
func (*Logic) GetSystemLoss(winArea []bool, userListAreaJetton sync.Map, userList sync.Map) (float32, map[int32]float32, map[int32]float32, float32) {

	userListLoss := make(map[int32]float32)
	tempUserListLoss := make(map[int32]float32)
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
			userListLoss[k.(int32)] -= v.([global.AreaCount]float32)[i]
			tempUserListLoss[k.(int32)] -= v.([global.AreaCount]float32)[i]
			if !userItem.(*user.Item).IsRobot() {
				userNoRootTax += userListLoss[k.(int32)]
			}
			if winArea[i] {
				userWins := v.([global.AreaCount]float32)[i] * global.AreaMultiple[i]

				if !userItem.(*user.Item).IsRobot() {
					userNoRootTax += userWins
				}
				// 记录税收
				userTax[k.(int32)] += userWins * store.GameControl.GetGameInfo().RevenueRatio
				//判断是否>0 赢了 要扣税
				userListLoss[k.(int32)] += userWins - userTax[k.(int32)]

				tempUserListLoss[k.(int32)] += userWins
			}

		}
		//机器人批次号为 > 0
		if userItem.(*user.Item).BatchID != -1 {
			return true
		}
		systemScore -= tempUserListLoss[k.(int32)]
		return true
	})

	return systemScore, userListLoss, userTax, userNoRootTax
}
