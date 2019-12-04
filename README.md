# pgrpc
Passive grpc means the grpc server connect to the client.

It is useful while grpc server hidding behind NATs or gateways.

## Usage:
**grpc server side**
```go
	s := grpc.NewServer()
	// register handler as usual

	ln, err := pgrpc.Listen("127.0.0.1:50052", "example_server")
	if err != nil {
		log.Fatalln(err)
	}

	if err := s.Serve(ln); err != nil {
		log.Fatalln(err)
	}
```

**grpc client side**
``` go
	if err := pgrpc.InitClient(":50052",
		pgrpc.WithGrpcDialOpt(grpc.WithInsecure())); err != nil {
		log.Fatalln(err)
	}

	// loop all connected grpc servers
	pgrpc.Each(func(id string, cc *grpc.ClientConn, err error) error {
		if err != nil {
			log.Println("connect to id: %s, addr: %s, fail:", id, cc.Target(), err)
			return err
		}

		// grpc actions as usual
	}

	// call the specified grpc server
	conn, err := pgrpc.Dial("example_server")
	if err != nil {
		log.Println("connect to id: %s, addr: %s, fail:", id, cc.Target(), err)
	}

	// grpc actions as usual
```
