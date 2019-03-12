package proxy

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestServer_ListenAndServe(t *testing.T) {
	type fields struct {
		addr           string
		listener       net.Listener
		mu             sync.Mutex
		waitConns      sync.WaitGroup
		inShutdown     int32
		doneChan       chan struct{}
		authenticators map[AuthType]Authenticator
		DialTimeout    time.Duration
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
			srv := &Server{
				addr:           tt.fields.addr,
				listener:       tt.fields.listener,
				mu:             tt.fields.mu,
				waitConns:      tt.fields.waitConns,
				inShutdown:     tt.fields.inShutdown,
				doneChan:       tt.fields.doneChan,
				authenticators: tt.fields.authenticators,
				DialTimeout:    tt.fields.DialTimeout,
			}
			if err := srv.ListenAndServe(); (err != nil) != tt.wantErr {
				t.Errorf("Server.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
