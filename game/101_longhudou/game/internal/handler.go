package internal

import (
	"fmt"
	"reflect"
	"time"
	"xj_game_server/game/101_longhudou/conf"
	"xj_game_server/game/101_longhudou/game/table"
	"xj_game_server/game/101_longhudou/global"
	"xj_game_server/game/101_longhudou/msg"
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
	handler(&msg.Game_C_UserJetton{}, handlerUserJetton)
	handler(&msg.Game_C_UserList{}, handlerUserList)
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
			ErrorMsg:  err.Error(),
		})
		a.Close()
		return
	}

	//验证是否重复登陆/断线重连
	//userItem, ok := user.List[userID]
	userItem, ok := user.List.Load(userID)
	if ok && userItem.(*user.Item).TableID != -1 {
		if userID == 138684 || userID == 138677 {
			fmt.Println(userItem.(*user.Item).TableID, "  handlerTokenLogin 重连", "handlerTokenLogin重连", time.Now())
		}
		//绑定agent
		oldAgent := userItem.(*user.Item).Agent
		a.SetUserData(userID)
		userItem.(*user.Item).Agent = a

		//发送登陆成功
		var data = make([]*msg.Game_S_MsgHallHistory, 0)
		for _, v := range table.List {
			msgHall := &msg.Game_S_MsgHallHistory{
				TableID:         v.GetTableID(),
				GameJettonTime:  conf.GetServer().GameJettonTime,
				GameLotteryTime: conf.GetServer().GameLotteryTime,
				JettonList:      conf.GetServer().JettonList,
				LotteryRecord:   v.GetLotteryRecord(),
				GameStatus:      v.GetGameStatus(),
				SceneStartTime:  v.GetSceneStartTime(),
				UserCount:       v.GetUserCount(),
			}
			data = append(data, msgHall)
		}
		userItem.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
			Status:          2,
			TableID:         userItem.(*user.Item).TableID,
			GameJettonTime:  conf.GetServer().GameJettonTime,
			GameLotteryTime: conf.GetServer().GameLotteryTime,
			Data:            data,
		})

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
	//for _, v := range table.List {
	//	go v.SendLotteryRecord(value.(*user.Item), 0)
	//}
	//go table.List[userItem.(*user.Item).TableID].SendLotteryRecord(value.(*user.Item), 0)
	var data = make([]*msg.Game_S_MsgHallHistory, 0)
	for _, v := range table.List {
		msgHall := &msg.Game_S_MsgHallHistory{
			TableID:         v.GetTableID(),
			GameJettonTime:  conf.GetServer().GameJettonTime,
			GameLotteryTime: conf.GetServer().GameLotteryTime,
			JettonList:      conf.GetServer().JettonList,
			LotteryRecord:   v.GetLotteryRecord(),
			GameStatus:      v.GetGameStatus(),
			SceneStartTime:  v.GetSceneStartTime(),
			UserCount:       v.GetUserCount(),
		}
		data = append(data, msgHall)
	}
	if userID == 138684 || userID == 138677 {
		fmt.Println(value.(*user.Item).TableID, "  handlerTokenLogin 普通登录", "handlerTokenLogin普通登录", time.Now())
	}
	value.(*user.Item).WriteMsg(&msg.Game_S_LoginSuccess{
		Status:          1,
		TableID:         -1,
		GameJettonTime:  conf.GetServer().GameJettonTime,
		GameLotteryTime: conf.GetServer().GameLotteryTime,
		Data:            data,
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
		Data: nil,
	})
}

//用户坐下
func handlerUserSitDown(args []interface{}) {
	m := args[0].(*msg.Game_C_UserSitDown)
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

	//if value.(*user.Item).BatchID == -1 {
	//	fmt.Printf("%c[1;47;31m 用户坐下===== %v %c[0m\n", 0x1B, value.(*user.Item), 0x1B)
	//	fmt.Printf("%c[1;47;31m 用户坐下状态===== %d %c[0m\n", 0x1B, value.(*user.Item).Status, 0x1B)
	//}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusFree {
		_ = log.Logger.Errorf("handlerUserSitDown %s---%d----%d", "坐下失败, 用户状态异常", value.(*user.Item).Status, value.(*user.Item).UserID)
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.SitDownError2,
		//	ErrorMsg:  "坐下失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	//校验数据
	if m.TableID < 0 || m.TableID >= store.GameControl.GetGameInfo().TableCount {
		_ = log.Logger.Errorf("handlerUserSitDown %s---%d", "坐下失败, 无效的桌子号", m.TableID)
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.SitDownError3,
			ErrorMsg:  "坐下失败, 无效的桌子号",
		})
		a.Close()
		return
	}

	table.List[m.TableID].OnActionUserSitDown(value.(*user.Item))
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
	//if value.(*user.Item).BatchID == -1 {
	//	fmt.Printf("%c[1;47;31m 用户起立===== %v %c[0m\n", 0x1B, value.(*user.Item), 0x1B)
	//	fmt.Printf("%c[1;47;31m 用户起立状态===== %d %c[0m\n", 0x1B, value.(*user.Item).Status, 0x1B)
	//}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserStandUp %s====%d", "起立失败, 用户状态异常", value.(*user.Item).Status)
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.StandUpError2,
		//	ErrorMsg:  "起立失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnActionUserStandUp(value.(*user.Item), false)
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
		_ = log.Logger.Errorf("handlerUserJetton %s====%d", "下注失败, 用户状态异常", value.(*user.Item).Status)
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.JettonError2,
		//	ErrorMsg:  "下注失败, 用户状态异常",
		//})
		//a.Close()
		return
	}

	table.List[value.(*user.Item).TableID].OnUserPlaceJetton(value.(*user.Item), m)
}

//获取用户列表
func handlerUserList(args []interface{}) {
	m := args[0].(*msg.Game_C_UserList)
	a := args[1].(gate.Agent)

	//校验用户绑定
	if a.UserData() == nil {
		_ = log.Logger.Errorf("handlerUserList %s", "获取用户列表失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.UserListError1,
			ErrorMsg:  "获取用户列表失败, 用户未绑定",
		})
		a.Close()
		return
	}
	value, ok := user.List.Load(a.UserData().(int32))
	if !ok {
		_ = log.Logger.Errorf("handlerUserList %s", "获取用户列表失败, 用户未绑定")
		a.WriteMsg(&msg.Game_S_ReqlyFail{
			ErrorCode: global.UserListError1,
			ErrorMsg:  "获取用户列表失败, 用户未绑定",
		})
		a.Close()
		return
	}
	//校验用户状态
	if value.(*user.Item).Status != user.StatusPlaying {
		_ = log.Logger.Errorf("handlerUserList %s===%d", "获取用户列表失败, 用户状态异常", value.(*user.Item).Status)
		//a.WriteMsg(&msg.Game_S_ReqlyFail{
		//	ErrorCode: global.UserListError2,
		//	ErrorMsg:  "获取用户列表失败, 用户状态异常",
		//})
		//a.Close()
		return
	}
	table.List[value.(*user.Item).TableID].GetUserList(value.(*user.Item), m)
}
