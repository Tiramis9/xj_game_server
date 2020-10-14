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
	"fmt"
	"time"
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/204_redmahjong/msg"
	"xj_game_server/game/204_redmahjong/robot/robot"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/util"
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
	//游戏场景
	msg.Processor.SetHandler(&msg.Game_S_PlayScene{}, handlerPlayScene)
	//开始游戏
	msg.Processor.SetHandler(&msg.Game_S_GameStart{}, handlerGameStart)
	//结束游戏消息
	msg.Processor.SetHandler(&msg.Game_S_GameConclude{}, handlerGameConclude)
	//用户出牌
	msg.Processor.SetHandler(&msg.Game_S_UserOutCard{}, handlerUserOutCard)
	//用户操作通知
	msg.Processor.SetHandler(&msg.Game_S_UserOperate{}, handlerUserOperate)
	//发牌通知
	msg.Processor.SetHandler(&msg.Game_S_SendMj{}, handlerSendMj)
	//准备通知消息
	msg.Processor.SetHandler(&msg.Game_S_UserPrepare{}, handlerUserPrepare)
	//取消准备通知消息
	msg.Processor.SetHandler(&msg.Game_S_UserUnPrepare{}, handlerUserUnPrepare)
	//操作提示
	msg.Processor.SetHandler(&msg.Game_S_OperateNotify{}, handlerOperateNotify)
}

//请求失败
func handlerReqlyFail(args []interface{}) {
	m := args[0].(*msg.Game_S_ReqlyFail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = fmt.Errorf("%c[1;40;31m 红中麻将 请求失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil! 请求失败")
		a.Close()
		return
	}

	bytes, _ := json.Marshal(m)
	_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 请求失败=====%c[0m uid:%v  %v\n", 0x1B, 0x1B, a.UserData().(int32), string(bytes))

}

//登陆成功
func handlerLoginSuccess(args []interface{}) {
	//fmt.Printf("%c[1;40;31m登陆成功=====%c[0m\n", 0x1B, 0x1B)

	_ = args[0].(*msg.Game_S_LoginSuccess)
	agent := args[1].(gate.Agent)

	if agent.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 登录失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
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

//起立成功
func handlerUpNotify(args []interface{}) {
	//fmt.Printf("%c[1;40;31m起立成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_StandUpNotify)
	_ = args[1].(gate.Agent)
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

//空闲场景
func handlerFreeScene(args []interface{}) {
	m := args[0].(*msg.Game_S_FreeScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 空闲场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	//设置机器人椅子号
	for _, v := range m.UserList {
		if v.UserID == robot.GetUserID() {
			robot.SetUserChairID(v.ChairID)
			break
		}
	}

	//准备
	//time.Sleep(time.Second * time.Duration(util.RandInterval(1, 5)))
	//robot.Prepare()
}

//游戏场景
func handlerPlayScene(args []interface{}) {
	_ = args[0].(*msg.Game_S_PlayScene)
	_ = args[1].(gate.Agent)
}

//游戏开始
func handlerGameStart(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	m := args[0].(*msg.Game_S_GameStart)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)

	var userMjData = make(map[int32]int32, 0)
	for _, v := range m.UserMjData {
		userMjData[v.MjKey] = v.MjValue
	}
	//游戏开始
	robot.GameStart(userMjData)
}

//游戏结束
func handlerGameConclude(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏结束=====%c[0m\n ", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_GameConclude)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 炸金花 游戏结束=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	robot.GameEnd()
}

//用户出牌
func handlerUserOutCard(args []interface{}) {
	m := args[0].(*msg.Game_S_UserOutCard)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 发牌通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)
	robot.SetNearestOutCard(m.MjData)
	if m.ChairID != robot.GetUserChairID() {
		return
	}
	robot.SetUserOutCard(m.MjData, 1)
	robot.ADDRounds()
}

//用户操作
func handlerUserOperate(args []interface{}) {
	m := args[0].(*msg.Game_S_UserOperate)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 发牌通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)
	var CurrentChairID = m.OperateUser
	if m.OperateUser != robot.GetUserChairID() || m.OperateCode&global.WIK_HU != 0 {
		return
	}

	var cardNum int32
	if m.OperateCode&global.WIK_BU_GANG != 0 {
		cardNum = 1
	} else if m.OperateCode&global.WIK_MING_GANG != 0 {
		cardNum = 1
	} else if m.OperateCode&global.WIK_AN_GANG != 0 {
		cardNum = 4
	} else if m.OperateCode&global.WIK_PENG != 0 {
		cardNum = 2
	}
	robot.SetUserOutCard(m.OperateMj, cardNum)

	if m.OperateCode&global.WIK_BU_GANG != 0 || m.OperateCode&global.WIK_MING_GANG != 0 || m.OperateCode&global.WIK_AN_GANG != 0 {
		return
	}

	if m.OperateCode&global.WIK_PENG != 0 {
		robot.DiskMj(m.OperateCode, m.OperateMj)
		go func() {
			//t := time.NewTimer(time.Millisecond * time.Duration(util.RandInterval(1000, 4000)))
			time.Sleep(time.Millisecond * time.Duration(util.RandInterval(1000, 4000)))
			if CurrentChairID == robot.GetUserChairID() {
				robot.OutMj()
			}
		}()
	}
}

//用户摸牌
func handlerSendMj(args []interface{}) {
	m := args[0].(*msg.Game_S_SendMj)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 发牌通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}

	robot := value.(*robot.Robot)
	var CurrentChairID = m.CurrentChairID
	if m.CurrentChairID != robot.GetUserChairID() {
		return
	}

	// 摸牌有反应则不出牌
	if robot.CheckOutMj(m.MjData) {
		return
	}
	go func() {
		//	t := time.NewTimer(time.Millisecond * time.Duration(util.RandInterval(1000, 4000)))
		time.Sleep(time.Millisecond * time.Duration(util.RandInterval(1000, 3000)))
		if CurrentChairID == robot.GetUserChairID() {
			robot.OutMj()
		}
	}()
}

//操作提示
func handlerOperateNotify(args []interface{}) {
	m := args[0].(*msg.Game_S_OperateNotify)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 操作提示=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robot := value.(*robot.Robot)

	//time.Sleep(time.Second * time.Duration(util.RandInterval(1, conf.GetServer().GameOperateTime-2)))
	go func() {
		time.Sleep(time.Second * time.Duration(util.RandInterval(1, 3)))
		robot.Operate(m.Response)
	}()
}

//准备通知消息
func handlerUserPrepare(args []interface{}) {
	m := args[0].(*msg.Game_S_UserPrepare)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 准备通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	if m.ChairID != robot.GetUserChairID() {
		return
	}

	//go func() {
	//	//取消准备
	//	time.Sleep(time.Second * time.Duration(util.RandInterval(30, 60)))
	//	if robot.GetGameStatus() == global.GameStatusPlay {
	//		return
	//	}
	////	robot.UnPrepare()
	//}()
}

//取消准备通知消息
func handlerUserUnPrepare(args []interface{}) {
	m := args[0].(*msg.Game_S_UserUnPrepare)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 红中麻将 准备通知=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	if m.ChairID != robot.GetUserChairID() {
		return
	}

	//退出或准备
	robot.RandStandUp()
}
