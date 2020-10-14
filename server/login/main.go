package main

import (
	_ "github.com/go-sql-driver/mysql"
	"xj_game_server/server/login/gate"
	"xj_game_server/server/login/login"
	"xj_game_server/util/leaf"
)

func main() {
	leaf.Run(
		login.Module,
		gate.Module,
	)
}
