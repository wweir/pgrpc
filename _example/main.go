package main

import (
	"context"
	"log"
	"time"

	"github.com/wweir/pgrpc"
	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	go server()
	client()
}

// grpc handler
type pingServer struct{}

func (s *pingServer) Ping(ctx context.Context, in *PingMsg) (*PingMsg, error) {
	return &PingMsg{Msg: "Hello " + in.Msg}, nil
}

// pgrpc server
func server() {
	ln, err := pgrpc.Listen("127.0.0.1:50052", "example_server")
	if err != nil {
		log.Fatalln(err)
	}

	s := grpc.NewServer()
	RegisterPingServer(s, &pingServer{})
	if err := s.Serve(ln); err != nil {
		log.Fatalln(err)
	}
}

// pgrpc client
func client() {
	if err := pgrpc.InitClient(":50052",
		pgrpc.WithGrpcDialOpt(grpc.WithInsecure())); err != nil {
		log.Fatalln(err)
	}

	// waiting for server connection
	time.Sleep(time.Second)

	var oneID string
	{ // test each all server
		pgrpc.Each(func(id string, cc *grpc.ClientConn) error {
			oneID = id

			resp, err := NewPingClient(cc).Ping(context.Background(), &PingMsg{
				Msg: "pgrpc_each",
			})
			if err != nil {
				log.Fatalf("%+v", err)
				return err
			}

			log.Printf("remote ip: %s, msg: %s", cc.Target(), resp.Msg)

			oneID = id
			return nil
		})
	}

	{ // test dial one server
		cc, err := pgrpc.Dial(oneID)
		if err != nil {
			log.Fatalln(err)
		}
		resp, err := NewPingClient(cc).Ping(context.Background(), &PingMsg{
			Msg: "pgrpc_dial",
		})
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("remote ip: %s, msg: %s", cc.Target(), resp.Msg)
		cc.Close()
	}
}
