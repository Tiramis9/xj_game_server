package global

// table 相关
const (
	PokerCount     = 5 // 扑克计数 开奖结果扑克个数
	TablePlayCount = 3 //房间人数
)

// 游戏状态
const (
	GameStatusFree = iota //空闲状态
	GameStatusJF          //叫分状态
	GameStatusPlay        //下注状态
)

//通知用户上线

var NoticeRobotOnline = make(chan int32, 10000)

// 通知匹配 机器人
var NoticeLoadMath = make(chan int32, 100)

// 错误码
const (
	ServerError = 500

	// 登录
	LoginError      = 1001
	LoginTokenError = 1002

	//坐下
	SitDownError1 = 2001
	SitDownError2 = 2002
	SitDownError3 = 2003

	//起立
	StandUpError1 = 3001
	StandUpError2 = 3002

	//看牌
	KPError1 = 4001

	//出牌
	CPError1 = 5001
	CPError2 = 5002
	CPError3 = 5003

	//准备
	ZBError1 = 6001
	ZBError2 = 6002

	//进场
	JCError1 = 7001

	//叫分
	JFError1 = 8001
	JFError2 = 8002

	//过
	PassError1 = 9001
	PassError2 = 9002
)

// 错误码
const (
	CardTypeStatus      = iota
	CardTypeSINGLE      //单根
	CardTypeDOUBLE      //对子
	CardTypeSINGLEALONE //顺子
	CardTypeDOUBLEALONE //连对
	CardTypeTHREE       //三不带
	CardTypeTHREEONE    //三带一
	CardTypeTHREETWO    //三带对
	CardTypeFOURONE     //四带二单
	CardTypeFOURTWO     //四带二对
	CardTypeBOMB        //炸弹
	CardTypeKINGBOMB    //王炸
	//CardTypePLANE       //飞机
	//CardTypePLANEEMPTY  //三不带飞机
)
