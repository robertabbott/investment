package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	"time"

	"github.com/apourchet/investment"
	"github.com/apourchet/investment/lib/influx-session"
	pb "github.com/apourchet/investment/protos"
)

var (
	session *ix_session.Session
)

func quickOrder(units int32, side string) *pb.OrderCreationReq {
	o := &pb.OrderCreationReq{}
	o.InstrumentId = "EURUSD"
	o.Type = invt.TYPE_MARKET
	o.Side = side
	o.Units = units
	return o
}

func getStream(broker pb.BrokerClient) pb.Broker_StreamCandleClient {
	req := &pb.StreamCandleReq{&pb.AuthToken{}, "EURUSD"}
	stream, err := broker.StreamCandle(context.Background(), req)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
	return stream
}

// Create a point and add to batch
func mine(def *invt.DefaultBroker) {
	fmt.Println("Trader started")
	broker := def.GetClient()
	stream := getStream(broker)

	steps := 0
	for {
		c1, err := stream.Recv()
		if err == io.EOF || c1 == nil {
			fmt.Println("Candle stream has ended.")
			return
		}
		c := invt.CandleFromProto(c1)

		if steps%20 == 0 {
			req := &pb.AccountInfoReq{}
			resp, _ := broker.GetAccountInfo(context.Background(), req)
			fmt.Println(resp.Info.MarginAvail)
			session.Write("candle", c, c.Timestamp)
		}

		if c.Close-c.Low > (c.High-c.Close)*4 {
			o := quickOrder(3000, invt.StringOfSide(invt.SIDE_BUY))
			broker.CreateOrder(context.Background(), o)
			session.Write("order_buy", o, c.Timestamp)
		} else if (c.Close-c.Low)*4 < c.High-c.Close {
			o := quickOrder(3000, invt.StringOfSide(invt.SIDE_SELL))
			broker.CreateOrder(context.Background(), o)
			session.Write("order_sell", o, c.Timestamp)
		}
		steps++
	}
}

func main() {
	session = ix_session.NewSession(ix_session.DEFAULT_ADDRESS, "investment", "password", "testdb")
	datafile := "examples/data/largish.csv"
	if len(os.Args) >= 2 {
		datafile = os.Args[1]
	}

	simulator := invt.NewSimulator(invt.DATAFORMAT_CANDLE, datafile, 10)
	broker := invt.NewDefaultBroker(1027)
	go simulator.SimulateDataStream(broker)

	time.Sleep(time.Millisecond * 50)
	mine(broker)
}
