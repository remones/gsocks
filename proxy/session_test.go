package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setAuthenticator(method uint8, auth Authenticator) {
	authenticators[method] = auth
}

func resetAuthenticator() {
	authenticators = make(map[uint8]Authenticator)
}

func TestSessionAuthenticate(t *testing.T) {
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

func TestSession_handleCmdConnect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	backendAddr := ln.Addr().String()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		req := make([]byte, 13)
		conn.Read(req)
		assert.Equal(t, "hello, world!", string(req))
		_, err = conn.Write([]byte("hello, world!"))
		assert.NoError(t, err)
	}()

	server, client := net.Pipe()
	go func() {
		defer server.Close()
		s := &Session{
			Conn:        server,
			DialTimeout: 3 * time.Second,
		}
		s.ServeRequest(context.TODO())
	}()
	defer client.Close()

	// TODO: 封装这个 sendCmd 的函数
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

// TODO: 两组
func TestSession_handleCmdBind(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	clientListenAddr := ln.Addr().String()

	result := make(chan string)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		req := make([]byte, 13)
		conn.Read(req)
		result <- string(req)
	}()

	server, client := net.Pipe()
	go func() {
		defer server.Close()
		s := &Session{
			Conn:        server,
			DialTimeout: 3 * time.Second,
		}
		s.ServeRequest(context.TODO())
	}()
	defer client.Close()

	_, port, _ := net.SplitHostPort(clientListenAddr)
	nPort, _ := strconv.Atoi(port)
	bPort := make([]byte, 2)
	bPort[0] = uint8(nPort >> 8)
	bPort[1] = uint8(nPort & 255)
	cmd := []byte{5, 2, 0, 1, 127, 0, 0, 1, bPort[0], bPort[1]}
	_, err = client.Write(cmd)
	assert.NoError(t, err)

	reply1 := make([]byte, 10)
	client.Read(reply1)
	rport := (int(reply1[8])<<8 | int(reply1[9]))

	connSrv, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", rport))
	assert.NoError(t, err)

	connSrv.Write([]byte("hello, world!"))
	assert.Equal(t, "hello, world!", <-result)
}
