package logic

import (
	"reflect"
	"sync"
	"xj_game_server/game/203_doudizhu/global"
	publicLogic "xj_game_server/game/public/logic"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
)

var Client = new(Logic)

type Logic struct {
}

//发牌[global.PokerCount]uint8
func (*Logic) DispatchTableCard() map[int32][]int32 {

	var card = make(map[int32][]int32, 0)

	//0x01:第一位代表花色 第二位代表牌大小
	pokers := publicLogic.Poker{
		Count:        1,    //3副牌
		IsNeedGhost:  true, //不需要大小鬼
		ShuffleCount: 10,   //洗牌次数
	}.GetPokers()

	card[0] = pokers[:17]
	card[1] = pokers[17:34]
	card[2] = pokers[34:51]
	card[3] = pokers[51:]

	return card
}

func (*Logic) GetSystemLoss(dizChairID int32, userList sync.Map, currentMultiple int32, isDizWins bool) (float32, map[int32]float32, map[int32]float32) {
	userListLoss := make(map[int32]float32, 0)
	userTax := make(map[int32]float32, 0)
	var systemScore float32

	var sumMultiple = currentMultiple
	if isDizWins {
		dizScore := (store.GameControl.GetGameInfo().CellScore * 2) * float32(sumMultiple)
		userListLoss[dizChairID] = dizScore
		tempScore := float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore * 2
		var winScore float32 = -1
		userID, ok := userList.Load(dizChairID)
		if ok {
			userItem, ok := user.List.Load(userID)
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					if userItem.(*user.Item).UserGold-userListLoss[dizChairID] <= 0 {
						winScore = userItem.(*user.Item).UserGold
					}
				} else {
					if userItem.(*user.Item).UserDiamond-userListLoss[dizChairID] <= 0 {
						winScore = userItem.(*user.Item).UserDiamond
					}
				}
			}
		}
		userList.Range(func(key, value interface{}) bool {
			if key.(int32) == dizChairID {
				return true
			}
			userListLoss[key.(int32)] = -(float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore)
			userItem, ok := user.List.Load(value)
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					if userItem.(*user.Item).UserGold+userListLoss[key.(int32)] <= 0 {
						tempScore += userItem.(*user.Item).UserGold + userListLoss[key.(int32)]
						userListLoss[key.(int32)] = -userItem.(*user.Item).UserGold
					} else if winScore > 0 && (userItem.(*user.Item).UserGold+userListLoss[key.(int32)] >= 0 || winScore/2 >= userItem.(*user.Item).UserGold) {
						tempScore += (winScore / 2) + userListLoss[key.(int32)]
						userListLoss[key.(int32)] = -winScore / 2
					}
				} else {
					if userItem.(*user.Item).UserDiamond+userListLoss[key.(int32)] <= 0 {
						tempScore += userItem.(*user.Item).UserDiamond + userListLoss[key.(int32)]
						userListLoss[key.(int32)] = -userItem.(*user.Item).UserDiamond
					} else if winScore > 0 && (userItem.(*user.Item).UserDiamond > userListLoss[key.(int32)] || userItem.(*user.Item).UserDiamond >= winScore/2) {
						tempScore += winScore/2 + userListLoss[key.(int32)]
						userListLoss[key.(int32)] = -(winScore / 2)
					}
				}
			}
			return true
		})
		userListLoss[dizChairID] = tempScore

	} else {
		dizScore := float32(sumMultiple) * (store.GameControl.GetGameInfo().CellScore * 2)
		userListLoss[dizChairID] = -dizScore

		userID, ok := userList.Load(dizChairID)
		if ok {
			userItem, ok := user.List.Load(userID)
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {
					if userItem.(*user.Item).UserGold+userListLoss[dizChairID] <= 0 {
						userListLoss[dizChairID] = userItem.(*user.Item).UserGold
					}
				} else {
					if userItem.(*user.Item).UserDiamond+userListLoss[dizChairID] <= 0 {
						userListLoss[dizChairID] = userItem.(*user.Item).UserDiamond
					}
				}
			}
		}
		var tempScore = float32(sumMultiple) * (store.GameControl.GetGameInfo().CellScore * 2)
		userList.Range(func(key, value interface{}) bool {
			if key.(int32) == dizChairID {
				return true
			}
			userListLoss[key.(int32)] = float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore
			userItem, ok := user.List.Load(value)
			if ok {
				if store.GameControl.GetGameInfo().DeductionsType == 0 {

					if userListLoss[dizChairID] > 0 && (userItem.(*user.Item).UserGold > userListLoss[key.(int32)] || userListLoss[dizChairID]/2 >= userItem.(*user.Item).UserGold) {
						tempScore += userListLoss[dizChairID]/2 - userListLoss[key.(int32)]
						userListLoss[key.(int32)] = userListLoss[dizChairID] / 2
					} else if userItem.(*user.Item).UserGold <= userListLoss[key.(int32)] {
						tempScore += userItem.(*user.Item).UserGold - userListLoss[key.(int32)]
						userListLoss[key.(int32)] = userItem.(*user.Item).UserGold
					}
				} else {
					if userListLoss[dizChairID] > 0 && (userItem.(*user.Item).UserDiamond > userListLoss[key.(int32)] || userItem.(*user.Item).UserDiamond >= userListLoss[dizChairID]/2) {
						tempScore += userListLoss[dizChairID]/2 - userListLoss[key.(int32)]
						userListLoss[key.(int32)] = userListLoss[dizChairID] / 2
					} else if userItem.(*user.Item).UserDiamond <= userListLoss[key.(int32)] {
						tempScore += userItem.(*user.Item).UserDiamond - userListLoss[key.(int32)]
						userListLoss[key.(int32)] = userItem.(*user.Item).UserDiamond
					}
				}
			}
			return true
		})
		userListLoss[dizChairID] = -tempScore
	}

	for chairID, score := range userListLoss {
		uid, ok := userList.Load(chairID)
		if ok {
			userItem, ok := user.List.Load(uid.(int32))
			if ok {
				//	机器人
				if userItem.(*user.Item).BatchID != -1 {
					systemScore += score
				}
			}
		}

		// 记录税收
		if score > 0 {
			userTax[chairID] += score * store.GameControl.GetGameInfo().RevenueRatio
			userListLoss[chairID] = userListLoss[chairID] - userTax[chairID]
		}
	}

	return systemScore, userListLoss, userTax
}

func (*Logic) GetSystemLoss_bak(dizChairID int32, userList sync.Map, currentMultiple int32, isDizWins bool) (float32, map[int32]float32, map[int32]float32) {
	userListLoss := make(map[int32]float32)
	userTax := make(map[int32]float32)
	var systemScore float32

	var sumMultiple = currentMultiple
	if isDizWins {
		userListLoss[dizChairID] = float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore * 2
		userList.Range(func(key, value interface{}) bool {
			if key.(int32) == dizChairID {
				return true
			}

			userListLoss[key.(int32)] = -float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore
			return true
		})
	} else {
		userListLoss[dizChairID] = -float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore * 2
		userList.Range(func(key, value interface{}) bool {
			if key.(int32) == dizChairID {
				return true
			}

			userListLoss[key.(int32)] = float32(sumMultiple) * store.GameControl.GetGameInfo().CellScore
			return true
		})
	}

	for chairID, score := range userListLoss {
		uid, ok := userList.Load(chairID)
		if ok {
			userItem, ok := user.List.Load(uid.(int32))
			if ok {
				//	机器人
				if userItem.(*user.Item).BatchID != -1 {
					systemScore += score
				}
			}
		}

		// 记录税收
		if score > 0 {
			userTax[chairID] += score * store.GameControl.GetGameInfo().RevenueRatio
			userListLoss[chairID] = userListLoss[chairID] - userTax[chairID]
		}
	}

	return systemScore, userListLoss, userTax
}

//单牌	1
//对子	2
//顺子	3
//连对	4
//三带	5
//三带一	6
//三带对	7
//四带两单8
//四带两对9
//炸弹10
//王炸11
//类型  牌数  最大牌
func (l *Logic) GetPokerType(pokers []int32) (int32, int32) {
	switch len(pokers) {
	case 0: //错误
		return 0, 0
	case 1: //单张
		return 1, pokers[0]
	case 2: //对子或王炸
		value := publicLogic.GetValue(pokers[0])
		if value != publicLogic.GetValue(pokers[1]) {
			return 0, 0
		}

		//王炸
		if value == 0x0F {
			return 11, 0x5F
		}

		//对子
		return 2, pokers[0]
	}

	//获取数量对应的牌
	moldNum := l.conversionPokerType(pokers)

	//四张判断
	if len(moldNum[4]) == 1 && len(moldNum[3]) == 0 {
		if len(pokers) == 4 {
			//炸弹(四张)
			return 10, moldNum[4][0]
		} else if len(pokers) == 6 {
			//四带二(单)
			return 8, moldNum[4][0]
		} else if len(moldNum[2]) == 2 && len(pokers) == 8 {
			//四带二(双)
			return 9, moldNum[4][0]
		}

		return 0, 0
	}

	//三张判断
	if len(moldNum[3]) > 0 {
		if len(moldNum[4]) > 0 {
			moldNum[3] = append(moldNum[3], moldNum[4]...)
			for i := 0; i < len(moldNum[3])-1; i++ {
				for j := i + 1; j < len(moldNum[3]); j++ {
					if l.GetLogicValue(moldNum[3][i]) < l.GetLogicValue(moldNum[3][j]) {
						moldNum[3][i], moldNum[3][j] = moldNum[3][j], moldNum[3][i]
					}
				}
			}
		}
		for i := 0; i < len(moldNum[3])-1; i++ {
			if l.GetLogicValue(moldNum[3][i])-1 != l.GetLogicValue(moldNum[3][i+1]) || l.GetLogicValue(moldNum[3][i]) == 0x0F {
				return 0, 0
			}
		}

		if len(pokers) == len(moldNum[3])*3 {
			//三不带
			return 5, moldNum[3][0]
		} else if len(pokers) == len(moldNum[3])*3+len(moldNum[3])*1 {
			//三带单
			return 6, moldNum[3][0]
		} else if len(pokers) == len(moldNum[3])*3+len(moldNum[2])*2 && len(moldNum[3]) == len(moldNum[2]) {
			//三带双
			return 7, moldNum[3][0]
		}

		return 0, 0
	}

	//两张判断
	if len(moldNum[2]) >= 3 && len(pokers) == len(moldNum[2])*2 {
		for i := 0; i < len(moldNum[2])-1; i++ {
			if publicLogic.GetValue(moldNum[2][i]) == 0x02 {
				return 0, 0
			}

			if l.GetLogicValue(moldNum[2][i])-1 != l.GetLogicValue(moldNum[2][i+1]) {
				return 0, 0
			}
		}

		//连对
		return 4, moldNum[2][0]
	}

	//单张判断
	if len(moldNum[1]) >= 5 && len(pokers) == len(moldNum[1]) {
		for i := 0; i < len(moldNum[1])-1; i++ {
			if publicLogic.GetValue(moldNum[1][i]) == 0x02 {
				return 0, 0
			}

			if l.GetLogicValue(moldNum[1][i])-1 != l.GetLogicValue(moldNum[1][i+1]) {
				return 0, 0
			}
		}

		//顺子
		return 3, moldNum[1][0]
	}

	return 0, 0
}

// 分值	地主叫分
//≧10	3
//≧7	2
//≧ 4	1
//<4	不叫

//火箭	6（双王算火箭）
//炸弹	4
//大王	3
//小王	2
//2	1
func (l *Logic) GetJFMultiple(card []int32) int32 {
	var multiple int32
	for _, v := range card {
		if v == 0x4F {
			multiple += 2
		}
		if v == 0x5F {
			multiple += 3
		}
	}
	if multiple >= 5 {
		multiple++
	}
	mold := l.conversionPokerType(card)
	for k, v := range mold {
		if k == 4 {
			multiple += int32(len(v)) * 4
			continue
		}
		for _, val := range v {
			if l.GetLogicValue(val) == l.GetLogicValue(2) {
				multiple += k
				continue
			}
		}
	}
	if multiple >= 10 {
		multiple = 3
	} else if multiple >= 7 {
		multiple = 2
	} else if multiple >= 4 {
		multiple = 1
	} else {
		multiple = 0
	}

	return multiple
}

//获取逻辑值
func (l *Logic) GetLogicValue(poker int32) int32 {
	switch publicLogic.GetValue(poker) {
	case 0x01:
		return 0x0E
	case 0x02:
		return 0x0F
	case 0x0F:
		if publicLogic.GetColor(poker) == 4 {
			return 0x10
		} else {
			return 0x11
		}
	}
	return publicLogic.GetValue(poker)
}

func (l *Logic) conversionPokerType(pokers []int32) map[int32][]int32 {
	var mold = make(map[int32]int32, 0) //牌 数量

	for _, card := range pokers {
		mold[card&0x0F]++
	}

	var moldNum = make(map[int32][]int32, 0) //数量 牌

	for card, num := range mold {
		var cards = moldNum[num]
		moldNum[num] = append(cards, card)
	}

	//排序扑克
	for k, v := range moldNum {
		for i := 0; i < len(v)-1; i++ {
			for j := i + 1; j < len(v); j++ {
				if l.GetLogicValue(v[i]) < l.GetLogicValue(v[j]) {
					moldNum[k][i], moldNum[k][j] = moldNum[k][j], moldNum[k][i]
				}
			}
		}
	}

	return moldNum
}

// 托管首次出牌
func (l *Logic) GetAutoOutPokers(minPoker int32, pokers []int32) []int32 {
	var outPokers, tempPoker []int32
	var cardsNum, pokerNum int32
	pokersMold := l.conversionPokerType(pokers)
	for k, v := range pokersMold {
		for i := 0; i < len(v)-1; i++ {
			for j := i + 1; j < len(v); j++ {
				if l.GetLogicValue(v[i]) > l.GetLogicValue(v[j]) {
					pokersMold[k][i], pokersMold[k][j] = pokersMold[k][j], pokersMold[k][i]
				}
			}
		}
	}
	for num, cards := range pokersMold {
		for _, card := range cards {
			if card == 0x0F {
				pokerNum = num
			}
			if l.GetLogicValue(minPoker) == l.GetLogicValue(card) {
				tempPoker = append(tempPoker, card)
				cardsNum = num
				break
			}
		}
	}
	if pokerNum == int32(len(pokers)) {
		return pokers
	}

	if cardsNum == 3 {
		if len(pokersMold[1]) > 0 {
			tempPoker = append(tempPoker, pokersMold[1][0])
		} else if len(pokersMold[2]) > 0 && pokersMold[2][0] != 0x0F {
			tempPoker = append(tempPoker, pokersMold[2][0])
		}
	}
	if cardsNum == 4 {
		if len(pokersMold[1]) >= 2 {
			tempPoker = append(tempPoker, pokersMold[1][0], pokersMold[1][1])
		} else if len(pokersMold[2]) > 0 && pokersMold[2][0] != 0x0F {
			tempPoker = append(tempPoker, pokersMold[2][0])
		}
	}
	for _, card := range tempPoker {
		for _, poker := range pokers {
			if l.GetLogicValue(card) == l.GetLogicValue(poker) {
				outPokers = append(outPokers, poker)
			}
		}
	}
	return outPokers
}

// 托管接牌
func (l *Logic) GetAutoOutPokersByType(cardType int32, nearestPokers []int32, pokers []int32) []int32 {
	handCards := l.conversionPokerType(pokers)
	l.sort(nearestPokers)
	for k, v := range handCards {
		for i := 0; i < len(v)-1; i++ {
			for j := i + 1; j < len(v); j++ {
				if l.GetLogicValue(v[i]) > l.GetLogicValue(v[j]) {
					handCards[k][i], handCards[k][j] = handCards[k][j], handCards[k][i]
				}
			}
		}
	}
	getOriginData := func(card []int32, isSingle bool) []int32 {
		pokerVal := make([]int32, 0)
		size := len(pokers)
		for j := 0; j < len(card); j++ {
			for i := 0; i < size; i++ {
				if l.GetLogicValue(pokers[i]) == l.GetLogicValue(card[j]) {
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

	var outCard []int32
	var tempCards = make([]int32, 0)
	nearestCards := l.conversionPokerType(nearestPokers)
	switch cardType {
	case global.CardTypeSINGLE:
		tempCards = handCards[1]
		for i := range tempCards {
			if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestPokers[0]) {
				outCard = []int32{tempCards[i]}
				break
			}
		}
		if len(outCard) == 0 {
			var king int32
			for _, card := range handCards[1] {
				if card == 0x0F {
					king = card
				}
			}
			if king > 0 {
				outCard = getOriginData([]int32{0x4F, 0x5F}, false)
				if l.GetLogicValue(outCard[0]) < l.GetLogicValue(nearestPokers[0]) {
					outCard = make([]int32, 0)
				}
			}
		} else if outCard[0] == 0x0F {
			outCard = getOriginData([]int32{0x4F, 0x5F}, false)
			if l.GetLogicValue(outCard[0]) < l.GetLogicValue(nearestPokers[0]) {
				outCard = make([]int32, 0)
			}
		} else {
			outCard = getOriginData(outCard, true)
		}
	case global.CardTypeDOUBLE:
		tempCards = handCards[2]
		for i := range tempCards {
			if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestPokers[0]) && tempCards[i] != 0x0F {
				outCard = []int32{tempCards[i]}
				break
			}
		}
		outCard = getOriginData(outCard, false)
	case global.CardTypeSINGLEALONE:
		tempCards = l.getSingleContinuity(handCards, true)
		l.sort(tempCards)
		if len(tempCards) >= len(nearestPokers) {
			var index int
			for i := range tempCards {
				if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestPokers[0]) {
					index = i
					break
				}
			}
			if len(tempCards[index:]) >= len(nearestPokers) && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestPokers[0]) {
				n := len(tempCards[index:]) - len(nearestPokers)
				outCard = tempCards[index : len(tempCards)-n]
			}
		}
		outCard = getOriginData(outCard, true)
	case global.CardTypeDOUBLEALONE:
		tempCards = l.getDoubleContinuity(handCards[2])
		l.sort(tempCards)
		if len(tempCards) > 0 {
			var index int
			for i := range tempCards {
				if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestPokers[0]) {
					index = i
					break
				}
			}
			if len(tempCards[index:]) >= len(nearestCards[2]) && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestPokers[0]) {
				n := len(tempCards[index:]) - len(nearestCards[2])
				outCard = tempCards[index : len(tempCards)-n]
			}
		}
		outCard = getOriginData(outCard, false)
	case global.CardTypeTHREE:
		if len(nearestCards[3]) > 1 {
			outCard = l.getPlanes(cardType, nearestCards[3], handCards)
		} else {
			tempCards = handCards[3]
			if len(tempCards) > 0 {
				var index int
				for i := range tempCards {
					if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestCards[3][0]) {
						index = i
						break
					}
				}
				if len(tempCards[index:]) > 0 && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestCards[3][0]) {
					outCard = []int32{tempCards[index]}
				}
			}
		}
		outCard = getOriginData(outCard, false)
	case global.CardTypeTHREEONE:
		if len(nearestCards[3]) > 1 {
			outCard = l.getPlanes(cardType, nearestCards[3], handCards)
			outCard = getOriginData(outCard, false)
		} else {
			tempCards = handCards[3]
			if len(tempCards) > 0 {
				var index int
				for i := range tempCards {
					if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestCards[3][0]) {
						index = i
						break
					}
				}
				if len(tempCards[index:]) > 0 && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestCards[3][0]) {
					outCard = []int32{tempCards[index]}
				}
			}
			if len(outCard) > 0 {
				if len(handCards[1]) >= 1 {
					tempCards = []int32{handCards[1][0]}
					if handCards[1][0] == 0x0F && len(pokers) == len(nearestPokers) {
						tempCards = []int32{0x4F, 0x5F}
					}
				} else if len(handCards[2]) > 0 {
					tempCards = []int32{handCards[2][0]}
				}
				if len(tempCards) > 0 {
					tempCards = getOriginData(tempCards, true)
					outCard = getOriginData(outCard, false)
					outCard = append(outCard, tempCards...)
				}
			}
		}
	case global.CardTypeTHREETWO:
		if len(nearestCards[3]) > 1 {
			outCard = l.getPlanes(cardType, nearestCards[3], handCards)
			outCard = getOriginData(outCard, false)
		} else {
			tempCards = handCards[3]
			if len(tempCards) > 0 {
				var index int
				for i := range tempCards {
					if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestCards[3][0]) {
						index = i
						break
					}
				}
				if len(tempCards[index:]) > 0 && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestCards[3][0]) {
					outCard = []int32{tempCards[index]}
				}
			}
			if len(outCard) > 0 && len(handCards[2]) > 0 {
				outCard = append(outCard, handCards[2][0])
				outCard = getOriginData(outCard, false)
			}
		}
	case global.CardTypeFOURONE:
		tempCards = handCards[4]
		if len(tempCards) > 0 {
			var index int
			for i := range tempCards {
				if l.GetLogicValue(tempCards[i]) > l.GetLogicValue(nearestCards[4][0]) {
					index = i
					break
				}
			}
			if len(tempCards[index:]) > 0 && l.GetLogicValue(tempCards[index]) > l.GetLogicValue(nearestCards[4][0]) {
				outCard = []int32{tempCards[index]}
			}
		}
		if len(outCard) > 0 {
			if len(handCards[1]) >= 2 {
				outCard = append(outCard, handCards[1][0], handCards[1][1])
			} else if len(handCards[2]) > 1 {
				outCard = append(outCard, handCards[2][0])
			}
		}
		outCard = getOriginData(outCard, false)
	case global.CardTypeFOURTWO:
		outCard = handCards[4]
		if len(outCard) > 0 {
			var index int
			for i := range outCard {
				if l.GetLogicValue(outCard[i]) > l.GetLogicValue(nearestCards[4][0]) {
					index = i
					break
				}
			}
			if len(outCard[index:]) == 0 || l.GetLogicValue(outCard[0]) < l.GetLogicValue(nearestCards[4][0]) {
				outCard = make([]int32, 0)
			} else if len(outCard[index:]) > 0 {
				outCard = []int32{outCard[index]}
			}
		}
		if len(outCard) > 0 {
			if len(handCards[1]) >= 2 {
				tempCards = append(tempCards, handCards[1][0], handCards[1][1])
			}
			if len(tempCards) > 0 {
				outCard = append(outCard, tempCards...)
				outCard = getOriginData(outCard, false)
			}
		}
	case global.CardTypeBOMB:
		outCard = handCards[4]
		var index int
		for i := range outCard {
			if l.GetLogicValue(outCard[i]) > l.GetLogicValue(nearestCards[4][0]) {
				index = i
				break
			}
		}
		if len(outCard[index:]) == 0 || l.GetLogicValue(outCard[0]) < l.GetLogicValue(nearestCards[4][0]) {
			outCard = make([]int32, 0)
		} else if len(outCard[index:]) > 0 {
			outCard = []int32{outCard[index]}
		}
		if len(outCard) > 0 {
			outCard = getOriginData(outCard, false)
		}
	case global.CardTypeKINGBOMB:
		for _, card := range handCards[2] {
			if card == 15 {
				tempCards = append(tempCards, card)
			}
			if len(tempCards) == 2 {
				outCard = getOriginData([]int32{0x4F, 0x5F}, false)
			}
		}
	}
	if len(nearestPokers) != len(outCard) {
		outCard = make([]int32, 0)
	}
	return outCard
}

// 获取单顺子,true 拆炸弹，false 不拆
func (l *Logic) getSingleContinuity(poker map[int32][]int32, isBomb bool) []int32 {
	// 不拆炸弹
	if !isBomb {
		delete(poker, 4)
	}
	card := l.pokerMap2Sort(poker)
	var pokerSlice []int32
	var dataMap = make(map[int32]bool)
	var count int
	keys := card
	for i := 0; i < len(keys)-1; i++ {
		values := keys[i+1]
		if (keys[i] < values) && (keys[i] == values-1) {
			count++
			dataMap[keys[i]] = true
			dataMap[values] = true
		} else {
			if count < 4 {
				count = 0
				dataMap = make(map[int32]bool, 0)
			} else {
				break
			}
		}
	}
	if len(dataMap) >= 5 {
		for k, _ := range dataMap {
			pokerSlice = append(pokerSlice, k)
		}
	}
	return pokerSlice
}

func (l *Logic) sort(poker []int32) {
	for i := 0; i < len(poker)-1; i++ {
		for j := i + 1; j < len(poker); j++ {
			if l.GetLogicValue(poker[i]) > l.GetLogicValue(poker[j]) {
				poker[i], poker[j] = poker[j], poker[i]
			}
		}
	}
}

// 获取逻辑数，以便判断顺子
func (l *Logic) pokerMap2Sort(pokers map[int32][]int32) []int32 {
	var card []int32
	for k, v := range pokers {
		for i := range v {
			if pokers[k][i] == 0x01 {
				pokers[k][i] = 0x0E
			}
			if pokers[k][i] != 0x02 && pokers[k][i] != 0x0F {
				card = append(card, pokers[k][i])
			}
		}
	}
	for i := 0; i < len(card)-1; i++ {
		for j := i + 1; j < len(card); j++ {
			if card[i] > card[j] {
				card[i], card[j] = card[j], card[i]
			}
		}
	}
	return card
}

// 获取连对,三对
func (l *Logic) getDoubleContinuity(poker []int32) []int32 {
	l.sort(poker)
	var pokerSlice []int32
	var dataMap = make(map[int32]bool)
	var count int
	for i := 0; i < len(poker)-1; i++ {
		values := poker[i+1]
		if l.GetLogicValue(poker[i]) == 0x0F || l.GetLogicValue(poker[i]) == 0x10 || l.GetLogicValue(poker[i]) == 0x11 {
			continue
		}
		if (poker[i] < values) && (poker[i] == values-1) {
			count++
			dataMap[poker[i]] = true
			dataMap[values] = true
		} else {
			if count < 2 {
				count = 0
				dataMap = make(map[int32]bool, 0)
			} else {
				break
			}
		}
	}
	if len(dataMap) >= 3 {
		for k, _ := range dataMap {
			pokerSlice = append(pokerSlice, k)
		}
	}
	return pokerSlice
}

// 获取飞机
func (l *Logic) getPlanes(cardType int32, nearestCards []int32, handCards map[int32][]int32) []int32 {
	outCard := make([]int32, 0)
	ContinuityNum := make([]int32, 0)
	if len(handCards[3]) >= len(nearestCards) {
		for _, card := range handCards[3] {
			if l.GetLogicValue(card) > l.GetLogicValue(nearestCards[0]) && l.GetLogicValue(card) != 0x0F {
				outCard = append(outCard, card)
			}
		}
	}
	for i := 0; i < len(outCard)-1; i++ {
		for j := i; j < len(outCard); j++ {
			if l.GetLogicValue(outCard[i]) > l.GetLogicValue(outCard[j]) {
				outCard[i], outCard[j] = outCard[j], outCard[i]
			}
		}
	}

	for i := len(outCard) - 1; i > 0; i-- {
		if l.GetLogicValue(outCard[i-1]+1) == l.GetLogicValue(outCard[i]) {
			if len(ContinuityNum) > 0 {
				if ContinuityNum[len(ContinuityNum)-1] != outCard[i] {
					ContinuityNum = append(ContinuityNum, outCard[i])
				}
				if ContinuityNum[len(ContinuityNum)-1] != outCard[i-1] {
					ContinuityNum = append(ContinuityNum, outCard[i-1])
				}
			} else {
				ContinuityNum = append(ContinuityNum, outCard[i], outCard[i-1])
			}
		} else {
			if len(ContinuityNum) < 2 {
				ContinuityNum = make([]int32, 0)
			}
		}
	}

	if len(ContinuityNum) >= len(nearestCards) {
		n := len(ContinuityNum) - len(nearestCards)
		outCard = ContinuityNum[:len(ContinuityNum)-n]
	} else {
		outCard = make([]int32, 0)
	}
	if len(outCard) > 0 {
		switch cardType {
		case global.CardTypeTHREE: //三不带
		case global.CardTypeTHREEONE: //三带一
			if len(handCards[1]) >= len(outCard) {
				outCard = append(outCard, handCards[1][:len(outCard)]...)
			}

		case global.CardTypeTHREETWO: //三带对
			if len(handCards[2]) >= len(outCard) {
				outCard = append(outCard, handCards[2][:len(outCard)]...)
			}
		}
	}
	return outCard
}

// 获取炸弹
func (l *Logic) GetBomb(pokerType int32, poker []int32) []int32 {
	var outCards []int32
	handCards := l.conversionPokerType(poker)
	if pokerType < 10 {
		if len(poker) == len(handCards[4])*4 {
			outCards = poker
		} else if len(handCards[2]) > 0 && handCards[2][0] == 0x0F {
			if len(handCards[2])*2 == len(poker) || len(handCards[3])*3 == len(poker) || ((len(handCards[2])*2)+1 == len(poker) && len(handCards[2]) == 1) {
				for _, card := range poker {
					if publicLogic.GetValue(card) == 0x0F {
						outCards = append(outCards, card)
					}
				}
			}
		}
	}
	return outCards
}

// 三目运算符
func (l *Logic) TernaryOperator(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		switch reflect.ValueOf(trueVal).Type().String() {
		case "func()":
			runFunc := trueVal.(func())
			runFunc()
		}
		return trueVal
	}
	switch reflect.ValueOf(falseVal).Type().String() {
	case "func()":
		runFunc := falseVal.(func())
		runFunc()
	}
	return falseVal
}

// 获取顺子或者连队
func (l *Logic) GetContinuityByCardType(pokerType int32, pokers []int32) []int32 {
	var outCards []int32
	handCards := l.conversionPokerType(pokers)
	if pokerType == global.CardTypeSINGLEALONE {
		outCards = l.getSingleContinuity(handCards, false)
	} else if pokerType == global.CardTypeDOUBLEALONE {
		outCards = l.getDoubleContinuity(handCards[2])
	}
	return outCards
}
