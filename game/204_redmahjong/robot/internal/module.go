package internal

import (
	"xj_game_server/game/204_redmahjong/robot/logic"
	"xj_game_server/public/base"
	"xj_game_server/util/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	logic.Client.OnInit()
}

func (m *Module) OnDestroy() {

}