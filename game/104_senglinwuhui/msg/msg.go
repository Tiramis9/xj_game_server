/*
 * @Author: yhlyl
 * @Date: 2019-12-25 11:02:08
 * @LastEditTime: 2020-01-04 14:20:36
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/104_senglinwuhui/msg/msg.go
 * @https://github.com/android-coco
 */
package msg

import (
	"xj_game_server/util/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_LoginDown{})   //token登陆消息
	Processor.Register(&Game_C_UserSitDown{}) //用户坐下消息
	Processor.Register(&Game_C_UserStandUp{}) //用户起立消息
	Processor.Register(&Game_C_UserJetton{})  //用户下注消息

	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})    //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{}) //登陆成功消息
	Processor.Register(&Game_S_JettonScene{})  //下注场景消息
	Processor.Register(&Game_S_LotteryScene{}) //开奖场景消息
	Processor.Register(&Game_S_ConcludeScene{})
	Processor.Register(&Game_S_GameStart{})     //游戏开始消息
	Processor.Register(&Game_S_GameConclude{})  //游戏结束消息
	Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
	Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
	Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息
	Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息
	Processor.Register(&Game_S_UserJetton{})    //下注通知
	// Processor.Register(&Game_S_Hall{})          //游戏结束发送大厅场景

	//机器人-----
	Processor.Register(&Game_C_RobotLogin{}) //机器人登陆

	Processor.Register(&Game_S_AreaJetton{}) //当前下注状况,每个区域,每个玩家的下注情况
	Processor.Register(&Game_C_UserList{})   //获取用户列表客户端参数
	Processor.Register(&Game_S_UserList{})   //获取用户列表服务器返回
}
