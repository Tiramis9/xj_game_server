package mysql

import (
	"github.com/jinzhu/gorm"
	"time"
)

/*CREATE TABLE `confine_address` (
`addr_string` varchar(15) NOT NULL COMMENT '地址字符',
`enjoin_logon` tinyint(4) NOT NULL DEFAULT '0' COMMENT '限制登录',
`enjoin_register` tinyint(4) NOT NULL DEFAULT '0' COMMENT '限制注册',
`enjoin_over_date` datetime DEFAULT NULL COMMENT '过期时间',
`collect_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收集日期',
`collect_note` varchar(200) DEFAULT '' COMMENT '输入备注',
PRIMARY KEY (`addr_string`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='IP地址限制表';

*/
type ConfineAddress struct {
	AddrString     string    `json:"addr_string"`
	EnjoinLogon    int       `json:"enjoin_logon"`
	EnjoinRegister int       `json:"enjoin_register"`
	EnjoinOverDate time.Time `json:"enjoin_over_date"`
	CollectDate    time.Time `json:"collect_date"`
	CollectNote    string    `json:"collect_note"`
}

func CheckConfineAddressLimit(db *gorm.DB, addrString string) (bool, error) {
	var confAddr ConfineAddress
	err := db.Model(ConfineAddress{}).Where("addr_string=? AND (enjoin_over_date>now() OR enjoin_over_date IS NULL)", addrString).Find(&confAddr).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return confAddr.AddrString != "", err
}

/*
CREATE TABLE `confine_machine` (
  `machine_serial` varchar(64) NOT NULL COMMENT '机器序列',
  `enjoin_logon` tinyint(4) NOT NULL DEFAULT '0' COMMENT '限制登录',
  `enjoin_register` tinyint(4) NOT NULL DEFAULT '0' COMMENT '限制注册',
  `enjoin_over_date` datetime DEFAULT NULL COMMENT '过期时间',
  `collect_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收集日期',
  `collect_note` varchar(200) DEFAULT NULL COMMENT '输入备注',
  PRIMARY KEY (`machine_serial`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='机器限制表';
*/

type ConfineMachine struct {
	MachineSerial  string    `json:"machine_serial"`
	EnjoinLogon    int       `json:"enjoin_logon"`
	EnjoinRegister int       `json:"enjoin_register"`
	EnjoinOverDate time.Time `json:"enjoin_over_date"`
	CollectDate    time.Time `json:"collect_date"`
	CollectNote    string    `json:"collect_note"`
}

func CheckConfineMachineLimit(db *gorm.DB, machineSerial string) (bool, error) {
	var confMach ConfineMachine
	err := db.Model(ConfineMachine{}).Where("machine_serial=? AND (enjoin_over_date>now() OR enjoin_over_date IS NULL)", machineSerial).Find(&confMach).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return true, nil
	}
	return confMach.CollectNote != "", nil
}
