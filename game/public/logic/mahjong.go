package logic

import (
	golog "log"
	"xj_game_server/util/leaf/log"
)

var hzMahjong = []int32 {
	0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08,0x09,						//万子
	0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08,0x09,						//万子
	0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08,0x09,						//万子
	0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08,0x09,						//万子
	0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,						//索子
	0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,						//索子
	0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,						//索子
	0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,						//索子
	0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,						//同子
	0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,						//同子
	0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,						//同子
	0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,						//同子
	0x35,0x35,0x35,0x35,												//红中
}

type Mahjong struct {
	MahjongCount	int32
}

func (it *Mahjong) GetMahjongHeap() []int32 {
	var mahjongHeap []int32
	switch it.MahjongCount {
	case 112:
		mahjongHeap = randomShuffle(hzMahjong)
	default:
		_ = log.Logger.Errorf("getMahjongHeap \"MahjongCount\": %v is err", it.MahjongCount)
		golog.Fatal("getMahjongHeap \"MahjongCount\": %v is err", it.MahjongCount)
	}
	return mahjongHeap
}

// 获取单次麻将堆
func GetOneMahjongHeap() []int32 {
	mjData := make(map[int32]int32)
	for _, v := range hzMahjong {
		mjData[v] = v
	}
	var mjheap []int32
	for _, v := range mjData {
		mjheap = append(mjheap, v)
	}
	return mjheap
}
