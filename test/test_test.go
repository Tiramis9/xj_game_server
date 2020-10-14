package test

import (
	"sync"
	"testing"
	"xj_game_server/game/207_qiangzhuangniuniu_kansanzhang/msg"
	"xj_game_server/game/public/user"
)

/*
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
	{"0":23043,"1":23046,"2":1001030},"user_list_qz":{"0":3,"1":1,"2":0},

	"user_list_jetton":{"0":0,"1":5,"2":1},"user_list_poker":

		{"0":{"card_type":28,"lottery_poker":[28,38,9,11,22]},

			"1":{"card_type":541,"lottery_poker":[2,57,39,29,52]},

			"2":{"card_type":45,"lottery_poker":[33,3,34,45,18]}},

"2":[0x21,0x03,0x22,0x2D,0x12]
"1":[0x02,0x39,0x27,1D,0x34]
"0":[0x1C,0x26,0x9,B,0x26]
	"user_list_loss":{"0":324,"1":-300,"2":-60},"user_tax":{"0":36,"1":0,"2":0}
*/
func TestGetSystemLoss(t *testing.T) {

	userListJetton := sync.Map{}
	userListJetton.Store(int32(0), int32(0))
	userListJetton.Store(int32(1), int32(5))
	userListJetton.Store(int32(2), int32(1))
	userListPoker := make(map[int32]*msg.Game_S_LotteryPoker, 0)
	temp0 := new(msg.Game_S_LotteryPoker)
	temp0.LotteryPoker = []int32{28, 38, 9, 11, 22}
	temp0.PokerType = getPokerType(temp0.LotteryPoker)
	userListPoker[int32(0)] = temp0
	temp1 := new(msg.Game_S_LotteryPoker)
	temp1.LotteryPoker = []int32{2, 57, 39, 29, 52}
	temp1.PokerType = getPokerType(temp1.LotteryPoker)
	userListPoker[int32(1)] = temp1
	temp2 := new(msg.Game_S_LotteryPoker)
	temp2.LotteryPoker = []int32{33, 3, 34, 45, 18}
	temp2.PokerType = getPokerType(temp2.LotteryPoker)
	userListPoker[int32(2)] = temp2

	userList := sync.Map{}
	userList.Store(int32(0), int32(23043))
	userList.Store(int32(1), int32(23046))
	userList.Store(int32(2), int32(1001030))
	user0 := new(user.Item)
	user0.UserDiamond = 5000
	user0.BatchID = 11
	user.List.Store(int32(23043), user0)

	user1 := new(user.Item)
	user1.UserDiamond = 5000
	user1.BatchID = 11
	user.List.Store(int32(23046), user1)

	user2 := new(user.Item)
	user2.UserDiamond = 5000
	user2.BatchID = -1
	user.List.Store(int32(1001030), user2)
	systemScore, userListLoss, userTax, jackpots := GetSystemLoss(0, 1, userListJetton, userListPoker, userList)
	t.Log(systemScore, userListLoss, userTax, jackpots)
}
