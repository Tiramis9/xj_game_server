package internal

import (
	"xj_game_server/game/204_redmahjong/game/table"
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
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
}
