package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

/*CREATE TABLE `system_status_info` (
  `status_name` varchar(32) NOT NULL COMMENT '状态名字',
  `status_value` int(11) NOT NULL DEFAULT '0' COMMENT '状态数值',
  `status_string` varchar(50) NOT NULL DEFAULT '' COMMENT '状态字符',
  `status_tip` varchar(50) NOT NULL DEFAULT '' COMMENT '状态显示名称',
  `status_description` varchar(200) NOT NULL DEFAULT '' COMMENT '字符的描述',
  `sort_id` int(11) NOT NULL DEFAULT '99' COMMENT '排序，越小越靠前',
  PRIMARY KEY (`status_name`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='系统设置表';

*/

type SystemStatusInfo struct {
	StatusName        string `json:"status_name"`
	StatusValue       int    `json:"status_value"`
	StatusString      string `json:"status_string"`
	StatusDescription string `json:"status_description"`
	SortID            int    `json:"sort_id"`
}

func (s SystemStatusInfo) TableName() string {
	return "system_status_info"
}
func CheckSystemStatusInfoByName(db *gorm.DB, statusName string) (string, error) {
	var statusInfo SystemStatusInfo
	err := db.Model(SystemStatusInfo{}).Where("status_name= ?", statusName).Find(&statusInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return err.Error(), err
	}
	return statusInfo.StatusString, nil
}

/*
CREATE TABLE `system_stream_room_info` (
  `date_id` int(11) NOT NULL COMMENT '日期标识',
  `kind_id` int(11) NOT NULL COMMENT '类型标识',
  `game_id` int(11) NOT NULL COMMENT '游戏标识',
  `logon_count` int(11) NOT NULL DEFAULT '0' COMMENT '进入数目',
  `collect_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '录入时间',
  PRIMARY KEY (`date_id`,`kind_id`,`game_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='进入游戏次数日统计表';
*/

type SystemStreamRoomInfo struct {
	DateID      int       `json:"date_id"`
	KindID      int32     `json:"kind_id"`
	GameID      int32     `json:"game_id"`
	LogonCount  int       `json:"logon_count"`
	CollectDate time.Time `json:"collect_date"`
}

//insert into system_stream_room_info (date_id,kind_id,game_id,logon_count) values(0715,201,59,1) on DUPLICATE key update logon_count=logon_count+values(logon_count)
func CreateOrUpdateSystemStreamRoomInfo(db *gorm.DB, streamRoom SystemStreamRoomInfo) error {
	var sqlPre = "insert into system_stream_room_info (date_id,kind_id,game_id,logon_count,collect_date) VALUES("

	sqlPre += fmt.Sprintf("%v,%v,%v,%v,%v) ", streamRoom.DateID, streamRoom.KindID, streamRoom.GameID, streamRoom.LogonCount, "NOW()") +
		" on DUPLICATE key update logon_count=logon_count+values(logon_count) "
	return db.Debug().Exec(sqlPre).Error

}
