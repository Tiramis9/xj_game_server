package logic

import (
	"math"
	"sort"
	"sync"
	"xj_game_server/game/201_qiangzhuangniuniu/global"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
	publicLogic "xj_game_server/game/public/logic"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/log"
)

var Client = new(Logic)

type Logic struct {
}

//发牌
func (logic *Logic) DispatchTableCard(userPlaying map[int32]bool) map[int32]*msg.Game_S_LotteryPoker {
	//0x01:第一位代表花色 第二位代表牌大小
	pokers := publicLogic.Poker{
		Count:        1,     //1副牌
		IsNeedGhost:  false, //不需要大小鬼
		ShuffleCount: 100,   //洗牌次数
	}.GetPokers()
	//赋值
	userListPoker := make(map[int32]*msg.Game_S_LotteryPoker)

	var i int
	for k, v := range userPlaying {

		if v {

			var d msg.Game_S_LotteryPoker
			d.LotteryPoker = pokers[i*global.PokerCount : (i+1)*global.PokerCount]
			d.PokerType = logic.getPokerType(d.LotteryPoker)

			userListPoker[k] = &d
			i++
		}

	}

	return userListPoker
}

// 比牌
func (logic *Logic) compare(poker1 *msg.Game_S_LotteryPoker, poker2 *msg.Game_S_LotteryPoker) float32 {

	//return logic.compareType(logic.getPokerType(poker1), logic.getPokerType(poker2))
	return logic.compareType(poker1.PokerType, poker2.PokerType)

}

type SystemLoss struct {
	UserListPoker map[int32]*msg.Game_S_LotteryPoker
	UserListLoss  map[int32]float32
	UserTax       map[int32]float32
	SystemScore   float32
	Jackpots      float32
}

type SystemLosss []SystemLoss

func (s SystemLosss) Len() int           { return len(s) }
func (s SystemLosss) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SystemLosss) Less(i, j int) bool { return s[i].Jackpots < s[j].Jackpots }

func (logic *Logic) GetUserSystemLoss(bankerChairID int32, bankerMultiple int32, userListJetton sync.Map, userListPoker map[int32]*msg.Game_S_LotteryPoker, userList sync.Map) (float32, map[int32]float32, map[int32]float32, map[int32]*msg.Game_S_LotteryPoker) {
	var systemLosss = make([]SystemLoss, 0)
	var si int32

	log.Logger.Errorf("GetUserSystemLoss userListPoker.len=%d", len(userListPoker))

	for i := range userListPoker {
		if si == 0 {
			si = i
			continue
		}
		var userListPoker1 = make(map[int32]*msg.Game_S_LotteryPoker)

		for j, p := range userListPoker {
			var poker = new(msg.Game_S_LotteryPoker)
			poker.LotteryPoker = p.LotteryPoker
			poker.PokerType = p.PokerType
			userListPoker1[j] = poker
		}

		userListPoker1[i], userListPoker1[si] = userListPoker1[si], userListPoker1[i]

		var systemLoss SystemLoss

		loss, m, m2, j := logic.GetSystemLoss(bankerChairID, bankerMultiple, userListJetton, userListPoker1, userList)
		systemLoss.UserListPoker = userListPoker1
		systemLoss.SystemScore = loss
		systemLoss.UserListLoss = m
		systemLoss.UserTax = m2
		systemLoss.Jackpots = float32(math.Abs(float64(j)))

		systemLosss = append(systemLosss, systemLoss)
	}
	sort.Sort(SystemLosss(systemLosss))

	//用户输赢
	userListLoss := make(map[int32]float32)
	//税收
	userTax := make(map[int32]float32)
	//系统输赢
	var systemScore float32
	userPoker := make(map[int32]*msg.Game_S_LotteryPoker)

	if len(systemLosss) > 0 {
		userListLoss = systemLosss[0].UserListLoss
		userTax = systemLosss[0].UserTax
		systemScore = systemLosss[0].SystemScore
		userPoker = systemLosss[0].UserListPoker
	}

	return systemScore, userListLoss, userTax, userPoker

}

// 获取系统损耗与用户输赢
func (logic *Logic) GetSystemLoss(bankerChairID int32, bankerMultiple int32, userListJetton sync.Map, userListPoker map[int32]*msg.Game_S_LotteryPoker, userList sync.Map) (float32, map[int32]float32, map[int32]float32, float32) {
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

			if store.GameControl.GetGameInfo().DeductionsType == 0 {
				lowGrade = userItem.(*user.Item).UserGold
			} else {
				lowGrade = userItem.(*user.Item).UserDiamond
			}

		}
	}
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
				poker2 := userListPoker[chairID.(int32)]
				//用户是否输赢 返回赔率
				wins := logic.compare(poker1, poker2)

				var score float32
				//庄家倍数 * 自己倍数 * 赔率 * 底注
				win := float32(bankerMultiple*multiple.(int32)) * wins * store.GameControl.GetGameInfo().CellScore

				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					score = userItem.UserGold
				} else {
					score = userItem.UserDiamond
				}

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
			revenueRatio = win * store.GameControl.GetGameInfo().RevenueRatio
		}
		userTax[k] = revenueRatio
		userListLoss[k] = win - revenueRatio

		systemScore = systemScore * ratio
	}

	return systemScore, userListLoss, userTax, jackpots

}

// 没牛 0
// 牛1 2 3 4 5 6
// 牛7 8
// 牛9
// 牛牛 a 10
// 五花牛 b 11
// 炸弹 c 12
// 五小牛 d 13
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

	judge := logic.judge(num, lotteryPoker)

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

//比牌 反赔率
func (logic *Logic) compareType(pokerType1 int32, pokerType2 int32) float32 {

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
