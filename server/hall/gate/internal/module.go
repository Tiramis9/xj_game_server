package internal

import (
	"xj_game_server/server/hall/conf"
	"xj_game_server/server/hall/hall"
	"xj_game_server/server/hall/msg"
	"xj_game_server/util/leaf/gate"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.GetServer().MaxConnNum,
		PendingWriteNum: conf.PendingWriteNum,
		MaxMsgLen:       conf.MaxMsgLen,
		WSAddr:          conf.GetServer().WSAddr,
		HTTPTimeout:     conf.HTTPTimeout,
		CertFile:        conf.GetServer().CertFile,
		KeyFile:         conf.GetServer().KeyFile,
		TCPAddr:         conf.GetServer().TcpUrl,
		LenMsgLen:       conf.LenMsgLen,
		LittleEndian:    conf.LittleEndian,
		Processor:       msg.Processor,
		AgentChanRPC:    hall.ChanRPC,
	}
}
