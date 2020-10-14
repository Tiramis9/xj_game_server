package internal

import (
	"fmt"
	"reflect"
	"time"
	"xj_game_server/game/201_qiangzhuangniuniu/conf"
	"xj_game_server/game/201_qiangzhuangniuniu/game/table"
	"xj_game_server/game/201_qiangzhuangniuniu/global"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/mysql"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
)

func init() {
	//机器人消息处理绑定
	msg.Processor.SetHandler(&msg.Game_C_RobotLogin{}, handlerRobotLogin)

	//客户端消息处理绑定
	handler(&msg.Game_C_TokenLogin{}, handlerTokenLogin)
	handler(&msg.Game_C_UserSitDown{}, handlerUserSitDown)
	handler(&msg.Game_C_UserStandUp{}, handlerUserStandUp)
	handler(&msg.Game_C_UserQZ{}, handlerUserQZ)
	handler(&msg.Game_C_UserJetton{}, handlerUserJetton)
	handler(&msg.Game_C_UserTP{}, handlerUserTP)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//用户登陆(token)
func handlerTokenLogin(args []interface{}) {
	fmt.Print("handlerTokenLogin ")
	m := args[0].(*msg.Game_C_TokenLogin)
	a := args[1].(gate.Agent)

	//验证token
	userID, err := redis.GameClient.UnmarshalToken(m.Token)
	fmt.Println(userID, err)
	if err != nil {
		_ = log.Logger.Errorf("handlerTokenLogin登录失败token:%s userId:%d", m.Token, userID)
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.LoginTokenError,
			ErrorMsg:  "token is validation.",
		})
		a.Close()
		return
	}

	//验证是否重复登陆/断线重连
	userItem, ok := user.List.Load(userID)
	if ok && userItem.(*user.Item).TableID != -1 {
		//绑定agent
		oldAgent := userItem.(*user.Item).Agent
		a.SetUserData(userID)
		userItem.(*user.Item).Agent = a

		if table.List[userItem.(*user.Item).TableID].GetGameStatus() == global.GameStatusEnd || table.List[userItem.(*user.Item).TableID].GetGameStatus() == global.GameStatusFree {
			//发送登陆成功
			userItem.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
				GameQZTime:     conf.GetServer().GameQZTime - 1,
				GameJettonTime: conf.GetServer().GameJettonTime - 1,
				GameTPTime:     conf.GetServer().GameTPTime - 1,
				GameStartTime:  conf.GetServer().GameStartTime - 1,
				//MultipleList:   conf.GetServer().MultipleList,
				//JettonList:     conf.GetServer().JettonList,
				Status: 1,
			})
			return
		} else {
			//发送登陆成功
			userItem.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
				GameQZTime:     conf.GetServer().GameQZTime - 1,
				GameJettonTime: conf.GetServer().GameJettonTime - 1,
				GameTPTime:     conf.GetServer().GameTPTime - 1,
				GameStartTime:  conf.GetServer().GameStartTime - 1,
				//MultipleList:   conf.GetServer().MultipleList,
				//JettonList:     conf.GetServer().JettonList,
				//Status: 1,
			})
			if table.List[userItem.(*user.Item).TableID].GetGameStatus() != global.GameStatusTP {
				time.Sleep(time.Millisecond * 500) // 前端跑动画事件
			}
		}

		//通知上个用户下线
		if userItem.(*user.Item).Status != user.StatusOffline && (a.RemoteAddr().String() != userItem.(*user.Item).RemoteAddr().String()) {
			oldAgent.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.LoginError,
				ErrorMsg:  "你的账号已在其他设备登录, 如非本人操作请立即修改密码!",
			})
			oldAgent.Close()
		}

		//检测是不是空闲状态
		if userItem.(*user.Item).Status == user.StatusFree {
			return
		}

		table.List[userItem.(*user.Item).TableID].OnActionUserReconnect(userItem)
		return
	}

	//用户登陆
	errCode, errMsg := mysql.GameClient.UserLogin(a, userID, m.MachineID)
	// 登录失败
	if errCode != common.StatusOK {
		_ = log.Logger.Errorf("handlerTokenLogin 登录失败 mysql :%d---%s", errCode, errMsg)
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: errCode,
			ErrorMsg:  errMsg,
		})
		a.Close()
		return
	}
	a.SetUserData(userID)
	value, ok := user.List.Load(userID)
	if !ok {
		_ = log.Logger.Errorf("handlerTokenLogin user.List.Load err")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: common.StatusInternalServerFail,
			ErrorMsg:  common.Description(common.StatusInternalServerFail),
		})
		return
	}
	value.(*user.Item).Agent = a
	//发送登陆成功
	value.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
		GameQZTime:     conf.GetServer().GameQZTime - 1,
		GameJettonTime: conf.GetServer().GameJettonTime - 1,
		GameTPTime:     conf.GetServer().GameTPTime - 1,
		GameStartTime:  conf.GetServer().GameStartTime - 1,
		//MultipleList:   conf.GetServer().MultipleList,
		//JettonList:     conf.GetServer().JettonList,
		Status: 1,
	})
}

//机器人登陆
func handlerRobotLogin(args []interface{}) {

	m := args[0].(*msg.Game_C_RobotLogin)
	a := args[1].(gate.Agent)
	//用户登陆
	errCode, errMsg := mysql.GameClient.UserLogin(a, m.UserID, "")
	// 登录失败
	if errCode != common.StatusOK {
		_ = log.Logger.Errorf("handlerRobotLogin登录失败 mysql :%d---%s---%d", errCode, errMsg, m.UserID)
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ServerError,
			ErrorMsg:  errMsg,
		})
		a.Close()
		return
	}

	a.SetUserData(m.UserID)
	value, ok := user.List.Load(m.UserID)
	if !ok {
		_ = log.Logger.Errorf("handlerRobotLogin user.List.Load err")
		return
	}
	value.(*user.Item).Agent = a
	value.(*user.Item).BatchID = m.BatchID
	//登陆成功
	a.WriteMsg(&msg.Game_S_LoginSuccess{
		GameQZTime:     conf.GetServer().GameQZTime,
		GameJettonTime: conf.GetServer().GameJettonTime,
		GameTPTime:     conf.GetServer().GameTPTime,
		GameStartTime:  conf.GetServer().GameStartTime,
		//MultipleList:   conf.GetServer().MultipleList,
		//JettonList:     conf.GetServer().JettonList,
		Status: 2,
	})
}

//用户坐下
func handlerUserSitDown(args []interface{}) {
	fmt.Print("handlerUserSitDown ")
	_ = args[0].(*msg.Game_C_UserSitDown)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		_ = log.Logger.Errorf("handlerUserSitDown %s", "坐下失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError1,
			ErrorMsg:  "坐下失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusFree {
		_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 用户状态异常", value.(*user.Item).Status)
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.SitDownError2,
		//	ErrorMsg:  "坐下失败, 用户状态异常",
		//})
		//a.Close()
		return
	}
	// 判断金币是否足够
	if store.GameControl.GetGameInfo().DeductionsType == 0 {
		if value.(*user.Item).UserGold < store.GameControl.GetGameInfo().MinEnterScore {
			_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 余额不足!")
			value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.JCError1,
				ErrorMsg:  "退出房间, 余额不足!",
			})
			//value.(*user.Item).Close()
			//// map 中移除
			//user.List.Delete(value.(*user.Item).UserID)
			return
		}
	} else {
		if value.(*user.Item).UserDiamond < store.GameControl.GetGameInfo().MinEnterScore {
			_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 余额不足!")
			value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.JCError1,
				ErrorMsg:  "退出房间, 余额不足!",
			})
			//value.(*user.Item).Close()
			//// map 中移除
			//user.List.Delete(value.(*user.Item).UserID)
			return
		}
	}
	// 查询空闲座位
	//var tableId int32
	//for key, _ := range table.List {
	//	tableId = table.List[key].GetFreeTableID()
	//	if tableId >= 0 {
	//		break
	//	}
	//}
	//if tableId < 0 {
	//	_ = log.Logger.Errorf("handlerUserSitDown %s", "坐下失败, 没有空闲的桌子")
	//	a.WriteMsg(&msg.Game_S_ReqlyFail{
	//		ErrorCode: global.SitDownError2,
	//		ErrorMsg:  "坐下失败, 没有空闲的桌子",
	//	})
	//	a.Close()
	//	return
	//}
	//table.List[tableId].OnActionUserSitDown(value.(*user.Item))
	// 添加匹配池
	fmt.Printf("%v\n", value.(*user.Item).UserID)
	table.ADDQueueInfo(value.(*user.Item))
}

//用户起立
func handlerUserStandUp(args []interface{}) {
	_ = args[0].(*msg.Game_C_UserStandUp)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		_ = log.Logger.Errorf("handlerUserStandUp %s", "起立失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.StandUpError1,
			ErrorMsg:  "起立失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("handlerUserStandUp %s", "起立失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.StandUpError1,
			ErrorMsg:  "起立失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusFree {
		_ = log.Logger.Errorf("handlerUserStandUp %s", "起立失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.StandUpError2,
		//	ErrorMsg:  "起立失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	//table.List[value.(*user.Item).TableID].OnActionUserStandUp(value.(*user.Item), false)
	if value.(*user.Item).IsRobot() {
		//if _, exist := table.GetQueuePool().RobotQueue.Load(value.(*user.Item).UserID); exist {
		//	table.GetQueuePool().RobotQueue.Delete(value.(*user.Item).UserID)
		//	table.GetQueuePool().RobotCount--
		//	log.Logger.Debug("del to d UserID ", value.(*user.Item).UserID)
		//	mysql.InitRobotLockInfoByUID(mysql.GameClient.GetXJGameDB, value.(*user.Item).UserID)
		//	value.(*user.Item).Close()
		//	// map 中移除
		//	user.List.Delete(value.(*user.Item).UserID)
		//}
		table.GetQueuePool().RobotQueue.Range(func(key, userInfo interface{}) bool {
			if table.GetQueuePool().UserCount != 0 || userInfo.(*user.Item).ChairID >= 0 {
				return false
			}
			if value.(*user.Item).UserID == userInfo.(*user.Item).UserID {
				table.GetQueuePool().RobotQueue.Delete(value.(*user.Item).UserID)
				table.GetQueuePool().RobotCount--
				log.Logger.Debug("del to  UserID ", value.(*user.Item).UserID)
				mysql.InitRobotLockInfoByUID(mysql.GameClient.GetXJGameDB, value.(*user.Item).UserID)
				value.(*user.Item).Close()
				// map 中移除
				user.List.Delete(value.(*user.Item).UserID)
				return false
			}
			return true
		})
	}
}

//用户抢庄
func handlerUserQZ(args []interface{}) {
	m := args[0].(*msg.Game_C_UserQZ)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		log.Logger.Errorf("handlerUserQZ %s", "抢庄失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.QZError1,
			ErrorMsg:  "抢庄失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		log.Logger.Errorf("handlerUserQZ %s", "抢庄失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.QZError1,
			ErrorMsg:  "抢庄失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserQZ %s", "抢庄失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.QZError2,
		//	ErrorMsg:  "抢庄失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	//用户抢庄
	table.List[value.(*user.Item).TableID].OnUserQZ(value.(*user.Item), m)
}

//用户下注
func handlerUserJetton(args []interface{}) {
	m := args[0].(*msg.Game_C_UserJetton)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		_ = log.Logger.Errorf("handlerUserJetton %s", "下注失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("handlerUserJetton %s", "下注失败, 用户未绑定")

		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.JettonError1,
			ErrorMsg:  "下注失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserJetton %s", "下注失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError2,
		//	ErrorMsg:  "下注失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	if !value.(*user.Item).IsRobot() {
		log.Logger.Debugf("玩家下注 OnUserPlaceJetton :%v ", value.(*user.Item).UserID)
	}
	table.List[value.(*user.Item).TableID].OnUserPlaceJetton(value.(*user.Item), m)
}

//用户摊牌
func handlerUserTP(args []interface{}) {
	m := args[0].(*msg.Game_C_UserTP)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		_ = log.Logger.Errorf("handlerUserTP %s", "摊牌失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TPError1,
			ErrorMsg:  "摊牌失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("handlerUserTP %s", "摊牌失败, 用户未绑定")

		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TPError1,
			ErrorMsg:  "摊牌失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserTP %s", "摊牌失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.TPError2,
		//	ErrorMsg:  "摊牌失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	//摊牌
	table.List[value.(*user.Item).TableID].OnUserTP(value.(*user.Item), m)
}
