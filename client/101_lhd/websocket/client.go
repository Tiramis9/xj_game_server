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
	gamemsg "xj_game_server/client/101_lhd/101msg"
	"time"

	"github.com/gorilla/websocket"
)

//wss://mainnet.eos.dfuse.io/v1/stream?token=eyJ..YOURTOKENHERE...
//var addr = flag.String("addr", "47.75.218.79:8205", "http service address")
var addr = flag.String("addr", "47.107.188.43:10170", "http service address")

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
	msg = append(msg, data...) //数据
	return msg

}

func loginMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "8bced100f3a8c2413a0f3ab6d7c7ba9367efd92d"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg2MDkiLCJuYmYiOjE1ODIwMTQyMjh9.dqHrzpcBaVWl0QUWR6zz9UXOAH80aymuvdRBrgU6IPe6F70VAk-ZXHwjz-Emhkqj9fFddD6YPPfQ3c7D8MsbHcvzJgqlAK3TyPPe4X6QAT8kzvmHCNvIrfUKkh0WH4N-MJQBM7NjgS3zsFynbvSZ8CLC2wThf1FgwgtOhnWko6Vhjp8-ahdPA25DlTx_K6uQq-2uOggelPW9332_o0cHVwoU9sF5OK5T96Mb23q4g7MI3B7o9JHLaohaB--i-U20slnxwqgotcFQni8YdukTr4UfZ7qcO4u8j41RF1XPw5GNyyg56CJjN0s9wQP3ECgYYHf2i_uzRYZfjD2Kpf6Y3g"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiI0NCIsIm5iZiI6MTU3NzA5NTQ2OH0.ipk7z2Z_zV8jvyDfJg0CXxMpSJD3zGKW3WJJy_y3iJW2Sd_sGu10MzOm9-7xpBZg6AtZXs9AZIiR7GjyUXypE7qz0MJuYQsaeodOg6r-W9MKZZLzcJLqAtVDyzw-zGBhPxdVP51QUy9Y8VAJGw2-p54FTDj-BLFCtx6Uke_bwDj2LREl__X89FCuWKqeoSDNrCHMJPzCCWfX0ygSL71SFIqIhUanAI39ptrmy3hAHl3Noj48HtwCFNaqq96iHKiF_KA_RZIyOz8rG9OTJNXmBIkMwRQUBpjqpZFZ2fXPBGqalxYtAb_AhO9exIUmB7gH6M8A98g3dAU5htbvgu062w"
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

func sitDownMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserSitDown)
	loginVisitor.ChairID = 0
	loginVisitor.TableID = 1
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x01, data)
}

func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

func userJettonUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	loginVisitor.JettonArea = 0
	loginVisitor.JettonScore = 1
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x03, data)
}

func getUserList() []byte {
	userList := new(gamemsg.Game_C_UserList)
	userList.Page = 1
	userList.Size = 20
	data, _ := proto.Marshal(userList)
	return CreteCmd(0x12, data)
}

//整形转换成字节  大端模式   高位在前
func IntToBytes(n interface{}) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes(), err
}

var isSitDown = true
var int1 = 0

func createClient() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: ""}
	fmt.Printf("connecting to %s", u.String())

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
			var cmd int16
			err = BytesToInt(&cmd, message[0:2])
			if err != nil {
				log.Printf("cmd 错误 err:%d\n", cmd)
				return
			}
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

				if msg.Status == 2 {
					isSitDown = false
				}
				log.Print("登陆成功消息recvString:\n", string(bytes), len(message))
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
			case 15: //游戏结束大厅场景消息
				var msg gamemsg.Game_S_Hall
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Println("消息长度", len(message))
				log.Print("游戏结束大厅场景消息recvString:\n", string(bytes))
			case 17: //当前下注状况,每个区域,每个玩家的下注情况
				var msg gamemsg.Game_S_AreaJetton
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("每个区域,每个玩家的下注情况recvString:\n", string(bytes))
			case 19: //获取用户列表
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

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ticker.Stop()
			login := loginMsg()
			sitDown := sitDownMsg()

			if int1 == 0 {
				err := c.WriteMessage(websocket.TextMessage, login)
				if err != nil {
					log.Println("write:", err)
					return
				}
				int1++
			}

			//log.Println("login:", login, n)
			//坐下
			time.Sleep(1 * time.Second)
			if isSitDown {
				err := c.WriteMessage(websocket.TextMessage, sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}

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
