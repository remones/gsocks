package config

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Load("./config.toml.example")
	assert.NoError(t, err)
	runtime.Breakpoint()
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, uint(1080), cfg.Port)
}
