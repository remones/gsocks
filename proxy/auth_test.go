package proxy

import (
	"bytes"
	"io"
	"testing"
)

func TestUserPassAuthenticator_Authenticate(t *testing.T) {
	type fields struct {
		version  uint8
		accounts []*UserPasswd
	}
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantOk  bool
		wantW   string
		wantErr bool
	}{
		{
			name: "userpasswd_matched",
			fields: fields{
				version: uint8(0x05),
				accounts: []*UserPasswd{
					&UserPasswd{
						username: "san.zhang",
						password: "1234",
					},
					&UserPasswd{
						username: "si.li",
						password: "4321",
					},
				},
			},
			args: args{
				r: bytes.NewBuffer([]byte{5, 5, 's', 'i', '.', 'l', 'i', 4, '4', '3', '2', '1'}),
			},
			wantOk:  true,
			wantW:   string([]byte{5, 1}),
			wantErr: false,
		},
		{
			name: "userpasswd_unmatch",
			fields: fields{
				version: uint8(0x05),
				accounts: []*UserPasswd{
					&UserPasswd{
						username: "san.zhang",
						password: "1234",
					},
				},
			},
			args: args{
				r: bytes.NewBuffer([]byte{5, 5, 's', 'i', '.', 'l', 'i', 4, '4', '3', '2', '1'}),
			},
			wantOk:  false,
			wantW:   string([]byte{5, 0}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &UserPassAuthenticator{
				version:  tt.fields.version,
				accounts: tt.fields.accounts,
			}
			w := &bytes.Buffer{}
			gotOk, err := auth.Authenticate(tt.args.r, w)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserPassAuthenticator.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("UserPassAuthenticator.Authenticate() = %v, want %v", gotOk, tt.wantOk)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("UserPassAuthenticator.Authenticate() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
