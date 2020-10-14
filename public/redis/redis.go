package redis

import (
	"github.com/go-redis/redis/v7"
	golog "log"
	"xj_game_server/public/config"
	"xj_game_server/util/leaf/log"
)

var Client = new(Redis)

type Redis struct {
	Client *redis.Client
}

func (self *Redis) OnInit() {
	//连接redis
	client := redis.NewClient(&redis.Options{
		Password: config.GetRedis().PassWd,
		Addr:     config.GetRedis().Host,
		DB:       config.GetRedis().Db,
	})
	_, err := client.Ping().Result()
	if err != nil {
		_ = log.Logger.Errorf("redis err %v", err)
		golog.Fatalf("redis err %v", err)
		return
	}
	self.Client = client
}

func (self *Redis) OnDestroy() {
	//关闭redis连接
	err := self.Client.Close()
	if err != nil {
		_ = log.Logger.Errorf("redis err %v", err)
		golog.Fatalf("redis err %v", err)
		return
	}
}
