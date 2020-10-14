package conf

import (
	"log"
	"time"
)

var (
	// log conf
	LogFlag = log.LstdFlags

	// gate conf
	PendingWriteNum        = 100000
	MaxMsgLen       uint32 = 8192
	HTTPTimeout            = 10 * time.Second
	LenMsgLen              = 2
	LittleEndian           = false

	// skeleton conf
	GoLen              = 10000
	TimerDispatcherLen = 10000
	AsynCallLen        = 10000
	ChanRPCLen         = 10000
)

type Service struct {
	ServerUrl     string        `yaml:"server_url"`
	TCPUrl        string        `yaml:"tcp_url"`
	MaxConnNum    int           `yaml:"max_conn_num"`
	WSAddr        string        `yaml:"ws_addr"`
	WSUrl         string        `yaml:"ws_url"`
	CertFile      string        `yaml:"cert_file"`
	KeyFile       string        `yaml:"key_file"`
	ConsolePort   int           `yaml:"console_port"`
	ProfilePath   string        `yaml:"profile_path"`
	KindID        int32         `yaml:"kind_id"`        //游戏种类ID
	GameID        int32         `yaml:"game_id"`        //游戏ID
	GRPCAddr      string        `yaml:"grpc_addr"`      //grpc地址
	GRPcUrl       string        `yaml:"grpc_url"`       //grpc地址
	HeartbeatTime time.Duration `yaml:"heartbeat_time"` //心跳时间 单位秒
}

var s *Service

func InitService(service *Service) {
	s = service
}

func GetService() *Service {
	return s
}
