package gate

import (
	"xj_game_server/game/103_bairenniuniu/game"
	"xj_game_server/game/103_bairenniuniu/msg"
)

func init() {
	//客户端路由
	msg.Processor.SetRouter(&msg.Game_C_LoginDown{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_TokenLogin{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserSitDown{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserStandUp{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserJetton{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserList{}, game.ChanRPC)
}
