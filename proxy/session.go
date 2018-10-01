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
	return nil
}

// HandleRequest ...
func (s *Session) HandleRequest() error {
	return nil
}

func (s *Session) readClientAddr(ctx context.Context) error {
	return nil
}
