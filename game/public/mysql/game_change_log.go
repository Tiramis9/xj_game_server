package mysql

import (
	"github.com/jinzhu/gorm"
	"time"
)

/*
CREATE TABLE `game_coin_change_log` (
  `log_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '标识',
  `user_id` int(11) NOT NULL COMMENT '用户标识',
  `capital_type_id` int(11) NOT NULL DEFAULT '1' COMMENT '资金变动类型：1充值，2提现，3游戏比分，4赠送，5奖励，7用户转账，8代理分红，9平台扣减10其他',
  `log_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  `capital_amount` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '变动金额',
  `last_amount` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '剩余金额',
  `client_ip` varchar(15) NOT NULL DEFAULT '0.0.0.0' COMMENT '变更IP',
  `remark` varchar(200) DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`log_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户金币变动记录表';
*/
/*
CREATE TABLE `game_diamond_change_log` (
  `log_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '标识',
  `user_id` int(11) NOT NULL COMMENT '用户标识',
  `capital_type_id` int(11) NOT NULL DEFAULT '1' COMMENT '资金变动类型：1充值，2提现，3游戏比分，4赠送，5奖励，7用户转账，8代理分红，9平台扣减10其他',
  `log_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  `capital_amount` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '变动余额',
  `last_amount` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '剩余余额',
  `client_ip` varchar(15) NOT NULL DEFAULT '0.0.0.0' COMMENT '变更IP',
  `remark` varchar(200) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`log_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户余额变动记录表';
*/
type GameCoinChangeLog struct {
	LogID         int32     `json:"log_id"`
	UserID        int32     `json:"user_id"`
	CapitalTypeID int32     `json:"capital_type_id"`
	LogDate       time.Time `json:"log_date"`
	CapitalAmount float32   `json:"capital_amount"`
	LastAmount    float32   `json:"last_amount"`
	ClientIP      string    `json:"client_ip"`
	Remark        string    `json:"remark"`
}
type GameDiamondChangeLog struct {
	LogID         int32     `json:"log_id"`
	UserID        int32     `json:"user_id"`
	CapitalTypeID int32     `json:"capital_type_id"`
	LogDate       time.Time `json:"log_date"`
	CapitalAmount float32   `json:"capital_amount"`
	LastAmount    float32   `json:"last_amount"`
	ClientIP      string    `json:"client_ip"`
	Remark        string    `json:"remark"`
}

func CreateGameDiamondChangeLog(db *gorm.DB, log GameDiamondChangeLog) error {
	return db.Debug().Create(&log).Error
}
func CreateGameCoinChangeLog(db *gorm.DB, log GameCoinChangeLog) error {
	return db.Debug().Create(&log).Error
}
