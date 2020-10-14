package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"xj_game_server/public/mysql"
	"xj_game_server/server/login/conf"
	"xj_game_server/server/login/msg"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"strings"
	"time"
)

var LoginMysqlClient *LoginMysql

//var LoginMysqlClient *LoginMysql

type LoginMysql struct {
	*mysql.Mysql
	//*mysql.NativeMysql
}

func init() {
	mysql.Client.OnInit()
	LoginMysqlClient = &LoginMysql{
		Mysql: mysql.Client,
	}

}

func (self *LoginMysql) OnDestroy() {
	//关闭用户数据库连接
	self.Mysql.OnDestroy()
}

//微信登陆
func (self *LoginMysql) WechatLogin(agent gate.Agent, message *msg.Login_C_Wechat) interface{} {
	host, _, _ := net.SplitHostPort(agent.RemoteAddr().String())
	rows, err := self.Query(self.AccountDB, "Call LSP_WechatLogin(?, ?, ?, ?, ?, ?, ?, ?,?)", message.AgentID, message.UserUin, message.Gender, message.NikeName, message.HeadImageUrl,
		message.MachineID, message.DeviceType, host,strings.Split(conf.GetServer().ServerUrl, ":")[0])
	defer func() {
		if rows != nil {
			err := rows.Close()
			if err != nil {
				_ = log.Logger.Errorf("WechatLogin err%v", err)
			}
		}
	}()
	if err != nil {
		_ = log.Logger.Errorf("WechatLogin err %v", err)
		return &msg.Login_S_Fail{
			ErrorCode: -1,
			ErrorMsg:  err.Error(),
		}
	}
	return self.loadUserInfo(rows)
}

//手机登陆
func (self *LoginMysql) MobileLogin(agent gate.Agent, message *msg.Login_C_Mobile) interface{} {
	host, _, _ := net.SplitHostPort(agent.RemoteAddr().String())
	rows, err := self.Query(self.AccountDB, "Call LSP_MobileLogin(?, ?, ?, ?, ?,?)", message.PhoneNumber, message.Password,
		message.MachineID, message.DeviceType, host,strings.Split(conf.GetServer().ServerUrl, ":")[0])
	defer func() {
		if rows != nil {
			err := rows.Close()
			if err != nil {
				_ = log.Logger.Errorf("MobileLogin err%v", err)
			}
		}
	}()
	if err != nil {
		_ = log.Logger.Errorf("MobileLogin err %v", err)
		return &msg.Login_S_Fail{
			ErrorCode: -1,
			ErrorMsg:  err.Error(),
		}
	}
	return self.loadUserInfo(rows)
}

//游客登陆
func (self *LoginMysql) VisitorLogin(agent gate.Agent, message *msg.Login_C_Visitor) interface{} {
	var ip string
	if agent == nil {
		ip = "127.0.0.1:1001"
	} else {
		ip = agent.RemoteAddr().String()
	}

	rows, err := self.Query(self.AccountDB, " Call LSP_VisitorLogin(?, ?, ?,?)", message.MachineID, message.DeviceType, strings.Split(ip, ":")[0], strings.Split(conf.GetServer().ServerUrl, ":")[0])

	defer func() {
		if rows != nil {
			err := rows.Close()
			if err != nil {
				_ = log.Logger.Errorf("VisitorLogin err%v", err)
			}
		}
	}()
	if err != nil {
		_ = log.Logger.Errorf("VisitorLogin err %v", err)
		return &msg.Login_S_Fail{
			ErrorCode: -1,
			ErrorMsg:  err.Error(),
		}
	}

	return self.loadUserInfo(rows)
}

//加载用户信息
func (self *LoginMysql) loadUserInfo(rows *sql.Rows) interface{} {
	var errorCode int32
	var errorMsg string
	rows.Next()
	err := rows.Scan(&errorCode, &errorMsg)
	if err != nil {
		_ = log.Logger.Errorf("loadUserInfo err %v", err)
		return &msg.Login_S_Fail{
			ErrorCode: -1,
			ErrorMsg:  err.Error(),
		}
	}
	if errorCode != 0 {
		return &msg.Login_S_Fail{
			ErrorCode: errorCode,
			ErrorMsg:  errorMsg,
		}
	}

	rows.NextResultSet()
	rows.Next()
	loginSuccess := new(msg.Login_S_Success)
	err = rows.Scan(&loginSuccess.UserID, &loginSuccess.NikeName, &loginSuccess.UserGold, &loginSuccess.UserDiamonds, &loginSuccess.PhoneNumber, &loginSuccess.BinderCardNo, &loginSuccess.MemberOrder, &loginSuccess.FaceID,
		&loginSuccess.RoleID, &loginSuccess.SuitID, &loginSuccess.PhotoFrameID, &loginSuccess.Gender)
	if err != nil {
		_ = log.Logger.Errorf("loadUserInfo err %v", err)
		return &msg.Login_S_Fail{
			ErrorCode: -1,
			ErrorMsg:  err.Error(),
		}
	}
	lock, err := self.getGameID(loginSuccess.UserID)
	if err != nil {
		loginSuccess.KindID = -1
		loginSuccess.GameID = -1
	} else {
		loginSuccess.GameID = lock.GameID
		loginSuccess.KindID = lock.KindID
	}
	return loginSuccess
}

//锁定用户
type AccountPlayingLock struct {
	UserID      int32      `json:"user_id"`
	KindID      int32      `json:"kind_id"`
	GameID      int32      `json:"game_id"`
	EnterIP     string     `json:"enter_ip"`
	CollectDate *time.Time `json:"collect_date"`
}

func (a AccountPlayingLock) TableName() string {
	return "AccountPlayingLock"
}

//获取上次的游戏ID
func (self *LoginMysql) getGameID(userId int32) (AccountPlayingLock, error) {
	var lock AccountPlayingLock
	_, err := self.AccountDB.Where("UserID = ? ", userId).Get(&lock)
	//查询数据，指定字段名，返回sql.Rows结果集
	//rows, err := self.AccountDB.Query("select * from AccountPlayingLock Where UserID = ? ", userId)
	//defer rows.Close()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//for rows.Next() {
	//	rows.Scan(&lock.UserID, &lock.KindID, &lock.GameID, &lock.EnterIP, &lock.CollectDate)
	//}
	return lock, err
}
