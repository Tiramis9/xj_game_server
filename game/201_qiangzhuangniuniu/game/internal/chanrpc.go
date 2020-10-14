package internal

import (
	"xj_game_server/game/201_qiangzhuangniuniu/game/table"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
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
		if userItem.Status == user.StatusFree {
			user.List.Delete(a.UserData().(int32))
		} else {
			if userItem.Status == user.StatusPlaying && !table.List[userItem.TableID].IsInGame(userItem) {
				table.List[userItem.TableID].OnActionUserStandUp(userItem, true)
				user.List.Delete(a.UserData().(int32))
				return
			}
			table.List[userItem.TableID].OnActionUserOffLine(userInfo)
		}
	}
}
