package mysql

import (
	"github.com/jinzhu/gorm"
	"time"
)

//锁定用户
type AccountPlayingLock struct {
	UserID      int32     `json:"user_id"`
	KindID      int32     `json:"kind_id"`
	GameID      int32     `json:"game_id"`
	EnterIP     string    `json:"enter_ip"`
	CollectDate time.Time `json:"collect_date"`
}

func (a AccountPlayingLock) TableName() string {
	return "account_playing_lock"
}

func CreateAccountPlayingLock(db *gorm.DB, accPlay AccountPlayingLock) error {
	return db.Create(&accPlay).Error
}
func DelAccountPlayingLock(db *gorm.DB, userID int32) error {
	return db.Where("user_id =?", userID).Delete(AccountPlayingLock{}).Error
}

func GetAccountPlayingLock(db *gorm.DB, userID int32) (lock AccountPlayingLock, err error) {
	err = db.Debug().Where("user_id =?", userID).Find(&lock).Error
	return
}
