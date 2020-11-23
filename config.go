package sailtrim

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/Songmu/prompter"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/kayac/go-config"
)

// Config represents configurations for SailTrim
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

func (c *Config) dumpService(sv *lightsail.ContainerService) error {
	b, err := MarshalJSON(&lightsail.ContainerService{
		ContainerServiceName: sv.ContainerServiceName,
		Power:                sv.Power,
		Scale:                sv.Scale,
	})
	if err != nil {
		return err
	}
	return saveFile(c.Service, b, os.FileMode(0644))
}

func (c *Config) dumpDeployment(dp *lightsail.ContainerServiceDeployment) error {
	b, err := MarshalJSON(&lightsail.ContainerServiceDeployment{
		Containers:     dp.Containers,
		PublicEndpoint: dp.PublicEndpoint,
	})
	if err != nil {
		return err
	}
	return saveFile(c.Deployment, b, os.FileMode(0644))
}

func saveFile(path string, b []byte, mode os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		ok := prompter.YN(fmt.Sprintf("Overwrite existing file %s?", path), false)
		if !ok {
			log.Println("[warn] skipping", path)
			return nil
		}
	}
	log.Printf("[info] writing file %s", path)
	return ioutil.WriteFile(path, b, mode)
}

func printFile(path string, w io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func (c *Config) printService(w io.Writer) error {
	return printFile(c.Service, w)
}

func (c *Config) printDeployment(w io.Writer) error {
	return printFile(c.Deployment, w)
}
