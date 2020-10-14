package test

import (
	"fmt"
	"math"
	"sync"
	"xj_game_server/game/207_qiangzhuangniuniu_kansanzhang/global"
	"xj_game_server/game/207_qiangzhuangniuniu_kansanzhang/msg"
	publicLogic "xj_game_server/game/public/logic"
	"xj_game_server/game/public/user"
)

// 比牌
func compare(poker1 *msg.Game_S_LotteryPoker, poker2 *msg.Game_S_LotteryPoker) float32 {

	//return logic.compareType(logic.getPokerType(poker1), logic.getPokerType(poker2))
	return compareType(poker1.PokerType, poker2.PokerType)

}

func getPokerType(lotteryPoker []int32) int32 {

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

		pokerType = 0xd00 + cardMax

		return pokerType
	}

	if friedNum >= 4 || friedNum1 >= 4 {
		pokerType = 0xc00 + cardMax

		return pokerType
	}

	if bigCowNum == 5 {
		pokerType = 0xb00 + cardMax

		return pokerType
	}

	if bigCowNum == 3 || bigCowNum == 4 {

		if num == 0 {
			pokerType = 0xa00 + cardMax
			return pokerType
		} else if num == 9 {
			pokerType = 0x900 + cardMax
			return pokerType
		} else if num == 8 {
			pokerType = 0x800 + cardMax
			return pokerType
		} else if num == 7 {
			pokerType = 0x700 + cardMax
			return pokerType
		} else {
			pokerType = (0x100 * num) + cardMax
			return pokerType
		}
	}

	judge := judge(num, lotteryPoker)

	if !judge {
		return cardMax
	} else {
		if num == 0 {
			pokerType = 0xa00 + cardMax
			return pokerType
		} else {
			pokerType = (0x100 * num) + cardMax
			return pokerType
		}

	}
}

// 求合 算牛
func judge(sum int32, lotteryPoker []int32) bool {

	if len(lotteryPoker) == 0 {
		return false
	}

	for i := 1; i < len(lotteryPoker); i++ {

		if publicLogic.GetCombined(lotteryPoker[0], lotteryPoker[i]) == sum {
			return true
		}

	}

	return judge(sum, lotteryPoker[1:])

}

//比牌 反赔率
func compareType(pokerType1 int32, pokerType2 int32) float32 {

	var area0, area1, area2, area3, area4, area5 int32

	area0 = pokerType1 & 0xf00 / 0x100
	area1 = pokerType2 & 0xf00 / 0x100
	area2 = pokerType1 & 0xf0 / 0x10
	area3 = pokerType2 & 0xf0 / 0x10
	area4 = pokerType1 & 0xf
	area5 = pokerType2 & 0xf

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

// 获取系统损耗与用户输赢
func GetSystemLoss(bankerChairID int32, bankerMultiple int32, userListJetton sync.Map, userListPoker map[int32]*msg.Game_S_LotteryPoker, userList sync.Map) (float32, map[int32]float32, map[int32]float32, float32) {
	//用户输赢
	userListLoss := make(map[int32]float32)
	//税收
	userTax := make(map[int32]float32)
	//系统输赢
	var systemScore float32

	//个人奖池
	var jackpots float32

	//庄家输赢
	var bankerScore float32
	//庄家金额
	var lowGrade float32
	//庄家是否机器人
	var isRobot bool

	uid, ok := userList.Load(bankerChairID)

	if ok {
		userItem, ok := user.List.Load(uid.(int32))
		if ok {
			isRobot = userItem.(*user.Item).BatchID != -1

			lowGrade = userItem.(*user.Item).UserDiamond

		}
	}
	userListPoker[bankerChairID].PokerType = getPokerType(userListPoker[bankerChairID].LotteryPoker)
	//庄家牌
	poker1 := userListPoker[bankerChairID]

	userListJetton.Range(func(chairID, multiple interface{}) bool {
		//跳过庄家下注
		if chairID == bankerChairID {
			return true
		}

		uid, ok := userList.Load(chairID)

		if ok {
			value, ok := user.List.Load(uid.(int32))
			if ok {

				userItem := value.(*user.Item)
				//用户牌

				userListPoker[chairID.(int32)].PokerType = getPokerType(userListPoker[chairID.(int32)].LotteryPoker)
				poker2 := userListPoker[chairID.(int32)]
				//用户是否输赢 返回赔率
				wins := compare(poker1, poker2)

				var score float32
				//庄家倍数 * 自己倍数 * 赔率 * 底注
				win := float32(bankerMultiple*multiple.(int32)) * wins * 10

				score = userItem.UserDiamond

				if float32(math.Abs(float64(win))) > score {

					if win > 0 {
						win = score
					} else {
						win = -score
					}

				}

				userListLoss[chairID.(int32)] = win

				bankerScore -= win

				//	机器人
				if userItem.BatchID != -1 {

					if isRobot {
						return true
					} else {
						systemScore += win
					}

				} else { //普通用户
					jackpots += win + userItem.Jackpot
					if isRobot {
						systemScore -= win
					} else {
						return true
					}
				}
			}
		}

		return true
	})

	userListLoss[bankerChairID] = bankerScore

	ratio := float32(1)

	// 用户金额上线
	if float32(math.Abs(float64(bankerScore))) > lowGrade {
		ratio = lowGrade / float32(math.Abs(float64(bankerScore)))
	}

	for k, win := range userListLoss {
		var revenueRatio float32
		win = ratio * win
		if win > 0 {
			revenueRatio = win * 0.1
		}
		userTax[k] = revenueRatio
		userListLoss[k] = win - revenueRatio

		systemScore = systemScore * ratio
	}
	fmt.Println(systemScore, userListLoss, userTax, jackpots)
	return systemScore, userListLoss, userTax, jackpots

}
