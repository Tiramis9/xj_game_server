package internal

import (
	"xj_game_server/game/104_senglinwuhui/conf"
	"xj_game_server/game/104_senglinwuhui/game/table"
	"xj_game_server/game/public/grpc"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/redis"
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

	mysql.GameClient.InitGameConfig() //初始化游戏配置
	redis.GameClient.RegisterGame()   //注册游戏
	redis.GameClient.RegisterGrpc()   //注册grpc

	grpc.OnInit(conf.GetServer().GRPcUrl)
	table.OnInit()
}

func (m *Module) OnDestroy() {
	mysql.GameClient.OnDestroy()
	redis.GameClient.OnDestroy()
}
