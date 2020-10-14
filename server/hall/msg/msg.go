package msg

import (
	"xj_game_server/util/leaf/network/json"
)

var Processor = json.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	Processor.Register(&Hall_C_Msg{}) //连接消息
	//Processor.Register(&Hall_S_Msg{})           //大厅消息
	Processor.Register(&Hall_S_Fail{}) //失败消息
	//Processor.Register(&Hall_Recharge_Notice{}) //充值通知
	Processor.Register(&HeartLoginInit{}) //充值通知
	Processor.Register(&UserInfoChange{}) //玩家信息变更通知
	Processor.Register(&GameInfoChange{}) //房间信息变更通知
	Processor.Register(&Notify{})         //通知
}
