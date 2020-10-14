package gate

import (
	"xj_game_server/server/hall/hall"
	"xj_game_server/server/hall/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Hall_C_Msg{}, hall.ChanRPC)  //微信登陆消息路由
}
