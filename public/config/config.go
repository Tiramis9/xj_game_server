package config

import (
	yaml2 "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"xj_game_server/public"
)

// DB 数据库配置
type Db struct {
	EnableLog          bool   `yaml:"enable_log" json:"enable_log"`
	Dialect            string `yaml:"dialect" json:"dialect"`
	Host               string `yaml:"host" json:"host"`
	User               string `yaml:"user" json:"user"`
	PassWd             string `yaml:"pass" json:"pass"`
	Db                 string `yaml:"db" json:"db"`
	MaxOpenConnections int    `yaml:"max_open_connections" json:"max_open_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections" json:"max_idle_connections"`
}
type LogConfig struct {
	Path string `yaml:"path"`
}

// redis
type Redis struct {
	Host   string `yaml:"host"`
	PassWd string `yaml:"pass"`
	Db     int    `yaml:"db"`
}

// 公用 配置
type Global struct {
	XJGameDB Db `yaml:"xj_game_db"`

	Redis Redis     `yaml:"redis"`
	Log   LogConfig `yaml:"log"`
}


func GetXJGameDB() Db {
	return config.XJGameDB
}


func GetRedis() Redis {
	return config.Redis
}

func GetLogConfigPath() LogConfig {
	return config.Log
}

/**
 * @description:
 * @param {type}
 * @return:
 */
var config Global

func init() {
	InitConfig(public.GlobalConfigYmlPath, &config)
}

// InitConfig 初始化config
func InitConfig(path string, config interface{}) {
	pathStr, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	yamlFile, err := ioutil.ReadFile(pathStr + path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml2.Unmarshal(yamlFile, config)
	if err != nil {
		log.Printf("Unmarshal: %v", err)
	}
}
