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

	GameStartQzTime int32 `yaml:"game_start_qz_time"` //抢庄持续时间单位s
	GameQZTime      int32 `yaml:"game_qz_time"`       //抢庄持续时间单位st
	GameJettonTime  int32 `yaml:"game_jetton_time"`   //下注持续时间单位s
	GameTPTime      int32 `yaml:"game_tp_time"`       //摊牌持续时间单位s
	GameStartTime   int32 `yaml:"game_start_time"`    //开始时间时间单位s 发牌 默认3s
	GameShuffleTime int32 `yaml:"game_shuffle_time"`  //开始游戏时间单位s 洗牌  默认8s

	MultipleList []int32 `yaml:"multiple_list"`
	JettonList   []int32 `yaml:"jetton_list"`
}
type Config struct {
	Service Service `yaml:"service"`
}

func GetServer() Service {
	return config.Service
}

var config Config

func init() {
	publicConfig.InitConfig(public.QiangZhuangNiuNiuConfigYmlPath201, &config)
	defaultInit()
}

func defaultInit() {
	conf.Post = config.Service.TCPUrl
	if config.Service.GameQZTime == 0 {
		config.Service.GameQZTime = 5
	}
	if config.Service.GameJettonTime == 0 {
		config.Service.GameJettonTime = 5
	}
	if config.Service.GameTPTime == 0 {
		config.Service.GameTPTime = 5
	}

	if config.Service.GameStartQzTime == 0 {
		config.Service.GameStartQzTime = 5
	}
	// 初始化抢庄倍数

	if len(config.Service.MultipleList) == 0 {
		config.Service.MultipleList = append(config.Service.MultipleList, 0, 1, 2, 3)
	}
	if len(config.Service.JettonList) == 0 {
		config.Service.JettonList = append(config.Service.JettonList, 1, 2, 5, 10, 15)
	}

	if config.Service.GameShuffleTime == 0 {
		config.Service.GameShuffleTime = 8
	}

	publicGameConfig.InitService(&config.Service.Service)
}
