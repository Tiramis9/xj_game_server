package internal

import (
	"xj_game_server/server/login/db"
	"xj_game_server/server/login/msg"
	"xj_game_server/util/leaf/gate"
	"reflect"
)

func init() {
	handler(&msg.Login_C_Wechat{}, handleWechatLogin)
	handler(&msg.Login_C_Mobile{}, handleMobileLogin)
	handler(&msg.Login_C_Visitor{}, handleVisitorLogin)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//处理微信登陆
func handleWechatLogin(args []interface{}) {
	m := args[0].(*msg.Login_C_Wechat)
	a := args[1].(gate.Agent)
	//用户登陆
	message := db.LoginMysqlClient.WechatLogin(a, m)
	if reflect.TypeOf(message) == reflect.TypeOf(&msg.Login_S_Success{}) {
		login_S_Success := message.(*msg.Login_S_Success)
		//生成Token
		login_S_Success.Token = db.LoginRedisClient.MakeToken(login_S_Success.UserID)
		//加载游戏列表
		login_S_Success.GameInfoList = db.LoginRedisClient.LoadGameList()
		a.WriteMsg(message)
		a.Close()
		return
	}
	a.WriteMsg(message)
	a.Close()
}

//处理手机登陆
func handleMobileLogin(args []interface{}) {
	m := args[0].(*msg.Login_C_Mobile)
	a := args[1].(gate.Agent)

	message := db.LoginMysqlClient.MobileLogin(a, m)
	if reflect.TypeOf(message) == reflect.TypeOf(&msg.Login_S_Success{}) {
		login_S_Success := message.(*msg.Login_S_Success)
		//生成Token
		login_S_Success.Token = db.LoginRedisClient.MakeToken(login_S_Success.UserID)
		//加载游戏列表
		login_S_Success.GameInfoList = db.LoginRedisClient.LoadGameList()
		a.WriteMsg(message)
		a.Close()
		return
	}
	a.WriteMsg(message)
	a.Close()
}

//处理游客登陆
func handleVisitorLogin(args []interface{}) {
	m := args[0].(*msg.Login_C_Visitor)
	a := args[1].(gate.Agent)
	message := db.LoginMysqlClient.VisitorLogin(a, m)
	if reflect.TypeOf(message) == reflect.TypeOf(&msg.Login_S_Success{}) {
		login_S_Success := message.(*msg.Login_S_Success)
		//生成Token
		login_S_Success.Token = db.LoginRedisClient.MakeToken(login_S_Success.UserID)
		//加载游戏列表
		login_S_Success.GameInfoList = db.LoginRedisClient.LoadGameList()
		a.WriteMsg(message)
		a.Close()
		return
	}
	a.WriteMsg(message)
	a.Close()
}
