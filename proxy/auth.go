package proxy

import (
	"fmt"
	"io"
	"strings"
)

const (
	AuthNoRequried  = uint8(0x00)
	AuthGSSAPI      = uint8(0x01)
	AuthUserPass    = uint8(0x02)
	AuthNoAccetable = uint8(0xFF)
)

const (
	UserPassSuccess = uint8(0x01)
	UserPassFailure = uint8(0x00)
)

var authenticators map[uint8]authenticator

type authenticator interface {
	Authenticate(r io.Reader) (ok bool, err error)
}

// GSSAPIAuthenticate ...
type GSSAPIAuthenticate struct{}

// UserPasswd ...
type UserPasswd struct {
	username string
	passwd   string
}

// UserPassAuthenticator ...
type UserPassAuthenticator struct {
	version  uint8
	accounts []*UserPasswd
}

// Authenticate ...
func (auth *UserPassAuthenticator) Authenticate(r io.Reader, w io.Writer) (ok bool, err error) {
	header := make([]byte, 0, 2)
	if _, err := r.Read(header); err != nil {
		return false, err
	}
	ver := uint8(header[0])
	if ver != auth.version {
		return false, fmt.Errorf("Invalid version")
	}
	ulen := int(header[1])
	user := make([]byte, 0, ulen)
	if _, err := io.ReadAtLeast(r, user, ulen); err != nil {
		return false, err
	}
	if _, err := r.Read(header[:1]); err != nil {
		return false, err
	}
	plen := int(header[0])
	passwd := make([]byte, 0, plen)
	if _, err := io.ReadAtLeast(r, passwd, plen); err != nil {
		return false, err
	}
	status := auth.verifyAccount(string(user), string(passwd))
	w.Write([]byte{auth.version, status})
	return status == UserPassSuccess, nil
}

func (auth *UserPassAuthenticator) verifyAccount(username, passwd string) (status uint8) {
	for _, account := range auth.accounts {
		if strings.Compare(username, account.username) == 0 && strings.Compare(passwd, account.passwd) == 0 {
			return UserPassSuccess
		}
	}
	return UserPassFailure
}
