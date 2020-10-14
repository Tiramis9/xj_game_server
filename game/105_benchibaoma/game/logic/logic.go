/*
 * @Author: yhlyl
 * @Date: 2019-12-25 11:02:08
 * @LastEditTime : 2020-01-04 15:59:08
 * @LastEditors  : yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/105_benchibaoma/game/logic/logic.go
 * @https://github.com/android-coco
 */
package logic

import (
	"xj_game_server/game/105_benchibaoma/global"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	rand "xj_game_server/util/leaf/util"
	"sync"
)

var Client = new(Logic)

type Logic struct {
}

//40, 5, 30, 5, 20, 5, 10, 5,
//奔驰宝马
var bcBmPk []int32

//var bcBmPk = []int32{
////大小  大小
//0x00, 0x01, 0x02, 0x03,
//0x04, 0x05, 0x06, 0x07,
////大小  大小
//0x00, 0x01, 0x02, 0x03,
//0x04, 0x05, 0x06, 0x07,
////大小  大小
//0x00, 0x01, 0x02, 0x03,
//0x04, 0x05, 0x06, 0x07,
////大小  大小
//0x00, 0x01, 0x02, 0x03,
//0x04, 0x05, 0x06, 0x07,
//}

func init() {
	//根据赔率生成开奖随机列表
	var sumAreaMultiple = 0
	for _, v := range global.AreaMultiple {
		sumAreaMultiple += int(v)
	}
	for i, v := range global.AreaMultiple {
		x := sumAreaMultiple / int(v)
		if i == 2 || i == 4 || i == 6 || i == 7 {
			x = x * 3
		}
		for j := 0; j <= x; j++ {
			bcBmPk = append(bcBmPk, int32(i))
		}
	}
}

// 区域
var area = []int32{
	//大小 大小
	0x00, 0x01, 0x02, 0x03,
	//大小 大小
	0x04, 0x05, 0x06, 0x07,
}

//发牌
func (*Logic) DispatchTableCard() int32 {
	lotteryNum := rand.RandInterval(0, int32(len(bcBmPk)-1))
	return bcBmPk[lotteryNum]
}

//获取获胜区域
//lotteryPoker 开奖扑克
//areaJetton 区域下注筹码
func (logic *Logic) GetWinArea(lotteryPoker int32) []bool {
	var winArea = make([]bool, global.AreaCount)
	for key, v := range area {
		if v == lotteryPoker {
			winArea[key] = true
		}
	}
	return winArea
}

//返回系统损耗
func (*Logic) GetSystemLoss(winArea []bool, userListAreaJetton sync.Map, userList sync.Map) (float32, map[int32]float32, map[int32]float32, float32) {

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
				userWins := v.([global.AreaCount]float32)[i] * global.AreaMultiple[i]
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
