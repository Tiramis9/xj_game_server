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
	"sync"
	"time"
	gameLogic "xj_game_server/game/204_redmahjong/game/logic"
	"xj_game_server/game/204_redmahjong/global"
	"xj_game_server/game/204_redmahjong/msg"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/util"
)

// RobotList 机器人列表 uid --> Robot
//var List = make(map[int32]*Robot)
var List sync.Map

// Robot 机器人
type Robot struct {
	*robot.Agent                         // 机器人连接代理
	batchID         int32                // 批次id
	userID          int32                // 用户id
	userChairID     int32                // 座位号
	isPrepare       bool                 //用户准备
	gameStatus      int32                //游戏状态
	gold            float32              //用户金币
	diamond         float32              //用户余额
	userMjData      map[int32]int32      //用户麻将
	withProbability int32                //退出概率
	userDiskMj      *msg.DiskMahjongList //玩家桌面麻将(已操作)
	rounds          int32                //出牌轮数
	nearestMjData   int32                // 最近出的麻将
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
func (r *Robot) getBatchID() int32 {
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
	batchId := r.getBatchID()
	maxCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxCoin
	minCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMinCoin

	maxDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxDiamond
	minDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMinDiamond
	if minDiamond < store.GameControl.GetGameInfo().MinEnterScore {
		minDiamond = store.GameControl.GetGameInfo().MinEnterScore
	}
	coin := float32(util.RandInterval(int32(minCoin), int32(maxCoin)-1))
	diamond := float32(util.RandInterval(int32(minDiamond), int32(maxDiamond)-1))

	value, ok := user.List.Load(r.userID)
	if ok {
		value.(*user.Item).UserGold = coin
		value.(*user.Item).UserDiamond = diamond
	}
	r.gold = coin
	r.diamond = diamond
}

//获取用户准备状态
func (r *Robot) GetUserPrepare() bool {
	return r.isPrepare
}

//修改用户座位号
func (r *Robot) SetUserChairID(userChairID int32) {
	r.userChairID = userChairID
}

//获取用户座位号
func (r *Robot) GetUserChairID() int32 {
	return r.userChairID
}

//SitDown 坐下
func (r *Robot) SitDown() {
	r.WriteMsg(&msg.Game_C_UserSitDown{
		TableID: -1,
		ChairID: -1,
	})
}

//StandUp 起立
func (r *Robot) standUp() {
	r.WriteMsg(&msg.Game_C_UserStandUp{})
}

//CheckBatchTimeOut 检查批次是否过期
func (r *Robot) CheckBatchTimeOut() bool {
	// 批次号是否过期，过期必退
	_, ok := robot.RobotConfigItem.GetConfig()[r.batchID]
	return ok
}

// 游戏开始
func (r *Robot) GameStart(userMjData map[int32]int32) {
	r.gameStatus = global.GameStatusPlay
	r.userMjData = userMjData
	r.userDiskMj = new(msg.DiskMahjongList)
	r.nearestMjData = 0
	r.rounds = 0
}

// 游戏结束，清空数据
func (r *Robot) GameEnd() {
	r.gameStatus = global.GameStatusFree
	r.isPrepare = false
	r.withProbability += 5
	r.userMjData = make(map[int32]int32, 0)
	r.RandStandUp()
}

// 获取游戏状态
func (r *Robot) GetGameStatus() int32 {
	return r.gameStatus
}

// 设置最近出牌
func (r *Robot) SetNearestOutCard(mjData int32) {
	r.nearestMjData = mjData
}

// 设置用户出牌
func (r *Robot) SetUserOutCard(mjData int32, num int32) {

	r.userMjData[mjData] -= num
	if r.userMjData[mjData] == 0 {
		delete(r.userMjData, mjData)
	}
}

func (r *Robot) GetUserCard() map[int32]int32 {
	return r.userMjData
}

// 随机退出或准备 百分百退出
func (r *Robot) RandStandUp() {
	t := time.NewTimer(time.Millisecond * time.Duration(util.RandInterval(500, 1000)))
	go func() {
		for {
			select {
			case <-t.C:
				r.SitDown()
				return
			}
		}
	}()
}

//准备
func (r *Robot) Prepare() {
	//准备消息
	r.Agent.WriteMsg(&msg.Game_C_UserPrepare{})
}

//取消准备
func (r *Robot) UnPrepare() {
	//准备消息
	r.Agent.WriteMsg(&msg.Game_C_UserUnPrepare{})
}

// 检查麻将出牌参数 如果是胡牌或者暗杠,补扛则不出牌
func (r *Robot) CheckOutMj(MjData int32) bool {
	r.userMjData[MjData]++
	response := gameLogic.Client.SendMjResponse(MjData, r.userMjData, r.userDiskMj, r.rounds == 0)
	return response != 0
}

//出牌
func (r *Robot) OutMj() {
	outData := gameLogic.Client.GetOutMjData(r.userMjData)
	log.Logger.Error("UID:", r.userID, r.userMjData, outData)
	r.Agent.WriteMsg(&msg.Game_C_UserOutCard{
		MjData: outData,
	})
}

//操作
func (r *Robot) Operate(response []int32) {
	//操作
	code := int32(0)
	for _, v := range response {
		if v == global.WIK_HU {
			code = global.WIK_HU
			break
		}
		if v == global.WIK_AN_GANG {
			code = global.WIK_AN_GANG
		} else if v == global.WIK_MING_GANG {
			code = global.WIK_MING_GANG
		} else if v == global.WIK_BU_GANG {
			code = global.WIK_BU_GANG
		} else if v == global.WIK_PENG {
			code = global.WIK_PENG
		}
	}

	if code == 0 {
		return
	} else if code == global.WIK_PENG && r.nearestMjData == global.MagicMahjong {
		r.Agent.WriteMsg(&msg.Game_C_UserOperate{
			OperateCode: global.WIK_NULL,
		})
		return
	}
	r.Agent.WriteMsg(&msg.Game_C_UserOperate{
		OperateCode: code,
	})
}

// 机器人出牌轮数
func (r *Robot) ADDRounds() {
	r.rounds++
}

// 添加桌面碰牌

func (r *Robot) DiskMj(OperateCode, OperateMj int32) {
	r.userDiskMj.Data = append(r.userDiskMj.Data, msg.DiskMahjong{Data: OperateMj, Code: OperateCode, ChairID: r.userChairID})
}
