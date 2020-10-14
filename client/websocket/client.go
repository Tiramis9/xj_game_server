package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"log"
	"net/url"
	"os"
	"os/signal"
	"xj_game_server/server/login/msg"
	"time"

	"github.com/gorilla/websocket"
)

//wss://mainnet.eos.dfuse.io/v1/stream?token=eyJ..YOURTOKENHERE...
//var addr = flag.String("addr", "47.75.218.79:8205", "http service address")
var addr = flag.String("addr", "47.107.188.43:8001", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)
	createClient()
	//for i := 0; i < 1000000; i++ {
	//	go createClient()
	//	time.Sleep(100 * time.Millisecond)
	//}
	//time.Sleep(100 * time.Hour)
}

func loginMsg() []byte {
	loginVisitor := new(msg.Login_C_Visitor)
	//8bced100f3a8c2413a0f3ab6d7c7ba9367efd92d
	loginVisitor.MachineID = "8bced100f3a8c2413a0f3ab6d7c7ba9367efd92d"
	loginVisitor.DeviceType = 7
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

//字节转换成整形
func BytesToInt(n interface{}, b []byte) error {
	bytesBuffer := bytes.NewBuffer(b)
	err := binary.Read(bytesBuffer, binary.BigEndian, n)
	return err
}
func CreteCmd(cmd int16, data []byte) []byte {
	var msg = make([]byte, 0)
	//var len = len(data)
	//lenByte, _ := IntToBytes(int16(len) + 2)
	cmdByte, _ := IntToBytes(cmd)
	msg = append(msg, cmdByte...) //命令
	//msg = append(msg, lenByte...) //长度
	msg = append(msg, data...)    //数据
	return msg

}

//整形转换成字节  大端模式   高位在前
func IntToBytes(n interface{}) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes(), err
}
func createClient() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: ""}
	fmt.Printf("connecting to %s \n", u.String())

	//s2 := rand.NewSource(time.Now().UnixNano()) //同前面一样的种子
	//r2 := rand.New(s2)
	//sj := utils.GetMd5([]byte(strconv.Itoa(r2.Intn(10000))))
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			//log.Printf("recv: %v", message)
			var cmd int16
			err = BytesToInt(&cmd, message[0:2])
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
	}()

	ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.BinaryMessage, loginMsg())
			if err != nil {
				log.Println("write:", err)
				return
			}
			ticker.Stop()
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bay bay"))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
