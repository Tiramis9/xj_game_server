/*
 * @Author: yhlyl
 * @Date: 2019-11-27 14:48:15
 * @LastEditTime: 2019-11-27 15:02:47
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/101_longhudou/robot/robot/robot.go
 * @https://github.com/android-coco
 */
package robot

import (
	"fmt"
	"sync"
	"time"
	"xj_game_server/game/201_qiangzhuangniuniu/conf"
	"xj_game_server/game/201_qiangzhuangniuniu/global"
	"xj_game_server/game/201_qiangzhuangniuniu/msg"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/util"
)

// RobotList 机器人列表 uid --> Robot
//var List = make(map[int32]*Robot)
var List sync.Map

// Robot 机器人
type Robot struct {
	*robot.Agent         // 机器人连接代理
	batchID      int32   // 批次id
	userID       int32   // 用户id
	userChairID  int32   // 座位号
	playStatus   bool    //桌子状态
	gold         float32 //用户金币
	diamond      float32 //用户余额
	isBanker     bool    //是否庄家
	jetton       []int32 // 可以下注倍数
	multiple     []int32 // 可以抢庄倍数

}

func (r *Robot) InitJetton(jetton []int32) {
	r.jetton = jetton
}
func (r *Robot) GetJetton()[]int32 {
	return r.jetton
}
func (r *Robot) InitMultiple(multiple []int32) {
	r.multiple = multiple
}
func (r *Robot) GetMultiple()[]int32 {
	return r.multiple
}
// OnInit 初始化
func (r *Robot) OnInit(userID int32, batchID int32, gate *gate.Gate, userCallBack func(args []interface{})) {
	r.Agent = new(robot.Agent)
	r.Agent.OnInit(gate, userCallBack)
	r.batchID = batchID
	r.userID = userID
	r.userChairID = -1
}

//GetUserID 获取用户id
func (r *Robot) GetUserID() int32 {
	return r.userID
}

//GetBatchID 获取批次号
func (r *Robot) GetBatchID() int32 {
	return r.batchID
}

//GetGold 获取金币
func (r *Robot) GetGold() float32 {
	return r.gold
}

//GetDiamond 获取余额
func (r *Robot) GetDiamond() float32 {
	return r.diamond
}

// Login 登录
func (r *Robot) Login() {
	r.WriteMsg(&msg.Game_C_RobotLogin{
		UserID:  r.userID,
		BatchID: r.batchID,
	})
}

// Assignment 给机器人给金币
func (r *Robot) Assignment() {
	//batchId := r.GetBatchID()
	//maxCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxCoin
	//minCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMinCoin
	//
	//maxDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxDiamond
	//minDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMinDiamond

	maxDiamond := float32(store.GameControl.GetGameInfo().ChairCount*conf.GetServer().MultipleList[len(conf.GetServer().MultipleList)-1]*
		conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) *
		store.GameControl.GetGameInfo().CellScore * global.AreaMultiple[len(global.AreaMultiple)-1] * 2

	minDiamond := float32(store.GameControl.GetGameInfo().ChairCount*conf.GetServer().MultipleList[1]*
		conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1]) *
		store.GameControl.GetGameInfo().CellScore * global.AreaMultiple[len(global.AreaMultiple)-1]

	//coin := float32(util.RandInterval(int32(minCoin), int32(maxCoin)-1))
	diamond := float32(util.RandInterval(int32(minDiamond), int32(maxDiamond)-1))

	value, ok := user.List.Load(r.userID)
	if ok {
		value.(*user.Item).UserGold = 0
		value.(*user.Item).UserDiamond = diamond
	}
	r.gold = 0
	r.diamond = diamond
}

//修改准备状态 坐下
func (r *Robot) SetPlayStatus(playStatus bool) {
	r.playStatus = playStatus
}

//获取准备状态
func (r *Robot) GetPlayStatus() bool {
	return r.playStatus
}

//修改用户座位号
func (r *Robot) SetUserChairID(userChairID int32) {
	r.userChairID = userChairID
}

//获取用户座位号
func (r *Robot) GetUserChairID() int32 {
	return r.userChairID
}

//修改庄家
func (r *Robot) SetBankerByChair(userChairID int32) {
	r.isBanker = r.userChairID == userChairID
}

//获取庄家
func (r *Robot) GetBanker() bool {
	return r.isBanker
}

//SitDown 坐下
func (r *Robot) SitDown() {
	r.WriteMsg(&msg.Game_C_UserSitDown{
		//ChairID: -1,
	})
}

//StandUp 起立
func (r *Robot) StandUp() {
	r.WriteMsg(&msg.Game_C_UserStandUp{})
}

//Jetton 下注
func (r *Robot) Jetton(multiple int32) {
	fmt.Println("=====:", multiple)
	r.WriteMsg(&msg.Game_C_UserJetton{
		Multiple: multiple,
	})
}

//UserQZ 抢庄
func (r *Robot) Qz(multiple int32) {
	r.WriteMsg(&msg.Game_C_UserQZ{
		Multiple: multiple,
	})
}

//UserTP 摊牌
func (r *Robot) TP() {
	r.WriteMsg(&msg.Game_C_UserTP{})
}

//CheckBatchTimeOut 检查批次是否过期
func (r *Robot) CheckBatchTimeOut() bool {
	// 批次号是否过期，过期必退
	_, ok := robot.RobotConfigItem.GetConfig()[r.batchID]
	return ok
}

// 随机起立在坐下
func (r *Robot) RandStandUp() {
	//var isStandUp = false
	t := time.NewTimer(time.Millisecond * 100)
	go func() {
		for {
			select {
			case <-t.C:
				//if !isStandUp {
				//	r.StandUp()
				//	isStandUp = true
				//	t.Reset(time.Second * 2)
				//} else {
				r.SitDown()
				return
				//	}
			}
		}
	}()
}
