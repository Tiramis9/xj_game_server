package main

import (
	_ "github.com/go-sql-driver/mysql"
	"xj_game_server/game/103_bairenniuniu/game"
	"xj_game_server/game/103_bairenniuniu/gate"
	"xj_game_server/game/103_bairenniuniu/robot"
	"xj_game_server/util/leaf"
)

func main() {
	leaf.Run(
		game.Module,  // 游戏逻辑模块
		gate.Module,  // 网关模块
		robot.Module, // 机器人模块
	)
}
