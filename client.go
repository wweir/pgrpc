package pgrpc

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Client struct {
	sync.Map

	clientOpts
}

var defaultClient *Client

// InitClient init the global client, it will tcp listen on the addr
func InitClient(addr string, opts ...ClientOpt) (err error) {
	defaultClient, err = NewClient(addr, opts...)
	return err
}

// NewClient init a new client, it will tcp listen on the addr
func NewClient(addr string, opts ...ClientOpt) (*Client, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	var c = &Client{}
	for _, opt := range opts {
		opt.applyClient(&c.clientOpts)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				c.Log("tcp listen fail: %s", err)
				continue
			}

			go func(conn net.Conn) {
				for _, fn := range c.onAccept {
					if conn, err = fn(conn); err != nil {
						c.Log("run tcp accept hook fail: %s", err)
						conn.Close()
						return
					}
				}

				// read connection id
				var id string
				{
					buf := make([]byte, MAX_ID_LEN)
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					if _, err := io.ReadFull(conn, buf); err != nil {
						c.Log("read connection fail: %s, readed: %s", err, buf)
						conn.Close()
						return
					}

					conn.SetDeadline(time.Time{})
					id = string(bytes.TrimRight(buf, string(0)))
				}

				if c, ok := conn.(*net.TCPConn); ok {
					c.SetKeepAlive(true)
					c.SetKeepAlivePeriod(5 * time.Second)
				}

				// cache connection
				if val, ok := c.LoadOrStore(id, &pool{id: id, conns: []net.Conn{conn}, Client: c}); ok {
					val.(*pool).PutConn(id, conn)
				}
			}(conn)
		}
	}()
	return c, nil
}

// Dial build a new connection from the global client
func Dial(key string) (*grpc.ClientConn, error) {
	return defaultClient.Dial(key)
}

// Dial build a new connection
func (c *Client) Dial(key string) (*grpc.ClientConn, error) {
	val, ok := c.Load(key)
	if !ok {
		return nil, errors.Errorf("connection point to %s not found", key)
	}

	cc, err := val.(*pool).Get()
	return cc, errors.Wrap(err, "pgrpc dial")
}

func Each(fn func(id string, cc *grpc.ClientConn) error) {
	defaultClient.Each(fn)
}
func (c *Client) Each(fn func(id string, cc *grpc.ClientConn) error) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	c.Range(func(key, val interface{}) bool {
		wg.Add(1)
		go func(id string, pool *pool) {
			defer wg.Done()

			cc, err := pool.Get()
			if err != nil {
				c.Log("pgrpc dial %s fail: %s", key, err)
				return
			}

			if err = fn(id, cc); err != nil {
				c.Log("pgrpc do each func for %s fail: %s", key, err)
			}
			pool.PutCC(cc, err)

		}(key.(string), val.(*pool))
		return true
	})
}

func PutCC(cc *grpc.ClientConn, err error) {
	defaultClient.PutCC(cc, err)
}
func (c *Client) PutCC(cc *grpc.ClientConn, err error) {
	val, ok := c.Load(cc.Target())
	if !ok {
		return
	}

	pool := val.(*pool)
	pool.PutCC(cc, err)
}

// pool maintain idle connections
type pool struct {
	id    string
	conns []net.Conn
	ccs   []*grpc.ClientConn

	*Client
	mu sync.Mutex
}

func (s *pool) Get() (cc *grpc.ClientConn, err error) {
	for i := 0; i < 10; i++ {
		s.mu.Lock()
		if len(s.ccs) != 0 {
			cc = s.ccs[0]
			s.ccs = s.ccs[1:]
			s.mu.Unlock()

			cc.Close()
			continue
		}

		// no avaiable ClientConn, try build from net.Conn
		if len(s.conns) == 0 {
			s.mu.Unlock()
			time.Sleep(200 * time.Millisecond)
			continue
		}

		conn := s.conns[0]
		s.conns = s.conns[1:]
		s.mu.Unlock()

		// dial client conn
		opts := append(s.grpcDialOpts, grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) { return conn, nil }))
		cc, err := grpc.DialContext(context.Background(), conn.RemoteAddr().String(), opts...)
		if err != nil {
			s.Log("grpc dail fail: %s", err)
			conn.Close()
			continue
		}
		{ // ping check
			for _, fn := range s.onGrpcDial {
				if err = fn(cc); err != nil {
					cc.Close()
					s.Log("build grpc client conn from tcp conn fail: %s", err)
					break
				}
			}
			if err == nil {
				return cc, nil
			}
		}
	}

	return nil, errors.New("no connection to " + s.id)
}

func (s *pool) PutCC(cc *grpc.ClientConn, err error) {
	if err != nil {
		if cc != nil {
			cc.Close()
		}
		return
	}

	s.mu.Lock()
	if len(s.ccs) == (MAX_IDLE - MIN_IDLE) {
		cc.Close()
	} else {
		s.ccs = append(s.ccs, cc)
	}
	s.mu.Unlock()
}

func (s *pool) PutConn(id string, conn net.Conn) {
	s.mu.Lock()
	s.conns = append(s.conns, conn)
	s.mu.Unlock()
}
