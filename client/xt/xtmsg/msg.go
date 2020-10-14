package msg

import (
	"xj_game_server/util/leaf/network/protobuf"
)

var Processor = protobuf.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	Processor.Register(&Hall_C_Msg{})  //连接消息
	Processor.Register(&Hall_S_Msg{})  //大厅消息
	Processor.Register(&Hall_S_Fail{}) //失败消息
	Processor.Register(&Hall_Recharge_Notice{}) //充值通知
}
