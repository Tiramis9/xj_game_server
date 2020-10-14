package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	gamemsg "xj_game_server/client/202msg"
	"time"
)

const (
	//addr = "47.56.172.167:2010"
	//addr = "192.168.1.43:2010"
	addr = "127.0.0.1:8011"
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

// 登录
func loginMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "88888888"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Nzg0NTI2MzUsImlzcyI6IjMwIiwibmJmIjoxNTc1ODYwNjM1fQ.mopif8V5a0filvK1EQ9lHMKMYozSf9NrS1iogLpSq7vHQMc4Ik0Q-KzV23hy17_h45jQ0AczwUDe3f0MIJCdiwF_QTSzMI14UcHdlopyzoFO-wQytauWaQ8xn43FO5GFYpF5Bucd52nddg4RVKG35EQDOVzaUGn_wB2TjotieAE78j6H1SmjOVm-kkslNwJjBdr9DM3E10xh9na3ozrKI31DmhYrJXSaNaC5E0yYx__olmxCCrsXN6tldKkGMf83XTvZZJFazWj0x8g5ZfNYlv3zATc_QmcU6CFPZp9Vx04wk19mYmggQS3CsBllfMJBZjZamLT8o8IZd5ExlUn-Xw"
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

//坐下
func sitDownMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserSitDown)
	loginVisitor.ChairID = 0
	loginVisitor.TableID = 0
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x01, data)
}

//起立
func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

//下注
func userJettonUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	loginVisitor.Multiple = 5
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x04, data)
}

//摊牌
func userTpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserTP)
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x05, data)
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		sitDown := sitDownMsg()
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

		break

	}
}

func read(conn net.Conn) {
	for {
		var message = make([]byte, 1024*5)
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
		fmt.Println("cmd:", cmd)
		switch cmd {
		case 8: //请求失败消息
			var msg gamemsg.Game_S_ReqlyFail
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("请求失败消息recvString:\n", string(bytes))
		case 9: //登陆成功消息
			var msg gamemsg.Game_S_LoginSuccess
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("登陆成功消息recvString:\n", string(bytes))
		case 10: //空闲场景消息
			var msg gamemsg.Game_S_FreeScene
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("空闲场景消息recvString:\n", string(bytes))
		case 11:
			var msg gamemsg.Game_S_JettonScene
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注场景消息recvString:\n", string(bytes))
		case 12:
			var msg gamemsg.Game_S_TPScene
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("摊牌场景消息recvString:\n", string(bytes))
		case 13:
			var msg gamemsg.Game_S_OnLineNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户上线通知消息recvString:\n", string(bytes))
		case 14:
			var msg gamemsg.Game_S_OffLineNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户掉线通知消息recvString:\n", string(bytes))
		case 15:
			var msg gamemsg.Game_S_StandUpNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("起立通知消息15recvString:\n", string(bytes))
		case 16:
			var msg gamemsg.Game_S_SitDownNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("坐下通知消息16recvString:\n", string(bytes))
		case 17:
			var msg gamemsg.Game_S_StartTime
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("开始定时器通知消息17recvString:\n", string(bytes))
		case 18:
			var msg gamemsg.Game_S_GameConclude
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("结束游戏通知消息18recvString:\n", string(bytes))
		case 19:
			var msg gamemsg.Game_S_UserKP
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("抢庄通知消息19recvString:\n", string(bytes))
		case 20:
			var msg gamemsg.Game_S_UserJetton
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注通知消息20recvString:\n", string(bytes))
		case 21:
			var msg gamemsg.Game_S_UserTP
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("摊牌通知消息21recvString:\n", string(bytes))
		case 22:
			var msg gamemsg.Game_S_GameBP
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("比牌通知消息22recvString:\n", string(bytes))
		case 23:
			var msg gamemsg.Game_S_UserQP
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("弃牌通知消息23recvString:\n", string(bytes))
		}

	}
}

// 服务端-----
//Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息8
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_FreeScene{})     //空闲场景消息10
//Processor.Register(&Game_S_JettonScene{})   //下注场景消息
//Processor.Register(&Game_S_TPScene{})       //摊牌场景消息
//Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
//Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
//Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息
//Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息
//Processor.Register(&Game_S_StartTime{})     //开始定时器通知消息
//Processor.Register(&Game_S_GameConclude{})  //结束游戏通知消息
//Processor.Register(&Game_S_UserKP{})        //抢庄通知消息
//Processor.Register(&Game_S_UserJetton{})    //下注通知消息
//Processor.Register(&Game_S_UserTP{})        //摊牌通知消息
//Processor.Register(&Game_S_GameBP{})        //比牌通知消息
//Processor.Register(&Game_S_UserQP{})        //摊牌通知消息
