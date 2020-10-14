package global

import "strings"

var gameName = map[string]int32{
	"baijiale": 102,
	"bell":     403,
	"bcbm":     105,
	"dahuatou": 206,

	"dezhou":    205,
	"doudizhu":  203,
	"hongbao":   403,
	"jxlw":      401,
	"hzmajiang": 204,
	"longhudou": 101,

	"niuniubairen": 103,
	"qpby":         301,
	"qznn":         201,
	"slwh":         104,
	"zhajinhua":    202,
	"ksznn":        207,
}

func DecGameName2KindID(name string) int32 {
	name = strings.Replace(name, " ", "", -1)

	result, ok := gameName[name]
	if !ok {
		return 0
	}
	return result
}
