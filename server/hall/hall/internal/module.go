package internal

import (
	"xj_game_server/public/base"
	"xj_game_server/server/hall/db"
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
	db.HallRedisClient.RegisterLogin()
}

func (m *Module) OnDestroy() {
	db.HallRedisClient.CancelLogin()
	db.HallMysqlClient.OnDestroy()
	db.HallRedisClient.OnDestroy()
}
