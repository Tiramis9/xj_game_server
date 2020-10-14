package global

// table 相关
const (
	PokerCount     = 5 // 扑克计数 开奖结果扑克个数
	TablePlayCount = 2 //房间最少人数
)

// 游戏状态
const (
	GameStatusFree   = iota //空闲状态
	GameStatusStart         //游戏开始
	GameStatusQZ            //抢庄状态
	GameStatusJetton        //下注状态
	GameStatusTP            //摊牌状态
	GameStatusEnd           //结束状态
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

	//抢庄
	QZError1 = 4001
	QZError2 = 4002

	//下注
	JettonError1 = 5001
	JettonError2 = 5002

	//摊牌
	TPError1 = 6001
	TPError2 = 6002

	//进场
	JCError1 = 7001
)

//区域赔率
// 没牛
// 牛1 2 3 4 5 6
// 牛7 8 9
// 牛牛
// 五花牛 炸弹 五小牛
var AreaMultiple = [14]float32{
	1,
	1, 1, 1, 1, 1, 1,
	2, 2, 2,
	3,
	4, 4, 4,
}

//区域赔率
// 没牛
// 牛1 2 3 4 5 6
// 牛7 8
// 牛9
// 牛牛
// 五花牛
// 炸弹
// 五小牛
//var AreaMultiple = [14]float32{
//	1,
//	1, 1, 1, 1, 1, 1,
//	2, 2,
//	3,
//	4,
//	5,
//	8,
//	10,
//}
