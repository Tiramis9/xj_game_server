package store

import (
	"math"
	"sync"
)

var GameControl Item

/**
CREATE TABLE `GameRoomConfig` (
  `GameID` int(11) NOT NULL AUTO_INCREMENT COMMENT '游戏ID(自增)',
  `KindID` int(11) NOT NULL COMMENT '游戏种类',
  `KindName` varchar(50) NOT NULL COMMENT '游戏种类名称',
  `GameName` varchar(50) NOT NULL COMMENT '游戏名',
  `SortID` int(11) NOT NULL DEFAULT '0' COMMENT '排序id',
  `TableCount` int(11) NOT NULL COMMENT '桌子数量',
  `ChairCount` int(11) NOT NULL COMMENT '椅子数量',
  `CellScore` decimal(18,3) NOT NULL COMMENT '游戏底分',
  `RevenueRatio` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '税收比例',
  `MinEnterScore` decimal(18,3) NOT NULL DEFAULT '0.000' COMMENT '最低进入积分',
  `DeductionsType` tinyint(4) NOT NULL DEFAULT '0' COMMENT '扣费类型（0：金币， 1：余额）',
  `StoresDecay` decimal(5,2) NOT NULL COMMENT '库存衰减比例',
  `StartStores` decimal(18,3) NOT NULL COMMENT '起始库存',
  `StartWinRate` decimal(5,2) NOT NULL COMMENT '起始胜率',
  `Threshold1` decimal(18,3) NOT NULL COMMENT '库存阈值1',
  `WinRate1` decimal(5,2) NOT NULL COMMENT '库存阈值1胜率',
  `Threshold2` decimal(18,3) NOT NULL COMMENT '库存阈值2',
  `WinRate2` decimal(5,2) NOT NULL COMMENT '库存阈值2胜率',
  `Threshold3` decimal(18,3) NOT NULL COMMENT '库存阈值3',
  `WinRate3` decimal(5,2) NOT NULL COMMENT '库存阈值3胜率',
  `Threshold4` decimal(18,3) NOT NULL COMMENT '库存阈值4',
  `WinRate4` decimal(5,2) NOT NULL COMMENT '库存阈值4胜率',
  `Threshold5` decimal(18,3) NOT NULL COMMENT '库存阈值5',
  `WinRate5` decimal(5,2) NOT NULL COMMENT '库存阈值5胜率',
  PRIMARY KEY (`GameID`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COMMENT='游戏房间配置表';
*/
type GameInfo struct {
	GameID         int32   `json:"game_id,omitempty"`          // 游戏ID
	KindID         int32   `json:"kind_id,omitempty"`          //游戏类型id
	KindName       string  `json:"kind_name,omitempty"`        //游戏类型名称
	GameName       string  `json:"game_name,omitempty"`        //游戏名
	SortID         int32   `json:"sort_id,omitempty"`          //排序id
	TableCount     int32   `json:"table_count,omitempty"`      //桌子数量
	ChairCount     int32   `json:"chair_count,omitempty"`      //椅子数量
	CellScore      float32 `json:"cell_score,omitempty"`       //游戏底分
	RevenueRatio   float32 `json:"revenue_ratio,omitempty"`    //税收比例
	UmRevenueRatio float32 `json:"um_revenue_ratio,omitempty"` //税收比例
	MinEnterScore  float32 `json:"min_enter_score,omitempty"`  //最低进入积分
	DeductionsType int32   `json:"deductions_type,omitempty"`  //扣费类型

	StoresDecay  float32 `json:"stores_decay,omitempty"`   //库存衰减比例
	StartStores  float32 `json:"start_stores,omitempty"`   //起始库存
	StartWinRate float32 `json:"start_win_rate,omitempty"` //起始胜率
	Threshold1   float32 `json:"threshold_1,omitempty"`
	WinRate1     float32 `json:"win_rate_1,omitempty"`
	Threshold2   float32 `json:"threshold_2,omitempty"`
	WinRate2     float32 `json:"win_rate_2,omitempty"`
	Threshold3   float32 `json:"threshold_3,omitempty"`
	WinRate3     float32 `json:"win_rate_3,omitempty"`
	Threshold4   float32 `json:"threshold_4,omitempty"`
	WinRate4     float32 `json:"win_rate_4,omitempty"`
	Threshold5   float32 `json:"threshold_5,omitempty"`
	WinRate5     float32 `json:"win_rate_5,omitempty"`
	Stores       float32
}

type WinAreas struct {
	LotteryPoker   []int32
	WinArea        []bool
	SystemScore    float32
	UserListLoss   map[int32]float32
	UserTax        map[int32]float32
	Stores         float32
	ColorIndex     int32
	LotterySpecial int32
}

type IntSlice []WinAreas

func (s IntSlice) Len() int { return len(s) }

func (s IntSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s IntSlice) Less(i, j int) bool {
	return math.Abs(float64(s[i].Stores)) < math.Abs(float64(s[j].Stores))
}

func (g GameInfo) TableName() string {
	return "game_room_config"
}

type Item struct {
	storeInfo *GameInfo
	l         sync.Mutex
}

//初始化
func (self *Item) OnInit(storeInfo *GameInfo) {
	self.l.Lock()
	defer self.l.Unlock()

	self.storeInfo = storeInfo
	self.storeInfo.Stores = storeInfo.StartStores
}

//获取用户胜率
func (self *Item) GetUserWinRate() float32 {
	self.l.Lock()
	defer self.l.Unlock()

	return self.storeInfo.StartWinRate

	//if self.storeInfo.StartStores < self.storeInfo.Threshold1 {
	//	return self.storeInfo.StartWinRate
	//}
	//
	//if self.storeInfo.StartStores >= self.storeInfo.Threshold1 && self.storeInfo.StartStores < self.storeInfo.Threshold2 {
	//	return self.storeInfo.WinRate1
	//}
	//
	//if self.storeInfo.StartStores >= self.storeInfo.Threshold2 && self.storeInfo.StartStores < self.storeInfo.Threshold3 {
	//	return self.storeInfo.WinRate2
	//}
	//
	//if self.storeInfo.StartStores >= self.storeInfo.Threshold3 && self.storeInfo.StartStores < self.storeInfo.Threshold4 {
	//	return self.storeInfo.WinRate3
	//}
	//
	//if self.storeInfo.StartStores >= self.storeInfo.Threshold4 && self.storeInfo.StartStores < self.storeInfo.Threshold5 {
	//	return self.storeInfo.WinRate4
	//}
	//
	//return self.storeInfo.WinRate5
}

//修改当前库存
func (self *Item) ChangeStore(score float32) {
	self.l.Lock()
	defer self.l.Unlock()

	self.storeInfo.StartStores += score
}

//获取当前库存
func (self *Item) GetStore() float32 {
	self.l.Lock()
	defer self.l.Unlock()

	return self.storeInfo.StartStores
}

//获取当前库存
func (self *Item) GetStore1() float32 {
	self.l.Lock()
	defer self.l.Unlock()

	return self.storeInfo.StartStores - self.storeInfo.Stores
}

func (self *Item) GetGameInfo() *GameInfo {
	if self.storeInfo == nil {
		self.storeInfo = new(GameInfo)
	}
	return self.storeInfo
}
