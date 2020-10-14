package model

import (
	"xj_game_server/game/105_benchibaoma/msg"
	"xj_game_server/game/public/user"
)

// 登录 大厅场景
//type MsgHallHistory struct {
//	TableID         int32     `json:"table_id"`          //桌子id
//	GameJettonTime  int32     `json:"game_jetton_time"`  //下注时间
//	GameLotteryTime int32     `json:"game_lottery_time"` //开奖时间
//	JettonList      []float32 `json:"jetton_list"`       //筹码选项列表
//	LotteryRecord   []int8    `json:"lottery_record"`    //开奖记录 最近20局
//	GameStatus      int32     `json:"game_status"`       //游戏状态
//	SceneStartTime  int64     `json:"scene_start_time"`  //场景开始时间
//	UserCount       int32     `json:"user_count"`        //玩家数量
//}

// 游戏结束
//type GameConclude struct {
//	LotteryPoker   [global.PokerCount]uint8 `json:"lottery_poker"`   //开奖颜色和动物
//	LotterySpecial uint8                    `json:"lottery_special"` //特殊开奖
//	WinArea        [global.AreaCount]bool   `json:"win_area"`        //输赢区域
//	UserListLoss   map[int32]float32        `json:"user_list_loss"`  //桌面用户盈亏
//}

//下注场景消息
//type JettonScene struct {
//	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info `json:"user_list"`        //桌面用户
//	LotteryRecord  []int8               `json:"lottery_record"`   //开奖记录
//	UserArraJetton []AreaJetton         `json:"user_arra_jetton"` //下注区域,总下注数
//}

// 开奖场景
//type LotteryScene struct {
//	SceneStartTime int64                    `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                    `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info     `json:"user_list"`        //桌面用户
//	LotteryRecord  []int8                   `json:"lottery_record"`   //开奖记录
//	UserArraJetton []AreaJetton             `json:"user_arra_jetton"` //下注区域,总下注数
//	LotteryPoker   [global.PokerCount]uint8 `json:"lottery_poker"`    //开奖扑克
//	WinArea        [global.AreaCount]bool   `json:"win_area"`         //输赢区域
//}

// 每个区域的下注数
//type AreaJetton struct {
//	Area   int     `json:"area"`
//	Jetton float32 `json:"jetton"`
//}
//
//type AreaJettonMsg struct {
//	TableID     int32        `json:"table_id"`
//	AreaJettons []AreaJetton `json:"area_jettons"`
//}

//机器人记录
type RobotSceneData struct {
	UserGold      float32              `json:"user_gold"`
	UserDiamond   float32              `json:"user_diamond"`
	UserChairID   int32                `json:"user_chair_id"`  //用户椅子号
	UserList      map[int32]*user.Info `json:"user_list"`      //玩家列表
	LotteryRecord []int8               `json:"lottery_record"` //开奖记录
}

//富豪榜
type Special []*msg.Game_S_User

func (ms Special) Len() int {
	return len(ms)
}

func (ms Special) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

//按键排序
func (ms Special) Less(i, j int) bool {
	// 先比砖石
	if ms[i].UserDiamond > ms[j].UserDiamond {
		return true
	}
	// 如果砖石相同比金币
	if ms[i].UserDiamond == ms[j].UserDiamond {
		return ms[i].UserGold > ms[j].UserGold
	}
	return false
}

//神算子
type Operator struct {
	User  *msg.Game_S_User
	Count int //赢的次数
}

type SpecialOperator []*Operator

func (ms SpecialOperator) Len() int {
	return len(ms)
}

func (ms SpecialOperator) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

//按键排序
func (ms SpecialOperator) Less(i, j int) bool {
	return ms[i].Count > ms[j].Count
}
