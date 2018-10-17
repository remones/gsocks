package proxy

import (
	"context"
	"io"
	"net"
)

// Session is the session of negotiation
type Session struct {
	net.Conn
	ver uint8
}

func newSession(c net.Conn, version uint8) *Session {
	return &Session{
		Conn: c,
		ver:  version,
	}
}

// Version returns the version of socks support
func (s *Session) Version() uint8 {
	return s.ver
}

// Authenticate ...
func (s *Session) Authenticate() (ok bool, err error) {
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
			return auth.Authenticate(s.Conn)
		}
	}
	return false, nil
}

func (s *Session) ackMethod(method byte) error {
	_, err := s.Write([]byte{s.ver, method})
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

	switch req.Cmd {
	case CmdConnect:
		err = s.handleCmdConnect(ctx, req)
	case CmdBind:
		err = s.handleCmdBind()
	case CmdUDPProcess:
		err = s.handleCmdUDPProcess()
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
	addr := net.JoinHostPort(string(req.DstAddr), string(req.DstPort))
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
