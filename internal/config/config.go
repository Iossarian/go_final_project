package config

import (
	"fmt"
	"log"
	"net"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	GRPC struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"grpc"`

	Service struct {
		LoginMaxLoad    int `yaml:"loginMaxLoad"`
		PasswordMaxLoad int `yaml:"passwordMaxLoad"`
		IPMaxLoad       int `yaml:"ipMaxLoad"`
	} `yaml:"service"`

	Redis struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"redis"`
}

func NewConfig(configPath string) *Config {
	f, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("fail to open config fail: %w", err))
	}

	config := &Config{}
	if err := yaml.Unmarshal(f, config); err != nil {
		log.Println(fmt.Errorf("fail to decode config file: %w", err))
	}

	return config
}

func (c *Config) GetRedisAddr() string {
	return net.JoinHostPort(c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetGRPCAddr() string {
	return net.JoinHostPort(c.GRPC.Host, c.GRPC.Port)
}

func (c *Config) GetLoginMaxLoad() int {
	return c.Service.LoginMaxLoad
}

func (c *Config) GetPasswordMaxLoad() int {
	return c.Service.PasswordMaxLoad
}

func (c *Config) GetIPMaxLoad() int {
	return c.Service.IPMaxLoad
}
