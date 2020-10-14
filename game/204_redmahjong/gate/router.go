package gate

import (
	"xj_game_server/game/204_redmahjong/game"
	"xj_game_server/game/204_redmahjong/msg"
)

func init() {
	//客户端路由
	msg.Processor.SetRouter(&msg.Game_C_TokenLogin{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserSitDown{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserStandUp{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserOutCard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserOperate{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserPrepare{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserUnPrepare{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UserTing{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_ChangeTable{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_AutoManage{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Game_C_UnAutoManage{}, game.ChanRPC)
}
