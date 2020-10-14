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
	gamemsg "xj_game_server/client/402_ld/402msg"
)

const (
	//addr = "47.56.172.167:2010"
	//addr = "192.168.1.43:2010"
	addr = "127.0.0.1:4020"
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

//Processor.Register(&Game_C_TokenLogin{})  //token登陆消息
//Processor.Register(&Game_C_UserStandUp{}) //用户起立消息
//Processor.Register(&Game_C_UserJetton{})  //用户下注消息
//Processor.Register(&Game_C_UserCompare{}) ////比较大小
//Processor.Register(&Game_C_UserList{})    //获取用户列表客户端参数
//// 服务端-----
//Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_GameResult{})    //铃铛结果
//Processor.Register(&Game_S_CompareResult{}) //比大小结果
//Processor.Register(&Game_S_UserList{})      //获取用户列表服务器返回
// 登录
func loginMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "88888888"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMCIsIm5iZiI6MTU3NzA5NTk2N30.IKY42fGV8g-PF4m4H84HLQ4jZl6idbSYgmOhzXZV9ltd368tflvctdi913HLLVQbejIn6b1ePE_VkNLyTysntUS1jX1SZoT07UVCeTxOBy7L08b2TEj6H725GbtgbvqlqQeEF3hhjGuoubuiJvDk800cEYjJXcwy6kM282iGkk3UQYv7bzAd1fG9enEfFZ5Eu_RL_OL5Uyqv9vWkHoW353K_nrFiVtxW0fIx4zrYivurs2-iXLGSngSP2BGxI9NzNzuQyc_icCx_psre7oS1XheD05-1AeBN2mleWin89-HkgYdwzmmuZOXvwEQKcEZvoF8DPVlEpKdb6d9bfQRZ1Q"
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

//起立
func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x01, data)
}

//下注
func userJettonUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	loginVisitor.JettonAreaAndSocre = map[int32]float32{int32(0): float32(1),
		int32(1): float32(1),
		int32(2): float32(1),
		int32(3): float32(1),
		int32(4): float32(1),
		int32(5): float32(1),
		int32(6): float32(1),
		int32(7): float32(1)}
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

//起立
func compare() []byte {
	loginVisitor := new(gamemsg.Game_C_UserCompare)
	loginVisitor.JettonArea = 1
	loginVisitor.JettonSocre = 20
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x03, data)
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		//sitDown := userJettonUpMsg()
		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}
		//log.Println("login:", login, n)
		//下注
		//time.Sleep(1 * time.Second)
		//_, err = conn.Write(sitDown)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}

		break

	}
}

//Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_GameResult{})    //铃铛结果
//Processor.Register(&Game_S_CompareResult{}) //比大小结果
//Processor.Register(&Game_S_UserList{})      //获取用户列表服务器返回
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
		case 5: //请求失败消息
			var msg gamemsg.Game_S_ReqlyFail
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("请求失败消息recvString:\n", string(bytes))
		case 6: //登陆成功消息
			var msg gamemsg.Game_S_LoginSuccess
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("登陆成功消息recvString:\n", string(bytes))
		case 7: //空闲场景消息
			var msg gamemsg.Game_S_GameResult
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("铃铛结果recvString:\n", string(bytes))
		case 8:
			var msg gamemsg.Game_S_CompareResult
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("比大小结果消息recvString:\n", string(bytes))
		case 9:
			var msg gamemsg.Game_S_UserList
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户列表消息recvString:\n", string(bytes))
		}
	}
}
