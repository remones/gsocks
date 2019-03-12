package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config ...
type Config struct {
	Host        string `toml:"host"`
	Port        uint   `toml:"port"`
	DialTimeout int    `toml:"dial_timeout"`
	Auth        Auth   `toml:"auth"`
}

// Auth ...
type Auth struct {
	*UserPasswd `toml:"username_password"`
	*GssAPI     `toml:"gss_api"`
	*NoRequired `toml:"no_required"`
}

// GssAPI ...
type GssAPI struct {
	Enable bool `toml:"enable"`
}

// NoRequired ...
type NoRequired struct {
	Enable bool `toml:"enable"`
}

// UserPasswd ...
type UserPasswd struct {
	Enable  bool      `toml:"enable"`
	Account []Account `toml:"account"`
}

// Account ...
type Account struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var defaultConf = Config{
	Host: "0.0.0.0",
	Port: 1080,
	Auth: Auth{
		NoRequired: &NoRequired{
			Enable: true,
		},
	},
}

// NewConfig ...
func NewConfig() *Config {
	conf := defaultConf
	return &conf
}

// Load config with a file
func (c *Config) Load(confFile string) error {
	_, err := toml.DecodeFile(confFile, c)
	if err != nil {
		return fmt.Errorf("Config file decode error: %#v", err)
	}
	return c.validate()
}

func (c *Config) validate() error {
	if c.Auth.UserPasswd != nil {
		for _, account := range c.Auth.UserPasswd.Account {
			if account.Username == "" {
				return fmt.Errorf("[auth]: account username can not be empty string")
			}
		}
	}
	return nil
}
