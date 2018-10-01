package proxy

// Command type of request
type CmdType uint8

const (
	CmdConnect = CmdType(0x01)
	CmdBind    = CmdType(0x02)
	CmdUDP     = CmdType(0x03)
)

// AType represent address type
type AType uint8

const (
	TypeIPV4       = AType(0x01)
	TypeDomainName = AType(0x03)
	TypeIPV6       = AType(0x04)
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
	Cmd     CmdType
	Rsv     uint8
	DstAddr []byte
	DstPort []byte
}

// NewRequest ...
func NewRequest() {}

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
func NewReply() {}
