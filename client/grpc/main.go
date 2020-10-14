package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	msg "xj_game_server/client/grpc/msg"
	"time"
)

const (
	//address = "localhost:8002"
	address = "47.107.188.43:10180"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := msg.NewGameGrpcClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ChangeDiamond(ctx, &msg.ChangeReq{UserID: 10111, Score: 10})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %d---%s", r.GetErrorCode(), r.GetErrorMsg())
}
