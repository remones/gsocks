package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// ReplyCode ...
type ReplyCode byte

// ReplyCodes ...
const (
	ReplySuccessed          = ReplyCode(0x00)
	ReplyFailure            = ReplyCode(0x01)
	ReplyNotAllowed         = ReplyCode(0x02)
	ReplyNetworkUnreachable = ReplyCode(0x03)
	ReplyHostUnreachable    = ReplyCode(0x04)
	ReplyConnectionRefused  = ReplyCode(0x05)
	ReplyTTLExpired         = ReplyCode(0x06)
	ReplyInvalidCommand     = ReplyCode(0x07)
	ReplyInvalidAddressType = ReplyCode(0x08)
	ReplyUnassigned         = ReplyCode(0x09)
)

// errors ...
var (
	ErrSendReplyFailed  = errors.New("sends a reply failed")
	ErrBindSocketFailed = errors.New("binds a socket failed")
	ErrResolverFailed   = errors.New("resolve remote address failed")
)

// Session is the session of negotiation
type Session struct {
	srv *Server
	net.Conn
}

func (srv *Server) newSession(c net.Conn) *Session {
	return &Session{
		srv:  srv,
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
		if auth, found := s.srv.authenticators[AuthType(method)]; found {
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
	case CmdUDP:
		err = s.handleCmdUDP(ctx, req)
	default:
		s.sendReply(ReplyInvalidCommand, nil)
		err = fmt.Errorf("Invalid Request Command: %#x", req.Command)
	}
	return err
}

func startProxy(rw1, rw2 io.ReadWriter, errCh chan<- error) {
	proxy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		if errCh != nil {
			errCh <- err
		}
	}
	go proxy(rw1, rw2)
	go proxy(rw2, rw1)
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
	target, err := s.resolverAndDialAddr(ctx, req.DestAddr)
	if err != nil {
		return err
	}
	defer target.Close()

	errCh := make(chan error)
	startProxy(target, s.Conn, errCh)

	select {
	case <-ctx.Done():
		s.Close()
		err = ctx.Err()
	case nErr := <-errCh:
		err = nErr
	}
	return err
}

func (s *Session) handleCmdBind(ctx context.Context, req *Request) error {
	target, err := s.resolverAndDialAddr(ctx, req.DestAddr)
	if err != nil {
		return err
	}
	defer target.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if rErr := s.sendReply(ReplyUnassigned, nil); rErr != nil {
			return ErrSendReplyFailed
		}
		return err
	}

	lnhost, lnport, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(lnport)
	as := &AddrSpec{
		IP:   net.ParseIP(lnhost),
		Port: port,
		Type: TypeIPV4,
	}
	s.sendReply(ReplySuccessed, as)

	conn, err := ln.Accept()
	if err != nil {
		if rErr := s.sendReply(ReplyFailure, nil); rErr != nil {
			return ErrSendReplyFailed
		}
		return err
	}

	errCh := make(chan error)
	startProxy(conn, target, errCh)
	select {
	case <-ctx.Done():
		s.Close()
		err = ctx.Err()
	case nErr := <-errCh:
		err = nErr
	}
	return err
}

/*
Support for UDP, UDP connection lifetime must be as same as the TCP.
The UDP request header like it:
+----+------+------+----------+----------+----------+
|RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
+----+------+------+----------+----------+----------+
| 2  |  1   |  1   | Variable |    2     | Variable |
+----+------+------+----------+----------+----------+
*/
func (s *Session) handleCmdUDP(ctx context.Context, req *Request) error {
	dest, err := req.DestAddr.Resolve(ctx)
	if err != nil {
		s.sendReply(ReplyInvalidAddressType, nil)
		return err
	}
	assignAddr, err := net.ResolveUDPAddr("udp", dest)
	if err != nil {
		s.sendReply(ReplyUnassigned, nil)
		return err
	}

	uAddr := net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("127.0.0.1"),
	}
	udpSrvConn, err := net.ListenUDP("udp", &uAddr)
	if err != nil {
		s.sendReply(ReplyFailure, nil)
		return err
	}
	defer udpSrvConn.Close()

	srvAddr := udpSrvConn.LocalAddr()
	lnhost, lnport, _ := net.SplitHostPort(srvAddr.String())
	port, _ := strconv.Atoi(lnport)
	as := AddrSpec{
		IP:   net.ParseIP(lnhost),
		Port: port,
		Type: TypeIPV4,
	}
	s.sendReply(ReplySuccessed, &as)

	doneCh := make(chan error)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	// used to close spawned goroutine
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// keep TCP alive
	go func() {
		kaReply := []byte{0x00}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.Write(kaReply); err != nil {
					doneCh <- err
					return
				}
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			s.Close()
			return ctx.Err()
		case err := <-doneCh:
			return err
		default:
		}

		_, cAddr, err := udpSrvConn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		if cAddr.Port != assignAddr.Port {
			return fmt.Errorf("UDP unknow port")
		}
	}
}

func (s *Session) resolverAndDialAddr(ctx context.Context, as *AddrSpec) (net.Conn, error) {
	addr, err := as.Resolve(ctx)
	if err != nil {
		if rErr := s.sendReply(ReplyHostUnreachable, nil); rErr != nil {
			return nil, ErrSendReplyFailed
		}
		return nil, ErrResolverFailed
	}

	dialer := net.Dialer{Timeout: s.srv.DialTimeout}
	target, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		errMsg := err.Error()
		resp := ReplyHostUnreachable
		if strings.Contains(errMsg, "refused") {
			resp = ReplyConnectionRefused
		} else if strings.Contains(errMsg, "network is unreachable") {
			resp = ReplyNetworkUnreachable
		}
		if rErr := s.sendReply(resp, nil); rErr != nil {
			// TODO: just log here
			return nil, ErrSendReplyFailed
		}
		return nil, ErrResolverFailed
	}
	return target, nil
}

func (s *Session) sendReply(code ReplyCode, addr *AddrSpec) error {
	var (
		addrType uint8
		addrBody []byte
		addrPort uint16
	)
	switch {
	case addr == nil:
		addrType = TypeIPV4
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = addr.Type
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = addr.Type
		addrBody = addr.IP.To4()
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = addr.Type
		addrBody = addr.IP.To16()
		addrPort = uint16(addr.Port)
	default:
		return fmt.Errorf("Invalid replied address: %v", addr)
	}

	reply := make([]byte, 6+len(addrBody))
	n := len(reply)
	reply[0] = Socks5Version
	reply[1] = byte(code)
	reply[2] = 0
	reply[3] = addrType

	copy(reply[4:n-2], addrBody)
	copy(reply[n-2:], []byte{uint8(addrPort >> 8), uint8(addrPort) & 255})
	_, err := s.Write(reply)
	return err
}
