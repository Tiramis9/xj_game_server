package logic

import (
	"xj_game_server/game/105_benchibaoma/game/table"
	"xj_game_server/game/105_benchibaoma/gate"
	gameRobot "xj_game_server/game/105_benchibaoma/robot/robot"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/user"
	leafGate "xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"time"
)

var Client = new(Logic)

func init() {
	Client.config = make(map[int32]robot.Config)
}

type Logic struct {
	config map[int32]robot.Config // 批次号加配置
}

func (l *Logic) OnInit() {
	go func() {
		//加载机器人配置
		mysql.GameClient.InitRobotConfig()
		//加载机器人
		l.loadRobot()

		for {
			t := time.NewTimer(10 * time.Second)
			select {
			case <-t.C:
				//加载机器人配置
				mysql.GameClient.InitRobotConfig()
				//加载机器人
				l.loadRobot()
			}
		}
	}()
}

//加载机器人
func (l *Logic) loadRobot() {
	robotConfig := robot.RobotConfigItem.GetConfig()
	// 循环批次号
	for k, v := range robotConfig {
		// 判断批次是否使用中
		_, ok := l.config[k]
		if ok {
			continue
		}
		// 获取机器人user id
		userIDList := mysql.GameClient.LoadRobotUser(&v)
		if userIDList == nil {
			_ = log.Logger.Error("loadRobot err: Lack of robots")
			return
		}
		// 模拟真实用户登陆
		go func() {
			for _, userID := range userIDList {
				//随机3-10秒登录
				robotUser := new(gameRobot.Robot)
				robotUser.OnInit(userID, k, gate.Module.Gate, callbackCloseAgent)
				gameRobot.List.Store(userID, robotUser)
				robotUser.Login()
				time.Sleep(time.Millisecond * 500)
			}
		}()
	}

	l.config = robotConfig
}

// 关闭回调
func callbackCloseAgent(args []interface{}) {
	a := args[0].(leafGate.Agent)
	if a.UserData() == nil {
		return
	}
	userInfo, ok := user.List.Load(a.UserData().(int32))
	if ok {
		userItem := userInfo.(*user.Item)
		//同一用户挤掉线 不走这方法
		if a.RemoteAddr().String() != userItem.RemoteAddr().String() {
			return
		}
		if userItem.Status == user.StatusFree {
			user.List.Delete(a.UserData().(int32))
		} else {
			if !table.List[userItem.TableID].IsUserBet(userItem) {
				//没有下注直接踢掉,强制起立
				// 起立 强制退出
				table.List[userItem.TableID].OnActionUserStandUp(userItem, true)
				// map 中移除
				user.List.Delete(a.UserData().(int32))
				return
			}
			table.List[userItem.TableID].OnActionUserOffLine(userInfo)
		}
	}
	gameRobot.List.Delete(a.UserData().(int32))
}
