package robot

import (
	"sync"
)

//机器人配置信息
type Config struct {
	BatchID     int32 `json:"batch_id"`
	GameID      int32 `json:"game_id"`
	ServiceMode int32 `json:"service_mode"`
	RobotCount  int64 `json:"robot_count"`
	//EnterTime      int64   `json:"enter_time"`
	//LeaveTime      int64   `json:"leave_time"`
	TakeMinCoin    float32 `json:"take_min_coin"`
	TakeMaxCoin    float32 `json:"take_max_coin"`
	TakeMinDiamond float32 `json:"take_min_diamond"`
	TakeMaxDiamond float32 `json:"take_max_diamond"`
}

func (self *Config) TableName() string {
	return "robot_config"
}

var RobotConfigItem = new(ConfigItem)

//机器人配置接口
type ConfigItem struct {
	config map[int32]Config //批次id----> 配置
	l      sync.Mutex
}

//初始化
func (self *ConfigItem) OnInit(config []Config) {
	//fmt.Println("加载机器人")
	self.l.Lock()
	defer self.l.Unlock()

	self.config = make(map[int32]Config)
	for _, v := range config {
		self.config[v.BatchID] = v
	}
}

//获取机器人配置
func (self *ConfigItem) GetConfig() map[int32]Config {
	self.l.Lock()
	defer self.l.Unlock()

	return self.config
}
