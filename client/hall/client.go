package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"time"

	//hallmsg "xj_game_server/client/hall/msg"
	hallmsg "xj_game_server/server/hall/msg"
)

const (
	//addr = "192.168.0.105:15000"
	//addr = "192.168.1.149:3010"
	//addr = "127.0.0.1:3010"
	addr = "47.113.94.16:15000"
	//addr = "216.118.243.18:15000"
)

func main() {
	fmt.Print(time.Now())
	tcpClient()
}
func tcpClient() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return
	}
	exit := make(chan bool)
	fmt.Println("已连接服务器")
	go sender(conn)
	read(conn, exit)

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

func CreteCmd1(data []byte) []byte {
	var msg = make([]byte, 0)
	var len = len(data)
	lenByte, _ := IntToBytes(int16(len))
	//cmdByte, _ := IntToBytes(cmd)
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
	loginVisitor := new(hallmsg.Hall_C_Msg)
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDcyIiwibmJmIjoxNTk1NTc2NjIyfQ.D5LgEqBAA8EvaXrPMu1npiFmDMJZuDobnFykbz-fn3CkGuav1bRrMO9Gf5BqA_Q9tfWaBd9Hhm1-gu6C72NScXeJF3t11PggG0Nf_dMDQwJfgg2oMphFivyIW1Dlkt3dFmADPFHJT5Yyf794_7YHZnq52AwofW79X3ZF1FcEG_jwcF6AtfuG_Ofj9fhUhDaorclzySCKFnfQRz4lFZHM3XJOmqdfAPOuVw2QZXCP-PWxEYFAHd8HJO5p9cX6eB6JJAUDBf1lf_KM4CJLB5-kRUvU_5W9hZ3uDBhlRinYJ7XeMdoXfbGCi10rok34BqT_4m12rQ5XwoXM48sxVGmhbQ"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDM0IiwibmJmIjoxNTk1NTAzOTMzfQ.vd_9FA7a3-1SH-Llo5z__ZrcOBfT-SLCXHIDsd5vzkwO5ccyPxYVkjBtQ5IbVRQTvoXVaAAH5jEYdmyI1rgupW1cMzW6dLCsucLAjMKbpq_yvRHw6yWO5Q05GKMdrJ360vO3v1vhPjNvG2z3E0SSk0MbHZtmN1k8gCTiS_CaWY9dykuv4ZnXo18pDppwwjGTjxp-Nzrcf1uynyitUo-3jlpblZiyaTyRBpbKYjozH3hYNChewz89vjNCYlrcB2cRmGMgUpl7zYR3liijxHhCaQBUT5pboP2I2aQPVPRZl9eZz-FvW54_C24kMoa3AGw_sKyack013tNx1lgqaC1xXw"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwNDU1IiwibmJmIjoxNTk2MDM0NTgzfQ.sv26QPOkOhAOceRj5Ef0EZR0x6iSs_qY-Y_tOFIRj5pB-tH9EALfHNm6nSeNIhEiN_lS6HNA2tV5SNg-S_nrhtOS2yUk9g3wXTPal5L97ToWKYmx_oB1STH0iTF_dnkJGH95Nh76-SlsLYxm10w0qCrFy-sqTmR_xmMSTrCdLiHMGGINZX8wEYEh00t134iTmc8LJbqnCr4mn1kBKzBYjmRbwGKBI-bRyGAHucDuDEijnmB_4-dJ8E-qbUsoSBlMUdNo16R2lhvBMaeFnD3FUNCIEPW3AfqKgKjnLKqNDhMKFFvJArOHvrG7l6osrPhM_EWHPH6oCBYBuYBFX_nnLA"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMzc0IiwibmJmIjoxNTk2MDk0MDY4fQ.Bj34LuKibinoMRqmqNcGekI7145ARybHc36TNLuDZ9FcAEaDWPNls8Jf4VX5SkT0I0HeFBUhnAWBdqSGUwhstyGBBQ-hMMS2zZkRn_8aKI8ASUDXsADF0dYFhUrtRD8kPLzyYWEpc5t_4kQlXiiED8gDQA-AbyJigFsAt8IOyDf1fkHJkEgkYP9fuflZT1U-gn4gtyYIydyA8xpl6M9gBBIcTWdHJgvDoY8T8mg6hg-zcCDTQwllCZiQAKnDHopF0UdtH0KkaJ46uGe68Xixy2n2fqESictQn9sspmqEWFkq4gdlsoZe-vtBcknfMJ-mgkChE-FfV_Bbz7sPiEFaIA"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NjkiLCJuYmYiOjE1OTQ4OTQ0MTZ9.SSya8gK_xEvXcsCpsbDciUGN5glfQK4TH19kDKDK0sNYhXRZdl6eCz4tGW3-sg5nbtGCbwz_-TmaRTijFvodYZOa23tgJTcTqiv6KqQ5Y4vNTS2T3ebaBSevr9I0i5wXKxAw6-JoiiWSjHR8mtV5samcaLUd6oOw0mXrdApirL_BJCirframW642JmZLpUGfRjLPG1E7yzcxWwHvcuZ7PwUbGtsptvP6PD3fwUPj809zxqL0Y0md0sYvi3qpFCvC4ir5r2jijk1w_q-hdGgF9xllV2HoJjTnaoxPJSKRXJ_q85raC1z0HoCA5S7ccGFzlipahFYSJ3qojZATr8wsew"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg4NjkiLCJuYmYiOjE1OTQ5OTE3Nzl9.Qa96oS_ii9O-TWaG1zHXUGCiSx4q_Z3PSLuYa7o3I8kuY6qfbLXtw9BUGen2Np213I8pFGvV78XAMl-ajecRVKspSPh-peZRa2NlD12xL65PdLFU0X74Y0dLLtOFPVk0wf7v8HZ9N8OQU6GvmeMjG0aoDLpSvMprcquqe2xu6bfAfn02j6BS3Imz1jFxNifaXm8Cu3NjirdNUSP1OK69xrka-mMUHFTTebAwyU4keuSBKwHtASCjhaj8n-DXobxMGdPmuZ1GBjliywSLvpxVsG2ciUT9ED5KhheBOjQ5hRxCmHUcwMc8PqMf35bom2oeFQCMIVjh8doJJ2CuNBnyrQ"
	//
	//// 阿里云
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwODY4IiwibmJmIjoxNTk3MDM5OTM4fQ.ZvoNGBoRJFenjzvPJUQT5skEEZ9fz2ggLE3e9x0OJ_mOq8ozdmTC8q8U8yBFWSQnXVJzPxm75VdLSTsNy2AhH9hIchCqPgk3c9X5l9yS-VMkU-jNvv4pPPwYPL7F4Fvv70ZUHu91hDtVIF6wkgIfd8hDGd7GzQHktRpqjo4VE1Be2xL3LZCbtVEBdaswEaiUlsyjHpiayvMLsaODeEJgRvjQpIAchqN3kuV64lJ8-HISQAu-sNRfjGJmy06xj-YB82Sz7QL9ogkr3cd0mx-pTXYi1bnp0jB27LtoVZh8rKUOmXqs4Iqxki1n3K1ubfa4XJ5FyBmsKEKjO4eTCVRopQ"
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NjYiLCJuYmYiOjE1OTQ4OTQwNDh9.ZCOWA36SomAcP7q4qxPibpW9h-LNTfarMz7LvhfvrlwYksFO34NuS87YD05VTxPqV4cvFGFEjR-3UqOlsMqiapf9KZVzAbfE1-REYYGaRMfw-TmL8tEP4K0p9n2Gt418csraRRI0ZKxCZz2iwbg7XR2w65roeMoDIzmahV9_qurMnRSxKjIpdhACSu5gw067wS_enU9n2SlpYpai-viOf3qfnlf2Z4gaVtp0QO6aw_SOiIw8AljXzpBZBW4REN-DVz8SvVD84LHemitoJ9cwwvUeo9RwkBcAsgD4gGY_H6AWgEWJcewZxYW2TnL95JrPYgMdjnpjMjSfJdYtUCg_VQ"
	//
	////	data, _ := json.Marshal(loginVisitor)
	////return CreteCmd(0x00, data)
	//
	url := "192.168.0.105"
	if addr != "192.168.0.105:15000" {
		url = "47.113.94.16"
		body, err := HttpPostJson("http://"+url+":8000/v1/user/mobile_login", map[string]interface{}{
			"platform":            "WindowsEditor",
			"app_version":         "0.1.53",
			"app_name":            "newxingjing",
			"num_register_origin": 8,
			"machine_id":          "51676087823dfaa801ce203dac0970460",
			"account":             "张三",
			"mobile":              "13200001111",
			"password":            "123456",
			"device_type":         8,
			"code":                "0000",
		})
		if err != nil {
			log.Fatalln("get login err:", string(body), err)
		}
		var resp RespResult
		err = json.Unmarshal(body, &resp)
		if err != nil {
			log.Fatalln("Unmarshal err:", err)
		}
		loginVisitor.Token = resp.Data.Token
	}
	//fmt.Println("token:", resp.Data.Token, resp)

	//fmt.Println("token 1:", loginVisitor.Token)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

type RespResult struct {
	ErrorCode int              `json:"errno"`
	ErrMsg    string           `json:"errmsg"`
	Data      VisitorLoginResp `json:"data"`
}
type VisitorLoginResp struct {
	UserInfo    *LoginUserInfo `json:"user_info"`
	Token       string         `json:"token"`
	HeartServer string         `json:"heart_server"`
}

type LoginUserInfo struct {
	UserID        int     `json:"user_id"`
	NikeName      string  `json:"user_nickname"`
	UserGold      float64 `json:"user_gold"`
	UserDiamonds  float64 `json:"user_diamond"`
	UserVipLevel  int     `json:"user_vip_level"`
	UserVipExp    int     `json:"user_vip_exp"`
	UserPhone     string  `json:"user_phone"`
	UserHeadUrl   string  `json:"user_head_url"`
	UserHeadFrame int     `json:"user_head_frame"`
	UserModel     int     `json:"user_model"`
	IsInsurePass  bool    `json:"is_insure_pass"`
	UserGameLock  []int   `json:"user_game_lock"`
	//	WalletInfo    []UserWalletInfo `json:"user_wallet"`
}

//Post 请求方法
func HttpPostJson(url string, data interface{}) ([]byte, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 登录
func loginMsgWeb() []byte {
	loginVisitor := new(hallmsg.Hall_C_Msg)
	//loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDcyIiwibmJmIjoxNTk1NTc2NjIyfQ.D5LgEqBAA8EvaXrPMu1npiFmDMJZuDobnFykbz-fn3CkGuav1bRrMO9Gf5BqA_Q9tfWaBd9Hhm1-gu6C72NScXeJF3t11PggG0Nf_dMDQwJfgg2oMphFivyIW1Dlkt3dFmADPFHJT5Yyf794_7YHZnq52AwofW79X3ZF1FcEG_jwcF6AtfuG_Ofj9fhUhDaorclzySCKFnfQRz4lFZHM3XJOmqdfAPOuVw2QZXCP-PWxEYFAHd8HJO5p9cX6eB6JJAUDBf1lf_KM4CJLB5-kRUvU_5W9hZ3uDBhlRinYJ7XeMdoXfbGCi10rok34BqT_4m12rQ5XwoXM48sxVGmhbQ"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMDM0IiwibmJmIjoxNTk1NTAzOTMzfQ.vd_9FA7a3-1SH-Llo5z__ZrcOBfT-SLCXHIDsd5vzkwO5ccyPxYVkjBtQ5IbVRQTvoXVaAAH5jEYdmyI1rgupW1cMzW6dLCsucLAjMKbpq_yvRHw6yWO5Q05GKMdrJ360vO3v1vhPjNvG2z3E0SSk0MbHZtmN1k8gCTiS_CaWY9dykuv4ZnXo18pDppwwjGTjxp-Nzrcf1uynyitUo-3jlpblZiyaTyRBpbKYjozH3hYNChewz89vjNCYlrcB2cRmGMgUpl7zYR3liijxHhCaQBUT5pboP2I2aQPVPRZl9eZz-FvW54_C24kMoa3AGw_sKyack013tNx1lgqaC1xXw"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwNDU1IiwibmJmIjoxNTk2MDM0NTgzfQ.sv26QPOkOhAOceRj5Ef0EZR0x6iSs_qY-Y_tOFIRj5pB-tH9EALfHNm6nSeNIhEiN_lS6HNA2tV5SNg-S_nrhtOS2yUk9g3wXTPal5L97ToWKYmx_oB1STH0iTF_dnkJGH95Nh76-SlsLYxm10w0qCrFy-sqTmR_xmMSTrCdLiHMGGINZX8wEYEh00t134iTmc8LJbqnCr4mn1kBKzBYjmRbwGKBI-bRyGAHucDuDEijnmB_4-dJ8E-qbUsoSBlMUdNo16R2lhvBMaeFnD3FUNCIEPW3AfqKgKjnLKqNDhMKFFvJArOHvrG7l6osrPhM_EWHPH6oCBYBuYBFX_nnLA"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwMzc0IiwibmJmIjoxNTk2MDk0MDY4fQ.Bj34LuKibinoMRqmqNcGekI7145ARybHc36TNLuDZ9FcAEaDWPNls8Jf4VX5SkT0I0HeFBUhnAWBdqSGUwhstyGBBQ-hMMS2zZkRn_8aKI8ASUDXsADF0dYFhUrtRD8kPLzyYWEpc5t_4kQlXiiED8gDQA-AbyJigFsAt8IOyDf1fkHJkEgkYP9fuflZT1U-gn4gtyYIydyA8xpl6M9gBBIcTWdHJgvDoY8T8mg6hg-zcCDTQwllCZiQAKnDHopF0UdtH0KkaJ46uGe68Xixy2n2fqESictQn9sspmqEWFkq4gdlsoZe-vtBcknfMJ-mgkChE-FfV_Bbz7sPiEFaIA"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg3NjkiLCJuYmYiOjE1OTQ4OTQ0MTZ9.SSya8gK_xEvXcsCpsbDciUGN5glfQK4TH19kDKDK0sNYhXRZdl6eCz4tGW3-sg5nbtGCbwz_-TmaRTijFvodYZOa23tgJTcTqiv6KqQ5Y4vNTS2T3ebaBSevr9I0i5wXKxAw6-JoiiWSjHR8mtV5samcaLUd6oOw0mXrdApirL_BJCirframW642JmZLpUGfRjLPG1E7yzcxWwHvcuZ7PwUbGtsptvP6PD3fwUPj809zxqL0Y0md0sYvi3qpFCvC4ir5r2jijk1w_q-hdGgF9xllV2HoJjTnaoxPJSKRXJ_q85raC1z0HoCA5S7ccGFzlipahFYSJ3qojZATr8wsew"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMzg4NjkiLCJuYmYiOjE1OTQ5OTE3Nzl9.Qa96oS_ii9O-TWaG1zHXUGCiSx4q_Z3PSLuYa7o3I8kuY6qfbLXtw9BUGen2Np213I8pFGvV78XAMl-ajecRVKspSPh-peZRa2NlD12xL65PdLFU0X74Y0dLLtOFPVk0wf7v8HZ9N8OQU6GvmeMjG0aoDLpSvMprcquqe2xu6bfAfn02j6BS3Imz1jFxNifaXm8Cu3NjirdNUSP1OK69xrka-mMUHFTTebAwyU4keuSBKwHtASCjhaj8n-DXobxMGdPmuZ1GBjliywSLvpxVsG2ciUT9ED5KhheBOjQ5hRxCmHUcwMc8PqMf35bom2oeFQCMIVjh8doJJ2CuNBnyrQ"
	loginVisitor.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIxMDAwNDE0IiwibmJmIjoxNTk2MDM3NjcxfQ.tktPM4wv14x_gpQhp87_E0PyPoJydq5n6SoZdSq_O43t7DLaI7QeZhqr4qtB2ihs5oq9hJ4ATnb5uvvnCwHGjOvtWIuzY0fawK_lucgca0K-Zz-rFDYPhNXs_MObhs7FEsCPKTxAthwH0yOQ6TfztgjGIyJDRyVQD51j4KYwmHqwaJN8PIBEUh-W7ctRkgE0lYU_qmwqPET1YbeRaR6nnDxHA4m2sZnWXimQbOPQiTfzKEk7YMkPSfw7MsxGgyTSEuhcDHLnLRK5eZ-X97W47F5G175btis5MvBwcYObgg58DNkE5ZbreFRbE-pigUDhtwf8rnyx7oy81x8UZ150Rw"

	//	data, _ := json.Marshal(loginVisitor)
	//return CreteCmd(0x00, data)
	data, _ := json.Marshal(map[string]interface{}{reflect.TypeOf(loginVisitor).Elem().Name(): loginVisitor})
	return CreteCmd1(data)
}

//下注
func userJettonUpMsg() []byte {
	//loginVisitor := new(gamemsg.Game_C_UserJetton)
	//loginVisitor.Multiple = 5
	//data, _ := proto.Marshal(loginVisitor)
	//return CreteCmd(0x04, data)
	return nil
}

//摊牌
func userTpMsg() []byte {
	//loginVisitor := new(gamemsg.Game_C_UserTP)
	//data, _ := proto.Marshal(loginVisitor)
	//return CreteCmd(0x05, data)
	return nil
}

func sender(conn net.Conn) {
	for {
		login := loginMsg()
		fmt.Println("send:", string(login))
		fmt.Println(login)

		_, err := conn.Write(login)
		if err != nil {
			log.Println("write:", err)
			return
		}

		time.Sleep(time.Second * 3)

		break

	}
}

////机器人-----
//Processor.Register(&Game_C_RobotLogin{}) //机器人登陆
func read(conn net.Conn, close chan bool) {
	for {
		var message = make([]byte, 1024*10)
		n, err := conn.Read(message)
		if err != nil && err != io.EOF || len(message) == 0 {
			log.Println("read:", err)
			return
		}
		index := bytes.IndexByte(message, 0)
		if index > 0 {
			message = message[:index]
		}

		message = message[:n]
		if len(message) == 0 {
			break
		}

		//fmt.Println("msg:", string(message[2:]), "len", n)
		//fmt.Println("msg:", string(message), "len", n)
		var m = make(map[string]json.RawMessage)
		err = json.Unmarshal(message[2:], &m)
		if err != nil {
			log.Printf("Unmarshal 错误 err:%s\n", err)
			return
		}
		if len(m) != 1 {
			log.Printf("message 错误 err:%v\n", m)
			return
		}

		for msgID, data := range m {
			switch msgID {
			case "HeartLoginInit": //登陆成功消息
				var msg hallmsg.HeartLoginInit
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求成功recvString:\n", string(bytes))
			case "Hall_S_Fail": //请求失败消息
				var msg hallmsg.Hall_S_Fail
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求失败消息recvString:\n", string(bytes))

				time.Sleep(time.Second * 2)
				_, err := conn.Write(loginMsg())
				fmt.Println("wrie:", err)
				if err != nil {
					log.Println("err:", err)
				}
			case "UserInfoChange": //充值消息
				var msg hallmsg.UserInfoChange
				err = json.Unmarshal(data, &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("充值消息recvString:\n", string(bytes))

			}

		}

		//var cmd int16
		//err = BytesToInt(&cmd, message[2:4])
		//if err != nil {
		//	log.Printf("cmd 错误 err:%d\n", cmd)
		//	return
		//}
		//fmt.Println("cmd:", cmd)
		/*
			switch cmd {
			case 1: //登陆成功消息
				var msg hallmsg.Hall_S_Msg
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求成功recvString:\n", string(bytes))
			case 2: //请求失败消息
				var msg hallmsg.Hall_S_Fail
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("请求失败消息recvString:\n", string(bytes))
			case 3: //充值消息
				var msg hallmsg.Hall_Recharge_Notice
				err = proto.Unmarshal(message[4:], &msg)
				bytes, _ := json.Marshal(msg)
				log.Print("充值消息recvString:\n", string(bytes))

			}

		*/
	}

	go func() {
		for {
			select {
			case <-close:
				fmt.Print("close read")
				return
			}
		}
	}()
}
