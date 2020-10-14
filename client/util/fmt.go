package util

import (
	"encoding/json"
	"fmt"
	"log"
)

func JsonFmt(c interface{}) {
	// 将这个映射序列化到JSON 字符串
	data, err := json.MarshalIndent(c, "", "      ") //这里返回的data值，类型是[]byte
	if err != nil {
		log.Println("ERROR:", err)
	}

	fmt.Println(string(data))
}
