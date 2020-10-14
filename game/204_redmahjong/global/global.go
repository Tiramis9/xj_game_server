package global

// table 相关
const (
	TablePlayCount = 4    //房间最少人数
	MahjongCount   = 112  //麻将数量
	MagicMahjong   = 0x35 //红中癞子
)

//操作类型
const (
	WIK_NULL      = 0x01  //过
	WIK_LEFT      = 0x02  //左吃
	WIK_CENTER    = 0x04  //中吃
	WIK_RIGHT     = 0x08  //右吃
	WIK_PENG      = 0x10  //碰牌
	WIK_BU_GANG   = 0x20  //补杠
	WIK_MING_GANG = 0x40  //明杠
	WIK_AN_GANG   = 0x80  //暗杠
	WIK_HU        = 0x100 //胡牌
)

//胡牌类型
const (
	CHR_PING_HU       = 0x200 //平胡
	CHR_SI_HONG_ZHONG = 0x400 //四红中
	CHR_QI_DUI        = 0x800 //七对
)

//特殊加番
const (
	CHR_TIAN_HU       = 0x1000 //天胡
	CHR_DI_HU         = 0x2000 //地胡
	CHR_QIANG_GANG_HU = 0x4000 //抢杠胡
	CHR_GANG_KAI      = 0x8000 //杠开
)

//癞子
const (
	CHR_MAGIC = 0x10000 //癞子胡
)

// 游戏状态
const (
	GameStatusFree = iota //空闲状态
	GameStatusPlay        //游戏状态
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

	//出牌
	OutCardError1 = 4001
	OutCardError2 = 4002

	//操作
	OperateError1 = 5001
	OperateError2 = 5002

	//准备
	ZBError1 = 6001
	ZBError2 = 6002

	//听牌
	TingError1 = 7001
	TingError2 = 7002

	//换桌
	ChangeError1 = 8001
	ChangeError2 = 8002
)
