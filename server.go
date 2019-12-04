package pgrpc

import (
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
			for _, fn := range ln.onAccept {
				if conn, err = fn(conn); err != nil {
					ln.Log("run tcp accept hook fail: %s", err)
					return
				}
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

func (a *activeConn) Read(b []byte) (n int, err error) {
	select {
	case <-a.init:
		return a.Conn.Read(b)

	default:
		// wait 1min and redial, avoid aws load balancer 1m close policy
		a.Conn.SetDeadline(time.Now().Add(time.Minute))
		if n, err = a.Conn.Read(b); err != nil {
			return n, errors.Errorf("tcp on accept hook fail: %s", err)
		}
		a.Conn.SetDeadline(time.Time{})

		// detect handshake
		if n != 0 && (b[0] == 22 /* h2 tls handshake */ || b[0] == 50 /* h2c [P]RISM */) {
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
