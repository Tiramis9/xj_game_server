/*
 * @Author: yhlyl
 * @Date: 2019-11-26 15:17:38
 * @LastEditTime: 2019-11-27 15:01:00
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /xj_game_server/game/public/robot/agent.go
 * @https://github.com/android-coco
 */
package robot

import (
	"net"
	"xj_game_server/util/leaf/gate"
	"xj_game_server/util/leaf/log"
	"reflect"
)

//Agent 机器人代理
type Agent struct {
	gate     *gate.Gate               // 底层连接代理
	userData interface{}              // 用户数据
	callBack func(args []interface{}) //关闭回调
}

// OnInit 初始化
func (a *Agent) OnInit(gate *gate.Gate, userCallBack func(args []interface{})) {
	a.gate = gate
	a.callBack = userCallBack
}

// WriteMsg 发命令
func (a *Agent) WriteMsg(msg interface{}) {
	err := a.gate.Processor.Route(msg, a)
	if err != nil {
		_ = log.Logger.Errorf("route message %v error: %v", reflect.TypeOf(msg), err)
	}
}

//RobotAddr 机器人地址
type RobotAddr struct {
}

//Network 机器人网络
func (*RobotAddr) Network() string {
	return "robot"
}

//String 机器人ip地址
func (*RobotAddr) String() string {
	return "127.0.0.1"
}

//LocalAddr 本地连接地址
func (*Agent) LocalAddr() net.Addr {
	return &RobotAddr{}
}

//RemoteAddr 远程连接地址
func (*Agent) RemoteAddr() net.Addr {
	return &RobotAddr{}
}

// Close 关闭
func (a *Agent) Close() {
	if a.callBack != nil {
		args := make([]interface{}, 0)
		args = append(args, a)
		a.callBack(args)
	}
}

// Destroy 销毁
func (a *Agent) Destroy() {
	if a.callBack != nil {
		args := make([]interface{}, 0)
		args = append(args, a)
		a.callBack(args)
	}
}

// UserData 用户信息
func (a *Agent) UserData() interface{} {
	return a.userData
}

// SetUserData 设置用户信息
func (a *Agent) SetUserData(data interface{}) {
	a.userData = data
}
