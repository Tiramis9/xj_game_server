package logic

import (
	"xj_game_server/game/104_senglinwuhui/global"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	rand "xj_game_server/util/leaf/util"
	"sync"
)

var Client = new(Logic)

type Logic struct {
}

const (
	//颜色掩码
	//maskColor = 0xF0
	//动物掩码
	maskAnimal = 0x0F

	//大三元 48
	tripleDouble = 0x30
	//大四喜 64
	tripleFour = 0x40
	//森林闪电 80
	forestLightning = 0x50
)

func getAnimal(poker int32) int32 {
	return poker & maskAnimal
}

func getColor(poker int32) int32 {
	return poker - getAnimal(poker)
}

func GetColorName(poker int32) string {
	switch poker {
	case 0x00:
		return "红色"
	case 0x10:
		return "绿色"
	case 0x20:
		return "黄色"
	}
	return ""
}

func GetAnimalName(poker int32) string {
	switch poker {
	case 0x00:
		return "狮子"
	case 0x01:
		return "熊猫"
	case 0x02:
		return "猴子"
	case 0x03:
		return "兔子"
	}
	return ""
}

//红色、绿色、黄色
// 0 16 32
var forestColor = []int32{
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
	0x00, 0x10, 0x20,
}

//狮子，熊猫，猴子，兔子
//var forestAnimal = []int32{
//	0x00, 0x01, 0x02, 0x03,
//	0x00, 0x01, 0x02, 0x03,
//	0x00, 0x01, 0x02, 0x03,
//	0x00, 0x01, 0x02, 0x03,
//	0x00, 0x01, 0x02, 0x03,
//	0x00, 0x01, 0x02, 0x03,
//}

// 区域
var area = []int32{
	//红色
	0x00, 0x01, 0x02, 0x03,
	//绿色 16~19
	0x10, 0x11, 0x12, 0x13,
	//黄 32~35
	0x20, 0x21, 0x22, 0x23,
}
var slWhPk []int32

func init() {
	//根据赔率生成开奖随机列表
	var sumAreaMultiple = 0
	for _, v := range global.AreaMultiple {
		sumAreaMultiple += int(v)
	}
	for i, v := range global.AreaMultiple {
		x := sumAreaMultiple / int(v)
		if i == 3 || i == 7 || i == 10 || i == 11 {
			x = x * 3
		}
		for j := 0; j <= x; j++ {
			if i >= 4 && i <= 7 {
				i = i + 12
			}
			if i >= 8 && i <= 11 {
				i = i + 24
			}
			slWhPk = append(slWhPk, int32(i))
		}
	}
}

// 获取赢的区域id
func GetWinAreaId(area []bool) int32 {
	var newArea int32
	for k, v := range area {
		if v {
			newArea = int32(k)
		}
	}
	return newArea
}

// 随机颜色
func RandomColor() []int32 {

	var fColors = make([]int32,0)

	for i := len(forestColor) - 1; i > 0; i-- {
		num := rand.RandInterval(0, int32(i))
		forestColor[i], forestColor[num] = forestColor[num], forestColor[i]
	}

	fColors = append(fColors, forestColor...)

	return fColors
}

//发牌
func (*Logic) DispatchTableCard(randomColor []int32) ([]int32, int32, int32) {
	lotteryPokerIndex := rand.RandInterval(0, int32(len(slWhPk)-1))
	temp := slWhPk[lotteryPokerIndex]
	//颜色下标
	vColor := 0
	// 随机第几个
	colorIndex := rand.RandInterval(0, 7)
	//记录颜色出现的次数
	var j = -1
	for i, v := range randomColor {
		if getColor(temp) == v {
			j++
			//当次数等于随机值时跳出循环
			if j == int(colorIndex) {
				vColor = i
				break
			}
		}
	}
	//赋值
	lotteryPoker := make([]int32, 0)
	lotteryPoker = append(lotteryPoker, getColor(temp))
	lotteryPoker = append(lotteryPoker, getAnimal(temp))
	//随机大三元大,四喜,森林闪电
	sj := rand.RandInterval(0, 100)
	var special int32
	switch sj {
	case tripleDouble:
		special = tripleDouble
	case tripleFour:
		special = tripleFour
	case forestLightning:
		special = forestLightning
	}
	return lotteryPoker, special, int32(vColor)
}

//获取获胜区域
//lotteryPoker 开奖扑克
//areaJetton 区域下注筹码
//大三元  同一个动物的三个颜色都中奖
//大四喜  同一个颜色的4个动物都中奖
//森林闪电 压中后所有得奖一律*2
func (logic *Logic) GetWinArea(lotteryPoker []int32, special int32) []bool {
	var winArea = make([]bool, global.AreaCount)
	var sum = lotteryPoker[0] + lotteryPoker[1]
	for key, v := range area {
		var color = getColor(v)
		var animal = getAnimal(v)
		if sum == v {
			winArea[key] = true
		}
		if special == tripleDouble { //大三元
			if animal == lotteryPoker[1] { //动物相同
				winArea[key] = true
			}
		}
		if special == tripleFour { //大四喜
			if color == lotteryPoker[0] { //同一个颜色
				winArea[key] = true
			}
		}
	}
	return winArea
}

//返回系统损耗
func (*Logic) GetSystemLoss(winArea []bool, userListAreaJetton sync.Map, userList sync.Map, special int32) (float32, map[int32]float32, map[int32]float32, float32) {

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
			userListLoss[k.(int32)] -= v.([global.AreaCount]float32)[i]
			//机器人批次号为 > 0
			if !userItem.(*user.Item).IsRobot() {
				userNoRootTax += userListLoss[k.(int32)]
			}
			if winArea[i] {
				var userWins float32
				if special == forestLightning {
					userWins = v.([global.AreaCount]float32)[i] * global.AreaMultiple[i] * 2
				} else {
					userWins = v.([global.AreaCount]float32)[i] * global.AreaMultiple[i]
				}
				//机器人批次号为 > 0
				if !userItem.(*user.Item).IsRobot() {
					userNoRootTax += userWins
				}
				// 判断是否>0 赢了 要扣税
				userListLoss[k.(int32)] += userWins - userWins*store.GameControl.GetGameInfo().RevenueRatio
				// 记录税收
				userTax[k.(int32)] += userWins * store.GameControl.GetGameInfo().RevenueRatio
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
