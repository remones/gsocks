package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"strconv"
)

// Session is the session of negotiation
type Session struct {
	net.Conn
}

func newSession(c net.Conn, version uint8) *Session {
	return &Session{
		Conn: c,
	}
}

// Authenticate ...
func (s *Session) Authenticate() (bool, error) {
	methods, err := s.readMethods()
	if err != nil {
		return false, err
	}
	for _, method := range methods {
		if auth, found := authenticators[method]; found {
			err := s.ackMethod(method)
			if err != nil {
				return false, err
			}
			status, err := auth.Authenticate(s.Conn)
			return status, err
		}
	}
	return false, nil
}

func (s *Session) ackMethod(method byte) error {
	_, err := s.Write([]byte{Socks5Version, method})
	return err
}

func (s *Session) readMethods() ([]byte, error) {
	b := make([]byte, 1)
	n, err := s.Read(b)
	if err != nil {
		return nil, err
	}
	nMethods := int(n)
	buf := make([]byte, nMethods)
	_, err = io.ReadAtLeast(s.Conn, buf, nMethods)
	return buf, err
}

// ServeRequest ...
func (s *Session) ServeRequest(ctx context.Context) error {
	req, err := NewReuqest(s)
	if err != nil {
		return err
	}

	switch req.Command {
	case CmdConnect:
		err = s.handleCmdConnect(ctx, req)
	case CmdBind:
		err = s.handleCmdBind(ctx, req)
	case CmdUDPAssociate:
		err = s.handleCmdUDPProcess(ctx, req)
	default:
		err = fmt.Errorf("Invalid Request Command: %#x", req.Command)
	}
	return err
}

/*
   In the reply to a CONNECT, BND.PORT contains the port number that the
   server assigned to connect to the target host, while BND.ADDR
   contains the associated IP address.  The supplied BND.ADDR is often
   different from the IP address that the client uses to reach the SOCKS
   server, since such servers are often multi-homed.  It is expected
   that the SOCKS server will use DST.ADDR and DST.PORT, and the
   client-side source address and port in evaluating the CONNECT
   request.

        +----+-----+-------+------+----------+----------+
        |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
        +----+-----+-------+------+----------+----------+
        | 1  |  1  | X'00' |  1   | Variable |    2     |
        +----+-----+-------+------+----------+----------+
*/
func (s *Session) handleCmdConnect(ctx context.Context, req *Request) error {
	runtime.Breakpoint()
	addr := net.JoinHostPort(string(req.DestAddr.IP), strconv.Itoa(req.DestAddr.Port))
	target, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer target.Close()

	errCh := make(chan error, 2)
	proxy := func(dst io.Writer, src io.Reader) {
		defer target.Close()
		_, err := io.Copy(target, s.Conn)
		errCh <- err
	}
	go proxy(s.Conn, target)
	go proxy(target, s.Conn)

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO(remones)
func (s *Session) handleCmdBind(ctx context.Context, req *Request) error {
	return nil
}

// TODO(remones)
func (s *Session) handleCmdUDPProcess(ctx context.Context, req *Request) error {
	return nil
}
