package proxy

import (
	"io"
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

type ReplyType uint8

const (
	ReplySuccessed          = ReplyType(0x00)
	ReplyFailure            = ReplyType(0x01)
	ReplyNotAllowed         = ReplyType(0x02)
	ReplyNetworkUnreachable = ReplyType(0x03)
	ReplyHostUnreachable    = ReplyType(0x04)
	ReplyConnectionRefused  = ReplyType(0x05)
	ReplyTTLExpired         = ReplyType(0x06)
	ReplyInvalidCommand     = ReplyType(0x07)
	ReplyInvalidAddressType = ReplyType(0x08)
	ReplyUnassigned         = ReplyType(0x09)
)

// Request ...
type Request struct {
	Ver     uint8
	Cmd     uint8
	Rsv     uint8
	Atype   uint8
	DstAddr []byte
	DstPort []byte
}

func NewReuqest(r io.Reader) (*Request, error) {
	req := Request{}
	err := readRequest(r, &req)
	return &req, err
}

func readRequest(r io.Reader, req *Request) error {
	header := make([]byte, 4)
	_, err := r.Read(header)
	if err != nil {
		return err
	}
	req.Ver = header[0]
	req.Cmd = header[1]
	req.Rsv = header[2]
	req.Atype = header[3]

	var addr []byte
	switch req.Atype {
	case TypeIPV4:
		addr, err = parseIPV4(r)
	case TypeIPV6:
		addr, err = parseIPV6(r)
	case TypeFQDN:
		addr, err = parseFQDN(r)
	}
	if err != nil {
		return err
	}
	req.DstAddr = addr
	port := make([]byte, 2)
	if _, err := io.ReadAtLeast(r, port, 2); err != nil {
		return err
	}
	req.DstPort = port
	return nil
}

func parseIPV4(r io.Reader) ([]byte, error) {
	buf := make([]byte, 4)
	_, err := io.ReadAtLeast(r, buf, 4)
	return buf, err
}

func parseIPV6(r io.Reader) ([]byte, error) {
	buf := make([]byte, 16)
	_, err := io.ReadAtLeast(r, buf, 16)
	return buf, err
}

func parseFQDN(r io.Reader) ([]byte, error) {
	h := make([]byte, 1)
	_, err := r.Read(h)
	if err != nil {
		return nil, err
	}
	n := int(h[0])
	buf := make([]byte, int(h[0]))
	_, err = io.ReadAtLeast(r, buf, n)
	return nil, err
}

// Handle ...
func (req *Request) Handle() {
	switch req.Cmd {
	case CmdConnect:
		req.handleCmdConnect()
	case CmdBind:
		req.handleCmdBind()
	case CmdUDPAssociate:
		req.handleCmdUDPAssociate()
	}
}

func (req *Request) handleCmdConnect() {

}

func (req *Request) handleCmdBind() {}

func (req *Request) handleCmdUDPAssociate() {}

// Reply ...
type Reply struct {
	Ver      uint8
	Rep      uint8
	Rsv      []byte
	AType    uint8
	BindAddr []byte
	BindPort []byte
}

// NewReply ...
