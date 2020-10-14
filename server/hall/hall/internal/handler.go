package internal

import (
	"fmt"
	"reflect"
	"sync"
	"time"
	"xj_game_server/game/public/redis"
	"xj_game_server/public"
	"xj_game_server/server/hall/conf"
	"xj_game_server/server/hall/db"
	"xj_game_server/server/hall/msg"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
)

var ConnList sync.Map

type Client struct {
	gate.Agent
	Token      string
	Uid        int32
	Ticker     *time.Ticker
	Stop       chan struct{}
	PlatformID int32 // 登录设备
}

func init() {
	handler(&msg.Hall_C_Msg{}, handleHall)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//处理大厅tcp消息
func handleHall(args []interface{}) {
	m := args[0].(*msg.Hall_C_Msg)
	a := args[1].(gate.Agent)
	if m.Token == "" {
		_ = log.Logger.Errorf("handleHall登录失败token不能为空:%s", m.Token)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  "token 不能为空!",
		})
		a.Close()
		return
	}
	fmt.Println("token", m.Token, "--开始连接")
	//验证token
	userID, err := redis.GameClient.UnmarshalToken(m.Token)
	if err != nil {
		_ = log.Logger.Errorf("handleHall登录失败token:%s userId:%d err:%v", m.Token, userID,err)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  err.Error(),
		})
		a.Close()
		return
	}
	_, ok := ConnList.Load(userID)
	if ok {
		_ = log.Logger.Errorf("handleHall登录失败请勿重复登录token:%s userId:%d", m.Token, userID)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  "请勿重复登录",
		})
		a.Close()
		return
	}
	fmt.Println("user_id", userID, "--开始连接")
	accountInfo, errMysql := db.HallMysqlClient.GetAccountsInfoByUserID(userID)
	if errMysql != nil {
		_ = log.Logger.Errorf("handleHall登录失败token:%s userId:%d err:%v", m.Token, userID,errMysql)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  errMysql.Error(),
		})
		a.Close()
		return
	}

	fmt.Println("PlatformID", accountInfo.PlatformID)
	if accountInfo.PlatformID == 0 {
		_ = log.Logger.Errorf("handleHall登录失败token:%s userId:%d", m.Token, userID)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  "连接平台非法",
		})
		a.Close()
		return
	}

	msgData, errPlatform := db.HallMysqlClient.GetUerByUid(userID, accountInfo.PlatformID)
	if errPlatform != nil {
		_ = log.Logger.Errorf("handleHall登录失败token:%s userId:%d", m.Token, userID)
		a.WriteMsg(&msg.Hall_S_Fail{
			ErrorCode: 1001,
			ErrorMsg:  errPlatform.Error(),
		})
		a.Close()
		return
	}
	a.SetUserData(userID)
	a.WriteMsg(msgData)

	currentClient := &Client{
		Agent:      a,
		Token:      m.Token,
		Uid:        userID,
		Ticker:     time.NewTicker(time.Duration(conf.GetServer().HeartbeatTime) * time.Second),
		Stop:       make(chan struct{}),
		PlatformID: accountInfo.PlatformID,
	}
	ConnList.Store(currentClient.Uid, currentClient)
	go heartBeat(currentClient)
	fmt.Println("user_id", userID, "--心跳连接成功")
}

func heartBeat(currentClient *Client) {
	defer currentClient.Ticker.Stop()
	for {
		select {
		case <-currentClient.Ticker.C:
			//fmt.Printf("\n心跳检测:%d",currentClient.Uid)
			userID, err := redis.GameClient.UnmarshalToken(currentClient.Token)
			if err != nil {
				_ = log.Logger.Errorf("heartBeat登录失败token:%s userId:%d err:%v", currentClient.Token, currentClient.Uid,err)
				currentClient.WriteMsg(&msg.Hall_S_Fail{
					ErrorCode: 1001,
					ErrorMsg:  err.Error(),
				})
				goto STOP
			}
			_ = redis.GameClient.Client.Set(fmt.Sprintf(public.RedisKeyToken+"%d:", userID), currentClient.Token, 60*time.Second).Err()
			//if redis.GameClient.CheckGameVersionByPlatform(currentClient.PlatformID) {
			//	fmt.Println("游戏版本：更新")
			//	redis.GameClient.DelGameVersionByPlatform(currentClient.PlatformID)
			//	//心跳
			//	ConnList.Range(func(key, value interface{}) bool {
			//		if value.(*Client).PlatformID != currentClient.PlatformID {
			//			return true
			//		}
			//		change, err1 := db.HallMysqlClient.GetGameInfoChange(value.(*Client).PlatformID)
			//		//msgData, err1 := db.HallMysqlClient.GetUerByUid(userID, value.(*Client).PlatformID)
			//		if err1 != nil {
			//			_ = log.Logger.Errorf("handleHall心跳失败 userId:%d", currentClient.Uid)
			//			currentClient.WriteMsg(&msg.Hall_S_Fail{
			//				ErrorCode: 1003,
			//				ErrorMsg:  err1.Error(),
			//			})
			//			currentClient.Close()
			//			return false
			//		}
			//		//value.(*Client).WriteMsg(msgData)
			//		value.(*Client).WriteMsg(change)
			//		return true
			//	})
			//
			//}
			//// 游戏房间信息变更
			//if redis.GameClient.CheckRoomChange() {
			//	msgData, err1 := db.HallMysqlClient.GetUerByUid(userID, currentClient.PlatformID)
			//	if err1 != nil {
			//		_ = log.Logger.Errorf("heartBeat心跳失败 userId:%d", currentClient.Uid)
			//		currentClient.WriteMsg(&msg.Hall_S_Fail{
			//			ErrorCode: 1003,
			//			ErrorMsg:  err1.Error(),
			//		})
			//		currentClient.Close()
			//		currentClient.Stop <- struct{}{}
			//		return
			//	}
			//	//currentClient.WriteMsg(msgData)
			//	ConnList.Range(func(key, value interface{}) bool {
			//		value.(*Client).WriteMsg(msgData)
			//		return true
			//	})
			//	redis.GameClient.DelRoomChange()
			//}
			// 充值，结算金币变更
			//if redis.GameClient.GetRecharge(userID) {
			if !redis.GameClient.IsExistsDiamond(userID) {
				scoreInfo, _ := db.HallMysqlClient.GetGameScoreInfoByUserId(userID)
				redis.GameClient.SetDiamond(userID, float32(scoreInfo.Diamond))
			}

			userDiamond, _ := redis.GameClient.GetDiamond(userID)
			currentClient.WriteMsg(&msg.UserInfoChange{UserInfo: &msg.UserInfo{Diamond: userDiamond}, SendTime: time.Now().Format("2006-01-02 15:04:05")})
			//删除标记
			//redis.GameClient.DelRecharge(userID)
			//}
		case <-currentClient.Stop:
			goto STOP
		}
	}
STOP:
	_ = redis.GameClient.Client.Del(fmt.Sprintf(public.RedisKeyToken+"%d:", currentClient.Uid)).Err()
	ConnList.Delete(currentClient.Uid)
	fmt.Println("心跳结束", currentClient.Uid)
}
