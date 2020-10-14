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
	gamemsg "xj_game_server/game/203_doudizhu/msg"
	gameJson "xj_game_server/util/leaf/network/json"
	"xj_game_server/util/leaf/util"
)

const (
	//addr = "47.113.94.16:20380"
	addr = "192.168.0.105:20380"
	//addr = "127.0.0.1:20380"

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

	//本地105 138766
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NjYiLCJuYmYiOjE1OTQ4OTQwNDh9.ZCOWA36SomAcP7q4qxPibpW9h-LNTfarMz7LvhfvrlwYksFO34NuS87YD05VTxPqV4cvFGFEjR-3UqOlsMqiapf9KZVzAbfE1-REYYGaRMfw-TmL8tEP4K0p9n2Gt418csraRRI0ZKxCZz2iwbg7XR2w65roeMoDIzmahV9_qurMnRSxKjIpdhACSu5gw067wS_enU9n2SlpYpai-viOf3qfnlf2Z4gaVtp0QO6aw_SOiIw8AljXzpBZBW4REN-DVz8SvVD84LHemitoJ9cwwvUeo9RwkBcAsgD4gGY_H6AWgEWJcewZxYW2TnL95JrPYgMdjnpjMjSfJdYtUCg_VQ"
	// 阿里云 100414
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwNDE0IiwibmJmIjoxNTk2MDM3NjcxfQ.tktPM4wv14x_gpQhp87_E0PyPoJydq5n6SoZdSq_O43t7DLaI7QeZhqr4qtB2ihs5oq9hJ4ATnb5uvvnCwHGjOvtWIuzY0fawK_lucgca0K-Zz-rFDYPhNXs_MObhs7FEsCPKTxAthwH0yOQ6TfztgjGIyJDRyVQD51j4KYwmHqwaJN8PIBEUh-W7ctRkgE0lYU_qmwqPET1YbeRaR6nnDxHA4m2sZnWXimQbOPQiTfzKEk7YMkPSfw7MsxGgyTSEuhcDHLnLRK5eZ-X97W47F5G175btis5MvBwcYObgg58DNkE5ZbreFRbE-pigUDhtwf8rnyx7oy81x8UZ150Rw"
	// 阿里云 1000885
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwODg1IiwibmJmIjoxNTk3MTEzMzkwfQ.SEFFTeLdCgr8Cw6JCNE3KV_BnhvzF4WwZn3nMCj33wvBPG65ue0neBxVcTw6ZwPwx70nfYLh-BdORzYhEJd6YgzuWgV0zR1ZdG--cVFKHQpXWlxWRFgTk-We3UOsL3T7RDtdIoNCkeJkZ6gV-HYLFkbgeALB9WIu4LsyY7aMruGeOiWKOeVQTvmxkiLqlEQgmy3Y0YKETutUL9ushKi6_xlPxBy_XqEbcoG9atwTqjmMrJmF79xt3DIyJQkVxnIbU-kw6_gC750uXF1Q3zWWdx00eWm40CFGjBDyjNvwfyHIddehoMVvxnIWfSwrRoUl1BWaGxYRa_x5bCsqzo09MQ"
	//data, _ := json.Marshal(loginVisitor)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

// 登录
func loginAutoMange() []byte {
	loginVisitor := new(gamemsg.Game_C_AutoManage)
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

//过牌
func userOutPass(charID int32) []byte {
	loginVisitor := new(gamemsg.Game_S_UserPass)
	loginVisitor.ChairID = charID
	//loginVisitor.TableID = 0
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()

		//fmt.Println("login:", login)
		//login := loginMsg1()
		sitDown := sitDownMsg()
		//standUp := standUpMsg()
		//userJetton := userJettonUpMsg()
		//userList := getUserList()
		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}
		//	fmt.Println("login:", string(login), string(sitDown))

		//坐下
		time.Sleep(1 * time.Second)
		_, err = conn.Write(sitDown)
		if err != nil {
			fmt.Println("write:", err)
			return
		}

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
		var chairID int32 = -1

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

				sleep := util.RandInterval(1, 3)
				sitDown := sitDownMsg() //坐下
				time.Sleep(time.Second * time.Duration(sleep))
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
				//time.Sleep(time.Millisecond * 100)
				//sitDown1 := sitDownMsg() //坐下
				//_, err1 := conn.Write(sitDown1)
				//if err1 != nil {
				//	log.Println("write:", err)
				//	return
				//}
			case "Game_S_FreeScene": //空闲场景消息
				var msg gamemsg.Game_S_FreeScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("空闲场景消息recvString:\n", string(bytes))
				for _, v := range msg.UserList {
					if v.UserID == 1000414 {
						chairID = v.ChairID
					}
				}
				conn.Write(loginAutoMange())
			case "Game_S_GrabLandlordScene": //抢庄场景消息
				var msg gamemsg.Game_S_GrabLandlordScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("叫分场景消息recvString:\n", string(bytes), err)
			case "Game_S_PlayScene": //下注场景消息
				var msg gamemsg.Game_S_PlayScene
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("出牌场景消息recvString:\n", string(bytes))

			case "Game_S_UserGrabLandlord": //叫分场景消息
				var msg gamemsg.Game_S_UserGrabLandlord
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("叫分消息recvString:\n", string(bytes))
			case "Game_S_CurrentUser": //现在操作Charid
				var msg gamemsg.Game_S_CurrentUser
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("现在操作消息recvString:\n", string(bytes))
				if msg.CurrentChairID == chairID {

					log.Println("write: userOutPass", chairID, msg.CurrentChairID)
					_, err := conn.Write(userOutPass(chairID))
					if err != nil {
						log.Println("write:", err)
						return
					}
				}
			case "Game_S_UserCP": //出牌通知
				var msg gamemsg.Game_S_UserCP
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("玩家出牌通知recvString:\n", string(bytes))
			case "Game_S_UserPass": //过牌通知
				var msg gamemsg.Game_S_UserPass
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("玩家过牌通知recvString:\n", string(bytes))

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

				sitDown := sitDownMsg() //坐下
				_, err := conn.Write(sitDown)
				if err != nil {
					log.Println("write:", err)
					return
				}
				chairID = msg.Data.ChairID
			case "Game_S_StartGame": //发牌环节
				var msg gamemsg.Game_S_StartGame
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("开始游戏  recvString:\n", string(bytes))
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
			default:
				log.Println("无效命令", msgID)
			}
		}
	}
}
