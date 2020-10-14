package db

import (
	"encoding/json"
	"fmt"
	"xj_game_server/public"
	"xj_game_server/public/jwt"
	"xj_game_server/public/redis"
	"xj_game_server/server/login/conf"
	"xj_game_server/server/login/msg"
	"xj_game_server/util/leaf/log"
	"strconv"
	"strings"
	"time"
)

var LoginRedisClient *LoginRedis

type LoginRedis struct {
	*redis.Redis
	LoginID string
}

func init() {
	redis.Client.OnInit()
	LoginRedisClient = &LoginRedis{
		Redis: redis.Client,
	}
}

func (self *LoginRedis) OnDestroy() {
	self.Redis.OnDestroy()
}

//注册登陆服
func (self *LoginRedis) RegisterLogin() {
	self.LoginID = fmt.Sprintf(public.RedisKeyLoginServer+"%d:", time.Now().UnixNano())
	err := self.Client.LPush(public.RedisKeyLoginServerList, self.LoginID).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}

	var loginInfo struct {
		ServerAddr   string `json:"server_addr"`    //服务器地址
		WsServerAddr string `json:"ws_server_addr"` //websocket地址
	}
	loginInfo.ServerAddr = conf.GetServer().ServerUrl
	loginInfo.WsServerAddr = conf.GetServer().WSUrl
	data, err := json.Marshal(&loginInfo)
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}
	err = self.Client.Set(self.LoginID, string(data), 60*time.Second).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}
	go func() {
		for {
			t := time.NewTimer(time.Second * 30)
			select {
			case <-t.C:
				exists := self.Client.Exists(self.LoginID)
				if exists.Val() == 0 {
					self.Client.Set(self.LoginID, string(data), 60*time.Second).Err()
					continue
				}
				err = self.Client.Expire(self.LoginID, 60*time.Second).Err()
				if err != nil {
					_ = log.Logger.Errorf("RegisterLogin err%v", err)
					return
				}
			}
		}
	}()
}

//注销登录
func (self *LoginRedis) CancelLogin() {
	err := self.Client.LRem(public.RedisKeyLoginServerList, 0, self.LoginID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelLogin err%v", err)
		return
	}

	err = self.Client.Del(self.LoginID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelLogin err%v", err)
		return
	}
}

//生成token
func (self *LoginRedis) MakeToken(userID int32) string {
	token, _ := jwt.EasyToken{
		Username: strconv.Itoa(int(userID)),
	}.GetToken()
	err := self.Client.Set(fmt.Sprintf(public.RedisKeyToken+"%d:", userID), token, -1).Err()
	if err != nil {
		_ = log.Logger.Errorf("MakeToken err%v", err)
		return ""
	}
	return token
}

//加载游戏列表
func (self *LoginRedis) LoadGameList() []*msg.GameInfo {
	gameServerList, err := self.Client.LRange(public.RedisKeyGameServerList, 0, -1).Result()
	if err != nil {
		_ = log.Logger.Errorf("LoadGameList err%v", err)
		return nil
	}

	var gameInfo struct {
		GameID         int32   `json:"game_id"`         //游戏ID
		KindID         int32   `json:"kind_id"`         //游戏种类编号
		ServerAddr     string  `json:"server_addr"`     //服务器地址
		WSAddr         string  `json:"ws_addr"`         //Websocket地址
		GameName       string  `json:"game_name"`       //游戏名
		SortID         int32   `json:"sort_id"`         //排序id
		TableCount     int32   `json:"table_count"`     //桌子数量
		ChairCount     int32   `json:"chair_count"`     //椅子数量
		CellScore      float32 `json:"cell_score"`      //游戏底分
		RevenueRatio   float32 `json:"revenue_ratio"`   //税收比例
		MinEnterScore  float32 `json:"min_enter_score"` //最低进入积分
		DeductionsType int32   `json:"deductions_type"` //扣费类型
	}
	var gameInfoList []*msg.GameInfo
	for _, value := range gameServerList {
		if self.Client.Exists(value).Val() == 0 {
			//将已失效的游戏服务器移除游戏列表
			err := self.Client.LRem(public.RedisKeyGameServerList, 0, value).Err()
			if err != nil {
				_ = log.Logger.Errorf("LoadGameList err%v", err)
				return gameInfoList
			}
			//将已失效的grpc服务器移除grpc列表
			err = self.Client.LRem(public.RedisKeyGameGRPCServerList, 0, strings.Replace(value, "game", "grpc", -1)).Err()
			if err != nil {
				_ = log.Logger.Errorf("LoadGameList err%v", err)
				return gameInfoList
			}
			continue
		}
		data, err := self.Client.Get(value).Result()
		if err != nil {
			_ = log.Logger.Errorf("LoadGameList err%v", err)
			return gameInfoList
		}
		err = json.Unmarshal([]byte(data), &gameInfo)
		if err != nil {
			_ = log.Logger.Errorf("LoadGameList err%v", err)
			return gameInfoList
		}
		var tempMsg = new(msg.GameInfo)
		tempMsg.GameID = gameInfo.GameID
		tempMsg.KindID = gameInfo.KindID
		tempMsg.ServerAddr = gameInfo.ServerAddr
		tempMsg.WsAddr = gameInfo.WSAddr
		tempMsg.GameName = gameInfo.GameName
		tempMsg.SortID = gameInfo.SortID
		tempMsg.TableCount = gameInfo.TableCount
		tempMsg.ChairCount = gameInfo.ChairCount
		tempMsg.CellScore = gameInfo.CellScore
		tempMsg.RevenueRatio = gameInfo.RevenueRatio
		tempMsg.MinEnterScore = gameInfo.MinEnterScore
		tempMsg.DeductionsType = gameInfo.DeductionsType
		gameInfoList = append(gameInfoList, tempMsg)
	}
	return gameInfoList
}
