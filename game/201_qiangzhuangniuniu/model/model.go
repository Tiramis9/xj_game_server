package model

//空闲场景消息
//type FreeScene struct {
//	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
//}

//抢庄场景消息
//type QZScene struct {
//	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
//	UserListQZ     map[int32]int32      `json:"user_list_qz"`     //用户抢庄
//}

//下注场景消息
//type JettonScene struct {
//	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
//	UserListJetton map[int32]int32      `json:"user_list_jetton"` //玩家下注
//	BankerChairID  int32                `json:"banker_chair_id"`  //庄家椅子号
//	BankerMultiple int32                `json:"banker_multiple"`  //庄家抢庄倍数
//}

//摊牌场景
//type TPScene struct {
//	SceneStartTime int64                `json:"scene_start_time"` //场景开始时间
//	UserChairID    int32                `json:"user_chair_id"`    //用户椅子号
//	UserList       map[int32]*user.Info `json:"user_list"`        //用户列表
//	UserListTP     map[int32][]uint8    `json:"user_list_tp"`     //已摊牌玩家
//	BankerChairID  int32                `json:"banker_chair_id"`  //庄家椅子号
//	BankerMultiple int32                `json:"banker_multiple"`  //庄家抢庄倍数
//}

//结束游戏
//type GameConclude struct {
//}

//摊牌通知
//type UserTP struct {
//	ChairID int32   `json:"chair_id"` //用户椅子号
//	Poker   []uint8 `json:"user_tp"`  //用户扑克
//}
