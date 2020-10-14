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
	"xj_game_server/game/201_qiangzhuangniuniu/conf"
	"xj_game_server/game/201_qiangzhuangniuniu/global"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
	"xj_game_server/game/201_qiangzhuangniuniu/robot/robot"
	"xj_game_server/game/public/user"
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
	// 抢庄场景
	msg.Processor.SetHandler(&msg.Game_S_QZScene{}, handlerQZScene)
	// 下注场景
	msg.Processor.SetHandler(&msg.Game_S_JettonScene{}, handlerJettonScene)
	// 摊牌场景
	msg.Processor.SetHandler(&msg.Game_S_TPScene{}, handlerTPScene)
	//	//开始游戏
	//	msg.Processor.SetHandler(&msg.Game_S_StartTime{}, handlerGameStart)
	// 发牌环节
	msg.Processor.SetHandler(&msg.Game_S_CardRound{}, handlerCardRound)
	// 抢庄环节
	msg.Processor.SetHandler(&msg.Game_S_CallRound{}, handlerCallRound)
	// 叫倍环节
	msg.Processor.SetHandler(&msg.Game_S_BetRound{}, handlerBetRound)
	// 开牌环节
	msg.Processor.SetHandler(&msg.Game_S_ShowRound{}, handlerShowRound)

	//结束游戏消息
	msg.Processor.SetHandler(&msg.Game_S_GameConclude{}, handlerGameConclude)
	//抢庄通知
	msg.Processor.SetHandler(&msg.Game_S_UserQZ{}, handlerUserQZ)
	//定庄通知
	msg.Processor.SetHandler(&msg.Game_S_GameDZ{}, handlerGameDZ)
	//下注通知
	msg.Processor.SetHandler(&msg.Game_S_UserJetton{}, handlerUserJetton)
	//摊牌通知
	msg.Processor.SetHandler(&msg.Game_S_UserTP{}, handlerUserTP)

}

//请求失败
func handlerReqlyFail(args []interface{}) {
	m := args[0].(*msg.Game_S_ReqlyFail)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = fmt.Errorf("%c[1;40;31m 抢庄牛牛 请求失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil! 请求失败")
		a.Close()
		return
	}

	bytes, _ := json.Marshal(m)
	_ = fmt.Errorf("%c[1;40;31m 抢庄牛牛 请求失败=====%c[0m %v\n", 0x1B, 0x1B, string(bytes))

	if m.ErrorCode == global.JCError1 {
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

		robot.SetUserChairID(0)

		robot.SetPlayStatus(false)

		go func() {

			time.Sleep(time.Millisecond * time.Duration(util.RandInterval(500, 1500)))

			robot.Assignment()
			robot.SitDown()
		}()
	}
}

//登陆成功
func handlerLoginSuccess(args []interface{}) {
	//fmt.Printf("%c[1;40;31m登陆成功=====%c[0m\n", 0x1B, 0x1B)

	_ = args[0].(*msg.Game_S_LoginSuccess)
	agent := args[1].(gate.Agent)

	if agent.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 登录失败=====%c[0m %v\n", 0x1B, 0x1B, "a is nil!")
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

//下注场景
func handlerJettonScene(args []interface{}) {
	m := args[0].(*msg.Game_S_JettonScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 下注场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	//var msg model.JettonScene
	//err := json.Unmarshal(m.Data, &msg)
	//
	//if err != nil {
	//	_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 下注场景 解析失败=====%c[0m %v\n", 0x1B, 0x1B, err.Error())
	//	a.Close()
	//	return
	//}

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
	robot.SetPlayStatus(m.UserPlaying[m.UserChairID])
	robot.SetUserChairID(m.UserChairID)

}

//抢庄场景
func handlerQZScene(args []interface{}) {
	m := args[0].(*msg.Game_S_QZScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 抢庄场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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
	//robot.SetPlayStatus(m.UserPlaying[m.UserChairID])
	robot.SetUserChairID(m.UserChairID)

}

//摊牌场景
func handlerTPScene(args []interface{}) {
	m := args[0].(*msg.Game_S_TPScene)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 摊牌场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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
	//robot.SetPlayStatus(m.UserPlaying[m.UserChairID])
	robot.SetUserChairID(m.UserChairID)
}

//坐下通知
func handlerSitDownNotify(args []interface{}) {
	_ = args[0].(*msg.Game_S_SitDownNotify)
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
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 空闲场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	//var msg model.FreeScene
	//err := json.Unmarshal(m.Data, &msg)
	//
	//if err != nil {
	//	_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 空闲场景 解析失败=====%c[0m %v\n", 0x1B, 0x1B, err.Error())
	//	a.Close()
	//	return
	//}

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
	robot.SetPlayStatus(true)
	robot.SetUserChairID(m.UserChairID)
}

//抢庄通知
func handlerUserQZ(args []interface{}) {
	_ = args[0].(*msg.Game_S_UserQZ)
	_ = args[1].(gate.Agent)
}

// 定庄通知
func handlerGameDZ(args []interface{}) {
	m := args[0].(*msg.Game_S_GameDZ)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 摊牌场景=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	robot.SetBankerByChair(m.ChairID)
}

//摊牌通知
func handlerUserTP(args []interface{}) {
	_ = args[0].(*msg.Game_S_UserTP)
	_ = args[1].(gate.Agent)
}

//下注通知
func handlerUserJetton(args []interface{}) {
	//fmt.Printf("%c[1;40;31m下注成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_UserJetton)
	_ = args[1].(gate.Agent)
}

//起立成功
func handlerUpNotify(args []interface{}) {
	//fmt.Printf("%c[1;40;31m起立成功=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_StandUpNotify)
	_ = args[1].(gate.Agent)
}

//游戏开始
func handlerCardRound(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_CardRound)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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
	robot.SetPlayStatus(true)
	/*
		switch m.GameStatus {
		case global.GameStatusFree:
			robot.SetPlayStatus(true)
		case global.GameStatusQZ:
			go func() {
				time.Sleep(time.Duration(util.RandInterval(2, conf.GetServer().GameQZTime-3)) * time.Second)
				if robot.GetPlayStatus() {
					robot.Qz(conf.GetServer().MultipleList[util.RandInterval(0, int32(len(conf.GetServer().MultipleList))-1)])
				}
			}()
		case global.GameStatusJetton:
			go func() {
				time.Sleep(time.Duration(util.RandInterval(2, conf.GetServer().GameJettonTime-3)) * time.Second)
				if robot.GetPlayStatus() && !robot.GetBanker() {
					robot.Jetton(conf.GetServer().JettonList[util.RandInterval(0, int32(len(conf.GetServer().JettonList))-1)])
				}
			}()
		case global.GameStatusTP:
			go func() {
				time.Sleep(time.Duration(util.RandInterval(3, conf.GetServer().GameTPTime-3)) * time.Second)
				if robot.GetPlayStatus() {
					robot.TP()
				}
			}()
		}
	*/
}

//叫倍环节
func handlerBetRound(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	m := args[0].(*msg.Game_S_BetRound)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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
	robot.InitJetton(m.Multiple)
	go func() {
		time.Sleep(time.Duration(util.RandInterval(2, conf.GetServer().GameJettonTime-3)) * time.Second)
		if robot.GetPlayStatus() && !robot.GetBanker() {
			//robot.Jetton(conf.GetServer().JettonList[util.RandInterval(0, int32(len(conf.GetServer().JettonList))-1)])
			robot.Jetton(robot.GetJetton()[util.RandInterval(0, int32(len(robot.GetJetton()))-1)])
		}
	}()

}

//摊牌环节
func handlerShowRound(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_ShowRound)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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

	go func() {
		time.Sleep(time.Duration(util.RandInterval(3, conf.GetServer().GameTPTime-3)) * time.Second)
		if robot.GetPlayStatus() {
			robot.TP()
		}
	}()
}

//抢庄环节
func handlerCallRound(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏开始=====%c[0m\n", 0x1B, 0x1B)
	m := args[0].(*msg.Game_S_CallRound)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 游戏开始=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
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
	robot.InitMultiple(m.Multiple)
	go func() {
		time.Sleep(time.Duration(util.RandInterval(2, conf.GetServer().GameQZTime-3)) * time.Second)
		if robot.GetPlayStatus() {
			robot.Qz(robot.GetMultiple()[util.RandInterval(0, int32(len(robot.GetMultiple()))-1)])
			//robot.Qz(conf.GetServer().MultipleList[util.RandInterval(0, int32(len(conf.GetServer().MultipleList))-1)])
		}
	}()
}

//游戏结束
func handlerGameConclude(args []interface{}) {
	//fmt.Printf("%c[1;40;31m 游戏结束=====%c[0m\n ", 0x1B, 0x1B)
	_ = args[0].(*msg.Game_S_GameConclude)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		_ = log.Logger.Errorf("%c[1;40;31m 抢庄牛牛 游戏结束=====%c[0m %v\n", 0x1B, 0x1B, "a is nil！")
		a.Close()
		return
	}

	value, ok := robot.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	robotAct := value.(*robot.Robot)
	// 过期必须退出
	if !robotAct.CheckBatchTimeOut() {
		a.Close()
		return
	}
	robotAct.SetPlayStatus(false)
	robotAct.RandStandUp() // 随机起立坐下
	var duration = 1
	go func() {
		for {
			duration *= 2
			time.Sleep(time.Second * time.Duration(duration))
			userItem, exist := user.List.Load(robotAct.GetUserID())
			if exist {
				if userItem.(*user.Item).Status == user.StatusFree {
					robotAct.StandUp()
				}
			} else {
				robot.List.Delete(robotAct.GetUserID())
				return
			}
			if duration > 600 { // 大于10分钟 退出
				return
			}
		}
	}()
}
