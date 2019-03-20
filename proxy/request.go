package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
)

// Command type of request
const (
	CmdConnect = uint8(0x01)
	CmdBind    = uint8(0x02)
	CmdUDP     = uint8(0x03)
)

// AType represent address type
const (
	TypeIPV4 = uint8(0x01)
	TypeFQDN = uint8(0x03)
	TypeIPV6 = uint8(0x04)
)

// AddrSpec ...
type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
	Type uint8
}

// NewAddrSpec ...
func NewAddrSpec(r io.Reader) (*AddrSpec, error) {
	return readAddrSpec(r)
}

func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	addr := new(AddrSpec)
	h := make([]byte, 2)

	if _, err := r.Read(h[:1]); err != nil {
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
		return nil, fmt.Errorf("Unknow Address Type: %d", h[0])
	}
	// Read Port
	if _, err := io.ReadAtLeast(r, h, 2); err != nil {
		return nil, err
	}
	addr.Type = h[0]
	addr.Port = (int(h[0])<<8 | int(h[1]))
	return addr, nil
}

// Resolve ...
func (as *AddrSpec) Resolve(ctx context.Context) (string, error) {
	ip, err := as.resolveIPAddr()
	if err != nil {
		return "", err
	}
	addr := net.JoinHostPort(ip.String(), strconv.Itoa(as.Port))
	return addr, nil
}

func (as *AddrSpec) resolveIPAddr() (net.IP, error) {
	var ip = as.IP
	if as.FQDN != "" {
		ipAddr, err := net.ResolveIPAddr("ip", as.FQDN)
		if err != nil {
			return nil, err
		}
		ip = ipAddr.IP
	}
	return ip, nil
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
	dest, err := NewAddrSpec(r)
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

// NewReply ...
