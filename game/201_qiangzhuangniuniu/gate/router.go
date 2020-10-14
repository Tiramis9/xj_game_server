package gate

import (
	"xj_game_server/game/201_qiangzhuangniuniu/game"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
)

func init() {
	//客户端路由
	msg.Processor.SetRouter(&msg.Game_C_TokenLogin{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserSitDown{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserStandUp{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserQZ{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserJetton{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserTP{}, game.ChanRPC)
}
