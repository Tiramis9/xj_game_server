package msg

import (
	"xj_game_server/util/leaf/network/protobuf"
)

var Processor = protobuf.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_TokenLogin{})  //token登陆消息
	Processor.Register(&Game_C_UserSitDown{}) //用户坐下消息
	Processor.Register(&Game_C_UserStandUp{}) //用户起立消息
	Processor.Register(&Game_C_UserQZ{})      //用户抢庄消息
	Processor.Register(&Game_C_UserJetton{})  //用户下注消息
	Processor.Register(&Game_C_UserTP{})      //用户摊牌消息

	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
	Processor.Register(&Game_S_FreeScene{})     //空闲场景消息
	Processor.Register(&Game_S_QZScene{})       //抢庄场景消息
	Processor.Register(&Game_S_JettonScene{})   //下注场景消息
	Processor.Register(&Game_S_TPScene{})       //摊牌场景消息
	Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
	Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
	Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息
	Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息
	Processor.Register(&Game_S_StartTime{})     //开始定时器通知消息
	Processor.Register(&Game_S_GameConclude{})  //结束游戏通知消息
	Processor.Register(&Game_S_UserQZ{})        //抢庄通知消息
	Processor.Register(&Game_S_GameDZ{})        //定庄通知消息
	Processor.Register(&Game_S_UserJetton{})    //下注通知消息
	Processor.Register(&Game_S_UserTP{})        //摊牌通知消息

	//机器人-----
	Processor.Register(&Game_C_RobotLogin{}) //机器人登陆
}
