package proxy

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// errors
var (
	ErrUnsupportAuthType = errors.New("unsupported auth type")
)

// AuthType ...
type AuthType uint8

// AuthType
const (
	AuthNoRequried  = AuthType(0x00)
	AuthGSSAPI      = AuthType(0x01)
	AuthUserPass    = AuthType(0x02)
	AuthNoAccetable = AuthType(0xFF)
)

var authNames = map[AuthType]string{
	AuthNoRequried: "no_required",
	AuthGSSAPI:     "gss_api",
	AuthUserPass:   "username_password",
}

var authReverseMap = func() map[string]AuthType {
	m := make(map[string]AuthType)
	for key, val := range authNames {
		m[val] = key
	}
	return m
}()

func (at *AuthType) String() string {
	return authNames[*at]
}

// GetAuthType ...
func GetAuthType(name string) (AuthType, bool) {
	t, ok := authReverseMap[name]
	return t, ok
}

func newAuthenticator(name string, info map[string]interface{}) Authenticator {
	// TODO:
	switch name {
	case "username_password":
		// TODO:
		acnts := info["accounts"].([]map[string]string)
		accounts := make([]*userPasswd, len(info["accounts"]))
		for i, item := range info["accounts"].([]map[string]string) {
			accounts[i] = &userPasswd{
				username: item["username"],
				password: item["password"],
			}
		}
		return &UserPassAuthenticator{accounts}
	}
	return nil
}

// UserPass ...
const (
	UserPassSuccess = uint8(0x00)
	UserPassFailure = uint8(0x01)
)

// Authenticator ...
type Authenticator interface {
	Type() AuthType
	Authenticate(rw io.ReadWriter) (ok bool, err error)
}

// GSSAPIAuthenticate ...
type GSSAPIAuthenticate struct{}

// UserPasswd ...
type userPasswd struct {
	username string
	password string
}

// UserPassAuthenticator ...
type UserPassAuthenticator struct {
	accounts []*userPasswd
}

// Type ...
func (*UserPassAuthenticator) Type() AuthType {
	return AuthUserPass
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
