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
	DialTimeout time.Duration
	net.Conn
}

func newSession(c net.Conn, version uint8, timeout time.Duration) *Session {
	return &Session{
		DialTimeout: timeout,
		Conn:        c,
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
	case CmdUDP:
		err = s.handleCmdUDP(ctx, req)
	default:
		s.sendReply(ReplyInvalidCommand, nil)
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
	addr, err := req.DestAddr.Resolve(ctx)
	if err != nil {
		if rErr := s.sendReply(ReplyHostUnreachable, nil); rErr != nil {
			// TODO: add log here
			return ErrSendReplyFailed
		}
		// TODO: need more log
		return ErrResolverFailed
	}

	dialer := net.Dialer{Timeout: s.DialTimeout}
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
			// TODO: add log here
			return ErrSendReplyFailed
		}
		// TODO: add log here
		return ErrResolverFailed
	}
	defer target.Close()

	errCh := make(chan error)
	proxy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errCh <- err
	}
	go proxy(target, s.Conn)
	go proxy(s.Conn, target)

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
	addr, err := req.DestAddr.Resolve(ctx)
	if err != nil {
		if rErr := s.sendReply(ReplyHostUnreachable, nil); rErr != nil {
			return fmt.Errorf("faild to send reply: %v", rErr)
		}
		return fmt.Errorf("faild to resolve destination address: %v", err)
	}

	dialer := net.Dialer{Timeout: s.DialTimeout}
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
			return fmt.Errorf("faild to send reply: %v", rErr)
		}
		return fmt.Errorf("faild to dial destination(%v): %v", addr, err)
	}
	defer target.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		// TODO: should add log here
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
	// TODO: send first reply here
	conn, err := ln.Accept()
	if err != nil {
		// TODO: should add log here, and sendReply
		if rErr := s.sendReply(ReplyFailure, nil); rErr != nil {
			return ErrSendReplyFailed
		}
		return err
	}

	errCh := make(chan error)
	proxy := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errCh <- err
	}
	go proxy(conn, target)
	go proxy(target, conn)

	select {
	case <-ctx.Done():
		s.Close()
		err = ctx.Err()
	case nErr := <-errCh:
		err = nErr
	}
	return err
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

// TODO(remones)
func (s *Session) handleCmdUDP(ctx context.Context, req *Request) error {
	return nil
}
