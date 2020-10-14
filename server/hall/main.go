package main

import (
	_ "github.com/go-sql-driver/mysql"
	"xj_game_server/server/hall/gate"
	"xj_game_server/server/hall/hall"
	"xj_game_server/util/leaf"
)

func main() {
	leaf.Run(
		hall.Module,
		gate.Module,
	)
}
