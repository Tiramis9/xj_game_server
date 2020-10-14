package logic

import (
	"time"
	"xj_game_server/game/203_doudizhu/game/table"
	"xj_game_server/game/203_doudizhu/gate"
	"xj_game_server/game/203_doudizhu/global"
	gameRobot "xj_game_server/game/203_doudizhu/robot/robot"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/user"
	leafGate "xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/util"
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
		l.loadRobots()
		//l.loadRobot()
		//
		//for {
		//	t := time.NewTimer(10 * time.Second)
		//	select {
		//	case <-t.C:
		//		//加载机器人配置
		//		mysql.GameClient.InitRobotConfig()
		//		//加载机器人
		//		l.loadRobot()
		//	}
		//}
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
				time.Sleep(time.Second * time.Duration(util.RandInterval(3, 10)))
			}
		}()
	}
	l.config = robotConfig
}

// 加载游戏配置
func (l *Logic) loadRobots() {
	go func() {
		for {
			select {
			case onlineCount := <-global.NoticeRobotOnline:
				robotConfig := robot.RobotConfigItem.GetConfig()
				for k, v := range robotConfig {
					// 判断批次是否使用中
					_, ok := l.config[k]
					if ok {
						v.BatchID = 0
						//v.RobotCount = int64(onlineCount)
					}
					v.RobotCount = int64(onlineCount)
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
						}
					}()
					global.NoticeLoadMath <- int32(len(userIDList))
				}
				l.config = robotConfig
			}
		}
	}()
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
		_, o := table.UserQueue.Load(userItem.UserID)
		if o {
			table.UserQueue.Delete(userItem.UserID)
			table.UserCount--
			log.Logger.Debug("断开连接真人：", userItem.UserID)
		}
		_, oRobot := table.RobotQueue.Load(userItem.UserID)
		if oRobot {
			table.RobotQueue.Delete(userItem.UserID)
			table.RobotCount--
			log.Logger.Debug("断开连接机器人：", userItem.UserID)
		}
		if userItem.Status == user.StatusFree {
			if userItem.ChairID >= 0 {
				table.List[userItem.TableID].OnMoveUserByChairID(map[int32]int32{userItem.ChairID: userItem.UserID})
			}
			user.List.Delete(a.UserData().(int32))
		} else {
			if userItem.Status == user.StatusPlaying && table.List[userItem.TableID].GetGameStatus() == global.GameStatusFree {
				table.List[userItem.TableID].OnActionUserStandUp(userItem, true)
				user.List.Delete(a.UserData().(int32))
				return
			}
			table.List[userItem.TableID].OnActionUserOffLine(userInfo)
		}
	}
	gameRobot.List.Delete(a.UserData().(int32))
}
