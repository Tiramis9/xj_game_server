package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

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

func GetAccountsInfoByUserID(db *gorm.DB, userID int32) (AccountsInfo, error) {
	var user AccountsInfo
	err := db.Debug().Model(AccountsInfo{}).Where(" user_id = ? ", userID).Find(&user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return user, nil
	}

	return user, err
}

func UpdateAccountsLastLogonDateByUserID(db *gorm.DB, userID int32) error {
	return db.Exec("UPDATE accounts_info set last_logon_date= NOW() where user_id=?", userID).Error
}

type GameScoreInfo struct {
	UserID             int32   `json:"user_id" gorm:"user_id"`
	GoldCoin           float32 `json:"gold_coin" gorm:"gold_coin"`
	Diamond            float32 `json:"diamond" gorm:"diamond"`
	Revenue            float32 `json:"revenue" gorm:"revenue"`
	WinCount           int32   `json:"win_count" gorm:"win_count"`
	LostCount          int32   `json:"lost_count" gorm:"lost_count"`
	DrawCount          int32   `json:"draw_count" gorm:"draw_count"`
	FleeCount          int32   `json:"flee_count" gorm:"flee_count"`
	TotalDiamondStream float32 `json:"total_diamond_stream" gorm:"total_diamond_stream"`
	TotalCoinStream    float32 `json:"total_coin_stream" gorm:"total_coin_stream"`
	AllLogonTimes      int32   `json:"all_logon_times" gorm:"all_logon_times"`
	PlayTimeCount      int32   `json:"play_time_count" gorm:"play_time_count"`
	Jackpot            float32 `json:"jackpot" gorm:"jackpot"`
}

func GetGameScoreInfoByUserId(db *gorm.DB, userID int32) (GameScoreInfo, error) {

	var user GameScoreInfo
	err := db.Debug().Model(GameScoreInfo{}).Where(" user_id = ? ", userID).Find(&user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return user, nil
	}
	return user, err

}

// 更新用户进入游戏记录
func UpdateGameScoreInfoByUserID(db *gorm.DB, userID int32, data map[string]interface{}) error {
	if data == nil {
		return db.Model(GameScoreInfo{}).Where("user_id=?", userID).Set("all_logon_times", "all_logon_times+1").Error
	} else {
		return db.Debug().Model(GameScoreInfo{}).Where("user_id=?", userID).Updates(data).Error
	}
}

/*
UPDATE game_score_info SET total_coin_stream =total_coin_stream + ABS(dec_user_score), gold_coin=gold_coin+dec_user_score, revenue=revenue+ dec_revenue, win_count=win_count+ int_win_count, lost_count=lost_count+int_lost_count, draw_count=draw_count+int_draw_count,
			flee_count=flee_count+int_flee_count,play_time_count=play_time_count+int_play_time_count
			WHERE user_id=int_user_id;
*/

func UpdateGameScoreInfoUp(db *gorm.DB, userID int32, scoreInfo GameScoreInfo) error {
	sqlPre := "UPDATE game_score_info SET "
	sqlPre += fmt.Sprintf(" total_coin_stream =total_coin_stream + ABS(%v), gold_coin=gold_coin+%v, diamond=%v, "+
		"revenue=revenue+ %v, win_count=win_count+ %v, lost_count=lost_count+%v, "+
		"draw_count=draw_count+%v,"+
		"flee_count=flee_count+%v,play_time_count=play_time_count+%v,jackpot=%v"+
		" WHERE user_id=%v", scoreInfo.TotalCoinStream, scoreInfo.GoldCoin, scoreInfo.Diamond, scoreInfo.Revenue, scoreInfo.WinCount, scoreInfo.LostCount, scoreInfo.DrawCount, scoreInfo.FleeCount,
		scoreInfo.PlayTimeCount, scoreInfo.Jackpot, userID)
	return db.Debug().Exec(sqlPre).Error
}

type AccountsImage struct {
	UserID       int32 `json:"user_id"`        // 用户id
	RoleID       int32 `json:"role_id"`        // 角色标识
	SuitID       int32 `json:"suit_id"`        // 套装标识
	PhotoFrameID int32 `json:"photo_frame_id"` // 头像框标识
}

func (AccountsImage) TableName() string {
	return "accounts_image"
}

func GetAccountsImageByUserID(db *gorm.DB, userID int32) (AccountsImage, error) {
	var accountsImage AccountsImage

	err := db.Debug().Model(AccountsImage{}).Where(" user_id = ? ", userID).Find(&accountsImage).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return accountsImage, nil
	}

	return accountsImage, err
}
