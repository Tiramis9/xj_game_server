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
	gamemsg "xj_game_server/game/201_qiangzhuangniuniu/msg"
	gameJson "xj_game_server/util/leaf/network/json"
	"xj_game_server/util/leaf/util"
)

const (
	//addr = "47.56.172.167:2010"
	addr = "47.113.94.16:20180"
	//addr = "192.168.0.105:20180"
	//addr = "127.0.0.1:20180"
	//addr = "47.56.172.167:13001"
)

var jsonEn = gameJson.NewProcessor()

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

/*
// 对应游戏服务器 Hello 消息结构体
	data := []byte(`{
		"Hello": {
			"Name": "leaf"
		}
	}`)

	// len + data
	m := make([]byte, 2+len(data))

	// 默认使用大端序
	binary.BigEndian.PutUint16(m, uint16(len(data)))

	copy(m[2:], data)
*/
func CreteCmd1(data []byte) []byte {
	//m := make([]byte, 2+len(data))
	//binary.BigEndian.PutUint16(m, uint16(len(data)))
	//copy(m[2:], data)
	//fmt.Printf("data:%s\n", msg)
	//return m
	var msg = make([]byte, 0)
	var len = len(data)
	lenByte, _ := IntToBytes(int16(len))
	msg = append(msg, lenByte...) //长度
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
	//loginVisitor.MachineID = "88888888"
	//loginVisitor.MachineID = "2593f5c67cb883453083844c5cc7394eb31f9220"
	loginVisitor.MachineID = "51676087823dfaa801ce203dac0970522"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3MjIiLCJuYmYiOjE1OTUwNDE5NTZ9.sKlaj-E1pHifVOJ4DPo9aNJNbkSMz4yCS1_orwou87S3OXn6C6dpLGry9eBKPbvfyrtiHIHTEz8HnMnKbdcZ6z_1qkTB8n2sCqfFA08TUaiaMV3gKAD3ujZYjpD49b2JigGaU2fw8gmKi9r2CsdIVmCUqGMAureWn0XS6vfXAFKbM9JHInG2DHRfQ5FwU8bKKxUaCuZbT44hbXIIcwrV9iC8ssiOqUQALf1U21FcNIqwpvzcwtXY1_pN0Lzm5x-qaJrSofd9JToEQ8P_M-4ok5K9P1v9XvxKtgei-ZMnb_LN6o8pXlYt-nLN3yghV_S4Einyp-Bos0yLvXnh_4oolA"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDcyIiwibmJmIjoxNTk1NTc2NjIyfQ.D5LgEqBAA8EvaXrPMu1npiFmDMJZuDobnFykbz-fn3CkGuav1bRrMO9Gf5BqA_Q9tfWaBd9Hhm1-gu6C72NScXeJF3t11PggG0Nf_dMDQwJfgg2oMphFivyIW1Dlkt3dFmADPFHJT5Yyf794_7YHZnq52AwofW79X3ZF1FcEG_jwcF6AtfuG_Ofj9fhUhDaorclzySCKFnfQRz4lFZHM3XJOmqdfAPOuVw2QZXCP-PWxEYFAHd8HJO5p9cX6eB6JJAUDBf1lf_KM4CJLB5-kRUvU_5W9hZ3uDBhlRinYJ7XeMdoXfbGCi10rok34BqT_4m12rQ5XwoXM48sxVGmhbQ"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDM0IiwibmJmIjoxNTk1NTAzOTMzfQ.vd_9FA7a3-1SH-Llo5z__ZrcOBfT-SLCXHIDsd5vzkwO5ccyPxYVkjBtQ5IbVRQTvoXVaAAH5jEYdmyI1rgupW1cMzW6dLCsucLAjMKbpq_yvRHw6yWO5Q05GKMdrJ360vO3v1vhPjNvG2z3E0SSk0MbHZtmN1k8gCTiS_CaWY9dykuv4ZnXo18pDppwwjGTjxp-Nzrcf1uynyitUo-3jlpblZiyaTyRBpbKYjozH3hYNChewz89vjNCYlrcB2cRmGMgUpl7zYR3liijxHhCaQBUT5pboP2I2aQPVPRZl9eZz-FvW54_C24kMoa3AGw_sKyack013tNx1lgqaC1xXw"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwNDE1IiwibmJmIjoxNTk2MDgzNjQ5fQ.Er9N1I0FsQ_q4IMbP-pp1lyfIsU6Uac80NsOPlTcSP2xnh0UpMKirPF0E5qaJ9ne7wtDuIFDyhMTyKkiKJTkW3tovDRU5ijLUKDeCWDapCbcf8ELJBh6eyPGK2XRLHqpjOPqZh2ZXPITuHSatXiZI9drX2zUPQEXnLNU1an4rcwdQ3G1yql48z2lyY_RL7jSuvgdWivOv0yzJ33UtmiWvxTf_OBkKrdQAyzHU3tlARuxDV2kBmDs6pw8t20cShjeZRdW1OwjOS_rKfB83lgMOSXEjZAkHay03WX0hsPDLofVJc3fxPw5SatzWz0a_T6OjB022OuhnAZPUXm4Mm7UCg"

	//本地105 138766
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NjYiLCJuYmYiOjE1OTQ4OTQwNDh9.ZCOWA36SomAcP7q4qxPibpW9h-LNTfarMz7LvhfvrlwYksFO34NuS87YD05VTxPqV4cvFGFEjR-3UqOlsMqiapf9KZVzAbfE1-REYYGaRMfw-TmL8tEP4K0p9n2Gt418csraRRI0ZKxCZz2iwbg7XR2w65roeMoDIzmahV9_qurMnRSxKjIpdhACSu5gw067wS_enU9n2SlpYpai-viOf3qfnlf2Z4gaVtp0QO6aw_SOiIw8AljXzpBZBW4REN-DVz8SvVD84LHemitoJ9cwwvUeo9RwkBcAsgD4gGY_H6AWgEWJcewZxYW2TnL95JrPYgMdjnpjMjSfJdYtUCg_VQ"
	// 阿里云 1000886
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwODg2IiwibmJmIjoxNTk3MTE5MDE1fQ.UD6uCb1vcXQHApAxsD-Pm6IMH6scnqm_GPonxgyPh2zTFiSwkFCJqgrZzQWsUgMJJ_VQdQUvvq-zDnEYE6gyr17MiVmeZbohJxcap5ordwUKOePsQdcOa8gbMjBC4_-TgK0PDuy7rIdwYAFcaSSkNjJ4LxpbM2d35ZzwKD_lxHFa4ho5h9tPPXkRkgNb35mqYR8Q4ak84Bueds4sEZPibCaAUZP4QlX2Id-mOnjxjymuIzlvSCZeOsi80bm3kNunQC5PSyi-2bZM-3tvTl5Jej-wKHELbyp-egTOLM7bMrBMdEPyaN44L11XMQHkCdAh7mZ20UKkOadB79BsLcDzwQ"
	//data, _ := json.Marshal(loginVisitor)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

// 登录
func loginMsg1() []byte {
	loginVisitor := new(gamemsg.Game_C_TokenLogin)
	loginVisitor.MachineID = "88888888"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg2OTQiLCJuYmYiOjE1OTQxODAwODJ9.W5_UOSnhLnQtKARtXH6o9HMppEjaDEofjE4fW0WJ2VpVv0B37rLCecuWcWv3TtjoCIuPyaGvsV7KyOP_pn3VP6kyBFONObQ_k3BMEtkXKVekCRsSNdGScuF3aOE-CH8OjRw_kgJLSqs6fqw8CYRfIKh_oarjvBz1yXbIo7cRv-q01n1r2LIk4CJpmQezPz_ToiqhRWoVdlfI-JrrI9IrIbwyMyQz-l13XL-Uo-J-2OxBcewjYLaXHe5mdiRwBb9H7v6RpS7MvwC_ctMFdYLkO49SI1-XDx_QNXNQL0UiCjutd34PBVPkAr0B11s4SNl7VEBFvpWdPHJVI2OzUQ542A"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NzUiLCJuYmYiOjE1OTQ4OTUxOTV9.Jf_zUM3bj-jH_gymu-7M383_Q-6UfWj_rjOVbfLTgbypUYkIfc4xIej-0DpGgxNwlqGwLqeeBDvAVjXoglpxldwN3gDMWycpOlxW-cv7sr7vJUd6c33iHb-I9LCU7hbJ9PvKUIQOeXzfxNyyGL2v-29ptXay4kytBYTEdPOR9njk5EK0QPSAQio5FDv53y9J0JejplXSU1ZA6oURuoggRe-Yj6fd-gBy2GRiJeNzgWlFOeLR4haf5WuUST7UMRsmLiXYsytgwynIxe65LLcHdpvk7T6qvHfdHlB9ANQusx2HQpj8u7Uu_GpPY7yvNqZ5vD3k4L1QkV0a7BkA3fCVVw"
	data, _ := json.Marshal(loginVisitor)
	return CreteCmd(0x00, data)
}

//坐下
func sitDownMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserSitDown)
	//loginVisitor.ChairID = 0
	//loginVisitor.TableID = 0
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

//起立
func standUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserStandUp)
	data, _ := json.Marshal(loginVisitor)
	return CreteCmd(0x02, data)
}

//抢庄
func userQzMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserQZ)
	loginVisitor.Multiple = 40
	data, _ := json.Marshal(loginVisitor)
	return CreteCmd(0x03, data)
}

//下注
func userJettonUpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	var multiple = []int32{1, 2, 5, 10, 15}
	loginVisitor.Multiple = multiple[util.RandInterval(0, int32(len(multiple)-1))]
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

//下注
func userJettonUpMsg1() []byte {
	loginVisitor := new(gamemsg.Game_C_UserJetton)
	var multiple = []int32{1, 2, 5, 10, 15}
	loginVisitor.Multiple = multiple[util.RandInterval(0, int32(len(multiple)-1))]
	data, _ := json.Marshal(loginVisitor)
	return CreteCmd(0x04, data)
}

//摊牌
func userTpMsg() []byte {
	loginVisitor := new(gamemsg.Game_C_UserTP)
	data, _ := json.Marshal(loginVisitor)
	return CreteCmd(0x05, data)
}

//func getUserList() []byte {
//	userList := new(gamemsg.Game_C_UserList)
//	userList.Page = 1
//	userList.Size = 20
//	data, _ := json.Marshal(userList)
//	return CreteCmd(0x12, data)
//}

func sender(conn net.Conn) {
	for {
		login := loginMsg()

		fmt.Println("login:", login)
		//login := loginMsg1()
		//sitDown := sitDownMsg()
		//standUp := standUpMsg()
		//userJetton := userJettonUpMsg()
		//userList := getUserList()
		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}
		fmt.Println("login:", string(login))

		//坐下
		time.Sleep(1 * time.Second)
		//_, err = conn.Write(sitDown)
		//if err != nil {
		//	fmt.Println("write:", err)
		//	return
		//}

		//离开
		//time.Sleep(3 * time.Second)
		//_, err = conn.Write(standUp)
		//if err != nil {
		//	log.Println("write:", err)
		//	return
		//}
		//log.Println("写数据:", standUp, n)

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
		//// 去掉字节 0
		//index := bytes.IndexByte(message, 0)
		//message = message[:index]

		message = message[:n]
		if len(message) == 0 {
			break
		}
		//fmt.Println("msg:", string(message))

		var m = make(map[string]json.RawMessage)
		err = json.Unmarshal(message[2:], &m)
		if err != nil {
			log.Printf("Unmarshal 错误 err:%s\n", err)
			continue
			// return
		}
		if len(m) != 1 {
			log.Printf("message 错误 err:%v\n", m)
			continue
			// return
		}

		for msgID, data := range m {
			//fmt.Printf("%v\n", msgID)
			switch msgID {
			case "Game_S_ReqlyFail":
				var msg gamemsg.Game_S_ReqlyFail
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求失败消息recvString:\n", string(bytes))
			case "Game_S_LoginSuccess":
				var msg gamemsg.Game_S_LoginSuccess
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_LoginSuccess 登陆成功消息recvString:\n", string(bytes))

				sitDown := sitDownMsg() //坐下
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
				time.Sleep(time.Millisecond * 100)
				sitDown1 := sitDownMsg() //坐下
				_, err1 := conn.Write(sitDown1)
				if err1 != nil {
					log.Println("write:", err)
					return
				}

			case "Game_S_FreeScene": //空闲场景消息
				var msg gamemsg.Game_S_FreeScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("空闲场景消息recvString:\n", string(bytes))
			case "Game_S_QZScene": //抢庄场景消息
				var msg gamemsg.Game_S_QZScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("抢庄场景消息recvString:\n", string(bytes), err)
			case "Game_S_JettonScene": //下注场景消息
				var msg gamemsg.Game_S_JettonScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("下注场景消息recvString:\n", string(bytes))
			case "Game_S_TPScene": //摊牌场景消息
				var msg gamemsg.Game_S_TPScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("摊牌场景消息recvString:\n", string(bytes))
			case "Game_S_OnLineNotify": //用户上线通知消息
				var msg gamemsg.Game_S_OnLineNotify
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("用户上线消息recvString:\n", string(bytes))
			case "Game_S_OffLineNotify": //用户掉线通知消息
				var msg gamemsg.Game_S_OffLineNotify
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("用户掉线消息recvString:\n", string(bytes))
			case "Game_S_StandUpNotify": //起立通知消息
				var msg gamemsg.Game_S_StandUpNotify
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("起立通知消息recvString:\n", string(bytes))
			case "Game_S_SitDownNotify": //坐下通知消息
				var msg gamemsg.Game_S_SitDownNotify
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_SitDownNotify 坐下通知消息recvString:\n", string(bytes))

			case "Game_S_CardRound": //发牌环节
				var msg gamemsg.Game_S_CardRound
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_CardRound recvString:\n", string(bytes))
			//switch msg.GameStatus {
			//case global.GameStatusFree:
			//	fmt.Println("空闲倒计时开始：", string(bytes))
			//case global.GameStatusQZ:
			//	fmt.Println("抢庄倒计时开始：", string(bytes))
			//case global.GameStatusJetton:
			//	fmt.Println("下注倒计时开始：", string(bytes))
			//case global.GameStatusTP:
			//	fmt.Println("摊牌倒计时开始：", string(bytes))
			//}
			case "Game_S_ShowRound": //摊牌环节
				var msg gamemsg.Game_S_ShowRound
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_ShowRound recvString:\n", string(bytes))
			case "Game_S_CallRound": //抢庄环节
				var msg gamemsg.Game_S_CallRound
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_CallRound recvString:\n", string(bytes))
			case "Game_S_BetRound": //叫倍环节
				var msg gamemsg.Game_S_BetRound
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_BetRound recvString:\n", string(bytes))

				time.Sleep(time.Second * 1)
				_, err1 := conn.Write(userJettonUpMsg())
				if err1 != nil {
					log.Println("write:", err)
					return
				}
			case "Game_S_GameConclude": //结束游戏通知消息
				var msg gamemsg.Game_S_GameConclude
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_GameConclude 结束游戏通知消息recvString:\n", string(bytes))

				time.Sleep(time.Second * 1)
				// 循环测试
				_, err := conn.Write(sitDownMsg())
				if err != nil {
					log.Println("write:", err)
					return
				}
			case "Game_S_UserQZ": //抢庄通知
				var msg gamemsg.Game_S_UserQZ
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_UserQZ 抢庄通知recvString:\n", string(bytes))
			case "Game_S_GameDZ": //定庄通知消息
				var msg gamemsg.Game_S_GameDZ
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_GameDZ 定庄通知消息recvString:\n", string(bytes))

			case "Game_S_UserJetton": //下注通知消息
				var msg gamemsg.Game_S_UserJetton
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_UserJetton 下注通知消息recvString:\n", string(bytes))
			case "Game_S_UserTP": //摊牌通知
				var msg gamemsg.Game_S_UserTP
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("Game_S_UserTP 摊牌通知recvString:\n", string(bytes))

			default:
				log.Println("无效命令", msgID)
			}
		}
	}
}

func readbak(conn net.Conn) {
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
		case 6: //请求失败消息
			var msg gamemsg.Game_S_ReqlyFail
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("请求失败消息recvString:\n", string(bytes))
		case 7: //登陆成功消息
			var msg gamemsg.Game_S_LoginSuccess
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("登陆成功消息recvString:\n", string(bytes))
		case 8: //空闲场景消息
			var msg gamemsg.Game_S_FreeScene
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("空闲场景消息recvString:\n", string(bytes))
		case 9: //抢庄场景消息
			var msg gamemsg.Game_S_QZScene
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("抢庄场景消息recvString:\n", string(bytes), err)
		case 10: //下注场景消息
			var msg gamemsg.Game_S_JettonScene
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注场景消息recvString:\n", string(bytes))
		case 11: //摊牌场景消息
			var msg gamemsg.Game_S_TPScene
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("摊牌场景消息recvString:\n", string(bytes))
		case 12: //用户上线通知消息
			var msg gamemsg.Game_S_OnLineNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户上线消息recvString:\n", string(bytes))
		case 13: //用户掉线通知消息
			var msg gamemsg.Game_S_OffLineNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("用户掉线消息recvString:\n", string(bytes))
		case 14: //起立通知消息
			var msg gamemsg.Game_S_StandUpNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("起立通知消息recvString:\n", string(bytes))
		case 15: //坐下通知消息
			var msg gamemsg.Game_S_SitDownNotify
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("坐下通知消息recvString:\n", string(bytes))
		//case 16: //开始定时器通知消息
		//	var msg gamemsg.Game_S_StartTime
		//	err = json.Unmarshal(message[4:], &msg)
		//	bytes, _ := json.Marshal(msg)
		//	switch msg.GameStatus {
		//	case global.GameStatusFree:
		//		fmt.Println("空闲倒计时开始：", string(bytes))
		//	case global.GameStatusQZ:
		//		fmt.Println("抢庄倒计时开始：", string(bytes))
		//	case global.GameStatusJetton:
		//		fmt.Println("下注倒计时开始：", string(bytes))
		//	case global.GameStatusTP:
		//		fmt.Println("摊牌倒计时开始：", string(bytes))
		//	}
		case 17: //结束游戏通知消息
			var msg gamemsg.Game_S_GameConclude
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("结束游戏通知消息recvString:\n", string(bytes))
		case 18: //抢庄通知
			var msg gamemsg.Game_S_UserQZ
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("抢庄通知recvString:\n", string(bytes))
		case 19: //定庄通知消息
			var msg gamemsg.Game_S_GameDZ
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("定庄通知消息recvString:\n", string(bytes))

		case 20: //下注通知消息
			var msg gamemsg.Game_S_UserJetton
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("下注通知消息recvString:\n", string(bytes))
		case 21: //摊牌通知
			var msg gamemsg.Game_S_UserTP
			err = json.Unmarshal(message[4:], &msg)
			bytes, _ := json.Marshal(msg)
			log.Print("摊牌通知recvString:\n", string(bytes))
		default:
			log.Println("无效命令", cmd)
		}

	}
}
