package leaf

import (
	"fmt"
	golog "log"
	"os"
	"os/signal"
	"xj_game_server/util/leaf/cluster"
	"xj_game_server/util/leaf/conf"
	"xj_game_server/util/leaf/console"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/module"
)

func Run(mods ...module.Module) {

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// cluster
	cluster.Init()

	// console
	console.Init()

	log.Logger.Info("|-----------------------------------|")
	log.Logger.Info("|            Leaf " + version + "       |")
	log.Logger.Info("|-----------------------------------|")
	log.Logger.Info("|  Go Leaf Server Start Successful  |")
	log.Logger.Info("|  TcpPort" + conf.Post + "   Pid:" + fmt.Sprintf("%d", os.Getpid()) + "       |")
	log.Logger.Info("|-----------------------------------|")
	golog.Println("|-----------------------------------|")
	golog.Println("|            Leaf " + version + "       |")
	golog.Println("|-----------------------------------|")
	golog.Println("|  Go Leaf Server Start Successful  |")
	golog.Println("|    TcpPort" + conf.Post + "   Pid:" + fmt.Sprintf("%d", os.Getpid()) + "       |")
	golog.Println("|-----------------------------------|")
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Logger.Infof("Leaf closing down (signal: %v)", sig)

	console.Destroy()
	cluster.Destroy()
	module.Destroy()
}
