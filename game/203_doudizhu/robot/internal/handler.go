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
	"encoding/json"
	"xj_game_server/game/203_doudizhu/global"
	"xj_game_server/game/203_doudizhu/msg"
	"xj_game_server/game/203_doudizhu/robot/robot"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
)

func init() {

	//请求失败
	msg.Processor.SetHandler(&msg.Game_S_ReqlyFail{}, handlerReqlyFail)
	//登录成功
	msg.Processor.SetHandler(&msg.Game_S_LoginSuccess{}, handlerLoginSuccess)
	//坐下通知
	msg.Processor.SetHandler(&msg.Game_S_SitDownNotify{}, handlerSitDownNotify)
	//起立通知消息
	msg.Processor.SetHandler(&msg.Game_S_StandUpNotify{}, handlerUpNotify)
	//掉线通知消息
	msg.Processor.SetHandler(&msg.Game_S_OffLineNotify{}, handlerOffLineNotify)
	//上线通知消息
	msg.Processor.SetHandler(&msg.Game_S_OnLineNotify{}, handlerOnLineNotify)
	// 空闲场景
	msg.Processor.SetHandler(&msg.Game_S_FreeScene{}, handlerFreeScene)
	// 叫分场景
	msg.Processor.SetHandler(&msg.Game_S_GrabLandlordScene{}, handlerGrabLandlordScene)
	// 出牌场景
	msg.Processor.SetHandler(&msg.Game_S_PlayScene{}, handlerPlayScene)
	// 通知当前用户叫分、出牌
	msg.Processor.SetHandler(&msg.Game_S_CurrentUser{}, handlerCurrentUser)

	//开始游戏
	msg.Processor.SetHandler(&msg.Game_S_StartGame{}, handlerGameStart)
	//结束游戏消息
	msg.Processor.SetHandler(&msg.Game_S_GameConclude{}, handlerGameConclude)

	//	用户准备通知
	msg.Processor.SetHandler(&msg.Game_S_UserPrepare{}, handlerUserPrepare)
	//用户取消准备通知
	msg.Processor.SetHandler(&msg.Game_S_UserUnPrepare{}, handlerUnPrepare)
	//叫分通知消息
	msg.Processor.SetHandler(&msg.Game_S_UserGrabLandlord{}, handlerGrabLandlord)

	//用户出牌通知
	msg.Processor.SetHandler(&msg.Game_S_UserCP{}, handlerUserCP)
	//确认地主
	msg.Processor.SetHandler(&msg.Game_S_StartCPDetermine{}, handlerStartCP)
	// 玩家过牌
	msg.Processor.SetHandler(&msg.Game_S_UserPass{}, handlerUserPass)
}

//请求失败
func handlerReqlyFail(args []interface{}) {
	m := args[0].(*msg.Game_S_ReqlyFail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 请求失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil! 请求失败")
		a.Close()
		return
	}
	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	bytes, _ := json.Marshal(m)
	// _ = log.Logger.Errorf("%c[1;40;31m 斗地主 请求失败=====%c[0m %v\n", 0x1B, 0x1B, string(bytes))
	_ = log.Logger.Errorf("%c[1;40;31m 斗地主 请求失败=====%c[0m %v 用户id:%v,用户状态:%v\n", 0x1B, 0x1B, string(bytes),
		value.(*robot.Robot).GetUserID(), value.(*robot.Robot).GetUserStatus())

}

//登陆成功
func handlerLoginSuccess(args []interface{}) {
	//fmt.Printf("%c[1;40;31m登陆成功=====%c[0m\n", 0x1B, 0x1B)

	_ = args[0].(*msg.Game_S_LoginSuccess)
	agent := args[1].(gate.Agent)

	if agent.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 登录失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
		agent.Close()
		return
	}

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

//坐下通知
func handlerSitDownNotify(args []interface{}) {
	_ = args[0].(*msg.Game_S_SitDownNotify)
	_ = args[1].(gate.Agent)
}

//空闲场景
func handlerFreeScene(args []interface{}) {
	m := args[0].(*msg.Game_S_FreeScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 空闲场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}

	var ChairId int32
	for i := range m.UserList {
		if m.UserList[i].UserID == robot.GetUserID() {
			ChairId = m.UserList[i].ChairID
		}
	}
	if ChairId != robot.GetUserChairID() {
		robot.SetUserChairID(ChairId)
	}
	robot.SetTableStatus(false)
	robot.SetUserStatus(true)
	// 准备
	// robot.Prepare()
}

// 通知 出牌位置
func handlerCurrentUser(args []interface{}) {
	m := args[0].(*msg.Game_S_CurrentUser)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 下注场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robot.SetCurrentChairID(m.CurrentChairID)
	if m.CurrentChairID == robot.GetUserChairID() {
		robot.PokersCpORJF()
	}

}

//出牌场景
func handlerPlayScene(args []interface{}) {
	_ = args[0].(*msg.Game_S_PlayScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 下注场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robot.SetTableStatus(true)
	robot.SetUserStatus(false)
	robot.SetGameStatus(global.GameStatusPlay)
}

//叫分场景
func handlerGrabLandlordScene(args []interface{}) {
	_ = args[0].(*msg.Game_S_GrabLandlordScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 下注场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}

	//robot.SetUserChairID(m.UserChairID)
	robot.SetTableStatus(true)
	robot.SetUserStatus(true)
	robot.SetGameStatus(global.GameStatusJF)
}

//掉线通知消息
func handlerOffLineNotify(args []interface{}) {
	_ = args[0].(*msg.Game_S_OffLineNotify)
	_ = args[1].(gate.Agent)
}

//上线通知消息
func handlerOnLineNotify(args []interface{}) {
	_ = args[0].(*msg.Game_S_OnLineNotify)
	_ = args[1].(gate.Agent)
}

//起立成功
func handlerUpNotify(args []interface{}) {
	//fmt.Printf("%c[1;40;31m起立成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_StandUpNotify)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 起立通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	_, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	//robot := value.(*robot.Robot)
	//	robot.StandUp()
}

//游戏开始
func handlerGameStart(args []interface{}) {
	// fmt.Printf("%c[1;40;31m 机器人 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	m := args[0].(*msg.Game_S_StartGame)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}

	robot.SetGameStatus(global.GameStatusJF)
	// 这里拿不到现在的出牌人
	robot.StartGame(m.UserPoker)
}

//游戏结束
func handlerGameConclude(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏结束=====%c[0m\n ", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_GameConclude)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 游戏结束=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)
	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	if robot.GetDiamond() < 100 {
		robot.Assignment()
	}

	robot.GameEnd()
}

//用户准备通知
func handlerUserPrepare(args []interface{}) {
	_ = args[0].(*msg.Game_S_UserPrepare)
	_ = args[1].(gate.Agent)
}

//用户过牌通知
func handlerUserPass(args []interface{}) {
	_ = args[0].(*msg.Game_S_UserPass)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 开始叫分=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)
	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
}

//用户取消准备通知
func handlerUnPrepare(args []interface{}) {
	_ = args[0].(*msg.Game_S_UserUnPrepare)
	_ = args[1].(gate.Agent)
}

//用户叫分通知
func handlerGrabLandlord(args []interface{}) {
	m := args[0].(*msg.Game_S_UserGrabLandlord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 开始叫分=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)
	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robot.SetChip(m.Multiple)

}

//用户出牌通知 重新设置手牌
func handlerUserCP(args []interface{}) {
	m := args[0].(*msg.Game_S_UserCP)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 开始出牌=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)
	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robot.SetCurrentPokers(m.ChairID, m.PokerType, m.Pokers)
	robot.SetChip(m.CurrentMultiple)
	robot.AddOutCards(m.Pokers, m.ChairID)
	if robot.GetUserChairID() == m.ChairID {
		robot.ResetPoker()
	}

}

//确定地主，地主出牌（出牌接口：handlerCurrentUser）
func handlerStartCP(args []interface{}) {
	// fmt.Printf(" func Game_S_StartCPDetermine ")
	m := args[0].(*msg.Game_S_StartCPDetermine)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 斗地主 开始出牌=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)
	// 过期必须退出
	if !robot.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robot.SetGameStatus(global.GameStatusPlay)
	robot.SetChip(m.Multiple)
	robot.SetDizPokerAndChairId(m.CurrentChairID, m.LandlordPokers)
	robot.SetCurrentChairID(m.CurrentChairID)
	if robot.GetUserChairID() == robot.GetDizChairID() {
		robot.SetCurrentPokers(m.CurrentChairID, 0, []int32{})
		robot.AppendDzPokers()
		robot.PokersCpORJF()
	}
}
