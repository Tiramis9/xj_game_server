package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	golog "log"
	"strconv"
	"time"
	"xj_game_server/game/public/conf"
	"xj_game_server/game/public/store"
	"xj_game_server/public"
	"xj_game_server/public/jwt"
	"xj_game_server/public/redis"
	"xj_game_server/util/leaf/log"
)

var GameClient *GameRedis

type GameRedis struct {
	*redis.Redis
	GameID        string
	GRpcID        string
	HeartbeatTime int32
}

func init() {
	//初始化
	redis.Client.OnInit()
	GameClient = &GameRedis{
		Redis: redis.Client,
	}
}

func (self *GameRedis) OnDestroy() {
	//注销游戏
	self.CancelGame()
	//注销grpc
	self.CancelGRpc()

	//关闭redis连接
	self.Redis.OnDestroy()
}

//注册游戏
func (self *GameRedis) RegisterGame() {
	self.GameID = fmt.Sprintf(public.RedisKeyGameServer+"%d:%d:", conf.GetService().KindID, conf.GetService().GameID)
	//先删除
	self.Client.LRem(public.RedisKeyGameServerList, 0, self.GameID)
	err := self.Client.LPush(public.RedisKeyGameServerList, self.GameID).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterGame err %v", err)
		golog.Fatal(err.Error())
	}
	//查询数据库 得到游戏配置
	var gameInfo struct {
		GameID         int32   `json:"game_id"`         //游戏ID
		KindID         int32   `json:"kind_id"`         //游戏种类编号
		ServerAddr     string  `json:"server_addr"`     //服务器地址
		WSAddr         string  `json:"ws_addr"`         //webSocket地址
		GameName       string  `json:"game_name"`       //游戏名
		KindName       string  `json:"kind_name"`       //游戏类型名
		SortID         int32   `json:"sort_id"`         //排序id
		TableCount     int32   `json:"table_count"`     //桌子数量
		ChairCount     int32   `json:"chair_count"`     //椅子数量
		CellScore      float32 `json:"cell_score"`      //游戏底分
		RevenueRatio   float32 `json:"revenue_ratio"`   //税收比例
		MinEnterScore  float32 `json:"min_enter_score"` //最低进入积分
		DeductionsType int32   `json:"deductions_type"` //扣费类型
	}
	gameInfo.GameID = conf.GetService().GameID
	gameInfo.KindID = conf.GetService().KindID
	gameInfo.ServerAddr = conf.GetService().ServerUrl
	gameInfo.WSAddr = conf.GetService().WSUrl
	gameInfo.GameName = store.GameControl.GetGameInfo().GameName
	gameInfo.KindName = store.GameControl.GetGameInfo().KindName
	gameInfo.SortID = store.GameControl.GetGameInfo().SortID
	gameInfo.TableCount = store.GameControl.GetGameInfo().TableCount
	gameInfo.ChairCount = store.GameControl.GetGameInfo().ChairCount
	gameInfo.CellScore = store.GameControl.GetGameInfo().CellScore
	gameInfo.MinEnterScore = store.GameControl.GetGameInfo().MinEnterScore
	gameInfo.DeductionsType = store.GameControl.GetGameInfo().DeductionsType
	data, err := json.Marshal(&gameInfo)
	if err != nil {
		_ = log.Logger.Errorf("RegisterGame err %v", err)
		golog.Fatal(err.Error())
	}
	err = self.Client.Set(self.GameID, string(data), conf.GetService().HeartbeatTime*time.Second).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterGame err %v", err)
		golog.Fatal(err.Error())
	}
	go func() {
		for {
			t := time.NewTimer(conf.GetService().HeartbeatTime / 2 * time.Second)
			select {
			case <-t.C:
				exists := self.Client.Exists(self.GameID)
				if exists.Val() == 0 {
					self.Client.Set(self.GameID, string(data), conf.GetService().HeartbeatTime*time.Second).Err()
					continue
				}
				err = self.Client.Expire(self.GameID, conf.GetService().HeartbeatTime*time.Second).Err()
				if err != nil {
					_ = log.Logger.Errorf("RegisterGrpc err %v", err)
					golog.Fatal(err.Error())
				}
			}
		}
	}()
}

//注册grpc
func (self *GameRedis) RegisterGrpc() {
	self.GRpcID = fmt.Sprintf(public.RedisKeyGameGRPCServer+"%d:%d:", conf.GetService().KindID, conf.GetService().GameID)
	//先删除
	self.Client.LRem(public.RedisKeyGameGRPCServerList, 0, self.GRpcID)
	err := self.Client.LPush(public.RedisKeyGameGRPCServerList, self.GRpcID).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterGrpc err %v", err)
		golog.Fatal(err.Error())
	}

	var grpcInfo struct {
		GrpcAddr string `json:"grpc_addr"` //grpc地址
	}
	grpcInfo.GrpcAddr = conf.GetService().GRPCAddr
	data, err := json.Marshal(&grpcInfo)
	if err != nil {
		_ = log.Logger.Errorf("RegisterGrpc err %v", err)
		golog.Fatal(err.Error())
	}
	err = self.Client.Set(self.GRpcID, string(data), conf.GetService().HeartbeatTime*time.Second).Err()
	if err != nil {
		_ = log.Logger.Errorf("RegisterGrpc err %v", err)
		golog.Fatal(err.Error())
	}
	go func() {
		for {
			t := time.NewTimer(conf.GetService().HeartbeatTime / 2 * time.Second)
			select {
			case <-t.C:
				exists := self.Client.Exists(self.GRpcID)
				if exists.Val() == 0 {
					self.Client.Set(self.GRpcID, string(data), conf.GetService().HeartbeatTime*time.Second).Err()
					continue
				}
				err = self.Client.Expire(self.GRpcID, conf.GetService().HeartbeatTime*time.Second).Err()
				if err != nil {
					_ = log.Logger.Errorf("RegisterGrpc err %v", err)
					golog.Fatal(err.Error())
				}
			}
		}
	}()
}

//注销游戏
func (self *GameRedis) CancelGame() {
	err := self.Client.LRem(public.RedisKeyGameServerList, 0, self.GameID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelGame err %v", err)
		golog.Fatal(err.Error())
	}

	err = self.Client.Del(self.GameID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelGame err %v", err)
		golog.Fatal(err.Error())
	}
}

//注销grpc
func (self *GameRedis) CancelGRpc() {
	err := self.Client.LRem(public.RedisKeyGameGRPCServerList, 0, self.GRpcID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelGRpc err %v", err)
		golog.Fatal(err.Error())
	}

	err = self.Client.Del(self.GRpcID).Err()
	if err != nil {
		_ = log.Logger.Errorf("CancelGRpc err %v", err)
		golog.Fatal(err.Error())
	}
}

//

//解析token
func (self *GameRedis) UnmarshalToken(token string) (int32, error) {
	//解析token
	_, id, err := jwt.EasyToken{}.ValidateToken(token)
	if err != nil {
		_ = log.Logger.Errorf("UnmarshalToken1 err %v", err)
		return -1, err
	}

	//验证token
	userID, err1 := strconv.Atoi(id)
	if err1 != nil {
		_ = log.Logger.Errorf("UnmarshalToken2 err %v", err1)
		return -1, err1
	}
	tempToken, err2 := self.Client.Get(fmt.Sprintf(public.RedisKeyToken+"%d:", userID)).Result()
	if err2 != nil {
		_ = log.Logger.Errorf("UnmarshalToken3 err %v", err2)
		return -1, err2
	}
	if token != tempToken {
		_ = log.Logger.Errorf("UnmarshalToken4 err %s", "登陆过期, 请重新登陆")
		return -1, errors.New("登陆过期, 请重新登陆")
	}

	return int32(userID), nil
}

//充值通知
func (self *GameRedis) RegisterRecharge(uid int32) {
	rechargeKey := fmt.Sprintf(public.RedisKeyUserRecharge+"%d:", uid)
	self.Client.Set(rechargeKey, uid, -1)
}

//是否线下充值了
func (self *GameRedis) GetRecharge(uid int32) bool {
	rechargeKey := fmt.Sprintf(public.RedisKeyUserRecharge+"%d:", uid)
	isExists := self.Client.Exists(rechargeKey).Val()
	if isExists == 0 {
		return false
	}
	return true
}

// 金币是否变更
func (self *GameRedis) GetDiamond(uid int32) (float64, error) {
	diamondKey := fmt.Sprintf(public.RedisKeyUserDiamondRecharge+"%d:", uid)
	value := self.Client.Get(diamondKey).Val()
	return strconv.ParseFloat(value, 10)
}

func (self *GameRedis) IsExistsDiamond(uid int32) bool {
	rechargeKey := fmt.Sprintf(public.RedisKeyUserDiamondRecharge+"%d:", uid)
	isExists := self.Client.Exists(rechargeKey).Val()
	if isExists == 0 {
		return false
	}
	return true
}

// 金币是否变更
func (self *GameRedis) SetDiamond(uid int32, diamond float32) error {
	if diamond < 0 {
		diamond = 0
	}
	diamondKey := fmt.Sprintf(public.RedisKeyUserDiamondRecharge+"%d:", uid)
	return self.Client.Set(diamondKey, diamond, -1).Err()
}

//删除充值标记
func (self *GameRedis) DelRecharge(uid int32) {
	rechargeKey := fmt.Sprintf(public.RedisKeyUserRecharge+"%d:", uid)
	isExists := self.Client.Exists(rechargeKey).Val()
	if isExists == 1 {
		self.Client.Del(rechargeKey)
	}
}

// 游戏版本是否存在
func (self *GameRedis) CheckGameVersionByPlatform(Platform int32) bool {
	changeKey := fmt.Sprintf(public.RedisGameVersionChange+"%d:", Platform)
	isExists := self.Client.Exists(changeKey).Val()
	if isExists == 0 {
		return false
	}
	return true
}

//删除版本是否存在
func (self *GameRedis) DelGameVersionByPlatform(Platform int32) {
	rechargeKey := fmt.Sprintf(public.RedisGameVersionChange+"%d:", Platform)
	isExists := self.Client.Exists(rechargeKey).Val()
	if isExists == 1 {
		self.Client.Del(rechargeKey)
	}
}

// 游戏房间变更
func (self *GameRedis) CheckRoomChange() bool {
	isExists := self.Client.Exists(public.RedisRoomInfoChange).Val()
	if isExists == 0 {
		return false
	}
	return true
}

// 游戏房间变更
func (self *GameRedis) SetRoomChange() {
	self.Client.Set(public.RedisRoomInfoChange, true, -1)
}

//删除游戏房间变更
func (self *GameRedis) DelRoomChange() {
	isExists := self.Client.Exists(public.RedisRoomInfoChange).Val()
	if isExists == 1 {
		self.Client.Del(public.RedisRoomInfoChange)
	}
}
