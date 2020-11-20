package sailtrim

import (
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/kayac/go-config"
)

type Config struct {
	Service    string `json:"service"`
	Deployment string `json:"deployment"`
}

func loadConfig(path string) (*Config, error) {
	var c Config
	if err := config.LoadWithEnv(&c, path); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) loadService() (*lightsail.ContainerService, error) {
	var sv lightsail.ContainerService
	if err := config.LoadWithEnvJSON(&sv, c.Service); err != nil {
		return nil, err
	}
	return &sv, nil
}

func (c *Config) loadDeployment() (*lightsail.ContainerServiceDeployment, error) {
	var dp lightsail.ContainerServiceDeployment
	if err := config.LoadWithEnvJSON(&dp, c.Deployment); err != nil {
		return nil, err
	}
	return &dp, nil
}
