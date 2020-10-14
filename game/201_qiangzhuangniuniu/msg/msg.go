package msg

import (
	"xj_game_server/util/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_TokenLogin{})  //token登陆消息
	Processor.Register(&Game_C_UserSitDown{}) //用户坐下消息 2.1 坐下
	Processor.Register(&Game_C_UserStandUp{}) //用户起立消息 --未使用
	Processor.Register(&Game_C_UserQZ{})      //用户抢庄消息4.1
	Processor.Register(&Game_C_UserJetton{})  //用户下注消息5.1 叫倍环节
	Processor.Register(&Game_C_UserTP{})      //用户摊牌消息6.1  开牌环节

	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
	Processor.Register(&Game_S_FreeScene{})     //空闲场景消息8.1  发牌场景
	Processor.Register(&Game_S_QZScene{})       //抢庄场景消息8.2  抢庄场景
	Processor.Register(&Game_S_JettonScene{})   //下注场景消息8.3  叫倍场景
	Processor.Register(&Game_S_TPScene{})       //摊牌场景消息8.4  开牌场景
	Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
	Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
	Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息 --未使用
	Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息 // 2.2 坐下成功通知,匹配成功开始游戏

	Processor.Register(&Game_S_CardRound{}) //开始定时器通知消息 3  发牌环节
	Processor.Register(&Game_S_CallRound{}) //抢庄环节(可抢庄倍数配置)4.1
	Processor.Register(&Game_S_BetRound{})  //叫倍环节（可叫倍数配置） 5.1
	Processor.Register(&Game_S_ShowRound{}) //开牌环节 6.1

	Processor.Register(&Game_S_GameCard{})     //发牌通知 3.1  发牌通知(下注后)
	Processor.Register(&Game_S_UserQZ{})       //抢庄通知消息4.2
	Processor.Register(&Game_S_GameDZ{})       //定庄通知消息4.3
	Processor.Register(&Game_S_UserJetton{})   //下注通知消息5.2  叫倍通知
	Processor.Register(&Game_S_UserTP{})       //摊牌通知消息6.2  开牌通知
	Processor.Register(&Game_S_GameConclude{}) //结束游戏通知消息7.结算/结束游戏/起立
	//机器人-----
	Processor.Register(&Game_C_RobotLogin{}) //机器人登陆
}
