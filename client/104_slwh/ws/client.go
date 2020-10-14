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
	gamemsg "xj_game_server/client/104_slwh/msg"
	"time"

	"github.com/gorilla/websocket"
)

//wss://mainnet.eos.dfuse.io/v1/stream?token=eyJ..YOURTOKENHERE...
//var addr = flag.String("addr", "47.75.218.79:8205", "http service address")
var addr = flag.String("addr", "47.107.188.43:10470", "http service address")

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
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "88888888"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODUyODEyNjMsImlzcyI6IjQiLCJuYmYiOjE1ODI2ODkyNjN9.K96z5Yb_qPEazrkNVhdzPep544QgOOiKy8dE4X5K75THA4718Qb8q3PlouG_ttiSNYZGAGQjr3SExquEI3tJA9guWkEGupWNykqgiDDurlQ4CaE96QZbefKyyatnhZTlgie96ISy9BkoszyxZ_fdypMvgX8_Q4vNinJso8JIg9Psejo99sMeqNK2VBdz0Gei0J5OLOR0gqNcE7z3JT29tSEA2sZuSz71Smr3T-9iPfqQppGcuKfEtxMAeBYYo9lrsWikITiNBu7EYRUPjDTX9yrvrkxJtq3ORTEVqRmoLU5Jx80yf6L_BWlYzXoI7zXE8qPe6g9RbUrvsRGtdNaG2Q"
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

func sitDownMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserSitDown)
	loginVisitor.ChairID = 0
	loginVisitor.TableID = 0
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x01, data)
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
			case 4: //注册请求失败消息
				var msg gamemsg.Game_S_ReqlyFail
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求失败消息recvString:\n", string(bytes))
			case 5: //注册登陆成功消息
				var msg gamemsg.Game_S_LoginSuccess
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("登陆成功消息recvString:\n", string(bytes))
			case 6: //注册下注场景消息
				var msg gamemsg.Game_S_JettonScene
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("下注场景消息recvString:\n", string(bytes))
			case 7: //注册开奖场景消息
				var msg gamemsg.Game_S_LotteryScene
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("开奖场景消息recvString:\n", string(bytes))
			case 8: //注册游戏开始消息
				var msg gamemsg.Game_S_GameStart
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("游戏开始消息recvString:\n", string(bytes))
			case 9: //游戏结束消息
				var msg gamemsg.Game_S_GameConclude
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("游戏结束消息recvString:\n", string(bytes))
			case 10: //用户上线通知消息
				var msg gamemsg.Game_S_OnLineNotify
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("用户上线消息recvString:\n", string(bytes))
			case 11: //用户掉线通知消息
				var msg gamemsg.Game_S_OffLineNotify
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("用户掉线消息recvString:\n", string(bytes))
			case 12: //起立通知消息
				var msg gamemsg.Game_S_StandUpNotify
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("起立通知消息recvString:\n", string(bytes))
			case 13: //坐下通知消息
				var msg gamemsg.Game_S_SitDownNotify
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("坐下通知消息recvString:\n", string(bytes))
			case 14: //下注通知
				var msg gamemsg.Game_S_UserJetton
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("下注通知消息recvString:\n", string(bytes))
			case 16: //当前下注状况,每个区域,每个玩家的下注情况
				var msg gamemsg.Game_S_AreaJetton
				err = proto.Unmarshal(message[4:], &msg)
				log.Print("每个区域,每个玩家的下注情况recvString:\n", msg)
			case 18: //获取用户列表
				var msg gamemsg.Game_S_UserList
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("获取用户列表recvString:\n", string(bytes))
			default:
				log.Println("无效命令", cmd)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
	var login bool
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if !login {
				err := c.WriteMessage(websocket.BinaryMessage, loginMsg())
				if err != nil {
					log.Println("write:", err)
					return
				}
				login = true
			}

			err = c.WriteMessage(websocket.BinaryMessage, sitDownMsg())
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
