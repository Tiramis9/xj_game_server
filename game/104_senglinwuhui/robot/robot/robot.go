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
	"xj_game_server/game/104_senglinwuhui/conf"
	"xj_game_server/game/104_senglinwuhui/global"
	"xj_game_server/game/104_senglinwuhui/msg"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/util"
	"sync"
	"time"
)

// RobotList 机器人列表 uid --> Robot
//var List = make(map[int32]*Robot)
var List sync.Map

// Robot 机器人
type Robot struct {
	*robot.Agent            // 机器人连接代理
	batchID         int32   // 批次id
	userID          int32   // 用户id
	withProbability float32 //推出概率
	gold            float32 //用户金币
	diamond         float32 //用户余额
	gameStatus      bool    //游戏状态
	sitDownStatus   bool    //坐下状态
	betStatus       bool    //下注状态
}

// OnInit 初始化
func (r *Robot) OnInit(userID int32, batchID int32, gate *gate.Gate, userCallBack func(args []interface{})) {
	r.Agent = new(robot.Agent)
	r.Agent.OnInit(gate, userCallBack)
	r.batchID = batchID
	r.userID = userID
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

//GetGold 修改金币
func (r *Robot) SetGold(gold float32) {
	r.gold = gold
}

//GetDiamond 修改余额
func (r *Robot) SetDiamond(diamond float32) {
	r.diamond = diamond
}

func (r *Robot) GetWithProbability() float32 {
	return r.withProbability
}

func (r *Robot) SetWithProbability(w float32) {
	r.withProbability = w
}

func (r *Robot) AddWithProbability(w float32) {
	r.withProbability += w
}

func (r *Robot) GetGameStatus() bool {
	return r.gameStatus
}

func (r *Robot) SetGameStatus(w bool) {
	r.gameStatus = w
}

func (r *Robot) GetSitDownStatus() bool {
	return r.sitDownStatus
}

func (r *Robot) SetSitDownStatus(w bool) {
	r.sitDownStatus = w
}

func (r *Robot) GetBetStatus() bool {
	return r.betStatus
}

func (r *Robot) SetBetStatus(w bool) {
	r.betStatus = w
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
	batchId := r.GetBatchID()
	maxCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxCoin
	minCoin := robot.RobotConfigItem.GetConfig()[batchId].TakeMinCoin

	maxDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMaxDiamond
	minDiamond := robot.RobotConfigItem.GetConfig()[batchId].TakeMinDiamond

	coin := float32(util.RandInterval(int32(minCoin), int32(maxCoin)-1))
	diamond := float32(util.RandInterval(int32(minDiamond), int32(maxDiamond)-1))

	value, ok := user.List.Load(r.userID)
	if ok {
		value.(*user.Item).UserGold = coin
		value.(*user.Item).UserDiamond = diamond
	}
	//user.List[uid].UserGold = coin
	//user.List[uid].UserDiamond = diamond

	r.gold = coin
	r.diamond = diamond

}

//SitDown 坐下
func (r *Robot) SitDown() {
	r.WriteMsg(&msg.Game_C_UserSitDown{
		//随机一个桌子号
		TableID: util.RandInterval(0, store.GameControl.GetGameInfo().TableCount-1),
		ChairID: -1,
	})
}

//StandUp 起立
func (r *Robot) StandUp() {
	//发送命令后先修改游戏状态，禁止下注
	r.gameStatus = false

	r.WriteMsg(&msg.Game_C_UserStandUp{})
}

//Jetton 下注
func (r *Robot) jetton(area int32, score float32) {

	if !r.gameStatus || !r.sitDownStatus {
		return
	}
	r.betStatus = true

	r.WriteMsg(&msg.Game_C_UserJetton{
		JettonArea:  area,
		JettonScore: score,
	})
}

//CheckBatchTimeOut 检查批次是否过期
func (r *Robot) CheckBatchTimeOut() bool {
	// 批次号是否过期，过期必退
	_, ok := robot.RobotConfigItem.GetConfig()[r.batchID]
	return ok
}

//随机退出
func (r *Robot) RandStandUp() {
	time.Sleep(time.Duration(util.RandInterval(500, 3000)) * time.Millisecond)
	//起立不判断游戏场景  未下注 和 已经坐下
	if r.betStatus || !r.sitDownStatus {
		return
	}
	if float32(util.RandInterval(0, 100)) < r.GetWithProbability() {
		//_ = log.Logger.Errorf("%c[1;40;31m 龙虎斗 结束=====%c[0m  游戏状态：%v 下注状态：%v 坐下状态%v \n", 0x1B, 0x1B,r.GetGameStatus(),r.GetBetStatus(),r.GetSitDownStatus())
		r.StandUp()

		r.SetSitDownStatus(false)
		r.SetWithProbability(0)

		time.Sleep(time.Duration(util.RandInterval(1, 10)) * time.Second)

		//过期不在重进
		if !r.CheckBatchTimeOut() {
			r.Agent.Close()
			return
		}

		if r.sitDownStatus {
			return
		}

		//fmt.Printf("%c[1;47;31m 重新坐下uid===== %v %c[0m\n", 0x1B, uid, 0x1B)
		r.Assignment()
		r.SitDown()
	}
}

//随机下注
func (r *Robot) RobotLottery() {

	go func() {
		//下注定时器
		time.Sleep(time.Second * time.Duration(conf.GetServer().GameJettonTime-1))
		r.SetGameStatus(false)
	}()
	var score float32

	if store.GameControl.GetGameInfo().DeductionsType == 0 {
		score = r.GetGold()
	} else {
		score = r.GetDiamond()
	}

	if score < conf.GetServer().JettonList[len(conf.GetServer().JettonList)-1] {

		r.AddWithProbability(2)
		//没钱不下注，只看不买
		return
	}
	if !r.gameStatus || !r.sitDownStatus {
		return
	}
	time.Sleep(time.Second * 2)
	//for i := util.Int() % 2; i == 0; i = util.Int() % 5 {
	for i := int32(0); i < conf.GetServer().GameJettonTime; i++ {
		if !r.gameStatus || !r.sitDownStatus {
			break
		}

		time.Sleep(time.Millisecond * time.Duration(util.RandInterval(250, 1000)))

		betScore := conf.GetServer().JettonList[util.RandInterval(0, int32(len(conf.GetServer().JettonList)-1))] //最小下注

		var score float32

		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			score = r.GetGold()
		} else {
			score = r.GetDiamond()
		}

		if betScore > score {
			r.AddWithProbability(2)
			break
		}
		area, _ := getAreaMultiple()

		r.jetton(area, betScore)

		//_ = log.Logger.Errorf("%c[1;40;31m 龙虎斗 结束=====%c[0m  游戏状态：%v 下注状态：%v 坐下状态%v \n", 0x1B, 0x1B,r.GetGameStatus(),r.GetBetStatus(),r.GetSitDownStatus())
	}

	r.AddWithProbability(0.2)
}


// 获取赔率
func getAreaMultiple() (int32, float32) {

	area := util.RandInterval(0, int32(len(global.AreaMultiple))-1)
	multiple := global.AreaMultiple[area]

	return area, multiple

}