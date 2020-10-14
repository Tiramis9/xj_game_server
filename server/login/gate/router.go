package gate

import (
	"xj_game_server/server/login/login"
	"xj_game_server/server/login/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Login_C_Wechat{}, login.ChanRPC)  //微信登陆消息路由
	msg.Processor.SetRouter(&msg.Login_C_Mobile{}, login.ChanRPC)  //手机登陆消息路由
	msg.Processor.SetRouter(&msg.Login_C_Visitor{}, login.ChanRPC) //游客登陆消息路由
}
