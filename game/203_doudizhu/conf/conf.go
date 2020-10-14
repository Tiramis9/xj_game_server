package conf

import (
	publicGameConfig "xj_game_server/game/public/conf"
	"xj_game_server/public"
	publicConfig "xj_game_server/public/config"
	"xj_game_server/util/leaf/conf"
)

// Service 服务端配置
type Service struct {
	publicGameConfig.Service //游戏公用配置

	GameStartTime int32 `yaml:"game_start_time"` 		//开始时间单位s
	GameJFTime int32 `yaml:"game_jf_time"`		//叫分持续时间单位s
	GameCPTime  int32 `yaml:"game_cp_time"`  	//出牌持续时间单位s
}

type Config struct {
	Service Service `yaml:"service"`
}

func GetServer() Service {
	return config.Service
}

var config Config

func init() {
	publicConfig.InitConfig(public.DouDiZhuConfigYmlPath203, &config)
	defaultInit()
}

func defaultInit() {
	conf.Post = config.Service.TCPUrl
	if config.Service.GameStartTime == 0 {
		config.Service.GameStartTime = 5
	}

	if config.Service.GameJFTime == 0 {
		config.Service.GameJFTime = 5
	}

	if config.Service.GameCPTime == 0 {
		config.Service.GameCPTime = 15
	}


	publicGameConfig.InitService(&config.Service.Service)
}
