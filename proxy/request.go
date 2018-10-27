package proxy

import (
	"fmt"
	"io"
	"net"
)

// Command type of request
const (
	CmdConnect      = uint8(0x01)
	CmdBind         = uint8(0x02)
	CmdUDPAssociate = uint8(0x03)
)

// AType represent address type
const (
	TypeIPV4 = uint8(0x01)
	TypeFQDN = uint8(0x03)
	TypeIPV6 = uint8(0x04)
)

// Reply ...
const (
	ReplySuccessed          = uint8(0x00)
	ReplyFailure            = uint8(0x01)
	ReplyNotAllowed         = uint8(0x02)
	ReplyNetworkUnreachable = uint8(0x03)
	ReplyHostUnreachable    = uint8(0x04)
	ReplyConnectionRefused  = uint8(0x05)
	ReplyTTLExpired         = uint8(0x06)
	ReplyInvalidCommand     = uint8(0x07)
	ReplyInvalidAddressType = uint8(0x08)
	ReplyUnassigned         = uint8(0x09)
)

// AddrSpec ...
type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

// Request ...
type Request struct {
	Version    uint8
	Command    uint8
	RemoteAddr *AddrSpec
	DestAddr   *AddrSpec
}

// NewReuqest ...
func NewReuqest(r io.Reader) (*Request, error) {
	return readRequest(r)
}

func readRequest(r io.Reader) (*Request, error) {
	header := make([]byte, 3)
	_, err := r.Read(header)
	if err != nil {
		return nil, err
	}
	ver := header[0]
	cmd := header[1]
	dest, err := readAddrSpec(r)
	if err != nil {
		return nil, err
	}
	req := &Request{
		Version:  ver,
		Command:  cmd,
		DestAddr: dest,
	}
	return req, nil
}

func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	addr := new(AddrSpec)
	h := make([]byte, 1)

	if _, err := r.Read(h); err != nil {
		return nil, err
	}
	switch h[0] {
	case TypeIPV4:
		buf := make([]byte, 4)
		if _, err := io.ReadAtLeast(r, buf, 4); err != nil {
			return nil, err
		}
		addr.IP = buf
	case TypeIPV6:
		buf := make([]byte, 16)
		if _, err := io.ReadAtLeast(r, buf, 16); err != nil {
			return nil, err
		}
		addr.IP = buf
	case TypeFQDN:
		if _, err := r.Read(h); err != nil {
			return nil, err
		}
		n := int(h[0])
		buf := make([]byte, n)
		if _, err := io.ReadAtLeast(r, buf, n); err != nil {
			return nil, err
		}
		addr.FQDN = string(buf)
	default:
		return nil, fmt.Errorf("Unknow AType: %d", h[0])
	}
	return addr, nil
}

// NewReply ...
