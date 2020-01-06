package pgrpc

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

type clientOpts struct {
	logger
	dialTimeout time.Duration

	grpcDialOpts []grpc.DialOption
	onAccept     []func(net.Conn) (net.Conn, error)
	onGrpcDial   []func(*grpc.ClientConn) error
}

func (co *clientOpts) Log(format string, a ...interface{}) {
	if co.log != nil {
		co.log(format, a...)
	}
}

// ClientOpt is the client options
type ClientOpt interface {
	applyClient(*clientOpts)
}

type grpcDialOpt struct {
	grpcDialOpts []grpc.DialOption
}

// WithGrpcDialOpt set grpc dial options
func WithGrpcDialOpt(opts ...grpc.DialOption) ClientOpt {
	return &grpcDialOpt{grpcDialOpts: opts}
}
func (o *grpcDialOpt) applyClient(co *clientOpts) {
	co.grpcDialOpts = append(co.grpcDialOpts, o.grpcDialOpts...)
}

type proxyProtocol struct {
	net.Conn
	remoteAddr net.Addr
}

// WithProxyProtocol set proxy protocol support
func WithProxyProtocol() ClientOpt {
	return &proxyProtocol{}
}
func (o *proxyProtocol) applyClient(co *clientOpts) {
	co.onAccept = append(co.onAccept,
		func(conn net.Conn) (net.Conn, error) {
			c, ip, err := ParseProxyProto(o.Conn)
			if err != nil {
				return c, err
			}

			addr := c.RemoteAddr().(*net.TCPAddr)
			addr.IP = net.ParseIP(ip)
			return &proxyProtocol{Conn: c, remoteAddr: addr}, nil
		})
}
func (o *proxyProtocol) RemoteAddr() net.Addr {
	return o.remoteAddr
}
