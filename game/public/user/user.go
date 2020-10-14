package user

import (
	"sync"
	"xj_game_server/util/leaf/gate"
)

const (
	StatusFree    = iota //空闲状态
	StatusPlaying        //游戏状态
	StatusOffline        //断线状态
)

//var List = make(map[int32]*Item) // 用户id--->对应一个user
var List sync.Map // 用户id--->对应一个user

type Info struct {
	UserID       int32   `json:"user_id"`        //用户ID
	NikeName     string  `json:"nike_name"`      //网名
	UserGold     float32 `json:"user_gold"`      //用户金币
	UserDiamond  float32 `json:"user_diamond"`   //用户余额
	Jackpot      float32 `json:"jackpot"`        //个人奖池
	MemberOrder  int32   `json:"member_order"`   //会员等级
	HeadImageUrl string  `json:"head_image_url"` //微信头像url
	FaceID       int32   `json:"face_id"`        //头像ID
	RoleID       int32   `json:"role_id"`        //角色标识
	SuitID       int32   `json:"suit_id"`        //套装标识
	PhotoFrameID int32   `json:"photo_frame_id"` //头像框标识
	TableID      int32   `json:"table_id"`       //桌子号
	ChairID      int32   `json:"chair_id"`       //椅子号
	Status       int32   `json:"status"`         //用户状态
	Gender       int32   `json:"gender"`         //性别
}

type Item struct {
	gate.Agent       //agent接口 连接代理
	Info             //用户信息
	UserRight  int32 //用户权限
	BatchID    int32 //机器人批次ID
}

func (it *Item) OnInit() {
	it.BatchID = -1
	it.TableID = -1
	it.ChairID = -1
	it.Status = StatusFree
}

func (it *Item) GetUserInfo() *Info {
	return &it.Info
}

func (it *Item) SitDown(tableID int32, chairID int32) {
	it.TableID = tableID
	it.ChairID = chairID
	it.Status = StatusPlaying
}

func (it *Item) StandUp() {
	it.TableID = -1
	it.ChairID = -1
	it.Status = StatusFree
}

func (it *Item) IsRobot() bool {
	return it.BatchID != -1
}
