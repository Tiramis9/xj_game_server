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
	gamemsg "xj_game_server/client/xt/xtmsg"
)

const (
	//addr = "47.56.172.167:2010"
	//addr = "192.168.1.43:4020"
	//addr = "47.107.188.43:4020"
	//addr = "127.0.0.1:4020"
	addr = "47.107.188.43:15000"
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
	loginVisitor := new(gamemsg.Hall_C_Msg)
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg2NDMiLCJuYmYiOjE1ODQ1ODYyNDd9.kbpeCmBzu-m6gEIWrz_8wr-Mi9-ImE3iskAcgbKfn2GoJHu1A8ZIUPvRTzjX4S4FUvi8Kyz91dhmm8tLNeEcmH-0cQIZM9CLg71wh6Mqdd6VcP-CAMriugzvFzG5QDOOz9fsyS5NjIbjNA0EEIm5i1PMLK_cnV5Zw6iBY6SwBKDSwpnS67xgVpoVF5MPwXzJFS0gmz3KfUkf9Mn6YhhQrc4iZLEYeA5vDm8Cf7mEexwPB4itUbomNn3zVDuOhNeWnPLAJcoFEvxDTLmkH3dAqdqlmR7-e7ggU9Lq6P9Zt-j-MM6EGZOGod74UyUth22G0kgDe6JewrtBhXNi-aiWlA"
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

var isLogin = true

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		if isLogin {
			_, err := conn.Write(login)
			if err != nil {
				log.Println("write:", err)
				return
			}

			isLogin = false
		}

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
		case 1: //心跳
			var msg gamemsg.Hall_S_Msg
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("心跳recvString:\n", string(bytes))
		case 2: //请求失败消息
			var msg gamemsg.Hall_S_Fail
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("请求失败消息recvString:\n", string(bytes))
		case 3: //充值通知
			var msg gamemsg.Hall_Recharge_Notice
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("充值通知recvString:\n", string(bytes))
		}
	}
}
