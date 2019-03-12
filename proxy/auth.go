package proxy

import (
	"errors"
	"fmt"
	"io"

	"github.com/remones/gsocks/config"
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

func makeAuthsWithConfig(authCfg *config.Auth) map[AuthType]Authenticator {
	auths := make(map[AuthType]Authenticator)

	if authCfg.UserPasswd != nil && authCfg.UserPasswd.Enable {
		accounts := make(map[string]string)
		for _, account := range authCfg.UserPasswd.Account {
			accounts[account.Username] = account.Password
		}
		auths[AuthUserPass] = &UserPassAuthenticator{accounts}
	}

	if authCfg.NoRequired != nil && authCfg.NoRequired.Enable {
		auths[AuthNoRequried] = &AuthNoRequired{}
	}
	return auths
	// TODO: add gss_api
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

// UserPassAuthenticator ...
type UserPassAuthenticator struct {
	accounts map[string]string
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
	if p, ok := auth.accounts[username]; ok && passwd == p {
		return UserPassSuccess
	}
	return UserPassFailure
}

// AuthNoRequired ...
type AuthNoRequired struct{}

// Authenticate ...
func (auth *AuthNoRequired) Authenticate(rw io.ReadWriter) (ok bool, err error) {
	return true, nil
}

// Type ...
func (auth *AuthNoRequired) Type() AuthType {
	return AuthNoRequried
}
