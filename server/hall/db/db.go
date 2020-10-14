package db

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
	"xj_game_server/game/public/redis"
	"xj_game_server/public/mysql"
	"xj_game_server/server/hall/global"
	"xj_game_server/server/hall/msg"
)

var HallMysqlClient *HallMysql

//var LoginMysqlClient *LoginMysql

type HallMysql struct {
	*mysql.Mysql
	//*mysql.NativeMysql
}

func init() {
	mysql.Client.OnInit()
	HallMysqlClient = &HallMysql{
		Mysql: mysql.Client,
	}

}

func (self *HallMysql) OnDestroy() {
	//关闭用户数据库连接
	//self.Mysql.OnDestroy()
}

type NewsInfo struct {
	NewsID        int32
	Subject       string
	Body          string
	FormattedBody string
	ClassID       int32
}

type AccountHall_S_Msg struct {
	UserID         int     `json:"user_id" gorm:"user_id"`
	NickName       string  `json:"nick_name" gorm:"nick_name"`
	Gender         int     `json:"gender" gorm:"gender"`
	LevelNum       int     `json:"level_num" gorm:"level_num"`
	RegisterMobile string  `json:"register_mobile" gorm:"register_mobile"`
	GoldCoin       float64 `json:"gold_coin" gorm:"gold_coin"`
	Diamond        float64 `json:"diamond" gorm:"diamond"`
	FaceID         int     `json:"face_id" gorm:"face_id"`
	RoleID         int     `json:"role_id" gorm:"role_id"`
	SuitID         int     `json:"suit_id" gorm:"suit_id"`
	PhotoFrameID   int     `json:"photo_frame_id" gorm:"photo_frame_id"`
	BinderCardNo   string  `json:"binder_card_no" gorm:"binder_card_no"`
}

type RecordLoginResp struct {
	Terminal int `json:"terminal" gorm:"terminal"`
}

//
//type GameVersion struct {
//	Version   int32    `json:"version" gorm:"version"`
//	PokerName string `json:"poker_name" gorm:"poker_name"`
//}

type GameVersion struct {
	KindID    int32  `json:"-" gorm:"kind_id"`
	Version   int32  `json:"version" gorm:"version"`
	PokerName string `json:"poker_name"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	Platform  int    `json:"platform"`
}

type GameScoreInfo struct {
	GoldCoin float64 `json:"gold_coin" gorm:"gold_coin"`
	Diamond  float64 `json:"diamond" gorm:"diamond"`
}

func (self *HallMysql) GetUerByUid(uid, PlatformID int32) (*msg.HeartLoginInit, error) {

	var heart = &msg.HeartLoginInit{
		UserInfo: new(msg.UserInfo),
		GameInfo: make([]*msg.GameInfo, 0),
		Notify:   make([]*msg.Notify, 0),
	}
	byUid, err := self.GetAmountByUid(uid)

	if err != nil {
		return nil, err
	}
	heart.UserInfo = byUid.UserInfo
	heart.GameInfo = HallRedisClient.LoadGameList()

	//GameInfo []*GameInfo `json:"game_info"`
	//UserInfo UserInfo    `json:"user_info"`
	//Notify   []*Notify   `json:"notify"`

	GameVersion, err := self.GetAllGameVersionByPlatform(PlatformID)

	if len(GameVersion) > 0 {
		for k, v := range heart.GameInfo {
			for i := range GameVersion {
				if v.KindId == global.DecGameName2KindID(GameVersion[i].PokerName) {
					heart.GameInfo[k].GameVersion = GameVersion[i].Version
					break
				}
			}
		}
	}
	heart.SendTime = time.Now().Format("2006-01-02 15:04:05")
	return heart, nil
}

//func (self *HallMysql) GetUerByUid(uid int32) (*msg.Hall_S_Msg, error) {
//	//SELECT info.*,game.GoldCoin,game.Diamond FROM xjaccountsdb.AccountsInfo info INNER JOIN xjtreasuredb.GameScoreInfo game ON info.UserID = game.UserID WHERE info.UserID=18;
//	var announcement []NewsInfo
//	var accounthall_s_msg []AccountHall_S_Msg
//	//self.PlatformDB.Find(&announcement)
//	err := self.GetXJGameDB.Table("news_info").Find(&announcement).Error
//	//res, err := self.AccountDB.QueryString("SELECT info.user_id,info.nick_name,info.gender,info.level_num,info.register_mobile,game.gold_coin,game.diamond,info.face_id,image.role_id,image.suit_id,image.photo_frame_id,IFNULL(exchange.account_or_card,'') AS binder_card_no FROM accounts_info info INNER JOIN game_score_info game ON info.user_id = game.user_id INNER JOIN accounts_image image ON info.user_id = image.user_id LEFT JOIN exchange_account exchange ON info.user_id = exchange.user_id WHERE info.user_id=" + strconv.Itoa(int(uid)))
//	err = self.GetXJGameDB.Raw("SELECT info.user_id,info.nick_name,info.gender,info.level_num,info.register_mobile,game.gold_coin,game.diamond,info.face_id,image.role_id,image.suit_id,image.photo_frame_id,IFNULL(exchange.account_or_card,'') AS binder_card_no FROM accounts_info info INNER JOIN game_score_info game ON info.user_id = game.user_id INNER JOIN accounts_image image ON info.user_id = image.user_id LEFT JOIN exchange_account exchange ON info.user_id = exchange.user_id WHERE info.user_id=" + strconv.Itoa(int(uid))).Scan(&accounthall_s_msg).Error
//	//res, err := self.AccountDB.QueryString("SELECT info.UserID,info.NickName,info.Gender,info.LevelNum,info.RegisterMobile,game.GoldCoin,game.Diamond,info.FaceID,image.RoleID,image.SuitID,image.PhotoFrameID,IFNULL(exchange.AccountOrCard,'') AS BinderCardNo FROM xjaccountsdb.AccountsInfo info INNER JOIN xjtreasuredb.GameScoreInfo game ON info.UserID = game.UserID INNER JOIN xjaccountsdb.AccountsImage image ON info.UserID = image.UserID LEFT JOIN xjaccountsdb.ExchangeAccount exchange ON info.UserID = exchange.UserID WHERE info.UserID=" + strconv.Itoa(int(uid)))
//	if err != nil {
//		return nil, err
//	}
//	if len(accounthall_s_msg) <= 0 {
//		return nil, errors.New("用户不存在")
//	}
//	uidInt := accounthall_s_msg[0].UserID
//	levelNumInt := accounthall_s_msg[0].LevelNum
//	faceIDInt := accounthall_s_msg[0].FaceID
//	roleIDInt := accounthall_s_msg[0].RoleID
//	suitIDInt := accounthall_s_msg[0].SuitID
//	photoFrameIDInt := accounthall_s_msg[0].PhotoFrameID
//	genderInt := accounthall_s_msg[0].Gender
//	userGold := accounthall_s_msg[0].GoldCoin
//	userDiamonds := accounthall_s_msg[0].Diamond
//	//levelNumInt, _ := strconv.Atoi(res[0]["LevelNum"])
//	//faceIDInt, _ := strconv.Atoi(res[0]["FaceID"])
//	//roleIDInt, _ := strconv.Atoi(res[0]["RoleID"])
//	//suitIDInt, _ := strconv.Atoi(res[0]["SuitID"])
//	//photoFrameIDInt, _ := strconv.Atoi(res[0]["PhotoFrameID"])
//	//genderInt, _ := strconv.Atoi(res[0]["Gender"])
//	//userGold, _ := strconv.ParseFloat(res[0]["GoldCoin"], 64)
//	//userDiamonds, _ := strconv.ParseFloat(res[0]["Diamond"], 64)
//	var ans = make([]*msg.Announcement, 0)
//	for _, v := range announcement {
//		temp := &msg.Announcement{
//			NewsID:        v.NewsID,
//			Subject:       v.Subject,
//			Body:          v.Body,
//			FormattedBody: v.FormattedBody,
//			ClassID:       v.ClassID,
//		}
//		ans = append(ans, temp)
//	}
//	var recordloginresp []RecordLoginResp
//	//row, err := self.RecordDB.QueryString(fmt.Sprintf("select terminal from record_login WHERE user_id= %v order by terminal DESC limit 1", uid))
//	err = self.GetXJGameDB.Raw(fmt.Sprintf("select terminal from record_login WHERE user_id= %v order by terminal DESC limit 1", uid)).Scan(&recordloginresp).Error
//	//row, err := self.RecordDB.QueryString(fmt.Sprintf("select Terminal from RecordLogin WHERE UserID= %v order by Terminal DESC limit 1", uid))
//	if err != nil {
//		return nil, err
//	}
//
//	versionInfos := make([]*msg.VersionInfo, 0)
//	if len(recordloginresp) == 1 {
//		var deviceType string
//		switch recordloginresp[0].Terminal {
//		case 7:
//			deviceType = "WindowsEditor"
//		case 8:
//			deviceType = "IPhonePlayer"
//		case 11:
//			deviceType = "Android"
//		case 17:
//			deviceType = "WebGLPlayer"
//		}
//		if deviceType != "" {
//			var gameversion []GameVersion
//			err = self.GetXJGameDB.Raw(fmt.Sprintf("select kind_id,platform,version from game_version WHERE platform='%v' and poker_name NOT REGEXP '[.]'", deviceType)).Scan(&gameversion).Error
//			//row, err := self.PlatformDB.QueryString(fmt.Sprintf("select kind_id,platform,version from game_version WHERE platform='%v' and poker_name NOT REGEXP '[.]'", deviceType))
//			//row, err := self.PlatformDB.QueryString(fmt.Sprintf("select KindID,Platform,Version from GameVersion WHERE Platform='%v' and PokerName NOT REGEXP '[.]'", deviceType))
//			if err != nil {
//				return nil, err
//			}
//			for i := range gameversion {
//				Version := gameversion[i].Version
//				PokerName := gameversion[i].PokerName
//				temp := msg.VersionInfo{
//					Version:   int32(Version),
//					PokerName: PokerName,
//				}
//				versionInfos = append(versionInfos, &temp)
//			}
//		}
//	}
//
//	return &msg.Hall_S_Msg{
//		UserID:           int32(uidInt),
//		NikeName:         accounthall_s_msg[0].NickName,
//		MemberOrder:      int32(levelNumInt),
//		PhoneNumber:      accounthall_s_msg[0].RegisterMobile,
//		BinderCardNo:     accounthall_s_msg[0].BinderCardNo,
//		FaceID:           int32(faceIDInt),
//		RoleID:           int32(roleIDInt),
//		SuitID:           int32(suitIDInt),
//		PhotoFrameID:     int32(photoFrameIDInt),
//		Gender:           int32(genderInt),
//		TimeStamp:        time.Now().Unix(),
//		AnnouncementList: ans,
//		GameInfoList:     HallRedisClient.LoadGameList(),
//		VersionInfos:     versionInfos,
//		UserGold:         float32(userGold),
//		UserDiamonds:     float32(userDiamonds),
//	}, nil
//}

//func (self *HallMysql) GetAmountByUid(uid int32) (*msg.UserInfo, error) {
//	//SELECT info.*,game.GoldCoin,game.Diamond FROM xjaccountsdb.AccountsInfo info INNER JOIN xjtreasuredb.GameScoreInfo game ON info.UserID = game.UserID WHERE info.UserID=18;
//	var gamescoreinfo []GameScoreInfo
//	err := self.GetXJGameDB.Raw(fmt.Sprintf("SELECT gold_coin,diamond FROM `game_score_info` WHERE `user_id` = " + strconv.Itoa(int(uid)))).Scan(&gamescoreinfo).Error
//	//res, err := self.AccountDB.QueryString("SELECT gold_coin,diamond FROM `game_score_info` WHERE `UserID` = " + strconv.Itoa(int(uid)))
//	//res, err := self.AccountDB.QueryString("SELECT GoldCoin,Diamond FROM `xjtreasuredb`.`GameScoreInfo` WHERE `UserID` = " + strconv.Itoa(int(uid)))
//	if err != nil {
//		return nil, err
//	}
//	if len(gamescoreinfo) <= 0 {
//		return nil, errors.New("用户不存在")
//	}
//	userGold := gamescoreinfo[0].GoldCoin
//	userDiamonds := gamescoreinfo[0].Diamond
//	//userGold, _ := strconv.ParseFloat(gamescoreinfo[0].GoldCoin, 64)
//	//userDiamonds, _ := strconv.ParseFloat(res[0]["Diamond"], 64)
//	return &msg.UserInfo{
//		Gold:    userGold,
//		Diamond: userDiamonds,
//	}, nil
//}

func (self *HallMysql) GetAmountByUid(uid int32) (*msg.UserInfoChange, error) {
	//SELECT info.*,game.GoldCoin,game.Diamond FROM xjaccountsdb.AccountsInfo info INNER JOIN xjtreasuredb.GameScoreInfo game ON info.UserID = game.UserID WHERE info.UserID=18;
	var gamescoreinfo []GameScoreInfo
	err := self.GetXJGameDB.Raw(fmt.Sprintf("SELECT gold_coin,diamond FROM `game_score_info` WHERE `user_id` = " + strconv.Itoa(int(uid)))).Scan(&gamescoreinfo).Error
	//res, err := self.AccountDB.QueryString("SELECT gold_coin,diamond FROM `game_score_info` WHERE `UserID` = " + strconv.Itoa(int(uid)))
	//res, err := self.AccountDB.QueryString("SELECT GoldCoin,Diamond FROM `xjtreasuredb`.`GameScoreInfo` WHERE `UserID` = " + strconv.Itoa(int(uid)))
	if err != nil {
		return nil, err
	}
	if len(gamescoreinfo) <= 0 {
		return nil, errors.New("用户不存在")
	}
	//userGold := gamescoreinfo[0].GoldCoin
	userDiamonds := gamescoreinfo[0].Diamond
	//userGold, _ := strconv.ParseFloat(gamescoreinfo[0].GoldCoin, 64)
	//userDiamonds, _ := strconv.ParseFloat(res[0]["Diamond"], 64)

	if !redis.GameClient.IsExistsDiamond(uid) {
		redis.GameClient.SetDiamond(uid, float32(userDiamonds))
	}

	userDiamond, err := redis.GameClient.GetDiamond(uid)

	return &msg.UserInfoChange{
		UserInfo: &msg.UserInfo{
			Gold:    0,
			Diamond: userDiamond,
		},
	}, nil
	//return &msg.UserInfoChange{
	//	UserInfo: &msg.UserInfo{
	//		Gold:    userGold,
	//		Diamond: userDiamonds,
	//	},
	//}, nil
}

func (self *HallMysql) GetAllGameVersionByPlatform(platform int32) ([]GameVersion, error) {
	var gameVersions []GameVersion

	err := self.GetXJGameDB.Model(GameVersion{}).Where(" platform = ? ", platform).Find(&gameVersions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gameVersions, nil
		}
		return nil, err
	}
	return gameVersions, nil
}

type AccountsInfo struct {
	UserID           int32  `json:"user_id" gorm:"user_id"`
	FaceID           int32  `json:"face_id" gorm:"face_id"`
	Accounts         string `json:"accounts" gorm:"accounts"`
	NickName         string `json:"nick_name" gorm:"remarks"`
	UnderWrite       string `json:"under_write" gorm:"under_write"`
	IDCard           string `json:"id_card" gorm:"id_card"`
	RealName         string `json:"real_name" gorm:"real_name"`
	LogonPass        string `json:"logon_pass" gorm:"logon_pass"`
	InsurePass       string `json:"insure_pass" gorm:"insure_pass"`
	UserRight        int32  `json:"user_right" gorm:"user_right"`
	LevelNum         int32  `json:"level_num" gorm:"level_num"`
	Gender           int32  `json:"gender" gorm:"gender"`
	Nullity          int32  `json:"nullity" gorm:"nullity"`
	NullityOverDate  string `json:"nullity_over_date" gorm:"nullity_over_date"`
	NullityReasons   string `json:"nullity_reasons" gorm:"nullity_reasons"`
	MoorMachine      int32  `json:"moor_machine" gorm:"moor_machine"`
	UserType         int32  `json:"user_type" gorm:"user_type"`
	RegisterIP       string `json:"register_ip" gorm:"register_ip"`
	RegisterDate     string `json:"register_date" gorm:"register_date"`
	RegisterMobile   string `json:"register_mobile" gorm:"register_mobile"`
	RegisterMachine  string `json:"register_machine" gorm:"register_machine"`
	LastLogonMachine string `json:"last_logon_machine" gorm:"last_logon_machine"`
	UserUin          string `json:"user_uin" gorm:"user_uin"`
	RegisterOrigin   int32  `json:"register_origin" gorm:"register_origin"`
	HeadImageUrl     string `json:"head_image_url" gorm:"head_image_url"`
	GameLogonTimes   int32  `json:"game_logon_times" gorm:"game_logon_times"`
	LastLogonDate    string `json:"last_logon_date" gorm:"last_logon_date"`
	LastLogonIP      string `json:"last_logon_ip" gorm:"last_logon_ip"`
	CodeKey          string `json:"code_key" gorm:"code_key"`
	PlatformID       int32  `json:"platform_id" gorm:"platform_id"`
	Remarks          string `json:"remarks" gorm:"remarks"`
	SpreadChannelID  int32  `json:"spread_channel_id" gorm:"spread_channel_id"`
}

func (self *HallMysql) GetAccountsInfoByUserID(userID int32) (AccountsInfo, error) {
	var user AccountsInfo
	err := self.GetXJGameDB.Debug().Model(AccountsInfo{}).Where(" user_id = ? ", userID).Find(&user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return user, nil
	}

	return user, err
}

func (self *HallMysql) GetGameScoreInfoByUserId(userID int32) (GameScoreInfo, error) {

	var user GameScoreInfo
	err := self.GetXJGameDB.Model(GameScoreInfo{}).Where(" user_id = ? ", userID).Find(&user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return user, nil
	}
	return user, err

}

func (self *HallMysql) GetGameInfoChange(PlatformID int32) (*msg.GameInfoChange, error) {

	gameInfo := msg.GameInfoChange{}

	gameInfo.GameInfo = HallRedisClient.LoadGameList()

	GameVersion, err := self.GetAllGameVersionByPlatform(PlatformID)

	if len(GameVersion) > 0 {
		for k, v := range gameInfo.GameInfo {
			for i := range GameVersion {
				if v.KindId == global.DecGameName2KindID(GameVersion[i].PokerName) {
					gameInfo.GameInfo[k].GameVersion = GameVersion[i].Version
					break
				}
			}
		}
	}

	return &gameInfo, err

}
