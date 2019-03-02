package proxy

import (
	"context"
	"net"
	"sync"
	"testing"
)

func Test_server_ListenAndServe(t *testing.T) {
	type fields struct {
		addr           string
		mu             sync.Mutex
		waitConns      sync.WaitGroup
		inShutdown     int32
		doneChan       chan struct{}
		authenticators map[byte]Authenticator
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &server{
				addr:           tt.fields.addr,
				mu:             tt.fields.mu,
				waitConns:      tt.fields.waitConns,
				inShutdown:     tt.fields.inShutdown,
				doneChan:       tt.fields.doneChan,
				authenticators: tt.fields.authenticators,
			}
			if err := srv.ListenAndServe(); (err != nil) != tt.wantErr {
				t.Errorf("server.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_onceCloseListener_Close(t *testing.T) {
	type fields struct {
		Listener net.Listener
		once     sync.Once
		closeErr error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ln := &onceCloseListener{
				Listener: tt.fields.Listener,
				once:     tt.fields.once,
				closeErr: tt.fields.closeErr,
			}
			if err := ln.Close(); (err != nil) != tt.wantErr {
				t.Errorf("onceCloseListener.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_Close(t *testing.T) {
	type fields struct {
		addr           string
		listener       net.Listener
		mu             sync.Mutex
		waitConns      sync.WaitGroup
		inShutdown     int32
		doneChan       chan struct{}
		authenticators map[byte]Authenticator
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
			srv := &server{
				addr:           tt.fields.addr,
				listener:       tt.fields.listener,
				mu:             tt.fields.mu,
				waitConns:      tt.fields.waitConns,
				inShutdown:     tt.fields.inShutdown,
				doneChan:       tt.fields.doneChan,
				authenticators: tt.fields.authenticators,
			}
			if err := srv.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("server.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
