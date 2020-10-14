package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	golog "log"
	"strconv"
	"strings"
	"time"
	"xj_game_server/game/public/common"
	"xj_game_server/game/public/conf"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/robot"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/public"
	"xj_game_server/public/mysql"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/util"
	rand "xj_game_server/util/leaf/util"
)

var GameClient *GameMysql

type GameMysql struct {
	*mysql.Mysql
}

func init() {
	mysql.Client.OnInit()
	GameClient = &GameMysql{Mysql: mysql.Client}
	//重启解锁所有用户
	GameClient.UnLockALL()
}

func (self *GameMysql) OnDestroy() {
	self.Mysql.OnDestroy()
}

//初始化房间配置
func (self *GameMysql) InitGameConfig() {
	var gameInfo = new(store.GameInfo)
	err := self.GetXJGameDB.Where("kind_id=? AND game_id=?", conf.GetService().KindID, conf.GetService().GameID).Find(gameInfo).Error
	if err != nil {
		_ = log.Logger.Errorf("InitGameConfig err: %v", "游戏初始化失败")
		golog.Fatalf("InitGameConfig err: %v", err)
		return
	}

	store.GameControl.OnInit(gameInfo)
}

//初始化机器人配置
func (self *GameMysql) InitRobotConfig() bool {
	robotConfig := make([]robot.Config, 0)
	err := self.GetXJGameDB.Where("game_id = ? AND ? BETWEEN enter_time AND leave_time", conf.GetService().GameID, time.Now().Format(public.FormatTime)).Find(&robotConfig).Error
	if err == gorm.ErrRecordNotFound || len(robotConfig) == 0 {
		_ = log.Logger.Errorf("InitRobotConfig 机器人初始化 RecordNotFound len: %v", len(robotConfig))
		return false
	}
	if err != nil {
		_ = log.Logger.Errorf("InitRobotConfig机器人初始化失败 err: %v", err)
		return false
	}
	robot.RobotConfigItem.OnInit(robotConfig)
	return true
}

//加载机器人
func (self *GameMysql) LoadRobotUser(config *robot.Config) []int32 {

	tx := self.GetXJGameDB.Begin()
	if config.BatchID > 0 {
		err := InitRobotLockInfo(tx, config.BatchID)
		if err != nil {
			tx.Rollback()
			_ = log.Logger.Errorf("UpdateRobotLockInfo err%v", err)
			golog.Fatal(err.Error())
		}
	}
	robotInfo, err := GetRobotLockInfoByLimit(tx, config.RobotCount)
	if err != nil {
		tx.Rollback()
		_ = log.Logger.Errorf("GetRobotLockInfoByLimit err%v", err)
		golog.Fatal(err.Error())
	}
	// commit

	var userIDList []int32
	for i := range robotInfo {
		userIDList = append(userIDList, robotInfo[i].UserID)
	}

	err = UpdateRobotLockInfoByUserID(tx, userIDList, map[string]interface{}{
		"robot_status":   "1",
		"game_id":        config.GameID,
		"batch_id":       config.BatchID,
		"lock_date_time": time.Now(),
	})
	if err != nil {
		tx.Rollback()
		_ = log.Logger.Errorf("UpdateRobotLockInfo err%v", err)
		golog.Fatal(err.Error())
	}
	// 从用户表中添加机器人信息数据
	haveCount := config.RobotCount - int64(len(userIDList))
	if haveCount > 0 {
		loadUserList, err := GetAccountInfoJoinRobotLockInfo(tx, haveCount)
		if len(loadUserList) >= 1 {
			for _, v := range loadUserList {
				err = CreateRobotLockInfo(tx, RobotLockInfo{
					UserID:       v,
					RobotStatus:  1,
					GameID:       config.GameID,
					BatchID:      config.BatchID,
					LockDateTime: time.Now(),
				})
			}
			userIDList = append(userIDList, loadUserList...)
		} else {
			loadUserList, err = GetAccountInfo2RobotLockInfo(tx, haveCount)
			if len(loadUserList) > 0 {
				err = UpdateRobotLockInfoByUserID(tx, loadUserList, map[string]interface{}{
					"robot_status":   "1",
					"game_id":        config.GameID,
					"batch_id":       config.BatchID,
					"lock_date_time": time.Now(),
				})
				userIDList = append(userIDList, loadUserList...)
			}
		}
		if err != nil {
			tx.Rollback()
			_ = log.Logger.Errorf("CreateRobotLockInfo err%v", err)
			golog.Fatal(err.Error())
		}
	}
	tx.Commit()
	return userIDList
}

//用户登陆
func (self *GameMysql) UserLogin(agent gate.Agent, userID int32, machineID string) (int32, string) {

	tx := GameClient.GetXJGameDB.Begin()
	var SystemStatus = "enjoin_logon"
	var err error
	var bool = false
	var resultString string
	userInfo, err := GetAccountsInfoByUserID(tx, userID)
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	if userInfo.UserID == 0 {
		tx.Rollback()
		resultString = "您的帐号不存在或者密码输入有误，请查证后再次尝试登录！"
		return common.StatusNotImplementedFail, resultString
	}

	userImage, err := GetAccountsImageByUserID(tx, userID)
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	scoreInfo, err := GetGameScoreInfoByUserId(tx, userID)
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}

	// 同步redis 缓存金币
	if !redis.GameClient.IsExistsDiamond(userInfo.UserID) && userInfo.UserType != 1 {
		redis.GameClient.SetDiamond(userInfo.UserID, scoreInfo.Diamond)
	}
	userDiamond, _ := redis.GameClient.GetDiamond(userInfo.UserID)

	if userInfo.UserType != 1 {
		if userInfo.Nullity != 0 {
			tx.Rollback()
			resultString = "您的帐号暂时处于冻结状态，请联系客户服务中心了解详细情况！"
			return common.StatusNotImplementedFail, resultString

		}
		//if userInfo.LastLogonMachine != machineID {
		//	tx.Rollback()
		//	resultString = "您的帐号使用固定机器登录功能，您现所使用的机器不是所指定的机器！"
		//	return common.StatusNotImplementedFail, resultString
		//}
		resultString, err = CheckSystemStatusInfoByName(tx, SystemStatus)
		if err != nil {
			tx.Rollback()
			return common.StatusInternalServerFail, err.Error()
		}
		if resultString != "" {
			tx.Rollback()
			return common.StatusNotImplementedFail, resultString

		}
		bool, err = CheckConfineMachineLimit(tx, strings.Split(agent.RemoteAddr().String(), ":")[0])

		if err != nil {
			tx.Rollback()
			return common.StatusInternalServerFail, err.Error()
		}
		if !bool {
			tx.Rollback()
			resultString = "抱歉地通知您，系统禁止了您所在的 IP 地址的登录功能，请联系客户服务中心了解详细情况！"
			return common.StatusNotImplementedFail, resultString

		}
		bool, err = CheckConfineMachineLimit(tx, machineID)

		if err != nil {
			tx.Rollback()
			return common.StatusInternalServerFail, err.Error()
		}
		if !bool {
			tx.Rollback()
			resultString = "抱歉地通知您，系统禁止了您的机器的登录功能，请联系客户服务中心了解详细情况！"
			return common.StatusNotImplementedFail, resultString

		}
		if store.GameControl.GetGameInfo().DeductionsType == 0 {
			if scoreInfo.GoldCoin < store.GameControl.GetGameInfo().MinEnterScore {
				tx.Rollback()
				resultString = "抱歉，您的金币数目不能低于最低限制额度！"
				return common.StatusMinEnterScore, resultString
			}

		} else {
			if float32(userDiamond) < store.GameControl.GetGameInfo().MinEnterScore {
				tx.Rollback()
				resultString = "抱歉，您的余额数目不能低于最低限制额度"
				return common.StatusMinEnterScore, resultString
			}
		}
		err = UpdateGameScoreInfoByUserID(tx, userID, nil)
		if err != nil {
			tx.Rollback()
			return common.StatusInternalServerFail, err.Error()
		}
		//NowTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

		err = CreateOrUpdateSystemStreamRoomInfo(tx, SystemStreamRoomInfo{
			DateID:      time.Now().Day(),
			KindID:      store.GameControl.GetGameInfo().KindID,
			GameID:      store.GameControl.GetGameInfo().GameID,
			LogonCount:  1,
			CollectDate: time.Now(),
		})
		if err != nil {
			tx.Rollback()
			return common.StatusInternalServerFail, err.Error()
		}
	}
	tx.Commit()
	// 机器人更新最后登录时间
	if userInfo.UserType == 1 {
		err = UpdateAccountsLastLogonDateByUserID(
			self.GetXJGameDB, userInfo.UserID,
		)
		if err != nil {
			return common.StatusInternalServerFail, err.Error()
		}
	}

	userItem := new(user.Item)
	userItem.OnInit()
	userItem.UserRight = userInfo.UserRight

	var roleId = int32(2001)

	if userInfo.UserType == 1 {
		randNumber := rand.RandInterval(0, 2)
		roleId = 2001 + randNumber
	} else {
		roleId = userImage.RoleID
	}

	userItem.Info = user.Info{
		UserID:       userInfo.UserID,
		NikeName:     userInfo.NickName,
		UserGold:     scoreInfo.GoldCoin,
		UserDiamond:  float32(userDiamond),
		Jackpot:      scoreInfo.Jackpot,
		MemberOrder:  userInfo.LevelNum,
		HeadImageUrl: userInfo.HeadImageUrl,
		FaceID:       userInfo.FaceID,
		RoleID:       roleId,
		SuitID:       userImage.SuitID,
		PhotoFrameID: userImage.PhotoFrameID,
		Gender:       userInfo.Gender,
	}
	user.List.Store(userID, userItem)
	return common.StatusOK, ""
}

//用户写分
func (self *GameMysql) WriteUserScore(uid int32, decUserScore float32, tintScoreType int32, decRevenue float32, intWinCount, intLostCount, intDrawCount, intFleeCount, intPlayTimeCount, tintTaskForward, intKindID, intGameID int32, decWaterScore float32, strClientIP string, timeEnterTime, timeLeaveTime string, tableID string, jackpot, newUserScore float32, recordID string) (int64, string) {
	//GSP_WriteGameScore
	if recordID == "" {
		randNum := util.Krand(6, 3)
		recordID = fmt.Sprintf("%v%v%s", conf.GetService().GameID, time.Now().Unix(), randNum)
	}
	if intPlayTimeCount >= 86400 {
		return common.StatusNotExtendedFail, "数据异常,游戏时间过长"
	}
	if newUserScore < 0 {
		newUserScore = 0
	}
	tx := self.GetXJGameDB.Begin()
	userInfo, err := GetAccountsInfoByUserID(tx, uid)
	if err != nil || userInfo.UserID == 0 {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error() + strconv.Itoa(int(userInfo.UserID))
	}
	if userInfo.UserType == 1 {
		return common.StatusOK, ""
	}

	scoreInfo, err := GetGameScoreInfoByUserId(tx, uid)
	if err != nil || scoreInfo.UserID == 0 {
		tx.Rollback()
		return common.StatusInternalServerFail, "用户信息不存在"
	}
	var decPreScore float32
	if tintScoreType == 0 {
		err = UpdateGameScoreInfoUp(tx, uid, GameScoreInfo{
			GoldCoin:  newUserScore,
			Jackpot:   jackpot,
			Revenue:   decRevenue,
			WinCount:  intWinCount,
			LostCount: intLostCount,
			DrawCount: intDrawCount,
			FleeCount: intFleeCount,
		})
		if err == nil {
			err = CreateGameCoinChangeLog(tx, GameCoinChangeLog{
				UserID:        uid,
				CapitalTypeID: 3,
				LogDate:       time.Now(),
				CapitalAmount: decUserScore,
				LastAmount:    decUserScore + scoreInfo.GoldCoin,
				ClientIP:      strClientIP,
				Remark:        "游戏比分输赢值",
			})
		}
		decPreScore = scoreInfo.GoldCoin
	} else {
		err = UpdateGameScoreInfoUp(tx, uid, GameScoreInfo{
			Diamond:   newUserScore,
			Jackpot:   jackpot,
			Revenue:   decRevenue,
			WinCount:  intWinCount,
			LostCount: intLostCount,
			DrawCount: intDrawCount,
			FleeCount: intFleeCount,
		})
		if err == nil {
			err = CreateGameDiamondChangeLog(tx, GameDiamondChangeLog{
				UserID:        uid,
				CapitalTypeID: 3,
				LogDate:       time.Now(),
				CapitalAmount: decUserScore,
				LastAmount:    decUserScore + scoreInfo.Diamond,
				ClientIP:      strClientIP,
				Remark:        "游戏比分输赢值",
			})
		}
		decPreScore = scoreInfo.Diamond
	}

	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	err = self.UpdateGameRoomConfigStartStores(intGameID, decUserScore)
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	timeEnter, _ := time.ParseInLocation("2006-01-02 15:04:05", timeEnterTime, time.Local)
	timeLeave, _ := time.ParseInLocation("2006-01-02 15:04:05", timeLeaveTime, time.Local)
	lockRecord, err := GetAccountPlayingLock(self.GetXJGameDB, userInfo.UserID)
	if err == nil {
		timeEnter = lockRecord.CollectDate
		UpdateRecordDiamondDrawInfo(self.GetXJGameDB, RecordDiamondDrawInfo{
			DrawID:  tableID,
			EndTime: timeEnter,
		})
	}
	err = CreateGameRecordInfo(tx, GameRecordInfo{
		RecordID:     recordID,
		UserID:       uid,
		UserScore:    decUserScore,
		UserRevenue:  decRevenue,
		PreScore:     decPreScore,
		ScoreType:    tintScoreType,
		KindID:       intKindID,
		KindName:     store.GameControl.GetGameInfo().KindName,
		GameID:       intGameID,
		GameName:     store.GameControl.GetGameInfo().GameName,
		UserType:     userInfo.UserType,
		RevenueRatio: store.GameControl.GetGameInfo().RevenueRatio,
		WaterScore:   decWaterScore,
		DateID:       time.Now().Day(),
		RecordDate:   time.Now(),
		EnterTime:    timeEnter,
		LeaveTime:    timeLeave,
		DrawID:       tableID,
	})
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	timeNowString := time.Now().Format("20060102")
	nowTime, _ := strconv.Atoi(timeNowString)
	streamInfo := &StreamScoreInfo{
		DateID:           nowTime,
		UserID:           uid,
		KindID:           intKindID,
		KindName:         store.GameControl.GetGameInfo().KindName,
		GameID:           intGameID,
		GameName:         store.GameControl.GetGameInfo().GameName,
		DataType:         tintScoreType,
		TotalScore:       decWaterScore,
		TotalRevenue:     decRevenue,
		WinCount:         intWinCount,
		LostCount:        intLostCount,
		PlayTimeCount:    intPlayTimeCount,
		OnlineTimeCount:  intPlayTimeCount,
		FirstCollectDate: time.Now(),
		LastCollectDate:  time.Now(),
		WaterScore:       decWaterScore,
		GameTotalCount:   1,
	}
	//err = CreateOrUpdateStreamScoreInfo(tx, streamInfo)
	if CheckStreamScoreInfoByDay(tx, streamInfo) {
		err = CreateStreamScoreInfo(tx, streamInfo)
	} else {
		err = UpdateStreamScoreInfoByID(tx, streamInfo)
	}
	if err != nil {
		tx.Rollback()
		return common.StatusInternalServerFail, err.Error()
	}
	tx.Commit()

	//TODO GSP_WriteGameScore 统计型数据库表未写入

	return common.StatusOK, ""
}

func (self *GameMysql) UpdateGameRoomConfigStartStores(gameID int32, decStartStores float32) error {
	sqlPre := "UPDATE game_room_config SET "
	sqlPre += fmt.Sprintf("start_stores=start_stores+%v"+
		" WHERE game_id=%v", decStartStores, gameID)
	return self.GetXJGameDB.Exec(sqlPre).Error
}

/*
//用户写分
func (self *GameMysql) WriteUserScoreBak(uid int32, decUserScore float32, tintScoreType int32, decRevenue float32, intWinCount, intLostCount, intDrawCount, intFleeCount, intPlayTimeCount, tintTaskForward, intKindID, intGameID int32, decWaterScore float32, strClientIP string, timeEnterTime, timeLeaveTime string, tableID int32) (int64, string) {
	//GSP_WriteGameScore
	//rows, err := self.AccountDB.DB().DB.Query("CALL GSP_WriteGameScore(?, ?, ?, ?, ?,?, ?, ?, ?, ?,?, ?, ?, ?, ?,?)",
	//	userID, conf.GetServer().KindID,
	//	conf.GetServer().GameID, machineID,
	//	strings.Split(agent.RemoteAddr().String(), ":")[0])

	rows, err := self.Query(self.TreasureDB, "CALL GSP_WriteGameScore(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		uid,
		decUserScore,
		tintScoreType,
		decRevenue,
		intWinCount,
		intLostCount,
		intDrawCount,
		intFleeCount,
		intPlayTimeCount,
		tintTaskForward,
		intKindID,
		intGameID,
		decWaterScore,
		strClientIP,
		timeEnterTime,
		timeLeaveTime,
		tableID,
	)

	defer func() {
		if rows != nil {
			err := rows.Close()
			if err != nil {
				_ = log.Logger.Errorf("GSP_WriteGameScore rows 关闭错误 出错 err: %s ", err.Error())
			}
		}
	}()

	if err != nil {
		return 1, err.Error()
	}

	var errorCode int64
	var errorMsg string

	rows.Next()

	err = rows.Scan(&errorCode, &errorMsg)
	if err != nil {
		_ = log.Logger.Errorf("GSP_WriteGameScore 接口  查询存储过程 err: %s ", err.Error())
		return 1, err.Error()
	}

	return errorCode, errorMsg
}
*/
//游戏记录
func (self *GameMysql) WriteGameRecord(intTableID, intUserCount, intAndroidCount int32, decWasteCount, decResveueCount float32, timeStartTime, timeEndTime string, tintScoreType int32, detail string) (int64, string, string) {
	//GSP_RecordDrawInfo
	var errorCode = int64(common.StatusOK)
	var errorMsg string
	var err error
	var DrawID string
	randNum := util.Krand(6, 3)

	DrawID = fmt.Sprintf("%v%v%s", conf.GetService().GameID, time.Now().UnixNano(), randNum)
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStartTime, time.Local)
	endTime, _ := time.ParseInLocation("2006-01-02 15:04:05", timeEndTime, time.Local)
	if tintScoreType == 0 {
		err = CreateRecordCoinDrawInfo(self.GetXJGameDB, RecordCoinDrawInfo{
			DrawID:    DrawID,
			KindID:    conf.GetService().KindID,
			GameID:    conf.GetService().GameID,
			TableID:   intTableID,
			UserCount: intUserCount,

			AndroidCount: intAndroidCount,
			Waste:        decWasteCount,
			Revenue:      decResveueCount,
			StartTime:    startTime,
			EndTime:      endTime,
			InsertTime:   time.Now(),
			DateID:       time.Now().Day(),
		})
	} else {
		err = CreateRecordDiamondDrawInfo(self.GetXJGameDB, RecordDiamondDrawInfo{
			DrawID:       DrawID,
			KindID:       conf.GetService().KindID,
			GameID:       conf.GetService().GameID,
			TableID:      intTableID,
			UserCount:    intUserCount,
			AndroidCount: intAndroidCount,
			Waste:        decWasteCount,
			Revenue:      decResveueCount,
			StartTime:    startTime,
			EndTime:      endTime,
			Detail:       detail,
			InsertTime:   time.Now(),
			DateID:       time.Now().Day(),
		})
	}
	if err != nil {
		errorMsg = err.Error()
		errorCode = common.StatusInternalServerFail
	}
	return errorCode, errorMsg, DrawID
}

//用户是否被锁定 不能同时进入2个游戏
func (self *GameMysql) IsLock(userId int32) bool {
	//	var lock = new(AccountPlayingLock)
	var count int
	self.GetXJGameDB.Model(AccountPlayingLock{}).Where("user_id = ? AND (kind_id <> ? OR game_id <> ?)", userId, conf.GetService().KindID, conf.GetService().GameID).Count(&count)
	return count != 0
}

//锁定用户
func (self *GameMysql) Lock(userId int32, ip string) error {
	err := DelAccountPlayingLock(self.GetXJGameDB, userId)
	if err != nil {
		return err
	}
	return CreateAccountPlayingLock(self.GetXJGameDB, AccountPlayingLock{
		UserID:      userId,
		KindID:      conf.GetService().KindID,
		GameID:      conf.GetService().GameID,
		EnterIP:     ip,
		CollectDate: time.Now(),
	})
}

//解锁用户
func (self *GameMysql) UnLock(userId int32) error {
	var unLock AccountPlayingLock
	err := self.GetXJGameDB.Where("user_id = ? ", userId).Delete(&unLock).Error
	return err
}

//进程重启解锁当前游戏所有用户
func (self *GameMysql) UnLockALL() {
	var unLock AccountPlayingLock
	self.GetXJGameDB.Where("kind_id = ?  AND game_id = ? ", conf.GetService().KindID, conf.GetService().GameID).Delete(&unLock)

	InitRobotLockInfoByGameID(self.GetXJGameDB, conf.GetService().GameID)
}

// 热更数据
func (self *GameMysql) WriteAgentData(userId int32, dateType int, decAmount, decPercentValue float32) (errorCode int32, errorMsg string) {
	//rows, err := self.Query(self.GetXJGameDB, "CALL WSP_UpdateAgentData(?, ?, ?, ?)", userId, dateType, decAmount, decPercentValue)
	//defer func() {
	//	if rows != nil {
	//		//err = rows.Close()
	//		if err != nil {
	//			_ = log.Logger.Errorf("WSP_UpdateAgentData rows 更新 出错 err: %s ", err.Error())
	//		}
	//	}
	//}()
	//if err != nil {
	//	return 1, "服务器繁忙"
	//}
	return
}
