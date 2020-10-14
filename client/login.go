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
	"xj_game_server/client/msg"
	"time"
)

const (
	addr = "47.107.188.43:8000"
	//addr = "47.56.172.167:8000"
	//addr = "192.168.1.43:8000"
	//addr = "192.168.1.43:8000"
	//addr = "47.107.188.43:8000"
	//addr = "127.0.0.1:8000"
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

//Processor.Register(&Login_C_Wechat{})  //微信登陆消息
//Processor.Register(&Login_C_Mobile{})  //手机登陆消息
//Processor.Register(&Login_C_Visitor{}) //游客登陆消息
//
//Processor.Register(&Login_S_Success{}) //登陆成功消息
//Processor.Register(&Login_S_Fail{})    //登陆失败消息
func loginMsg() []byte {
	loginVisitor := new(msg.Login_C_Visitor)
	//8bced100f3a8c2413a0f3ab6d7c7ba9367efd92d
	loginVisitor.MachineID = "8bced100f3a8c2413a0f3ab6d7c7ba9367efd92d"
	loginVisitor.DeviceType = 7
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		fmt.Println(login)
		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}
		break
		time.Sleep(10 * time.Millisecond)
	}
}

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
		case 3: //注册请求失败消息
			var msg msg.Login_S_Success
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			bytes1, _ := json.Marshal(msg.GameInfoList)
			//util.JsonFmt(msg)
			log.Println("登陆成功消息recvString:\n", string(bytes))
			log.Println("登陆成功消息recvString:\n", string(bytes1))
		case 4: //注册登陆成功消息
			var msg msg.Login_S_Fail
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("登陆失败消息recvString:\n", string(bytes))
		default:
			log.Println("无效命令", cmd)
		}

	}
}
