package msg

import (
	"xj_game_server/util/leaf/network/protobuf"
)

var Processor = protobuf.NewProcessor()

func init() {
	//命令号 和注册顺序一致从0开始
	// 客户端--------
	Processor.Register(&Game_C_TokenLogin{})  //token登陆消息
	Processor.Register(&Game_C_UserStandUp{}) //用户起立消息
	Processor.Register(&Game_C_UserJetton{})  //用户下注消息
	Processor.Register(&Game_C_UserCompare{}) ////比较大小
	Processor.Register(&Game_C_UserList{})    //获取用户列表客户端参数
	// 服务端-----
	Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
	Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
	Processor.Register(&Game_S_GameResult{})    //铃铛结果
	Processor.Register(&Game_S_CompareResult{}) //比大小结果
	Processor.Register(&Game_S_UserList{})      //获取用户列表服务器返回
}
