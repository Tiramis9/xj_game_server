package internal

import (
	"xj_game_server/game/203_doudizhu/conf"
	"xj_game_server/game/203_doudizhu/game"
	"xj_game_server/game/203_doudizhu/msg"
	publicConf "xj_game_server/game/public/conf"
	"xj_game_server/util/leaf/gate"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.GetServer().MaxConnNum,
		PendingWriteNum: publicConf.PendingWriteNum,
		MaxMsgLen:       publicConf.MaxMsgLen,
		WSAddr:          conf.GetServer().WSAddr,
		HTTPTimeout:     publicConf.HTTPTimeout,
		CertFile:        conf.GetServer().CertFile,
		KeyFile:         conf.GetServer().KeyFile,
		TCPAddr:         conf.GetServer().TCPUrl,
		LenMsgLen:       publicConf.LenMsgLen,    // 1,2,4
		LittleEndian:    publicConf.LittleEndian, // true false
		Processor:       msg.Processor,
		AgentChanRPC:    game.ChanRPC,
	}
}
