package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Load("./config.toml.example")
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, uint(1080), cfg.Port)
	assert.Equal(t, []string{"username_password"}, cfg.AuthMethods)
	assert.Equal(t, "test", cfg.AuthInfo.UserPasswd.Username)
	assert.Equal(t, "test", cfg.AuthInfo.UserPasswd.Password)
}
