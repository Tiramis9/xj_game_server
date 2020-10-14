package model

import (
	"xj_game_server/game/public/user"
)

//空闲场景消息
type FreeScene struct {
	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
}

//下注场景消息
type JettonScene struct {
	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
	UserListJetton map[int32][]int32    `json:"user_list_jetton"` //玩家下注
	CurrentChairID int32                `json:"current_chair_id"` //当前操作椅子号
	WinsChairID    int32                `json:"wins_chair_id"`    //胜利椅子号
	Rounds         int                  `json:"rounds"`           //轮数
}

//摊牌场景
type TPScene struct {
	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
	WinsChairID    int32                `json:"wins_chair_id"`    //胜利椅子号
}

//结束游戏
type GameConclude struct {
}

//摊牌通知
type UserTP struct {
	ChairID int32   `json:"chair_id"` //用户椅子号
	Poker   []uint8 `json:"user_tp"`  //用户扑克
}
