package proxy

import (
	"context"
	"net"
	"strconv"
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
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	backendAddr := ln.Addr().String()
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		req := make([]byte, 13)
		conn.Read(req)
		assert.Equal(t, "hello, world!", string(req))
		if _, err = conn.Write([]byte("hello, world!")); err != nil {
			assert.NoError(t, err)
		}
	}()

	server, client := net.Pipe()
	go func() {
		defer server.Close()
		s := &Session{
			Conn: server,
		}
		s.ServeRequest(context.TODO())
	}()

	defer client.Close()
	_, port, _ := net.SplitHostPort(backendAddr)
	nPort, _ := strconv.Atoi(port)
	bPort := make([]byte, 2)
	bPort[0] = uint8(nPort >> 8)
	bPort[1] = uint8(nPort & 255)
	cmd := []byte{5, 1, 0, 1, 127, 0, 0, 1, bPort[0], bPort[1]}
	_, err = client.Write(cmd)
	assert.NoError(t, err)

	req := []byte("hello, world!")
	_, err = client.Write(req)
	assert.NoError(t, err)

	rsp := make([]byte, 13)
	_, err = client.Read(rsp)
	assert.NoError(t, err)
	assert.Equal(t, "hello, world!", string(rsp))
}
