package public

const (
	SuccessCode    = 0
	ErrorLackCode  = 1
	ErrorSqlCode   = 2
	ErrorRidesCode = 3

	//参数签名秘钥
	DesKey = "r5k1*8a$@8dc!dytkcs2dqz!"

	//redis key
	RedisKeyTableServer        = "table:server:"           //游戏列表 kind:game:table
	RedisKeyTableServerList    = "table:server:list:"      //游戏服务器列表
	RedisKeyGameServer         = "game:server:"            //游戏列表
	RedisKeyGameServerList     = "game:server:list:"       //游戏服务器列表
	RedisKeyGameGRPCServerList = "grpc:server:list:"       //游戏服务器GRPC列表
	RedisKeyGameGRPCServer     = "grpc:server:"            //游戏服务器GRPC
	RedisKeyToken              = "user:login:token:"       //token 缓存
	RedisKeyLoginServerList    = "user:login:server:list:" //登录服务器

	RedisKeyLoginServer         = "login:server:"          //登录服务器名称
	RedisKeyHallServerList      = "user:hall:server:list:" //大厅服务器列表
	RedisKeyHallServer          = "hall:server:"           //大厅服务器名称
	RedisKeyUserRecharge        = "user:recharge:"         //充值通知 金币变更消息
	RedisKeyUserDiamondRecharge = "user:recharge:diamond:" //充值通知 金币变更消息
	RedisGameVersionChange      = "user:game:version:"     // 游戏版本变更通知
	RedisRoomInfoChange         = "user:game:room:info:"    // 游戏房间变更通知
	//time
	FormatTime      = "15:04:05"            //时间格式
	FormatDate      = "2006-01-02"          //日期格式
	FormatDateTime  = "2006-01-02 15:04:05" //完整时间格式
	FormatDateTime2 = "2006-01-02 15:04"    //完整时间格式

	//配置文件路径
	// 全局配置路径
	GlobalConfigYmlPath = "/../conf/global.yml"
	//登录服务器配置文件路径
	LoginConfigYmlPath = "/../conf/login.yml"
	//大厅服务器
	HallConfigYmlPath = "/../conf/hall.yml"
	//游戏服务器配置文件路径
	LongHuDouConfigYmlPath101         = "/../conf/101_longhudou.yml"
	BaiJiaLeYmlPath102                = "/../conf/102_baijiale.yml"
	BaiRenNiuNiuConfigYmlPath103      = "/../conf/103_bairenniuniu.yml"
	SengLinWuHuiConfigYmlPath104      = "/../conf/104_senglinwuhui.yml"
	BenChiBaoMaConfigYmlPath105       = "/../conf/105_benchibaoma.yml"
	QiangZhuangNiuNiuConfigYmlPath201 = "/../conf/201_qiangzhuangniuniu.yml"
	ZhaJinHuaConfigYmlPath202         = "/../conf/202_zhajinhua.yml"
	DouDiZhuConfigYmlPath203          = "/../conf/203_doudizhu.yml"
	HZMJConfigYmlPath204              = "/../conf/204_hzmj.yml"
	DeZhouPukeConfigYmlPath205        = "/../conf/205_dezhoupuke.yml"
	YaoTouZiConfigYmlPath206          = "/../conf/206_yaotouzi.yml"
	QiangZhuangNiuNiuConfigYmlPath207 = "/../conf/207_qiangzhuangniuniu_kansanzhang.yml"
	Fish3dConfigYmlPath301            = "/../conf/301_3dfish.yml"
	JiuXianLawangConfigYmlPath401     = "/../conf/401_jiuxianlawang.yml"
	LingDangConfigYmlPath402          = "/../conf/402_lingdang.yml"
	HongBaoSaoLeiConfigYmlPath403     = "/../conf/403_hongbaosaolei.yml"
)
