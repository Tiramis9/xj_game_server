package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xj_game_server/public"
	"xj_game_server/public/jwt"
	"xj_game_server/public/redis"
	"xj_game_server/server/hall/conf"
	"xj_game_server/server/hall/msg"
	"xj_game_server/util/leaf/log"
)

var HallRedisClient *HallRedis

type HallRedis struct {
	*redis.Redis
	HallID string
}

func init() {
	redis.Client.OnInit()
	HallRedisClient = &HallRedis{
		Redis: redis.Client,
	}
}

func (self *HallRedis) OnDestroy() {
	self.Redis.OnDestroy()
}

//注册登陆服
func (self *HallRedis) RegisterLogin() {
	self.HallID = fmt.Sprintf(public.RedisKeyHallServer+"%d:", time.Now().UnixNano())
	err := self.Client.LPush(public.RedisKeyHallServerList, self.HallID).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}

	var hallInfo struct {
		ServerAddr   string `json:"server_addr"`    //服务器地址
		WsServerAddr string `json:"ws_server_addr"` //websocket地址
	}
	hallInfo.ServerAddr = conf.GetServer().ServerUrl
	hallInfo.WsServerAddr = conf.GetServer().WSUrl
	data, err := json.Marshal(&hallInfo)
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}
	err = self.Client.Set(self.HallID, string(data), 60*time.Second).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterLogin err%v", err)
		return
	}
	go func() {
		for {
			t := time.NewTimer(time.Second * 30)
			select {
			case <-t.C:
				exists := self.Client.Exists(self.HallID)
				if exists.Val() == 0 {
					self.Client.Set(self.HallID, string(data), 60*time.Second).Err()
					continue
				}
				err = self.Client.Expire(self.HallID, 60*time.Second).Err()
				if err != nil {
					_ = log.Logger.Errorf("RegisterLogin err%v", err)
					return
				}
			}
		}
	}()
}

//注销登录
func (self *HallRedis) CancelLogin() {
	err := self.Client.LRem(public.RedisKeyHallServerList, 0, self.HallID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelLogin err%v", err)
		return
	}

	err = self.Client.Del(self.HallID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelLogin err%v", err)
		return
	}
}

//解析token
func (self *HallRedis) UnmarshalToken(token string) (int32, error) {
	//解析token
	_, id, err := jwt.EasyToken{}.ValidateToken(token)
	if err != nil {
		_ = log.Logger.Errorf("UnmarshalToken err %v", err)
		return -1, err
	}

	//验证token
	userID, err := strconv.Atoi(id)
	if err != nil {
		_ = log.Logger.Errorf("UnmarshalToken err %v", err)
		return -1, err
	}
	tempToken, err := self.Client.Get(fmt.Sprintf(public.RedisKeyToken+"%d:", userID)).Result()
	if err != nil {
		_ = log.Logger.Errorf("UnmarshalToken err %v", err)
		return -1, err
	}
	if token != tempToken {
		_ = log.Logger.Errorf("UnmarshalToken err %v", err)
		return -1, errors.New("登陆过期, 请重新登陆")
	}

	return int32(userID), nil
}

//加载游戏列表
func (self *HallRedis) LoadGameList() []*msg.GameInfo {
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
		KindName       string  `json:"kind_name"`       //游戏名
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
		tempMsg.KindId = gameInfo.KindID
		tempMsg.GameName = gameInfo.KindName
		tempMsg.GameStatus = 1

		var roomInfos []*msg.RoomInfo
		var roomInfo msg.RoomInfo

		roomInfo.RoomServer = gameInfo.ServerAddr
		roomInfo.RoomStatus = 1
		roomInfo.GameID = gameInfo.GameID
		roomInfo.SortID = gameInfo.SortID
		roomInfo.CellScore = gameInfo.CellScore
		roomInfo.EnterScore = gameInfo.MinEnterScore
		roomInfo.ScoreType = gameInfo.DeductionsType

		var isExist = false
		for k, v := range gameInfoList {
			if v.KindId == gameInfo.KindID {
				isExist = true
				gameInfoList[k].RoomInfo = append(gameInfoList[k].RoomInfo, &roomInfo)
				for i := 0; i < len(gameInfoList[k].RoomInfo)-1; i++ {
					for j := i + 1; j < len(gameInfoList[k].RoomInfo); j++ {
						if gameInfoList[k].RoomInfo[i].GameID > gameInfoList[k].RoomInfo[j].GameID {
							gameInfoList[k].RoomInfo[i], gameInfoList[k].RoomInfo[j] = gameInfoList[k].RoomInfo[j], gameInfoList[k].RoomInfo[i]
						}
					}
				}
				break
			}
		}
		if !isExist {
			roomInfos = append(roomInfos, &roomInfo)
			tempMsg.RoomInfo = roomInfos
			gameInfoList = append(gameInfoList, tempMsg)
		}
	}
	return gameInfoList
}
