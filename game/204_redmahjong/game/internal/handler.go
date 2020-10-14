package internal

import (
	"fmt"
	"reflect"
	"xj_game_server/game/204_redmahjong/conf"
	"xj_game_server/game/204_redmahjong/game/table"
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/204_redmahjong/msg"
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
	handler(&msg.Game_C_UserOutCard{}, handlerUserOutCard)
	handler(&msg.Game_C_UserOperate{}, handlerUserOperate)
	handler(&msg.Game_C_UserPrepare{}, handlerUserPrepare)
	handler(&msg.Game_C_UserUnPrepare{}, handlerUserUnPrepare)
	handler(&msg.Game_C_UserTing{}, handlerUserTing)
	handler(&msg.Game_C_ChangeTable{}, handlerChangeTable)
	handler(&msg.Game_C_AutoManage{}, handlerAutoManage)
	handler(&msg.Game_C_UnAutoManage{}, handlerUnAutoManage)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//用户登陆(token)
func handlerTokenLogin(args []interface{}) {
	m := args[0].(*msg.Game_C_TokenLogin)
	a := args[1].(gate.Agent)
	//验证token
	userID, err := redis.GameClient.UnmarshalToken(m.Token)
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
	fmt.Printf("玩家登陆=%v\n", ok)
	if ok {
		//绑定agent
		oldAgent := userItem.(*user.Item).Agent
		a.SetUserData(userID)
		userItem.(*user.Item).Agent = a

		//发送登陆成功
		userItem.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
			GameStartTime:   conf.GetServer().GameStartTime,
			GameOperateTime: conf.GetServer().GameOperateTime,
			GameOutCardTime: conf.GetServer().GameOutCardTime,
			MaCount:         conf.GetServer().MaCount,
			Status:          2,
		})

		//通知上个用户下线
		if userItem.(*user.Item).Status != user.StatusOffline && (a.RemoteAddr().String() != userItem.(*user.Item).RemoteAddr().String()) {
			oldAgent.WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.LoginError,
				ErrorMsg:  "你的账号已在其他设备登录, 如非本人操作请立即修改密码!",
			})
			oldAgent.Close()
		}
		fmt.Printf("玩家断线重连,状态:%v,t:%v,c:%v\n", userItem.(*user.Item).Status, userItem.(*user.Item).TableID, userItem.(*user.Item).ChairID)
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
			ErrorCode: global.ServerError,
			ErrorMsg:  errMsg,
		})
		a.Close()
		return
	}
	a.SetUserData(userID)
	value, ok := user.List.Load(userID)
	if !ok {
		_ = log.Logger.Errorf("handlerTokenLogin user.List.Load err")
		return
	}
	value.(*user.Item).Agent = a
	//发送登陆成功
	value.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
		GameStartTime:   conf.GetServer().GameStartTime,
		GameOperateTime: conf.GetServer().GameOperateTime,
		GameOutCardTime: conf.GetServer().GameOutCardTime,
		MaCount:         conf.GetServer().MaCount,
		Status:          1,
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
		GameStartTime:   conf.GetServer().GameStartTime,
		GameOperateTime: conf.GetServer().GameOperateTime,
		GameOutCardTime: conf.GetServer().GameOutCardTime,
		MaCount:         conf.GetServer().MaCount,
	})
}

//用户坐下
func handlerUserSitDown(args []interface{}) {
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
		_ = log.Logger.Errorf("handlerUserSitDown %s---", "坐下失败, 用户状态异常 xx")
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
	// 检查是否锁定
	lock := mysql.GameClient.IsLock(value.(*user.Item).UserID)
	if lock {
		_ = log.Logger.Errorf("err %v", "坐下失败, 上局游戏未结束")
		value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError1,
			ErrorMsg:  "坐下失败, 上局游戏未结束",
		})
		value.(*user.Item).Close()
		return
	}

	// 判断金币是否足够
	if store.GameControl.GetGameInfo().DeductionsType == 0 {
		if value.(*user.Item).UserGold < store.GameControl.GetGameInfo().MinEnterScore {
			_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 金币不足!")
			value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.SitDownError3,
				ErrorMsg:  "退出房间, 金币不足!",
			})
			value.(*user.Item).Close()
			// map 中移除
			user.List.Delete(value.(*user.Item).UserID)
			return
		}
	} else {
		if value.(*user.Item).UserDiamond < store.GameControl.GetGameInfo().MinEnterScore {
			_ = log.Logger.Errorf("OnActionUserSitDown err %s", "退出房间, 金币不足!")
			value.(*user.Item).WriteMsg(&msg.Game_S_ReqlyFail{
				ErrorCode: global.SitDownError3,
				ErrorMsg:  "退出房间, 余额不足!",
			})
			value.(*user.Item).Close()
			// map 中移除
			user.List.Delete(value.(*user.Item).UserID)
			return
		}
	}
	//1, 加入相应的队列
	//2, 如果有真人开始匹配定时器 3秒
	if value.(*user.Item).IsRobot() {
		_, ok := table.RobotQueue.Load(value.(*user.Item).UserID)
		if ok {
			_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 重复uid", value.(*user.Item).UserID)
			return
		}
		table.RobotQueue.Store(value.(*user.Item).UserID, value.(*user.Item))
		table.RobotCount++
	} else {

		_, ok := table.UserQueue.Load(value.(*user.Item).UserID)
		if ok {
			_ = log.Logger.Errorf("handlerUserSitDown %s---%d,ucount:%v", "坐下失败, 重复uid", value.(*user.Item).UserID, table.UserCount)
			return
		}
		table.UserQueue.Store(value.(*user.Item).UserID, value.(*user.Item))
		table.UserCount++
		log.Logger.Debugf("handlerUserSitDown uid:%v ucount:%v rcount:%v", value.(*user.Item).UserID, table.UserCount, table.RobotCount)
	}
}

//用户起立
func handlerUserStandUp(args []interface{}) {
	_ = args[0].(*msg.Game_C_UserStandUp)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		//_ = log.Logger.Errorf("handlerUserStandUp %s", "起立失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.StandUpError1,
			ErrorMsg:  "起立失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.StandUpError1,
			ErrorMsg:  "起立失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserStandUp %s", "起立失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.StandUpError2,
		//	ErrorMsg:  "起立失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnActionUserStandUp(value.(*user.Item), false)
}

//用户出牌
func handlerUserOutCard(args []interface{}) {
	m := args[0].(*msg.Game_C_UserOutCard)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		//_ = log.Logger.Errorf("handlerUserOutCard %s", "出牌失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OutCardError1,
			ErrorMsg:  "下注失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("handlerUserOutCard %s,%v", "用户未绑定", a.UserData())
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OutCardError1,
			ErrorMsg:  "出牌失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserOutCard %s", "出牌失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.OutCardError1,
		//	ErrorMsg:  "出牌失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnUserOutCard(value.(*user.Item), m)
}

//用户操作
func handlerUserOperate(args []interface{}) {
	m := args[0].(*msg.Game_C_UserOperate)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		//_ = log.Logger.Errorf("handlerUserOperate %s", "操作失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OperateError1,
			ErrorMsg:  "操作失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.OperateError1,
			ErrorMsg:  "操作失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserOperate %s", "操作失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.OperateError2,
		//	ErrorMsg:  "操作失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnUserOperate(value.(*user.Item), m)
}

//准备
func handlerUserPrepare(args []interface{}) {
	m := args[0].(*msg.Game_C_UserPrepare)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError1,
			ErrorMsg:  "准备失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError1,
			ErrorMsg:  "准备失败, 用户未绑定",
		})
		a.Close()
		return
	}
	log.Logger.Debugf("handlerUserPrepare 准备 桌子号:%v,椅子:%v,uid:%v", value.(*user.Item).TableID, value.(*user.Item).ChairID, value.(*user.Item).UserID)
	table.List[value.(*user.Item).TableID].OnUserPrepare(value.(*user.Item), m)
}

//取消准备
func handlerUserUnPrepare(args []interface{}) {
	m := args[0].(*msg.Game_C_UserUnPrepare)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError1,
			ErrorMsg:  "取消准备失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.ZBError1,
			ErrorMsg:  "取消准备失败, 用户未绑定",
		})
		a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnUserUnPrepare(value.(*user.Item), m)
}

//用户听牌
func handlerUserTing(args []interface{}) {
	m := args[0].(*msg.Game_C_UserTing)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "用户听牌失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "用户听牌失败, 用户未绑定",
		})
		a.Close()
		return
	}

	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserOperate %s", "取消准备失败, 用户状态异常")
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.OperateError2,
		//	ErrorMsg:  "取消准备失败, 用户状态异常",
		//})
		//a.Close()
		return
	}
	return
	table.List[value.(*user.Item).TableID].OnUserTing(value.(*user.Item), m)
}

//换桌
func handlerChangeTable(args []interface{}) {
	_ = args[0].(*msg.Game_C_ChangeTable)
	a := args[1].(gate.Agent)

	handlerUserStandUp([]interface{}{&msg.Game_C_UserStandUp{}, a})
	handlerUserSitDown([]interface{}{&msg.Game_C_UserSitDown{}, a})
}

// 托管
func handlerAutoManage(args []interface{}) {
	_ = args[0].(*msg.Game_C_AutoManage)
	a := args[1].(gate.Agent)
	//校验用户绑定
	if a.UserData() == nil {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "托管失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "托管失败, 用户未绑定",
		})
		a.Close()
		return
	}
	table.List[value.(*user.Item).TableID].OnUserAutoManage(value.(*user.Item))
}

// 取消托管
func handlerUnAutoManage(args []interface{}) {
	_ = args[0].(*msg.Game_C_UnAutoManage)
	a := args[1].(gate.Agent)
	//校验用户绑定
	if a.UserData() == nil {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "取消托管失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.TingError1,
			ErrorMsg:  "取消托管失败, 用户未绑定",
		})
		a.Close()
		return
	}
	table.List[value.(*user.Item).TableID].OnUserUnAutoManage(value.(*user.Item))
}
