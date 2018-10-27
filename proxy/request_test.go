package proxy

import (
	"bytes"
	"io"
	"testing"
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
			name: "success test",
			args: args{
				r:   bytes.NewBuffer([]byte{1, 1, 0, 1, '1', '2', '7', '.', '0', '.', '0', '.', '1', uint8(1080)}),
				req: new(Request),
			},
			wantReq: &Request{
				Ver:     Socks5Version,
				Cmd:     CmdConnect,
				Rsv:     uint8(0),
				Atype:   TypeIPV4,
				DstAddr: []byte("127.0.0.1"),
				DstPort: []byte("1080"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readRequest(tt.args.r, tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("readRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
