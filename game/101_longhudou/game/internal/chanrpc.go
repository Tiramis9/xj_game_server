package internal

import (
	"xj_game_server/game/101_longhudou/game/table"
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
}
