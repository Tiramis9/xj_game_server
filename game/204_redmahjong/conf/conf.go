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

	GameStartTime  	int32 	`yaml:"game_start_time"`  	//开始时间时间单位s
	GameOutCardTime int32	`yaml:"game_out_card_time"`	//出牌持续时间单位s
	GameOperateTime int32	`yaml:"game_operate_time"` 	//操作持续时间单位s
	MaCount			int32	`yaml:"ma_count"`			//码数
}
type Config struct {
	Service Service `yaml:"service"`
}

func GetServer() Service {
	return config.Service
}

var config Config

func init() {
	publicConfig.InitConfig(public.HZMJConfigYmlPath204, &config)
	defaultInit()
}

func defaultInit() {
	conf.Post = config.Service.TCPUrl

	if config.Service.GameStartTime == 0 {
		config.Service.GameStartTime = 5
	}

	if config.Service.GameOutCardTime == 0 {
		config.Service.GameOutCardTime = 10
	}

	if config.Service.GameOperateTime == 0 {
		config.Service.GameOperateTime = 5
	}

	if config.Service.MaCount == 0 {
		config.Service.MaCount = 8
	}

	publicGameConfig.InitService(&config.Service.Service)
}
