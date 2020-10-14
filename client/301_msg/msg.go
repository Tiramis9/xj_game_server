package msg

import (
	"xj_game_server/util/leaf/network/protobuf"
)

var Processor = protobuf.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_TokenLogin{})   //token登陆消息
	Processor.Register(&Game_C_UserSitDown{})  //用户坐下消息
	Processor.Register(&Game_C_UserStandUp{})  //用户起立消息
	Processor.Register(&Game_C_UserFire{})     //用户开炮消息
	Processor.Register(&Game_C_CatchFish{})    //捕获消息
	Processor.Register(&Game_C_ChangeBullet{}) //变炮消息
	Processor.Register(&Game_C_LockFish{})     //用户锁定消息

	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})      //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{})   //登陆成功消息
	Processor.Register(&Game_S_FreeScene{})      //空闲场景消息
	Processor.Register(&Game_S_PlayScene{})      //游戏场景消息
	Processor.Register(&Game_S_GroupFishScene{}) //鱼群场景消息
	Processor.Register(&Game_S_OnLineNotify{})   //用户上线通知消息
	Processor.Register(&Game_S_OffLineNotify{})  //用户掉线通知消息
	Processor.Register(&Game_S_StandUpNotify{})  //起立通知消息
	Processor.Register(&Game_S_UserFire{})       //用户开炮通知消息
	Processor.Register(&Game_S_CatchFish{})      //捕获通知消息
	Processor.Register(&Game_S_ChangeBullet{})   //变炮通知消息
	Processor.Register(&Game_S_LockFish{})       //用户锁定消息
	Processor.Register(&Game_S_FishList{})       //生成鱼消息
	Processor.Register(&Game_S_GroupFish{})      //鱼群通知消息

	//机器人-----
	Processor.Register(&Game_C_RobotLogin{}) //机器人登陆
}
