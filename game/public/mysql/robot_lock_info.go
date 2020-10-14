package mysql

import (
	"github.com/jinzhu/gorm"
	"time"
)

/*
CREATE TABLE `robot_lock_info` (
  `user_id` int(11) NOT NULL COMMENT '机器标识',
  `logon_pass` char(32) NOT NULL DEFAULT '' COMMENT '机器密码',
  `robot_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0未使用，1使用中',
  `game_id` int(11) NOT NULL DEFAULT '0' COMMENT '游戏ID(服务器生成)',
  `batch_id` int(11) NOT NULL DEFAULT '0' COMMENT '批次标识',
  `lock_date_time` datetime DEFAULT NULL COMMENT '使用日期',
  `member_order` tinyint(4) NOT NULL DEFAULT '0' COMMENT '机器人等级',
  PRIMARY KEY (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='机器人信息表';
*/

type RobotLockInfo struct {
	UserID       int32     `json:"user_id"`
	LogonPass    string    `json:"logon_pass"`
	RobotStatus  int32     `json:"robot_status"`
	GameID       int32     `json:"game_id"`
	BatchID      int32     `json:"batch_id"`
	LockDateTime time.Time `json:"lock_date_time"`
	MemberOrder  int32     `json:"member_order"`
}

func InitRobotLockInfo(db *gorm.DB, batchID int32) error {
	return db.Exec("UPDATE robot_lock_info SET robot_status=0,game_id=0,batch_id=0 Where batch_id=?", batchID).Error
}

func InitRobotLockInfoByGameID(db *gorm.DB, GameID int32) error {
	return db.Exec("UPDATE robot_lock_info SET robot_status=0,game_id=0,batch_id=0 Where game_id=?", GameID).Error
}

func InitRobotLockInfoByUID(db *gorm.DB, userID int32) error {
	return db.Debug().Exec("UPDATE robot_lock_info SET robot_status=0,game_id=0,batch_id=0 Where user_id =?", userID).Error
}

func UpdateRobotLockInfoByUserID(db *gorm.DB, userID []int32, dataMap map[string]interface{}) error {
	return db.Model(RobotLockInfo{}).Where("user_id IN (?)", userID).Updates(dataMap).Error
}

func GetRobotLockInfoByLimit(db *gorm.DB, limitCount int64) ([]RobotLockInfo, error) {
	var robotInfo []RobotLockInfo
	err := db.Debug().Model(RobotLockInfo{}).Where("robot_status=0").Limit(limitCount).Find(&robotInfo).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return robotInfo, err
	}
	return robotInfo, nil
}

/*
select a.user_id,r.user_id FROM robot_lock_info as r
RIGHT join accounts_info as a on a.user_id=r.user_id
where r.user_id is NULL AND a.user_type=1 LIMIT 5
*/
func GetAccountInfoJoinRobotLockInfo(db *gorm.DB, limit int64) ([]int32, error) {
	type userIDList struct {
		UserID int32 `json:"user_id"`
	}
	list := make([]userIDList, 0)
	err := db.Debug().Table("robot_lock_info as r").Select("a.user_id as user_id").Joins("RIGHT join accounts_info as a on a.user_id=r.user_id").Where("r.user_id is NULL AND a.user_type=1").Limit(limit).Find(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []int32{}, nil
		}
		return []int32{}, err
	}
	var userList []int32
	for i := range list {
		userList = append(userList, list[i].UserID)
	}
	return userList, nil
}

/*
select a.user_id,r.user_id,a.last_logon_date FROM robot_lock_info as r
RIGHT join accounts_info as a on a.user_id=r.user_id
where  a.user_type=1  ORDER BY a.last_logon_date desc  LIMIT 5
*/
func GetAccountInfo2RobotLockInfo(db *gorm.DB, limit int64) ([]int32, error) {
	type userIDList struct {
		UserID        int32     `json:"user_id"`
		lastLogonDate time.Time `json:"last_logon_date"`
	}
	list := make([]userIDList, 0)
	err := db.Debug().Table("robot_lock_info as r").Select("a.user_id as user_id,a.last_logon_date AS last_logon_date").Joins("RIGHT join accounts_info as a on a.user_id=r.user_id").Where("a.user_type=1").Order("a.last_logon_date ASC").Limit(limit).Find(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []int32{}, nil
		}
		return []int32{}, err
	}
	var userList []int32
	for i := range list {
		userList = append(userList, list[i].UserID)
	}
	return userList, nil
}

func CreateRobotLockInfo(db *gorm.DB, robot RobotLockInfo) error {
	return db.Debug().Create(&robot).Error
}
