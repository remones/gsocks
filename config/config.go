package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Host        string    `toml:"host"`
	Port        uint      `toml:"port"`
	AuthMethods []string  `toml:"auth_methods"`
	AuthInfo    *authInfo `toml:"auth"`
}

type authInfo struct {
	UserPasswd *userpasswd `toml:"username_password"`
}

type userpasswd struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var defaultConf = Config{
	Host:        "0.0.0.0",
	Port:        1080,
	AuthMethods: []string{"no_required"},
}

// NewConfig ...
func NewConfig() *Config {
	conf := defaultConf
	return &conf
}

// Load ...
func (c *Config) Load(confFile string) error {
	_, err := toml.DecodeFile(confFile, c)
	return err
}
