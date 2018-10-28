package proxy

import (
	"bytes"
	"io"
	"testing"
)

func TestUserPassAuthenticator_Authenticate(t *testing.T) {
	type fields struct {
		accounts []*UserPasswd
	}
	type args struct {
		rw io.ReadWriter
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
				accounts: []*UserPasswd{
					&UserPasswd{
						username: "san.zhang",
						password: "1234",
					},
					&UserPasswd{
						username: "si.li",
						password: "1234",
					},
				},
			},
			args: args{
				rw: bytes.NewBuffer([]byte{5, 5, 's', 'i', '.', 'l', 'i', 4, '1', '2', '3', '4'}),
			},
			wantOk:  true,
			wantW:   string([]byte{5, 0}),
			wantErr: false,
		},
		{
			name: "userpasswd_unmatch",
			fields: fields{
				accounts: []*UserPasswd{
					&UserPasswd{
						username: "san.zhang",
						password: "1234",
					},
				},
			},
			args: args{
				rw: bytes.NewBuffer([]byte{5, 5, 's', 'i', '.', 'l', 'i', 4, '4', '3', '2', '1'}),
			},
			wantOk:  false,
			wantW:   string([]byte{5, 1}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &UserPassAuthenticator{
				accounts: tt.fields.accounts,
			}
			gotOk, err := auth.Authenticate(tt.args.rw)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserPassAuthenticator.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("UserPassAuthenticator.Authenticate() = %v, want %v", gotOk, tt.wantOk)
			}
			buf := tt.args.rw.(*bytes.Buffer)
			if gotW := buf.String(); gotW != tt.wantW {
				t.Errorf("UserPassAuthenticator.Authenticate() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
