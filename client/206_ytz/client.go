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
	gamemsg "xj_game_server/client/206_ytz/msg"
	rand "xj_game_server/util/leaf/util"
	"time"
)

const (
	addr = "47.107.188.43:2060"
	//addr = "127.0.0.1:2060"
	//addr = "47.107.188.43:13001"
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
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg2NTIiLCJuYmYiOjE1OTExNjYzMTh9.XpcvJPRrpSWDqUN-iYdxae8onf42qW90bLwRLHaT813N61hMZS6k8l9fFHM_g8uiAB1V2EudEyP_7CLV2BCnFfUU6LIjvbWwHQ6vzHJf4uZ83akd29Wq4qVcINs25ZpGu3_zblQZCUzTtJvQ9_lzFDp8GR7yPJif3xokwxUQZMYCl-VnGv0OuUJfpb8oCr29PRTn7tp8VGfn2ytxCMB4bDdPX8o1PKO3du2unqy5D11OFyuyHENNaJhr5LttBiXOISE2ao1PWGh79o_e0X_AuOy9K5nw7odesH8KuKfS3fYfs56ygs7u2EL7ERidyO9Q-9lvqiiNfqKf0DOOCMJmBw"
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

func jgMsg(ChairID, Multiple, Point int32, isZai bool) []byte {
	loginVisitor := gamemsg.Game_C_UserHP{
		Type:       0,
		ChairID:    ChairID,
		Multiple:   Multiple,
		Point:      Point,
		NetChairID: 0,
		IsZai:      isZai,
	}

	data, _ := proto.Marshal(&loginVisitor)
	return CreteCmd(0x03, data)
}

//起立
func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

//下注
func userJettonUpMsg() []byte {
	//int32 Type  = 1;            //类型0 叫点  1 指定人叫点 3 反转
	//int32 ChairID = 2;          //椅子号
	//int32 Multiple = 3;         //多少个
	//int32 Point = 4;            //点数
	//int32 NetChairID = 5;       //指定下个叫点人
	//bool IsReverse = 6;         //是否反转
	//bool IsZai = 7;             //是否栽
	loginVisitor := new(gamemsg.Game_C_UserHP)
	loginVisitor.Type = 0
	loginVisitor.Multiple = 2
	loginVisitor.Point = 5
	data, _ := proto.Marshal(loginVisitor)
	return CreteCmd(0x04, data)
}

func sender(conn net.Conn) {
	var isDown = false

	for {
		for i := 0; i < 10; i++ {
			login := loginMsg()

			_, err := conn.Write(login)
			if err != nil {
				log.Println("write:", err)
				return
			}

			if !isDown {
				sitDown := sitDownMsg()
				//log.Println("login:", login, n)
				//坐下
				time.Sleep(1 * time.Second)
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
				isDown = true
			}
		}
		break
	}
}

//Processor.Register(&Game_C_TokenLogin{})        //token登陆消息
//Processor.Register(&Game_C_UserSitDown{})       //用户坐下消息
//Processor.Register(&Game_C_UserStandUp{})       //用户起立消息
//Processor.Register(&Game_C_UserHP{})            //用户叫骰
//Processor.Register(&Game_C_UserKP{})            //用户开牌
//Processor.Register(&Game_C_UserP{})             //用户劈
//Processor.Register(&Game_C_UserFP{})            //用户反劈
//Processor.Register(&Game_C_UserQP{})            //用户弃牌消息
//
// 服务端-----
//Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_FreeScene{})     //空闲场景消息
//Processor.Register(&Game_S_JettonScene{})   //下注场景
//Processor.Register(&Game_S_LotteryScene{})  //开奖场景
//Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息
//Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息
//Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息
//Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息
//Processor.Register(&Game_S_StartTime{})     //开始定时器通知消息
//Processor.Register(&Game_S_GameConclude{})  //结束游戏通知消息
//Processor.Register(&Game_S_PNotify{})       //劈通知
//Processor.Register(&Game_S_FPNotify{})      //反劈通知
//Processor.Register(&Game_S_UserJetton{})    //叫骰通知
//Processor.Register(&Game_S_UserQP{})        //弃牌通知
////机器人-----
//Processor.Register(&Game_C_RobotLogin{}) //机器人登陆
func read(conn net.Conn) {
	var chid int32
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
			//记录我的椅子号
			chid = msg.UserChairID
			log.Print("空闲场景消息recvString:\n", string(bytes))
		case 11:
			var msg gamemsg.Game_S_JettonScene
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注场景消息recvString:\n", string(bytes))
		case 12:
			var msg gamemsg.Game_S_PScene
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("结束场景消息recvString:\n", string(bytes))
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
			if msg.CurrentChairID == chid {
				sitDown := jgMsg(chid, 5, 2, false)
				time.Sleep(5 * time.Second)
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			log.Print("开始游戏通知消息17recvString:\n", string(bytes))
		case 18:
			var msg gamemsg.Game_S_GameConclude
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("结束游戏通知消息18recvString:\n", string(bytes))
		case 19:
			var msg gamemsg.Game_S_PNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("抢庄通知消息19recvString:\n", string(bytes))
		case 20:
			var msg gamemsg.Game_S_FPNotify
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("反劈通知消息20recvString:\n", string(bytes))
		case 21:
			var msg gamemsg.Game_S_UserJetton
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			if msg.NetChairID == chid && msg.ChairID != chid {
				//当前我要叫号
				if msg.Point == 1 {
					msg.Multiple = msg.Multiple + 1
					msg.Point = 1
					msg.IsZai = true
				}
				if msg.Point+1 > 6 {
					msg.Point = 1
					msg.IsZai = true
				} else {
					msg.Point = msg.Point + 1
				}
				sitDown := jgMsg(chid, msg.Multiple, msg.Point, msg.IsZai)
				time.Sleep(time.Duration(rand.RandInterval(2, 5)) * time.Second)
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			log.Print("叫骰通知消息21recvString:\n", string(bytes))
		case 22:
			var msg gamemsg.Game_S_UserQP
			err = proto.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("弃牌通知消息22recvString:\n", string(bytes))
			//case 23:
			//	var msg gamemsg.Game_S_UserQP
			//	err = proto.Unmarshal(message[4:], &msg)
			//	bytes, _ := json.Marshal(msg)
			//	log.Print("弃牌通知消息23recvString:\n", string(bytes))
			//case 24:
			//	var msg gamemsg.Game_S_StartJetton
			//	err = proto.Unmarshal(message[4:], &msg)
			//	log.Print("通知下注通知消息23recvString:\n", msg)
		}

	}
}

// 服务端-----
//Processor.Register(&Game_S_ReqlyFail{})     //请求失败消息8
//Processor.Register(&Game_S_LoginSuccess{})  //登陆成功消息
//Processor.Register(&Game_S_FreeScene{})     //空闲场景消息10
//Processor.Register(&Game_S_JettonScene{})   //下注场景消息
//Processor.Register(&Game_S_TPScene{})       //摊牌场景消息
//Processor.Register(&Game_S_OnLineNotify{})  //用户上线通知消息13
//Processor.Register(&Game_S_OffLineNotify{}) //用户掉线通知消息

//Processor.Register(&Game_S_StandUpNotify{}) //起立通知消息15
//Processor.Register(&Game_S_SitDownNotify{}) //坐下通知消息16

//Processor.Register(&Game_S_StartTime{})     //开始定时器通知消息17
//Processor.Register(&Game_S_GameConclude{})  //结束游戏通知消息18

//Processor.Register(&Game_S_UserKP{})        //抢庄通知消息
//Processor.Register(&Game_S_UserJetton{})    //下注通知消息
//Processor.Register(&Game_S_UserTP{})        //摊牌通知消息
//Processor.Register(&Game_S_GameBP{})        //比牌通知消息
//Processor.Register(&Game_S_UserQP{})        //弃牌通知消息
