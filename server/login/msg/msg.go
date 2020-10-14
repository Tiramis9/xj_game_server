package msg

import (
	"xj_game_server/util/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	Processor.Register(&Login_C_Wechat{})  //微信登陆消息
	Processor.Register(&Login_C_Mobile{})  //手机登陆消息
	Processor.Register(&Login_C_Visitor{}) //游客登陆消息

	Processor.Register(&Login_S_Success{}) //登陆成功消息
	Processor.Register(&Login_S_Fail{})    //登陆失败消息

}
