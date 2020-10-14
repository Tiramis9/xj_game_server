/*
 * @Author: yhlyl
 * @Date: 2019-11-27 14:48:15
 * @LastEditTime: 2019-11-27 15:11:35
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/101_longhudou/robot/internal/handler.go
 * @https://github.com/android-coco
 */
package internal

import (
	"xj_game_server/game/104_senglinwuhui/conf"
	"xj_game_server/game/104_senglinwuhui/msg"
	"xj_game_server/game/104_senglinwuhui/robot/robot"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
)

//var robotSceneData model.RobotSceneData

func init() {
	//消息处理绑定
	//请求失败
	msg.Processor.SetHandler(&msg.Game_S_ReqlyFail{}, handlerReqlyFail)
	//登录成功
	msg.Processor.SetHandler(&msg.Game_S_LoginSuccess{}, handlerLoginSuccess)
	//下注场景消息
	msg.Processor.SetHandler(&msg.Game_S_JettonScene{}, handlerJettonScene)
	// 开奖场景消息
	msg.Processor.SetHandler(&msg.Game_S_LotteryScene{}, handlerLotteryScene)
	// 游戏开始
	msg.Processor.SetHandler(&msg.Game_S_GameStart{}, handlerGameStart)
	// 游戏结束
	msg.Processor.SetHandler(&msg.Game_S_GameConclude{}, handlerGameConclude)
	// 下注成功
	msg.Processor.SetHandler(&msg.Game_S_UserJetton{}, handlerUserJetton)
	// 起立成功
	msg.Processor.SetHandler(&msg.Game_S_StandUpNotify{}, handlerUpNotify)
}

//请求失败
func handlerReqlyFail(args []interface{}) {
	//m := args[0].(*msg.Game_S_ReqlyFail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 请求失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil! 请求失败")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 请求失败=====%c[0m %v\n", 0x1B, 0x1B, "robot.List.Load(a.UserData().(int32)) is not ok!")
		a.Close()
		return
	}

	//a.Close()
	//bytes, _ := json.Marshal(m)
	//_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 请求失败=====%c[0m %v 游戏状态：%v 下注状态：%v 坐下状态%v \n", 0x1B, 0x1B, string(bytes), value.(*robot.Robot).GetGameStatus(), value.(*robot.Robot).GetBetStatus(), value.(*robot.Robot).GetSitDownStatus())
	value.(*robot.Robot).Assignment()
}

//登陆成功
func handlerLoginSuccess(args []interface{}) {

	_ = args[0].(*msg.Game_S_LoginSuccess)
	agent := args[1].(gate.Agent)

	if agent.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 登录失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
		agent.Close()
		return
	}
	//
	// 给机器人赋值金币
	uid := agent.UserData().(int32)

	value, ok := robot.List.Load(uid)
	if !ok {
		agent.Close()
		return
	}

	value.(*robot.Robot).Assignment()
	value.(*robot.Robot).SitDown()

}

//下注场景
func handlerJettonScene(args []interface{}) {
	//fmt.Printf("%c[1;40;31m下注场景=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_JettonScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 下注场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	// 过期必须退出
	if !value.(*robot.Robot).CheckBatchTimeOut() {
		a.Close()
		return
	}
	value.(*robot.Robot).SetGameStatus(true)
	value.(*robot.Robot).SetSitDownStatus(true)

	//fmt.Println(string(m.Data))

	//var jettonScene model.JettonScene
	//var jettonScene msg.Game_S_JettonScene

	//err := json.Unmarshal(m.Data, &jettonScene)

	//if err != nil {
	//	_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 下注场景解析数据失败=====%c[0m  %c[1;40;34m %v  %c[0m\n", 0x1B, 0x1B, 0x1B, err, 0x1B)
	//
	//	return
	//}

	go value.(*robot.Robot).RobotLottery()
}

//开奖场景
func handlerLotteryScene(args []interface{}) {
	//fmt.Printf("%c[1;40;31m开奖场景=====%c[0m\n", 0x1B, 0x1B)

	_ = args[0].(*msg.Game_S_LotteryScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 开奖场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	// 过期必须退出
	if !value.(*robot.Robot).CheckBatchTimeOut() {
		a.Close()
		return
	}
	value.(*robot.Robot).SetGameStatus(false)
	value.(*robot.Robot).SetSitDownStatus(true)

}

//下注成功
func handlerUserJetton(args []interface{}) {
	//fmt.Printf("%c[1;40;31m下注成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_UserJetton)
	_ = args[1].(gate.Agent)
	//randStandUp(a)
}

//起立成功
func handlerUpNotify(args []interface{}) {
	//fmt.Printf("%c[1;40;31m起立成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_StandUpNotify)
	_ = args[1].(gate.Agent)

}

//游戏开始
func handlerGameStart(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_GameStart)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	// 过期必须退出
	if !value.(*robot.Robot).CheckBatchTimeOut() {
		a.Close()
		return
	}
	value.(*robot.Robot).SetGameStatus(true)

	//下注定时器
	go value.(*robot.Robot).RobotLottery()
}

//游戏结束
func handlerGameConclude(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏结束=====%c[0m\n ", 0x1B, 0x1B)
	m := args[0].(*msg.Game_S_GameConclude)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 游戏结束=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	// 过期必须退出
	if !value.(*robot.Robot).CheckBatchTimeOut() {
		a.Close()
		return
	}

	value.(*robot.Robot).SetBetStatus(false)
	value.(*robot.Robot).SetGameStatus(false)

	//var gameConclude model.GameConclude
	//
	//err := json.Unmarshal(m.Data, &gameConclude)
	//
	//if err != nil {
	//	_ = log.Logger.Errorf("%c[1;40;31m 森林舞会 游戏结束解析数据失败=====%c[0m  %c[1;40;34m %v  %c[0m\n", 0x1B, 0x1B, 0x1B, err, 0x1B)
	//
	//	return
	//}

	//uid := a.UserData().(int32)

	var score float32
	score = value.(*robot.Robot).GetGold() + m.Money
	value.(*robot.Robot).SetDiamond(score)
	//if store.GameControl.GetGameInfo().DeductionsType == 0 {
	//
	//	if m.UserListLoss[uid] > 0 {
	//		score = value.(*robot.Robot).GetGold() + m.UserListLoss[uid]
	//
	//		value.(*robot.Robot).SetGold(score)
	//	}
	//
	//} else {
	//	if m.UserListLoss[uid] > 0 {
	//		score = value.(*robot.Robot).GetDiamond() + m.UserListLoss[uid]
	//		value.(*robot.Robot).SetDiamond(score)
	//	}
	//
	//}

	if score < conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1] {
		//大幅增加推出概率
		value.(*robot.Robot).AddWithProbability(3)
	} else {
		value.(*robot.Robot).AddWithProbability(0.2)
	}

	go value.(*robot.Robot).RandStandUp()

}
