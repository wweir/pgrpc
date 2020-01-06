package pgrpc

import (
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type listener struct {
	addr net.Addr
	serverOpts

	connCh chan *activeConn
	stopCh chan struct{}
}

func (ln *listener) Accept() (net.Conn, error) {
	for {
		select {
		case conn := <-ln.connCh:
			return conn, nil
		case <-ln.stopCh:
			return nil, errors.New("listener has been stoped")
		}
	}
}
func (ln *listener) Close() error {
	close(ln.stopCh)
	return nil
}
func (ln *listener) Addr() net.Addr {
	return ln.addr
}

//Listen start a pgrpc server with session id, it will tcp connect to the address
func Listen(address, id string, opts ...ServerOpt) (net.Listener, error) {
	if idLen := len(id); len(id) > MAX_ID_LEN {
		return nil, errors.Errorf("id(%s) is too long", id)
	} else if idLen == 0 {
		return nil, errors.Errorf("id is empty")
	}

	ln := &listener{
		connCh: make(chan *activeConn, MIN_IDLE-1),
		stopCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt.applyServer(&ln.serverOpts)
	}

	var err error
	if ln.addr, err = net.ResolveTCPAddr("tcp", address); err != nil {
		return nil, err
	}

	go func() {
	DIAL_LOOP:
		for {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				ln.Log("tcp dial %s fail: %s", address, err)
				continue
			}

			for _, fn := range ln.onAccept {
				if conn, err = fn(conn); err != nil {
					ln.Log("tcp on accept hook fail: %s", err)
					time.Sleep(5 * time.Second)
					continue DIAL_LOOP
				}
			}

			aConn, err := newActiveConn(ln, conn, id)
			if err != nil {
				ln.Log("init tcp conn fail: %s", err)
				time.Sleep(time.Second)
				continue
			}

			ln.connCh <- aConn
			select {
			case <-aConn.init:
			case <-ln.stopCh:
				aConn.Close()
				return
			}
		}
	}()
	return ln, nil
}

// activeConn is a net.Conn that can detect conn first activaty
type activeConn struct {
	init chan struct{}
	once sync.Once

	*listener
	net.Conn
}

func newActiveConn(ln *listener, conn net.Conn, id string) (*activeConn, error) {
	if c, ok := conn.(*net.TCPConn); ok {
		c.SetLinger(1)
		c.SetKeepAlive(true)
		c.SetKeepAlivePeriod(5 * time.Second)
	}

	aConn := activeConn{listener: ln, Conn: conn, init: make(chan struct{})}

	// write id
	buf := make([]byte, MAX_ID_LEN)
	copy(buf, []byte(id))

	aConn.SetDeadline(time.Now().Add(5 * time.Second))
	var err error
	for n, nn := 0, 0; nn < MAX_ID_LEN; nn += n {
		if n, err = aConn.Write(buf[nn:]); err != nil {
			aConn.Close()
			return nil, errors.Errorf("write id fail: %s", err)
		}
	}

	aConn.SetDeadline(time.Time{})
	return &aConn, nil
}

var tlsHandshark = []byte{0x16, 0x03, 0x01} // h2 tls handshake TLS 1.0
var h2cHeader = []byte("PRISM")             // h2c [P]RISM

func (a *activeConn) Read(b []byte) (n int, err error) {
	select {
	case <-a.init:
		return a.Conn.Read(b)

	default:
		// redial after 30s
		a.Conn.SetDeadline(time.Now().Add(30 * time.Second))
		if n, err = a.Conn.Read(b); err != nil {
			return n, errors.Errorf("tcp on accept hook fail: %s", err)
		}
		a.Conn.SetDeadline(time.Time{})

		// detect handshake
		if bytes.HasPrefix(b, tlsHandshark) || bytes.HasPrefix(b, h2cHeader) {
			a.once.Do(func() {
				close(a.init)
			})
		}

		return
	}
}
func (a *activeConn) Close() error {
	a.once.Do(func() {
		close(a.init)
	})
	return a.Conn.Close()
}
