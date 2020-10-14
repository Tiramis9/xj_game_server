package mysql

import (
	"database/sql"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/jinzhu/gorm"
	golog "log"
	"xj_game_server/public/config"
	"xj_game_server/util/leaf/log"
	"xorm.io/core"
)

var Client *Mysql

//xj_game_db
type Mysql struct {
	GetXJGameDB *gorm.DB
}

func init() {
	Client = new(Mysql)
}


func (self *Mysql) OnInit() {
	//连接用户数据库
	var err error

	// 连接数据库(新)
	args := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.GetXJGameDB().User, config.GetXJGameDB().PassWd, config.GetXJGameDB().Host, config.GetXJGameDB().Db)
	//连接用户数据库
	self.GetXJGameDB, err = gorm.Open(config.GetXJGameDB().Dialect, args)
	if err != nil {
		golog.Fatalf("init db err %v \n", err)
	}
	self.GetXJGameDB.Model(config.GetXJGameDB().EnableLog)
	self.GetXJGameDB.SetLogger(dbLogger{})
	self.GetXJGameDB.SingularTable(true)
	// //用于设置最大打开的连接数，默认值为0表示不限制。
	self.GetXJGameDB.DB().SetMaxOpenConns(config.GetXJGameDB().MaxOpenConnections)
	//设置连接池的空闲数大小
	self.GetXJGameDB.DB().SetMaxIdleConns(config.GetXJGameDB().MaxIdleConnections)


	syncTable()
}


func (self *Mysql) OnDestroy() {
	//关闭用户数据库连接
	err := self.GetXJGameDB.Close()
	if err != nil {
		_ = log.Logger.Errorf("OnDestroy GetXJGameDB err %v", err)
		return
	}
}
func (self *Mysql) Query(db *xorm.Engine, query string, args ...interface{}) (*sql.Rows, error) {
	db.DB().DB.Query("delimiter //")
	rows, err := db.DB().DB.Query(query, args...)
	db.DB().DB.Query("//")
	return rows, err
}

func syncTable() {
	//同步表结构
}

type dbLogger struct {
	isShow bool
}

func (logger dbLogger) Print(v ...interface{}) {
	log.Logger.Info(v)
}

func (logger dbLogger) Level() core.LogLevel {
	return core.LOG_INFO
}

func (logger dbLogger) SetLevel(l core.LogLevel) {
}

func (logger dbLogger) ShowSQL(show ...bool) {
}

func (logger dbLogger) IsShowSQL() bool {
	return logger.isShow
}

func (logger dbLogger) Debug(v ...interface{}) {
	log.Logger.Debug(v)
}

func (logger dbLogger) Debugf(format string, v ...interface{}) {
	log.Logger.Debugf(format, v)
}

func (logger dbLogger) Error(v ...interface{}) {
	_ = log.Logger.Error(v)
}

func (logger dbLogger) Errorf(format string, v ...interface{}) {
	_ = log.Logger.Errorf(format, v)
}

func (logger dbLogger) Info(v ...interface{}) {
	log.Logger.Info(v)
}

func (logger dbLogger) Warn(v ...interface{}) {
	_ = log.Logger.Warn(v)
}

func (logger dbLogger) Warnf(format string, v ...interface{}) {
	_ = log.Logger.Warnf(format, v)
}

func (logger dbLogger) Infof(format string, v ...interface{}) {
	log.Logger.Infof(format, v)
}
