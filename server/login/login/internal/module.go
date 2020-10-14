package internal

import (
	"xj_game_server/public/base"
	"xj_game_server/server/login/db"
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
	db.LoginRedisClient.RegisterLogin()
}

func (m *Module) OnDestroy() {
	db.LoginRedisClient.CancelLogin()
	db.LoginMysqlClient.OnDestroy()
	db.LoginRedisClient.OnDestroy()
}
