package proxy

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Socks5Version ...
const Socks5Version = uint8(0x5)

// Errors ...
var (
	ErrServerClosed       = errors.New("socks: Server closed")
	ErrProtoNotSupport    = errors.New("socks: only support SOCKS5 for now")
	ErrAuthenticateFailed = errors.New("socks: authenticate failed")
)

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (ln *onceCloseListener) Close() error {
	ln.once.Do(ln.close)
	return ln.closeErr
}

func (ln *onceCloseListener) close() {
	ln.closeErr = ln.Listener.Close()
}

type server struct {
	addr       string
	listener   net.Listener
	mu         sync.Mutex
	waitConns  sync.WaitGroup
	inShutdown int32
	doneChan   chan struct{}

	authenticators map[byte]Authenticator
}

// ListenAndServe serve the socks server
func (srv *server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	ln = &onceCloseListener{Listener: ln}
	defer ln.Close()
	srv.listener = ln

	return srv.serve()
}

func (srv *server) serve() (err error) {
	if srv.shuttingDown() {
		return ErrServerClosed
	}

	var tempDelay time.Duration
	ctx := context.Background()
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		srv.waitConns.Add(1)
		go srv.serveSession(ctx, conn)
	}
}

func (srv *server) serveSession(ctx context.Context, conn net.Conn) error {
	defer srv.waitConns.Done()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	b := make([]byte, 1)
	_, err := conn.Read(b)
	if err != nil {
		return err
	}
	ver := uint8(b[0])
	if ver != Socks5Version {
		return ErrProtoNotSupport
	}

	sess := newSession(conn, ver, 30*time.Second)
	authentic, err := sess.Authenticate()
	if err != nil {
		return err
	}
	if !authentic {
		return ErrAuthenticateFailed
	}
	return sess.ServeRequest(ctx)
}

// Close the server.
func (srv *server) Close(ctx context.Context) error {
	atomic.StoreInt32(&srv.inShutdown, 1)

	ch := srv.getDoneChan()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		close(ch)
	}
	lnerr := srv.listener.Close()

	srv.waitConns.Wait()
	return lnerr
}

func (srv *server) shuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *server) getDoneChan() chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}
