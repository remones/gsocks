package proxy

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setAuthenticator(method uint8, auth Authenticator) {
	authenticators[method] = auth
}

func resetAuthenticator() {
	authenticators = make(map[uint8]Authenticator)
}

func TestSession_Authenticate(t *testing.T) {
	setAuthenticator(AuthUserPass, &UserPassAuthenticator{
		accounts: []*UserPasswd{
			&UserPasswd{
				username: "si.li",
				password: "1234",
			},
		},
	})
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		defer client.Close()
		_, err := client.Write([]byte{1, uint8(0x02)})
		assert.NoError(t, err)

		rsp := make([]byte, 2)
		n, err := client.Read(rsp)
		assert.Equal(t, 2, n)
		assert.NoError(t, err)
		assert.Equal(t, uint8(5), rsp[0])
		assert.Equal(t, uint8(2), rsp[1])

		_, err = client.Write([]byte{5, 5, 's', 'i', '.', 'l', 'i', 4, '1', '2', '3', '4'})
		assert.NoError(t, err)
		// verify the status of response
		n, err = client.Read(rsp)
		assert.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.Equal(t, uint8(5), rsp[0])
		assert.Equal(t, uint8(0), rsp[1])
	}()

	s := &Session{
		Conn: server,
	}
	gotOk, err := s.Authenticate()
	assert.NoError(t, err)
	assert.Equal(t, true, gotOk)
}

func TestSession_ServeRequest(t *testing.T) {
	type fields struct {
		Conn net.Conn
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Conn: tt.fields.Conn,
			}
			if err := s.ServeRequest(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Session.ServeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
