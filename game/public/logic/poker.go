package logic

import (
	rand "xj_game_server/util/leaf/util"
	"sort"
)

type Poker struct {
	Count        int  //多少副牌
	ShuffleCount int  //洗牌次数
	IsNeedGhost  bool // 是否需要大小鬼
}

const (
	//花色掩码
	maskColor = 0xF0
	//数值掩码
	maskValue = 0x0F
)

var pokerNoGhost = []int32{
	//0方块: 2 3 4 5 6 7 8 9 10 J Q K A
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x01,
	//1梅花: 2 3 4 5 6 7 8 9 10 J Q K A
	0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x11,
	//2红桃: 2 3 4 5 6 7 8 9 10 J Q K A
	0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x21,
	//3黑桃: 2 3 4 5 6 7 8 9 10 J Q K A
	0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x31,
}

var pokerHaveGhost = []int32{
	//0方块: 2 3 4 5 6 7 8 9 10 J Q K A
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x01,
	//1梅花: 2 3 4 5 6 7 8 9 10 J Q K A
	0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x11,
	//2红桃: 2 3 4 5 6 7 8 9 10 J Q K A
	0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x21,
	//3黑桃: 2 3 4 5 6 7 8 9 10 J Q K A
	0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0x31,
	//4鬼：大鬼 小鬼 十进制 79 95
	0x4F, 0x5F,
}

func (p Poker) GetPokers() []int32 {
	var porkers []int32
	if !p.IsNeedGhost {
		for i := 0; i < p.Count; i++ {
			porkers = append(porkers, pokerNoGhost...)
		}
	} else {
		for i := 0; i < p.Count; i++ {
			porkers = append(porkers, pokerHaveGhost...)
		}
	}
	for i := 0; i < p.ShuffleCount; i++ {
		porkers = randomShuffle(porkers)
	}
	return porkers
}

func GetValue(poker int32) int32 {
	return poker & maskValue
}

func GetColor(poker int32) int32 {
	return (poker & maskColor) / 16
}

func ContrastSize(poker int32, poker1 int32) int32 {

	if CompareK(poker, poker1) {
		return poker
	} else {
		return poker1
	}

}

//去花色 合计点数
func GetCombined(cards ...int32) int32 {

	var card int32

	for _, c := range cards {
		//去花色
		c = GetValue(c)

		if c >= 10 {
			c = 0
		}
		card += c
	}

	return card % 10
}

//获取对子数 和 相同花色数
func GetPairsColorNum(lotteryPoker []int32, page, pairs, color int) (int, int) {
	if len(lotteryPoker) == page {
		return pairs, color
	}

	for i := page + 1; i < len(lotteryPoker); i++ {
		if GetValue(lotteryPoker[page]) == GetValue(lotteryPoker[i]) {
			pairs++
		}
		if GetColor(lotteryPoker[page]) == GetColor(lotteryPoker[i]) {
			color++
		}
	}

	return GetPairsColorNum(lotteryPoker, page+1, pairs, color)
}

//获取对子数
func GetPairsNum(lotteryPoker []int32, page, num int) int {

	if len(lotteryPoker) == page {
		return num
	}

	for i := page + 1; i < len(lotteryPoker); i++ {
		if GetValue(lotteryPoker[page]) == GetValue(lotteryPoker[i]) {
			num++
		}
	}

	return GetPairsNum(lotteryPoker, page+1, num)

}

//获取相同花色数
func GetColorNum(lotteryPoker []int32, page, num int) int {

	if len(lotteryPoker) == page {
		return num
	}

	for i := page + 1; i < len(lotteryPoker); i++ {
		if GetColor(lotteryPoker[page]) == GetColor(lotteryPoker[i]) {
			num++
		}
	}

	return GetColorNum(lotteryPoker, page+1, num)

}

func IsFlushNum(lotteryPoker []int32) (bool, int32) {

	var shunza []int
	var card int32

	for _, v := range lotteryPoker {
		shunza = append(shunza, int(GetValue(v)))

		if card < v&maskValue {
			card = v & maskValue
		}
	}

	sort.Ints(shunza)

	for i := 1; i < len(shunza); i++ {
		if shunza[i-1]-shunza[i] != -1 {
			return false, 0
		}
	}

	return true, card

}

func IsColor(lotteryPoker []int32) (bool, int32) {

	for i := 1; i < len(lotteryPoker); i++ {
		if GetColor(lotteryPoker[i-1]) != GetColor(lotteryPoker[i]) {
			return false, 0
		}
	}
	return true, GetColor(lotteryPoker[0])
}

func GetColorString(poker int32) string {
	switch poker {
	case 0:
		return "方块"
	case 1:
		return "梅花"
	case 2:
		return "红桃"
	case 3:
		return "黑桃"
	default:
		return "未知"
	}
}

//比大小 K最大
func CompareK(poker int32, poker1 int32) bool {

	if GetValue(poker) > GetValue(poker1) {
		return true
	} else if GetValue(poker) == GetValue(poker1) {

		if GetColor(poker) > GetColor(poker1) {
			return false
		} else {
			return true
		}

	} else {
		return false
	}

}

//比大小 A最大
func CompareA(poker int32, poker1 int32) bool {

	if GetValue(poker) == 0x01 {
		poker = 0x0e
	}
	if GetValue(poker1) == 0x01 {
		poker1 = 0x0e
	}
	return CompareK(poker, poker1)
}

func randomShuffle(pokers []int32) []int32 {
	for i := len(pokers) - 1; i > 0; i-- {
		num := rand.RandInterval(0, int32(i))
		pokers[i], pokers[num] = pokers[num], pokers[i]
	}

	return pokers
}

func ReorderPokerK(pokers []int32, number int) []int32 {

	if number == len(pokers) {
		return pokers
	}

	for i := number + 1; i < len(pokers); i++ {

		if CompareK(pokers[number], pokers[i]) {
			pokers[number], pokers[i] = pokers[i], pokers[number]
		}

	}

	return ReorderPokerK(pokers, number+1)

}

func ReorderPokerA(pokers []int32, number int) []int32 {

	if number == len(pokers) {
		return pokers
	}

	for i := number + 1; i < len(pokers); i++ {

		if CompareA(pokers[number], pokers[i]) {
			pokers[number], pokers[i] = pokers[i], pokers[number]
		}

	}

	return ReorderPokerA(pokers, number+1)

}
