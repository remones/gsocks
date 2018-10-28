package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Socks5Version ...
const Socks5Version = uint8(0x5)

type server struct {
	wg             sync.WaitGroup
	quitOnce       sync.Once
	authenticators map[byte]Authenticator
}

func newServer() *server {
	return &server{}
}

func (s *server) ListenAndServe(network, addr string) error {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func (s *server) Serve(ln net.Listener) (err error) {
	ctx := context.Background()
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.serveSession(ctx, c) // # TODO: log the error
		}(conn)
	}
	return
}

func (s *server) serveSession(ctx context.Context, conn net.Conn) error {
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
		return fmt.Errorf("Only support SOCKS5 for now")
	}

	sess := newSession(conn, ver)
	authentic, err := sess.Authenticate()
	if err != nil {
		return err
	}
	if !authentic {
		return fmt.Errorf("AuthenticateFailed")
	}
	return sess.ServeRequest(ctx)
}
