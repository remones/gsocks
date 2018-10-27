package proxy

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_readRequest(t *testing.T) {
	type args struct {
		r   io.Reader
		req *Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantReq *Request
	}{
		{
			name: "ipv4_success",
			args: args{
				r: bytes.NewBuffer([]byte{5, 1, 0, 1, 127, 0, 0, 1, uint8(4), uint8(56)}),
			},
			wantReq: &Request{
				Version:    Socks5Version,
				Command:    CmdConnect,
				RemoteAddr: &AddrSpec{},
				DestAddr: &AddrSpec{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 1080,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tReq, err := readRequest(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("readRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantReq.DestAddr.IP.String(), tReq.DestAddr.IP.String())
			assert.Equal(t, tt.wantReq.DestAddr.FQDN, tReq.DestAddr.FQDN)
			assert.Equal(t, tt.wantReq.DestAddr.Port, tReq.DestAddr.Port)
		})
	}
}
