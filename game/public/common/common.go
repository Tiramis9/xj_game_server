package common

import "fmt"

const (
	/*	请求成功*/
	StatusOK                   = 200 // 请求执行成功并返回相应数据
	StatusCreated              = 201 //创建成功并返回相应资源数据
	StatusAccepted             = 202 // 接受请求，但无法立即完成创建行为
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // 请求执行成功，不返回相应资源数据
	/*客户端错误*/
	StatusBadRequest                   = 400 // 请求体包含语法错误
	StatusUnauthorized                 = 401 // 需要验证用户身份
	StatusForbidden                    = 403 // 服务器拒绝执行
	StatusNotFound                     = 404 // 找不到目标资源
	StatusMethodNotAllowed             = 405 // 不允许执行目标方法，响应中应该带有 Allow 头，内容为对该资源有效的 HTTP 方法
	StatusNotAcceptable                = 406 // 服务器不支持客户端请求的内容格式
	StatusGone                         = 410 // 被请求的资源已被删除
	StatusRequestEntityTooLarge        = 413 // POST 或者 PUT 请求的消息实体过大
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         = 415 // 服务器不支持请求中提交的数据的格式
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusMisdirectedRequest           = 421 // RFC 7540, 9.1.2
	StatusUnprocessableEntity          = 422 //  请求格式正确，但是由于含有语义错误，无法响应
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusTooEarly                     = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // 要求先决条件，如果想要请求能成功必须满足一些预设的条件要求先决条件，
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3
	StatusNetWork                      = 430 //网络异常
	StatusAlreadyExist                 = 440 //手机号码已注册
	StatusNotExist                     = 441 ////账号未注册
	StatusRequestParam                 = 442 //请求参数错误
	StatusMinEnterScore                = 443 //最低入场金额不足
	StatusAccountOrPassword            = 444 ////账号或密码错误
	StatusAccountExist                 = 445 ////账号已注册
	StatusUnknownPhone                 = 446 ////无效的手机号码
	StatusAccountErr                   = 447 ////账号异常，请联系客服
	StatusCodeLimit                    = 448 ////验证码获取次数上限
	StatusGameClose                    = 450 ////游戏已关闭
	StatusDownBreak                    = 451 //下载中断
	StatusUnknownRoom                  = 452 //房间不存在
	SMSSendingFailed                   = 453 //
	LoginHasFailed                     = 454 //
	StatusUnknownAgentCode             = 455 //

	StatusUnknownCode       = 456 //无效的验证码和验证码过期
	PasswordInconsistency   = 457 //支付密码不一致
	PaymentPasswordError    = 458 //支付密码错误
	ClaimFailedError        = 459 //领取失败
	SetPaymentPasswordError = 460 //请先设置支付密码

	/* 	服务端错误 */
	StatusInternalServerFail = 500 //服务器异常
	StatusNotImplementedFail = 501 //数据库异常 系统暂停
	StatusServiceUnavailable = 503 // 由于临时的服务器维护或者过载，服务器当前无法处理请求
	StatusNotExtendedFail    = 510 //验证失败
	StatusTokenTimeOut       = 511 //token过期
	StatusTokenParamFail     = 512 //token解析异常
	StatusTokenInvalid       = 513 //token无效
	StatusTokenEmpty         = 514 //token为空
	StatusHttpFail           = 520 //http异常
	StatusUserIDInvalid      = 521 //uid无效
	StatusParamFail          = 522 //参数异常

)

var statusDesc = map[int64]string{
	StatusOK:                    "成功",
	StatusCreated:               "创建成功并返回相应资源数据",
	StatusAccepted:              "接受请求，但无法立即完成创建行为",
	StatusNoContent:             "请求执行成功，不返回相应资源数据",
	StatusBadRequest:            "请求体包含语法错误",
	StatusUnauthorized:          "需要验证用户身份",
	StatusForbidden:             "服务器拒绝执行",
	StatusNotFound:              "找不到目标资源",
	StatusMethodNotAllowed:      "不允许执行目标方法，响应中应该带有 Allow 头，内容为对该资源有效的 HTTP 方法",
	StatusNotAcceptable:         "请求的内容格式错误",
	StatusGone:                  "被请求的资源已被删除",
	StatusRequestEntityTooLarge: "POST 或者 PUT 请求的消息实体过大",
	StatusUnsupportedMediaType:  "服务器不支持请求中提交的数据的格式",
	StatusUnprocessableEntity:   "请求格式正确，但是由于含有语义错误，无法响应",
	StatusPreconditionRequired:  "要求先决条件，如果想要请求能成功必须满足一些预设的条件要求先决条件，",
	StatusNetWork:               "网络异常",
	StatusAlreadyExist:          "手机号码已注册",
	StatusNotExist:              "账号未注册",
	StatusRequestParam:          "请求参数错误",
	StatusMinEnterScore:         "最低入场金额不足",
	StatusAccountOrPassword:     "账号或密码错误",
	StatusAccountExist:          "账号已注册",
	StatusUnknownPhone:          "无效的手机号码",
	StatusAccountErr:            "账号异常，请联系客服",
	StatusGameClose:             "游戏已关闭",
	StatusCodeLimit:             "验证码获取次数上限",
	StatusDownBreak:             "下载中断",
	StatusUnknownRoom:           "房间不存在",
	SMSSendingFailed:            "短信发送失败",
	LoginHasFailed:              "注册失败",
	StatusInternalServerFail:    "服务器异常",
	StatusNotImplementedFail:    "数据库异常",
	StatusServiceUnavailable:    "由于临时的服务器维护或者过载，服务器当前无法处理请求",
	StatusNotExtendedFail:       "验证失败",
	StatusTokenTimeOut:          "token过期",
	StatusTokenParamFail:        "token解析异常",
	StatusTokenInvalid:          "token无效",
	StatusTokenEmpty:            "token为空",
	StatusHttpFail:              "http异常",
	StatusUserIDInvalid:         "uid无效",
	StatusParamFail:             "参数异常",
	StatusUnknownAgentCode:      "无效的推荐码",

	StatusUnknownCode:       "无效的验证码",
	PasswordInconsistency:   "支付密码不一致",
	PaymentPasswordError:    "支付密码错误",
	ClaimFailedError:        "领取失败",
	SetPaymentPasswordError: "请先设置支付密码",
}

func Description(code int64) string {
	if msg, ok := statusDesc[code]; ok {
		return msg
	}

	return fmt.Sprintf("unknown error[%v]", code)
}
