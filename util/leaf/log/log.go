package log

import (
	"github.com/cihub/seelog"
	"os"
	"path/filepath"
	"xj_game_server/public/config"
)

var Logger seelog.LoggerInterface

func init() {
	defer seelog.Flush()
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//初始化全局变量Logger为seelog的禁用状态，主要为了防止Logger被多次初始化
	_ = seelog.ReplaceLogger(Logger)
	Logger, _ = seelog.LoggerFromConfigAsFile(path + config.GetLogConfigPath().Path)
}
