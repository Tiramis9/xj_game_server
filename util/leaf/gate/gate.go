package gate

import (
	"net"
	"reflect"
	"time"
	"xj_game_server/util/leaf/chanrpc"
	"xj_game_server/util/leaf/log"
	"xj_game_server/util/leaf/network"
)

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	Processor       network.Processor
	AgentChanRPC    *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool
}

func (gate *Gate) Run(closeSig chan bool) {
	//var wsServer *network.WSServer
	//if gate.WSAddr != "" {
	//	wsServer = new(network.WSServer)
	//	wsServer.Addr = gate.WSAddr
	//	wsServer.MaxConnNum = gate.MaxConnNum
	//	wsServer.PendingWriteNum = gate.PendingWriteNum
	//	wsServer.MaxMsgLen = gate.MaxMsgLen
	//	wsServer.HTTPTimeout = gate.HTTPTimeout
	//	wsServer.CertFile = gate.CertFile
	//	wsServer.KeyFile = gate.KeyFile
	//	wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
	//		a := &agent{conn: conn, gate: gate}
	//		if gate.AgentChanRPC != nil {
	//			gate.AgentChanRPC.Go("NewAgent", a)
	//		}
	//		return a
	//	}
	//}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
			}
			return a
		}
	}

	//if wsServer != nil {
	//	wsServer.Start()
	//}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	//if wsServer != nil {
	//	wsServer.Close()
	//}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

type agent struct {
	conn     network.Conn
	gate     *Gate
	userData interface{}
}

func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			_ = log.Logger.Errorf("read message: %v", err)
			break
		}
		if a.gate.Processor != nil {
			msg, err := a.gate.Processor.Unmarshal(data)
			if err != nil {
				_ = log.Logger.Errorf("unmarshal message error: %v,%v", err, string(data))
				break
			}
			err = a.gate.Processor.Route(msg, a)
			if err != nil {
				_ = log.Logger.Errorf("route message error: %v", err)
				break
			}
		}
	}
}

func (a *agent) OnClose() {
	if a.gate.AgentChanRPC != nil {
		err := a.gate.AgentChanRPC.Call0("CloseAgent", a)
		if err != nil {
			_ = log.Logger.Error("chanrpc error:", err)
			return
		}
	}
}

func (a *agent) WriteMsg(msg interface{}) {
	if a.gate.Processor != nil {
		data, err := a.gate.Processor.Marshal(msg)
		if err != nil {
			_ = log.Logger.Errorf("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			_ = log.Logger.Errorf("write message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	time.Sleep(500 * time.Millisecond)
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}

func (a *agent) UserData() interface{} {
	return a.userData
}

func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
