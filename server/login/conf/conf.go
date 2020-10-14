package conf

import (
	"log"
	"xj_game_server/public"
	publicConfig "xj_game_server/public/config"
	"xj_game_server/util/leaf/conf"
	"time"
)

var (
	// log conf
	LogFlag = log.LstdFlags

	// gate conf
	PendingWriteNum        = 2000
	MaxMsgLen       uint32 = 10 * 1024
	HTTPTimeout            = 10 * time.Second
	LenMsgLen              = 2
	LittleEndian           = false

	// skeleton conf
	GoLen              = 10000
	TimerDispatcherLen = 10000
	AsynCallLen        = 10000
	ChanRPCLen         = 10000
)

// Service 服务端配置
type Service struct {
	ServerUrl   string `yaml:"server_url"`
	TcpUrl      string `yaml:"tcp_url"`
	MaxConnNum  int    `yaml:"max_conn_num"`
	WSAddr      string `yaml:"ws_addr"`
	WSUrl       string `yaml:"ws_url"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
	ConsolePort int    `yaml:"console_port"`
	ProfilePath string `yaml:"profile_path"`
}
type Config struct {
	Service Service `yaml:"service"`
}

func GetServer() Service {
	return config.Service
}

var config Config

func init() {
	publicConfig.InitConfig(public.LoginConfigYmlPath, &config)
	conf.Post = config.Service.TcpUrl
}
