package pgrpc

import (
	"net"
)

type serverOpts struct {
	logger

	log      func(string, ...interface{})
	onAccept []func(net.Conn) (net.Conn, error)
}

// ServerOpt is the client options
type ServerOpt interface {
	applyServer(*serverOpts)
}
