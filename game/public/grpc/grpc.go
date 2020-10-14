package grpc

import (
	"context"
	"google.golang.org/grpc"
	golog "log"
	"net"
	"xj_game_server/game/public/redis"
	"xj_game_server/game/public/store"
	"xj_game_server/game/public/user"
	"xj_game_server/util/leaf/log"
)

//初始化grpc服务
func OnInit(rpcAddr string) {
	go func() {
		//注册Grpc
		serverLis, err := net.Listen("tcp", rpcAddr)
		if err != nil {
			_ = log.Logger.Errorf("RPC OnInit %v", err)
			golog.Fatalf("RPC OnInit %v", err)
			return
		}

		s := grpc.NewServer()
		//注册服务
		RegisterGameGrpcServer(s, &GameGrpc{})
		if err := s.Serve(serverLis); err != nil {
			_ = log.Logger.Errorf("RPC OnInit %v", err)
			golog.Fatalf("RPC OnInit %v", err)
			return
		}
	}()
}

type GameGrpc struct {
}

//金币变动
func (self *GameGrpc) ChangeGold(ctx context.Context, in *ChangeReq) (*ChangeReply, error) {
	//userItem, ok := user.List[in.UserID]
	userItem, ok := user.List.Load(in.UserID)
	redis.GameClient.RegisterRecharge(in.UserID)
	if !ok {
		return &ChangeReply{
			ErrorCode: -1,
			ErrorMsg:  "用户不在此游戏中!",
		}, nil
	}

	userItem.(*user.Item).UserGold = in.Score
	return &ChangeReply{
		ErrorCode: 0,
		ErrorMsg:  "",
	}, nil
}

//余额变动
func (self *GameGrpc) ChangeDiamond(ctx context.Context, in *ChangeReq) (*ChangeReply, error) {
	//userItem, ok := user.List[in.UserID]
	userItem, ok := user.List.Load(in.UserID)
	redis.GameClient.RegisterRecharge(in.UserID)
	if !ok {
		return &ChangeReply{
			ErrorCode: -1,
			ErrorMsg:  "用户不在此游戏中!",
		}, nil
	}

	userItem.(*user.Item).UserDiamond = in.Score
	return &ChangeReply{
		ErrorCode: 0,
		ErrorMsg:  "",
	}, nil
}

// 修改游戏配置
func (self *GameGrpc) ModifyStock(ctx context.Context, in *ModifyStockReq) (*ModifyStockReply, error) {
	log.Logger.Debugf("game config ModifyStock  do 更改前 原参数:%v,新的参数:%v ", store.GameControl.GetGameInfo(), in)
	if in.KindID == 0 || in.GameID == 0 {
		return &ModifyStockReply{}, nil
	}
	store.GameControl.OnInit(&store.GameInfo{
		GameID:         in.GameID,
		KindID:         in.KindID,
		KindName:       in.KindName,
		GameName:       in.GameName,
		SortID:         in.SortID,
		TableCount:     in.TableCount,
		ChairCount:     in.ChairCount,
		CellScore:      in.CellScore,
		RevenueRatio:   in.RevenueRatio,
		MinEnterScore:  in.MinEnterScore,
		DeductionsType: in.DeductionsType,
		StoresDecay:    in.StoresDecay,
		StartStores:    in.StartStores,
		StartWinRate:   in.StartWinRate,
		Threshold1:     in.Threshold1,
		WinRate1:       in.WinRate1,
		Threshold2:     in.Threshold2,
		WinRate2:       in.WinRate2,
		Threshold3:     in.Threshold3,
		WinRate3:       in.WinRate3,
		Threshold4:     in.Threshold4,
		WinRate4:       in.WinRate4,
		Threshold5:     in.Threshold5,
		WinRate5:       in.WinRate5,
	})
	log.Logger.Debugf("game config ModifyStock end 更改后 原参数:%v,新的参数:%v ", store.GameControl.GetGameInfo(), in)
	// 房间通知
	redis.GameClient.SetRoomChange()
	return &ModifyStockReply{}, nil
}

//获取当前库存
func (self *GameGrpc) QueryStock(ctx context.Context, in *QueryStockReq) (*QueryStockReply, error) {
	return &QueryStockReply{
		NowStores: store.GameControl.GetStore(),
	}, nil
}
