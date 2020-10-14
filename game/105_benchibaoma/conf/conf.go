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

	GameJettonTime  int32     `yaml:"game_jetton_time"`  //下注持续时间单位s
	GameLotteryTime int32     `yaml:"game_lottery_time"` //开奖持续时间单位s
	JettonList      []float32 `yaml:"jetton_list,flow"`  //筹码选项列表
}
type Config struct {
	Service Service `yaml:"service"`
}

func GetServer() Service {
	return config.Service
}

var config Config

func init() {
	publicConfig.InitConfig(public.BenChiBaoMaConfigYmlPath105, &config)
	defaultInit()
}

func defaultInit() {
	conf.Post = config.Service.TCPUrl
	if config.Service.GameJettonTime == 0 {
		config.Service.GameJettonTime = 20
	}
	if config.Service.GameLotteryTime == 0 {
		config.Service.GameLotteryTime = 10
	}
	publicGameConfig.InitService(&config.Service.Service)
}
