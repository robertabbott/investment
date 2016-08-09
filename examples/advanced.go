package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/apourchet/investment"
	pb "github.com/apourchet/investment/protos"
)

type Strat func(broker pb.BrokerClient, stream pb.Broker_StreamPricesClient)

func mine(broker pb.BrokerClient, stream pb.Broker_StreamPricesClient) {
	for {
		q, err := stream.Recv()
		if err == io.EOF || q == nil {
			return
		}

	}
}

func startTrader(strat Strat) {
	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	exitOnError(err)
	defer conn.Close()

	broker := pb.NewBrokerClient(conn)
	req := &pb.StreamPricesReq{}
	req.InstrumentId = "EURUSD"
	stream, err := broker.StreamPrices(context.Background(), req)
	exitOnError(err)

	strat(broker, stream)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
}

func main() {
	broker := invt.NewDefaultBroker()
	go broker.Start()
	time.Sleep(time.Millisecond * 50)

	milliStep := 20
	go startTrader(mine)
	invt.SimulateDataStream(broker, "examples/data/medium.csv", milliStep)

	req := &pb.AccountInfoReq{}
	resp, _ := broker.GetAccountInfo(context.Background(), req)
	fmt.Println(resp.Info.Balance)
}