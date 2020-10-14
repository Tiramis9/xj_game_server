package msg

import (
	"xj_game_server/util/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_TokenLogin{})       //token登陆消息
	Processor.Register(&Game_C_UserSitDown{})      //用户坐下消息
	Processor.Register(&Game_C_UserStandUp{})      //用户起立消息
	Processor.Register(&Game_C_UserPrepare{})      //用户准备消息
	Processor.Register(&Game_C_UserUnPrepare{})    //用户取消准备消息
	Processor.Register(&Game_C_UserGrabLandlord{}) //用户叫分消息
	Processor.Register(&Game_C_UserCP{})           //用户出牌消息
	Processor.Register(&Game_C_UserPass{})         //过
	Processor.Register(&Game_C_ChangeTable{})      //换桌
	Processor.Register(&Game_C_AutoManage{})       // 托管
	Processor.Register(&Game_C_UnAutoManage{})     // 取消托管
	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})         //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{})      //登陆成功消息
	Processor.Register(&Game_S_FreeScene{})         //空闲场景消息
	Processor.Register(&Game_S_PlayScene{})         //出牌场景消息
	Processor.Register(&Game_S_GrabLandlordScene{}) //摊牌场景消息
	Processor.Register(&Game_S_OnLineNotify{})      //用户上线通知消息
	Processor.Register(&Game_S_OffLineNotify{})     //用户掉线通知消息
	Processor.Register(&Game_S_StandUpNotify{})     //起立通知消息
	Processor.Register(&Game_S_SitDownNotify{})     //坐下通知消息
	Processor.Register(&Game_S_StartGame{})         //开始游戏通知消息
	Processor.Register(&Game_S_StartCPDetermine{})  //开始出牌通知消息
	Processor.Register(&Game_S_GameConclude{})      //结束游戏通知消息
	Processor.Register(&Game_S_UserPrepare{})       //准备通知消息
	Processor.Register(&Game_S_UserUnPrepare{})     //准备通知消息
	Processor.Register(&Game_S_UserGrabLandlord{})  //叫分通知消息
	Processor.Register(&Game_S_UserCP{})            //出牌通知消息
	Processor.Register(&Game_S_UserPass{})          //过牌通知消息
	Processor.Register(&Game_S_CurrentUser{})       //当前用户通知消息
	Processor.Register(&Game_S_AutoManage{})        // 通知用户托管消息
	Processor.Register(&Game_S_UnAutoManage{})      // 通知用户取消托管消息
	Processor.Register(&Game_S_GameRestart{})       // 通知用户没人叫分,重新发牌
	//机器人-----
	Processor.Register(&Game_C_RobotLogin{}) //机器人登陆

}
