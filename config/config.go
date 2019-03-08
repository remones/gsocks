package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config ...
type Config struct {
	Host string `toml:"host"`
	Port uint   `toml:"port"`
	Auth auth   `toml:"auth"`
}

type auth struct {
	Methods []string  `toml:"methods"`
	Info    *AuthInfo `toml:"info"`
}

// AuthInfo ...
type AuthInfo struct {
	*UserPasswd `toml:"username_password"`
}

// UserPasswd ...
type UserPasswd struct {
	Account []*account `toml:"account"`
}

type account struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var defaultConf = Config{
	Host: "0.0.0.0",
	Port: 1080,
	Auth: auth{
		Methods: []string{"no_required"},
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
	methods := c.Auth.Methods
	i := 0
	flag := make(map[string]struct{})
	for _, m := range methods {
		switch m {
		case "no_required", "gss_api", "username_password":
			if _, ok := flag[m]; !ok {
				methods[i] = m
				i++
			}
		}
	}
	c.Auth.Methods = methods[:i]

	if _, exists := flag["username_password"]; exists {
		if c.Auth.Info == nil || c.Auth.Info.UserPasswd == nil {
			return fmt.Errorf("[auth]: username_password should be set account information")
		}
		for _, account := range c.Auth.Info.UserPasswd.Account {
			if account.Username == "" {
				return fmt.Errorf("[auth]: account username can not be empty string")
			}
		}
	}
	return nil
}
