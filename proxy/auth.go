package proxy

import (
	"fmt"
	"io"
	"strings"
)

// Auth ...
const (
	AuthNoRequried  = uint8(0x00)
	AuthGSSAPI      = uint8(0x01)
	AuthUserPass    = uint8(0x02)
	AuthNoAccetable = uint8(0xFF)
)

// UserPass ...
const (
	UserPassSuccess = uint8(0x00)
	UserPassFailure = uint8(0x01)
)

var authenticators = make(map[uint8]Authenticator)

// Authenticator ...
type Authenticator interface {
	Authenticate(rw io.ReadWriter) (ok bool, err error)
}

// GSSAPIAuthenticate ...
type GSSAPIAuthenticate struct{}

// UserPasswd ...
type UserPasswd struct {
	username string
	password string
}

// UserPassAuthenticator ...
type UserPassAuthenticator struct {
	accounts []*UserPasswd
}

// Authenticate ...
func (auth *UserPassAuthenticator) Authenticate(rw io.ReadWriter) (ok bool, err error) {
	header := make([]byte, 2)
	if _, err := rw.Read(header); err != nil {
		return false, err
	}
	ver := uint8(header[0])
	if ver != Socks5Version {
		return false, fmt.Errorf("Invalid version")
	}
	ulen := int(header[1])
	user := make([]byte, ulen)
	if _, err := io.ReadAtLeast(rw, user, ulen); err != nil {
		return false, err
	}
	if _, err := rw.Read(header[:1]); err != nil {
		return false, err
	}
	plen := int(header[0])
	passwd := make([]byte, plen)
	if _, err := io.ReadAtLeast(rw, passwd, plen); err != nil {
		return false, err
	}
	status := auth.verifyAccount(string(user), string(passwd))
	rw.Write([]byte{Socks5Version, status})
	return status == UserPassSuccess, nil
}

func (auth *UserPassAuthenticator) verifyAccount(username, passwd string) (status uint8) {
	for _, account := range auth.accounts {
		if strings.Compare(username, account.username) == 0 && strings.Compare(passwd, account.password) == 0 {
			return UserPassSuccess
		}
	}
	return UserPassFailure
}
