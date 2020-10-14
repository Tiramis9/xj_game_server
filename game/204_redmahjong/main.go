package main

import (
	_ "github.com/go-sql-driver/mysql"
	"xj_game_server/game/204_redmahjong/game"
	"xj_game_server/game/204_redmahjong/gate"
	"xj_game_server/game/204_redmahjong/robot"
	"xj_game_server/util/leaf"
)

func main() {
	leaf.Run(
		game.Module,  // 游戏逻辑模块
		gate.Module,  // 网关模块
		robot.Module, // 机器人模块
	)
}
