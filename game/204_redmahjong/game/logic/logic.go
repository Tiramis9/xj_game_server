package logic

import (
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/204_redmahjong/msg"
	publicLogic "xj_game_server/game/public/logic"
)

var Client = new(Logic)

type Logic struct {
}

//发牌
func (it *Logic) DispatchTableCard() []int32 {
	mahjong := &publicLogic.Mahjong{
		MahjongCount: 112,
	}
	return mahjong.GetMahjongHeap()
}

//摸牌是否有响应
func (it *Logic) SendMjResponse(mjData int32, userMahjong map[int32]int32, userDiskMjList *msg.DiskMahjongList, first bool) int32 {
	//操作码
	response := int32(0)

	//暗杠检测
	for k, v := range userMahjong {
		if v != 4 {
			continue
		}

		if !first && mjData != k {
			continue
		}

		response |= global.WIK_AN_GANG
		break
	}

	if response == 0 && userDiskMjList != nil {
		//补杠检测
		for _, v := range userDiskMjList.Data {
			if v.Code != global.WIK_PENG {
				continue
			}

			if mjData != v.Data {
				continue
			}

			response |= global.WIK_BU_GANG
			break
		}
	}

	//红中胡检测
	if userMahjong[global.MagicMahjong] == 4 {
		response |= global.CHR_SI_HONG_ZHONG
		return response
	}

	//七对胡检测
	code := it.CheckQiDuiHu(userMahjong)
	if code != 0 {
		response |= code
		return response
	}

	//平胡检测
	response |= it.CheckPingHu(userMahjong)
	return response
}

//出牌是否有响应
func (it *Logic) OutMjResponse(chairID int32, mjData int32, userListMahjong map[int32]map[int32]int32) map[int32]int32 {
	//操作码
	response := make(map[int32]int32)

	for k, v := range userListMahjong {
		if k == chairID {
			continue
		}

		//胡牌检测
		//code := it.CheckPingHu(v)
		//if code != 0 {
		//	response[k] |= code
		//}

		if v[mjData] < 2 {
			continue
		} else if v[mjData] == 2 {
			response[k] |= global.WIK_PENG
		} else {
			response[k] |= global.WIK_MING_GANG
		}
	}

	return response
}

//判断是否听牌
func (it *Logic) CheckTing(mjData int32, userMahjong map[int32]int32) bool {
	//红中胡检测
	if userMahjong[global.MagicMahjong] >= 3 {
		return true
	}
	var temp = make(map[int32]int32)
	for k, v := range userMahjong {
		temp[k] = v
	}
	temp[mjData]++

	//七对胡检测
	if it.CheckQiDuiHu(temp)&global.WIK_HU != 0 {
		return true
	}

	//平胡检测
	if it.CheckPingHu(temp)&global.WIK_HU != 0 {
		return true
	}

	return false
}

//平胡
func (it *Logic) CheckPingHu(userMahjong map[int32]int32) int32 {
	var data [31]int32
	for k, v := range userMahjong {
		if k == global.MagicMahjong {
			continue
		}

		data[it.SwitchToIndex(k)] = v
	}

	var code int32
	for k, v := range data {
		//取出一对将
		tempData := data
		var queCount int32
		switch v {
		case 1:
			queCount++
			tempData[k]--
		case 2, 3, 4:
			tempData[k] -= 2
		default:
			continue
		}

		for i := 0; i < 31; {
			if tempData[i] == 0 {
				i++
				continue
			}

			//数值
			value := i%9 + 1
			for tempData[i] > 0 {
				//判断是否是刻字
				if tempData[i] == 3 {
					tempData[i] = 0
					continue
				}

				//判断是否是顺子
				if value <= 0x09-2 && tempData[i+1] > 0 && tempData[i+2] > 0 {
					for j := 0; j < 3; j++ {
						tempData[i+j]--
					}
					continue
				}

				switch tempData[i] {
				case 1:
					if value <= 0x08 && ((value == 0x08 && tempData[i+1] > 0) || tempData[i+1] > 0 || tempData[i+2] > 0) {
						if tempData[i+1] != 0 {
							tempData[i+1] --
						}
						if value != 0x08 && tempData[i+2] != 0 {
							tempData[i+2]--
						}
						queCount++
					} else {
						queCount += 2
					}
				case 2:
					queCount++
				case 4:
					if value <= 0x08 && ((value == 0x08 && tempData[i+1] > 0) || tempData[i+1] > 0 || tempData[i+2] > 0) {
						if tempData[i+1] != 0 {
							tempData[i+1] --
						}
						if value != 0x08 && tempData[i+2] != 0 {
							tempData[i+2]--
						}
						queCount++
					} else {
						queCount += 2
					}
				}

				tempData[i] = 0
			}
		}

		if queCount == 0 && userMahjong[global.MagicMahjong] != 1 || queCount > 0 && queCount == userMahjong[global.MagicMahjong] {
			code |= global.CHR_PING_HU
			code |= global.WIK_HU

			if queCount > 0 {
				code |= global.CHR_MAGIC
			}
			break
		}
	}

	return code
}

//七对胡
func (it *Logic) CheckQiDuiHu(userMahjong map[int32]int32) int32 {
	count := int32(0)
	queCount := int32(0)
	for k, v := range userMahjong {
		count += v
		if k == global.MagicMahjong {
			continue
		}

		if v != 2 && v != 4 {
			queCount++
		}
	}

	code := int32(0)
	if count == 14 && (queCount == 0 || queCount == userMahjong[global.MagicMahjong]) {
		code |= global.CHR_QI_DUI
		code |= global.WIK_HU

		if queCount > 0 {
			code |= global.CHR_MAGIC
		}
	}

	return code
}

//获取最低需要癞子数
func (it *Logic) GetMininum(userMahjong map[int32]int32) int32 {
	var data [31]int32
	for k, v := range userMahjong {
		if k == global.MagicMahjong {
			continue
		}

		data[it.SwitchToIndex(k)] = v
	}

	var Minimum int32
	for k, v := range data {
		var queCount int32
		//取出一对将
		tempData := data
		switch v {
		case 0:
			queCount += 2
		case 1:
			queCount++
			tempData[k] -= 1
		case 2:
			tempData[k] -= 2
		case 3:
			tempData[k] -= 2
		case 4:
			tempData[k] -= 2
		}

		for i := 0; i < 31; {
			if tempData[i] == 0 {
				i++
				continue
			}

			//数值
			value := i%9 + 1
			for tempData[i] > 0 {
				//判断是否是顺子
				if value <= 0x09-2 && tempData[i+1] > 0 && tempData[i+2] > 0 {
					for j := 0; j < 3; j++ {
						tempData[i+j]--
					}
					continue
				}

				switch tempData[i] {
				case 1:
					if value <= 0x08 && ((value == 0x08 && tempData[i+1] > 0) || tempData[i+1] > 0 || tempData[i+2] > 0) {
						if tempData[i+1] != 0 {
							tempData[i+1] --
						}
						if value != 0x08 && tempData[i+2] != 0 {
							tempData[i+2]--
						}
						queCount++
					} else {
						queCount += 2
					}
				case 2:
					queCount++
				case 3:
				case 4:
					queCount += 2
				}

				tempData[i] = 0
			}
		}

		if Minimum == 0 && queCount != 0 {
			Minimum = queCount
		}
		if queCount < Minimum {
			Minimum = queCount
		}
	}

	return Minimum
}

// 获取最佳出牌数据，出牌后缺的听胡数最少
func (it *Logic) GetOutMjData(MjData map[int32]int32) int32 {
	replaceCard := func(card int32, tempCards map[int32]int32) map[int32]int32 {
		if tempCards[card] != 0 {
			tempCards[card] --
		}
		tempCards[0x35]++
		return tempCards
	}
	var minFlag int32 = 0x0F // 再差的牌最多缺14张牌
	var userMjData = make(map[int32]int32)
	for k, v := range MjData {
		userMjData[k] = v
	}
	MagicMahjong := userMjData[global.MagicMahjong] //记录癞子
	var OutCardsAsc = make(map[int32]int32, 0)
	for k, num := range userMjData {
		if num <= 0 {
			delete(userMjData, k)
			continue
		}
		if k != 0x35 {
			queCount := it.GetMininum(replaceCard(k, userMjData))
			if minFlag > queCount {
				minFlag = queCount
				OutCardsAsc[minFlag] = k
			}
			userMjData[k]++
			userMjData[global.MagicMahjong] = MagicMahjong
			if MagicMahjong == 0 {
				delete(userMjData, global.MagicMahjong)
			}
		}
	}
	return OutCardsAsc[minFlag]
}

//补扛是否有响应
func (it *Logic) GetBuGangResponse(chairID int32, mjData int32, userListMahjong map[int32]map[int32]int32) map[int32]int32 {
	//操作码
	response := make(map[int32]int32)

	for k, v := range userListMahjong {
		if k == chairID {
			continue
		}
		v[mjData]++
		//胡牌检测
		code := it.CheckPingHu(v)
		if code != 0 {
			response[k] |= code
		}
		//红中胡检测
		if v[global.MagicMahjong] == 4 {
			response[k] |= global.CHR_SI_HONG_ZHONG
		}
		//七对胡检测
		code = it.CheckQiDuiHu(v)
		if code != 0 {
			response[k] |= code
			return response
		}
		v[mjData]--
		if v[mjData] == 0 {
			delete(v, mjData)
		}
	}

	return response
}

// 获取最佳出牌数据，获取听牌的数据
func (it *Logic) GetTingMjData(mjData map[int32]int32, mjDataNum map[int32]int32) map[int32]map[int32]int32 {
	replaceCard := func(card int32, tempCards map[int32]int32) {
		if tempCards[card] != 0 {
			tempCards[card] --
		}
		tempCards[global.MagicMahjong]++
	}
	userMjData := make(map[int32]int32)
	DataNum := make(map[int32]int32, 0) //未现的牌
	for k, v := range mjDataNum {
		DataNum[k] = v
	}
	for k, v := range mjData {
		userMjData[k] = v
		DataNum[k] -= v
	}
	MagicMahjong := userMjData[global.MagicMahjong] //记录癞子
	var OutCardsAsc = make(map[int32]int32, 0)
	var tingHuCards = make(map[int32]map[int32]int32, 0)
	for k, num := range userMjData {
		if num <= 0 {
			delete(userMjData, k)
			continue
		}
		if k != global.MagicMahjong {
			replaceCard(k, userMjData)
			queCount := it.GetMininum(userMjData)
			if queCount-MagicMahjong <= 1 {
				OutCardsAsc[k] = num
			}
			userMjData[k]++
			userMjData[global.MagicMahjong] = MagicMahjong
			if MagicMahjong == 0 {
				delete(userMjData, global.MagicMahjong)
			}
		}
	}
	heapMj := publicLogic.GetOneMahjongHeap()
	for k, _ := range OutCardsAsc {
		userMjData[k]--
		if userMjData[k] == 0 {
			delete(userMjData, k)
		}
		tingHuCards[k] = make(map[int32]int32, 0)
		for _, v := range heapMj {
			if it.CheckTing(v, userMjData) {
				tingHuCards[k][v] = DataNum[v]
			}
		}
		userMjData[k]++
	}
	return tingHuCards
}

// 获取逻辑值0-31
func (it *Logic) SwitchToIndex(mjData int32) int32 {
	return (mjData&0xF0>>4)*9 + (mjData & 0x0F) - 1
}
