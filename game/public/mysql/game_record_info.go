package mysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

/*
CREATE TABLE `game_record_info` (
  `record_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录标识',
  `user_id` int(11) NOT NULL COMMENT '用户标识',
  `user_score` decimal(18,3) NOT NULL COMMENT '用户输赢积分',
  `user_revenue` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '游戏税收',
  `pre_score` decimal(18,3) NOT NULL COMMENT '游戏前积分',
  `score_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '积分类型：0 金币，1 余额',
  `kind_id` int(11) NOT NULL COMMENT '游戏标识',
  `kind_name` varchar(50) NOT NULL COMMENT '游戏类型名称',
  `game_id` int(11) NOT NULL COMMENT '游戏ID',
  `game_name` varchar(50) NOT NULL DEFAULT '' COMMENT '游戏名称',
  `user_type` int(11) NOT NULL DEFAULT '0' COMMENT '用户类型：0 真实用户，1 机器人，2 虚拟号',
  `revenue_ratio` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '系统抽税比例（百分比）',
  `water_score` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '积分流水（打码额）',
  `date_id` int(11) NOT NULL DEFAULT '0' COMMENT '日期值',
  `record_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录日期',
  `enter_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '游戏开始时间',
  `leave_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '游戏结束时间',
  `draw_id` int(11) NOT NULL DEFAULT '0' COMMENT '局数记录ID',
  PRIMARY KEY (`record_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1001 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户游戏写分记录表';
*/

type GameRecordInfo struct {
	RecordID     string    `json:"record_id"`
	UserID       int32     `json:"user_id"`
	UserScore    float32   `json:"user_score"`
	UserRevenue  float32   `json:"user_revenue"`
	PreScore     float32   `json:"pre_score"`
	ScoreType    int32     `json:"score_type"`
	KindID       int32     `json:"kind_id"`
	KindName     string    `json:"kind_name"`
	GameID       int32     `json:"game_id"`
	GameName     string    `json:"game_name"`
	UserType     int32     `json:"user_type"`
	RevenueRatio float32   `json:"revenue_ratio"`
	WaterScore   float32   `json:"water_score"`
	DateID       int       `json:"date_id"`
	RecordDate   time.Time `json:"record_date"`
	EnterTime    time.Time `json:"enter_time"`
	LeaveTime    time.Time `json:"leave_time"`
	DrawID       string    `json:"draw_id"`
}

func CreateGameRecordInfo(db *gorm.DB, recordInfo GameRecordInfo) error {
	return db.Debug().Create(&recordInfo).Error
}

/*
CREATE TABLE `stream_score_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `date_id` int(11) NOT NULL COMMENT '日期标识',
  `user_id` int(11) NOT NULL COMMENT '用户标识',
  `kind_id` int(11) NOT NULL COMMENT '游戏类型标识',
  `kind_name` varchar(50) NOT NULL COMMENT '游戏类型名称',
  `game_id` int(11) NOT NULL COMMENT '游戏标识',
  `game_name` varchar(50) NOT NULL DEFAULT '' COMMENT '游戏名称',
  `data_type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '数据类型：0 金币，1 余额',
  `total_score` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '总积分',
  `total_revenue` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '总税收',
  `win_count` int(11) NOT NULL DEFAULT '0' COMMENT '赢局数',
  `lost_count` int(11) NOT NULL DEFAULT '0' COMMENT '输局数',
  `play_time_count` int(11) NOT NULL DEFAULT '0' COMMENT '游戏时长',
  `online_time_count` int(11) NOT NULL DEFAULT '0' COMMENT '在线时长',
  `first_collect_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始统计时间',
  `last_collect_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后统计时间',
  `water_score` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '积分流水（绝对值累加）',
  `game_total_count` int(11) NOT NULL DEFAULT '0' COMMENT '游戏房间局数',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='用户每个游戏输赢日统计表';
*/
type StreamScoreInfo struct {
	ID               int32     `json:"id"`
	DateID           int       `json:"date_id"`
	UserID           int32     `json:"user_id"`
	KindID           int32     `json:"kind_id"`
	KindName         string    `json:"kind_name"`
	GameID           int32     `json:"game_id"`
	GameName         string    `json:"game_name"`
	DataType         int32     `json:"data_type"`
	TotalScore       float32   `json:"total_score"`
	TotalRevenue     float32   `json:"total_revenue"`
	LostCount        int32     `json:"lost_count"`
	WinCount         int32     `json:"win_count"`
	PlayTimeCount    int32     `json:"play_time_count"`
	OnlineTimeCount  int32     `json:"online_time_count"`
	FirstCollectDate time.Time `json:"first_collect_date"`
	LastCollectDate  time.Time `json:"last_collect_date"`
	WaterScore       float32   `json:"water_score"`
	GameTotalCount   int32     `json:"game_total_count"`
}

func CheckStreamScoreInfoByDay(db *gorm.DB, stream *StreamScoreInfo) bool {
	var streamIn StreamScoreInfo
	err := db.Model(StreamScoreInfo{}).Debug().Where(" date_id= ? AND user_id= ? AND kind_id = ? AND game_id = ? AND data_type = ?", stream.DateID, stream.UserID, stream.KindID, stream.GameID, stream.DataType).Find(&streamIn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return true
		}
		return false
	}
	if streamIn.ID != 0 {
		stream.ID = streamIn.ID
		stream.TotalScore += streamIn.TotalScore
		stream.WaterScore += streamIn.WaterScore
		stream.TotalRevenue += streamIn.TotalRevenue
		stream.WinCount += streamIn.WinCount
		stream.LostCount += streamIn.LostCount
		stream.PlayTimeCount += streamIn.PlayTimeCount
		stream.OnlineTimeCount += streamIn.OnlineTimeCount
		stream.GameTotalCount += streamIn.GameTotalCount
		stream.FirstCollectDate = streamIn.FirstCollectDate
		stream.LastCollectDate = time.Now()

		return false
	}
	return true
}

func CreateStreamScoreInfo(db *gorm.DB, stream *StreamScoreInfo) error {
	return db.Debug().Create(stream).Error
}

func UpdateStreamScoreInfoByID(db *gorm.DB, stream *StreamScoreInfo) error {

	return db.Debug().Model(StreamScoreInfo{}).Where("id=?", stream.ID).Update(stream).Error
}
func CreateOrUpdateStreamScoreInfo(db *gorm.DB, stream StreamScoreInfo) error {
	sqlPre := "INSERT INTO stream_score_info(date_id, user_id,kind_id,kind_name,game_id,game_name,total_score, " +
		"total_revenue,data_type,  win_count, lost_count,play_time_count, online_time_count, first_collect_date, " +
		"last_collect_date,water_score,game_total_count)" +
		"VALUES("
	sqlPre += fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)", stream.DateID, stream.UserID, stream.KindID, stream.KindName, stream.GameID, stream.GameName, stream.TotalScore, stream.TotalRevenue,
		stream.DataType, stream.WinCount, stream.LostCount, stream.PlayTimeCount, stream.OnlineTimeCount, "NOW()", "NOW()", stream.WaterScore, stream.GameTotalCount)
	sqlPre += " on DUPLICATE key update total_score = total_score + values(total_score), water_score= water_score + values(water_score), " +
		"total_revenue=total_revenue+values(total_revenue),win_count=win_count+values(win_count),lost_count=lost_count+values(lost_count)," +
		" play_time_count= play_time_count+values(play_time_count), online_time_count = online_time_count +values(online_time_count), last_collect_date=NOW()"

	where := fmt.Sprintf(" WHERE date_id=%v AND user_id=%v kind_id = %v AND game_id = %v AND data_type = %v", stream.DateID, stream.UserID, stream.KindID, stream.GameID, stream.DataType)
	return db.Debug().Exec(sqlPre + where).Error
}

/*
CREATE TABLE `record_diamond_draw_info` (
  `draw_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '局数标识',
  `kind_id` int(11) NOT NULL COMMENT '类型标识',
  `game_id` int(11) NOT NULL COMMENT '房间标识',
  `table_id` int(11) NOT NULL COMMENT '桌子号码',
  `user_count` int(11) NOT NULL COMMENT '用户数目',
  `android_count` int(11) NOT NULL COMMENT '机器数目',
  `waste` decimal(18,3) NOT NULL COMMENT '损耗数目',
  `revenue` decimal(18,3) NOT NULL COMMENT '税收数目',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `insert_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '插入时间',
  `draw_course` longblob COMMENT '游戏过程',
  `date_id` int(11) NOT NULL DEFAULT '0' COMMENT '日期值',
  PRIMARY KEY (`draw_id`) USING BTREE,
  KEY `Index_DateID` (`date_id`) USING BTREE COMMENT '日期索引值'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='每局游戏记录表';
*/
type RecordDiamondDrawInfo struct {
	DrawID    string `json:"draw_id"`
	KindID    int32  `json:"kind_id"`
	GameID    int32  `json:"game_id"`
	TableID   int32  `json:"table_id"`
	UserCount int32  `json:"user_count"`

	AndroidCount int32     `json:"android_count"`
	Waste        float32   `json:"waste"`
	Revenue      float32   `json:"revenue"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	InsertTime   time.Time `json:"insert_time"`
	DrawCourse   int64     `json:"draw_course"`
	DateID       int       `json:"date_id"`
	Detail       string    `json:"detail"` //局数详细记录
}

func (r *RecordDiamondDrawInfo) TableName() string {
	return "record_diamond_draw_info"
}

type RecordCoinDrawInfo struct {
	DrawID    string `json:"draw_id"`
	KindID    int32  `json:"kind_id"`
	GameID    int32  `json:"game_id"`
	TableID   int32  `json:"table_id"`
	UserCount int32  `json:"user_count"`

	AndroidCount int32     `json:"android_count"`
	Waste        float32   `json:"waste"`
	Revenue      float32   `json:"revenue"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	InsertTime   time.Time `json:"insert_time"`
	DrawCourse   int64     `json:"draw_course"`
	DateID       int       `json:"date_id"`
}

func (r *RecordCoinDrawInfo) TableName() string {
	return "record_coin_draw_info"
}
func CreateRecordDiamondDrawInfo(db *gorm.DB, record RecordDiamondDrawInfo) error {
	return db.Create(&record).Error
}

func UpdateRecordDiamondDrawInfo(db *gorm.DB, data RecordDiamondDrawInfo) error {
	return db.Debug().Update(data).Error
}
func CreateRecordCoinDrawInfo(db *gorm.DB, record RecordCoinDrawInfo) error {
	return db.Create(&record).Error
}
