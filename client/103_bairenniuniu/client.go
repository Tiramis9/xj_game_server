package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"time"
	gamemsg "xj_game_server/client/103_bairenniuniu/103"
)

const (
	//addr = "47.56.172.167:8000"
	addr = "47.107.188.43:1033"
	//addr = "127.0.0.1:8005"
	//addr = "47.56.172.167:13001"
)

func main() {
	tcpClient()
}
func tcpClient() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return
	}
	fmt.Println("已连接服务器")
	defer conn.Close()
	go sender(conn)
	read(conn)

}

//字节转换成整形
func BytesToInt(n interface{}, b []byte) error {
	bytesBuffer := bytes.NewBuffer(b)
	err := binary.Read(bytesBuffer, binary.BigEndian, n)
	return err
}
func CreteCmd(cmd int16, data []byte) []byte {
	var msg = make([]byte, 0)
	var len = len(data)
	lenByte, _ := IntToBytes(int16(len) + 2)
	cmdByte, _ := IntToBytes(cmd)
	msg = append(msg, lenByte...) //长度
	msg = append(msg, cmdByte...) //命令
	msg = append(msg, data...)    //数据
	return msg

}

//整形转换成字节  大端模式   高位在前
func IntToBytes(n interface{}) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes(), err
}

//Processor.Register(&Game_S_ReqlyFail{})	 0		//注册请求失败消息
//Processor.Register(&Game_C_TokenLogin{})	1	//注册token登陆消息
//Processor.Register(&Game_S_LoginSuccess{})2		//注册登陆成功消息
//Processor.Register(&Game_C_UserSitDown{})	3	//注册用户坐下消息
//Processor.Register(&Game_S_JettonScene{})	4	//注册下注场景消息
//Processor.Register(&Game_S_LotteryScene{})5		//注册开奖场景消息
//Processor.Register(&Game_C_UserStandUp{})	6	//注册用户起立消息
//Processor.Register(&Game_S_GameStart{})	7		//注册游戏开始消息
//Processor.Register(&Game_S_GameConclude{})8		//注册游戏结束消息
//Processor.Register(&Game_C_UserJetton{})	9	//注册用户下注消息

func loginMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "88888888"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMCIsIm5iZiI6MTU3NzA5NTk2N30.IKY42fGV8g-PF4m4H84HLQ4jZl6idbSYgmOhzXZV9ltd368tflvctdi913HLLVQbejIn6b1ePE_VkNLyTysntUS1jX1SZoT07UVCeTxOBy7L08b2TEj6H725GbtgbvqlqQeEF3hhjGuoubuiJvDk800cEYjJXcwy6kM282iGkk3UQYv7bzAd1fG9enEfFZ5Eu_RL_OL5Uyqv9vWkHoW353K_nrFiVtxW0fIx4zrYivurs2-iXLGSngSP2BGxI9NzNzuQyc_icCx_psre7oS1XheD05-1AeBN2mleWin89-HkgYdwzmmuZOXvwEQKcEZvoF8DPVlEpKdb6d9bfQRZ1Q"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODUyODEyNjMsImlzcyI6IjQiLCJuYmYiOjE1ODI2ODkyNjN9.K96z5Yb_qPEazrkNVhdzPep544QgOOiKy8dE4X5K75THA4718Qb8q3PlouG_ttiSNYZGAGQjr3SExquEI3tJA9guWkEGupWNykqgiDDurlQ4CaE96QZbefKyyatnhZTlgie96ISy9BkoszyxZ_fdypMvgX8_Q4vNinJso8JIg9Psejo99sMeqNK2VBdz0Gei0J5OLOR0gqNcE7z3JT29tSEA2sZuSz71Smr3T-9iPfqQppGcuKfEtxMAeBYYo9lrsWikITiNBu7EYRUPjDTX9yrvrkxJtq3ORTEVqRmoLU5Jx80yf6L_BWlYzXoI7zXE8qPe6g9RbUrvsRGtdNaG2Q"

	//data, _ := proto.Marshal(loginVisitor)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd(0x00, data)
}

func sitDownMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserSitDown)
	loginVisitor.ChairID = 0
	loginVisitor.TableID = 0
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd(0x01, data)
}

func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd(0x02, data)
}

func userJettonUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	loginVisitor.JettonArea = 0
	loginVisitor.JettonScore = 1
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd(0x03, data)
}

func getUserList() []byte {
	userList := new(gamemsg.Game_C_UserList)
	userList.Page = 1
	userList.Size = 20
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(userList).Elem().Name(): userList})
	return CreteCmd(0x12, data)
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		sitDown := sitDownMsg()
		//standUp := standUpMsg()
		//userJetton := userJettonUpMsg()
		//userList := getUserList()
		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}
		//log.Println("login:", login, n)
		//坐下
		time.Sleep(1 * time.Second)
		_, err = conn.Write(sitDown)
		if err != nil {
			log.Println("write:", err)
			return
		}

		//_, err = conn.Write(userList)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}
		//fmt.Println("1111:", userList)
		//_,err
		//time.Sleep(3 * time.Second)
		//_, err = conn.Write(standUp)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}

		//time.Sleep(3 * time.Second)
		//_, err = conn.Write(sitDown)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}
		//log.Println("坐下:", sitDown, n)

		////下注
		//time.Sleep(3 * time.Second)
		//n, err = conn.Write(userJetton)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}
		//log.Println("下注:", userJetton, n)
		break
		//离开
		//time.Sleep(3 * time.Second)
		//n, err = conn.Write(standUp)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}
		//log.Println("写数据:", standUp, n)

	}
}

//Processor.Register(&Game_S_ReqlyFail{})   4  //请求失败消息
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_JettonScene{})   //下注场景消息
//Processor.Register(&Game_S_LotteryScene{})  //开奖场景消息
//Processor.Register(&Game_S_GameStart{})     //游戏开始消息
//Processor.Register(&Game_S_GameConclude{})  //游戏结束消息
//Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
//Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
//Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息
//Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息
//Processor.Register(&Game_S_UserJetton{})    //下注通知
//Processor.Register(&Game_S_Hall{})          //游戏结束发送大厅场景

func read(conn net.Conn) {
	for {
		var message = make([]byte, 102400)
		n, err := conn.Read(message)
		if err != nil && err != io.EOF || len(message) == 0 {
			log.Println("read:", err)
			return
		}
		message = message[:n]
		if len(message) == 0 {
			break
		}
		var cmd int16
		err = BytesToInt(&cmd, message[2:4])
		if err != nil {
			log.Printf("cmd 错误 err:%d\n", cmd)
			return
		}
		//fmt.Println("cmd:", cmd)
		switch cmd {
		case 4: //注册请求失败消息
			var msg gamemsg.Game_S_ReqlyFail
			err = json.Unmarshal(message[4:], &msg)
			log.Print("请求失败消息recvString:\n", msg)
		case 5: //注册登陆成功消息
			var msg gamemsg.Game_S_LoginSuccess
			err = json.Unmarshal(message[4:], &msg)
			//bytes, _ := json.Marshal(msg)
			log.Print("登陆成功消息recvString:\n", msg)
		case 6: //注册下注场景消息
			var msg gamemsg.Game_S_JettonScene
			err = json.Unmarshal(message[4:], &msg)
			//bytes, _ := json.Marshal(msg)
			log.Print("下注场景消息recvString:\n", msg)
		case 7: //注册开奖场景消息
			var msg gamemsg.Game_S_LotteryScene
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("开奖场景消息recvString:\n", string(bytes))
		case 8: //注册游戏开始消息
			var msg gamemsg.Game_S_GameStart
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("游戏开始消息recvString:\n", string(bytes))
		case 9: //游戏结束消息
			var msg gamemsg.Game_S_GameConclude
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("游戏结束消息recvString:\n", string(bytes))
		case 10: //用户上线通知消息
			var msg gamemsg.Game_S_OnLineNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户上线消息recvString:\n", string(bytes))
		case 11: //用户掉线通知消息
			var msg gamemsg.Game_S_OffLineNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户掉线消息recvString:\n", string(bytes))
		case 12: //起立通知消息
			var msg gamemsg.Game_S_StandUpNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("起立通知消息recvString:\n", string(bytes))
		case 13: //坐下通知消息
			var msg gamemsg.Game_S_SitDownNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("坐下通知消息recvString:\n", string(bytes))
		case 14: //下注通知
			var msg gamemsg.Game_S_UserJetton
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注通知消息recvString:\n", string(bytes))
		case 15: //游戏结束大厅场景消息
			var msg gamemsg.Game_S_Hall
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("游戏结束大厅场景消息recvString:\n", string(bytes))
		case 17: //当前下注状况,每个区域,每个玩家的下注情况
			var msg gamemsg.Game_S_AreaJetton
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("每个区域,每个玩家的下注情况recvString:\n", string(bytes))
		case 19: //获取用户列表
			var msg gamemsg.Game_S_UserList
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("获取用户列表recvString:\n", string(bytes))
		default:
			log.Println("无效命令", cmd)
		}

	}
}
