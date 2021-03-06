// Code generated by protoc-gen-go. DO NOT EDIT.
// source: msg/cmd.proto

package msg

//登陆消息
type Game_C_TokenLogin struct {
	Token                string   `json:"token"`
	MachineID            string   `json:"machine_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//机器人登陆
type Game_C_RobotLogin struct {
	UserID               int32    `json:"user_id"`
	BatchID              int32    `json:"batch_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户坐下
type Game_C_UserSitDown struct {
	TableID              int32    `json:"table_id"`
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户起立
type Game_C_UserStandUp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户准备
type Game_C_UserPrepare struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户取消准备
type Game_C_UserUnPrepare struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户叫分
type Game_C_UserGrabLandlord struct {
	Multiple             int32    `json:"multiple"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户出牌
type Game_C_UserCP struct {
	Pokers               []int32  `json:"poker"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户过牌
type Game_C_UserPass struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//换桌
type Game_C_ChangeTable struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 托管
type Game_C_AutoManage struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 取消托管
type Game_C_UnAutoManage struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//请求失败
type Game_S_ReqlyFail struct {
	ErrorCode            int32    `json:"errno"`
	ErrorMsg             string   `json:"errmsg"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//登陆成功
type Game_S_LoginSuccess struct {
	GameStartTime        int32    `json:"game_start_time"`
	GameJFTime           int32    `json:"game_jf_time"`
	GameCPTime           int32    `json:"game_cp_time"`
	Status               int32    `json:"status"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//坐下通知消息
type Game_S_SitDownNotify struct {
	Data                 *Game_S_User `json:"data"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

//起立通知消息
type Game_S_StandUpNotify struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//掉线通知消息
type Game_S_OffLineNotify struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//上线通知消息
type Game_S_OnLineNotify struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//user
type Game_S_User struct {
	UserID               int32    `json:"user_id"`
	NikeName             string   `json:"nike_name"`
	UserDiamond          float32  `json:"user_money"`
	MemberOrder          int32    `json:"member_order"`
	HeadImageUrl         string   `json:"head_image_url"`
	FaceID               int32    `json:"face_id"`
	RoleID               int32    `json:"role_id"`
	SuitID               int32    `json:"suit_id"`
	PhotoFrameID         int32    `json:"photo_frame_id"`
	TableID              int32    `json:"table_id"`
	ChairID              int32    `json:"chair_id"`
	Status               int32    `json:"status"`
	Gender               int32    `json:"gender"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//空闲场景
type Game_S_FreeScene struct {
	UserList             []Game_S_User    `json:"user_list"`
	PrepareList          []UserListStatus `json:"prepare_user_list"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

// 是否托管
type UserListTrusteeship struct {
	ChairID int32 `json:"chair_id"`
	Status  bool  `json:"is_auto"`
}

// 是否托管
type UserListStatus struct {
	ChairID int32 `json:"chair_id"`
	Status  bool  `json:"status"`
}

//叫分场景
type Game_S_GrabLandlordScene struct {
	SceneStartTime       int64                  `json:"scene_start_time"`
	UserList             []Game_S_User          `json:"user_list"`
	UserPoker            []int32                `json:"user_poker"`
	UserListGrabLandlord []UserListGrabLandlord `json:"user_list_grablandlord"`

	CurrentChairID       int32    `json:"current_chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 叫分
type UserListGrabLandlord struct {
	ChairID int32 `json:"chair_id"`
	Score   int32 `json:"score"`
}

// 出牌
type UserListPokerCount struct {
	ChairID int32 `json:"chair_id"`
	Count   int32 `json:"poker_num"`
}

//出牌场景
type Game_S_PlayScene struct {
	SceneStartTime       int64                 `json:"scene_start_time"`
	UserList             []Game_S_User         `json:"user_list"`
	UserPoker            []int32               `json:"user_poker"`
	UserListPokerCount   []UserListPokerCount  `json:"user_list_poker_count"` // 手上剩余牌
	BankerChairID        int32                 `json:"banker_chair_id"`
	Multiple             int32                 `json:"multiple"`
	LandlordPokers       []int32               `json:"landlord_poker"`
	SumMultiple          int32                 `json:"sum_multiple"`
	CurrentChairID       int32                 `json:"current_chair_id"`
	NearestChairID       int32                 `json:"nearest_chair_id"`
	NearestPokers        []int32               `json:"nearest_poker"`
	NearestCardType      int32                 `json:"nearest_card_type"`
	PokerCount           []int32               `json:"poker_count"`
	UserListTrusteeship  []UserListTrusteeship `json:"user_list_trusteeship"` // 是否托管
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

//开始游戏
type Game_S_StartGame struct {
	UserPoker            []int32  `json:"user_poker"`
	RecordID             string   `json:"record_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//当前用户
type Game_S_CurrentUser struct {
	CurrentChairID       int32    `json:"current_chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 没人叫分,重新发牌
type Game_S_GameRestart struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//确定地主开始出牌
type Game_S_StartCPDetermine struct {
	CurrentChairID       int32    `json:"current_chair_id"`
	Multiple             int32    `json:"multiple"`
	LandlordPokers       []int32  `json:"landlord_poker"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//结束游戏消息
type Game_S_GameConclude struct {
	UserListLoss    []Game_S_GameResult `json:"user_list_loss"`
	UserListMoney   []Game_S_GameResult `json:"user_list_money"`
	UserHandPoker   []Game_S_HandPoker  `json:"user_hand_poker"`
	SpringType      int32               `json:"spring_type"`
	CurrentMultiple int32               `json:"current_multiple"`
}

// 结束游戏
type Game_S_GameResult struct {
	ChairID int32   `json:"chair_id"`
	Result  float32 `json:"user_money"`
}

// 手上剩牌
type Game_S_HandPoker struct {
	Poker   []int32 `json:"poker"`
	ChairID int32   `json:"chair_id"`
}

//准备通知
type Game_S_UserPrepare struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//取消准备通知
type Game_S_UserUnPrepare struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//叫分通知
type Game_S_UserGrabLandlord struct {
	ChairID              int32    `json:"chair_id"`
	Multiple             int32    `json:"multiple"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户出牌通知
type Game_S_UserCP struct {
	ChairID              int32    `json:"chair_id"`
	Pokers               []int32  `json:"poker"`
	PokerType            int32    `json:"poker_type"`
	CurrentMultiple      int32    `json:"current_multiple"`
	PokerCount           int32    `json:"poker_count"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

//用户过牌通知
type Game_S_UserPass struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 用户托管通知
type Game_S_AutoManage struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// 用户取消托管通知
type Game_S_UnAutoManage struct {
	ChairID              int32    `json:"chair_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
