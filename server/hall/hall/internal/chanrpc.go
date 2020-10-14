package internal

import (
	"fmt"
	"xj_game_server/util/leaf/gate"
)

var Agents = make(map[gate.Agent]struct{})

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
	userInfo, ok := ConnList.Load(a.UserData())
	if ok {
		userItem := userInfo.(*Client)
		//同一用户挤掉线 不走这方法
		//if a.RemoteAddr().String() != userItem.RemoteAddr().String() {
		//	return
		//}
		fmt.Println("心跳结束rpcCloseAgent", userItem.Uid)
		ConnList.Delete(a.UserData())
		userItem.Stop <- struct{}{}
	}
}
