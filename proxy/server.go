package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/remones/gsocks/config"
)

// Socks5Version ...
const Socks5Version = uint8(0x5)

// Errors ...
var (
	ErrServerClosed       = errors.New("socks: server closed")
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

// Server ...
type Server struct {
	addr           string
	listener       net.Listener
	mu             sync.Mutex
	waitConns      sync.WaitGroup
	inShutdown     int32
	doneChan       chan struct{}
	authenticators map[AuthType]Authenticator
	DialTimeout    time.Duration
}

// ListenAndServe serve the socks server
func (srv *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	ln = &onceCloseListener{Listener: ln}
	defer ln.Close()
	srv.listener = ln

	return srv.serve()
}

func (srv *Server) serve() (err error) {
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

func (srv *Server) serveSession(ctx context.Context, conn net.Conn) error {
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

	sess := srv.newSession(conn)
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
func (srv *Server) Close(ctx context.Context) error {
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

func (srv *Server) shuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *Server) getDoneChan() chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

// NewServer ...
func NewServer(cfg *config.Config) *Server {
	srv := new(Server)
	srv.addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv.authenticators = makeAuthsWithConfig(&cfg.Auth)
	srv.DialTimeout = time.Millisecond * time.Duration(cfg.DialTimeout)
	return srv
}
