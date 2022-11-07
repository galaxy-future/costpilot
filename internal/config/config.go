package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/galaxy-future/costpilot/internal/constants/cloud"
	"github.com/galaxy-future/costpilot/internal/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CloudAccounts []types.CloudAccount `json:"cloud_accounts" yaml:"cloud_accounts"`
}

var globalConfig *Config

const (
	COSTPILOT_PROVIDER  = "COSTPILOT_PROVIDER"
	COSTPILOT_AK        = "COSTPILOT_AK"
	COSTPILOT_SK        = "COSTPILOT_SK"
	COSTPILOT_REGION_ID = "COSTPILOT_REGION_ID"
)

func Init() error {
	if err := InitFromEnvConfig(); err == nil { // find variables in environment firstly
		return nil
	}
	if err := InitFileConfig(); err != nil {
		return err
	}
	return nil
}

// InitFromEnvConfig lood configuration from environment
func InitFromEnvConfig() error {
	p := os.Getenv(COSTPILOT_PROVIDER)
	ak := os.Getenv(COSTPILOT_AK)
	sk := os.Getenv(COSTPILOT_SK)
	r := os.Getenv(COSTPILOT_REGION_ID)
	globalConfig = &Config{
		CloudAccounts: []types.CloudAccount{
			{
				Provider: cloud.Provider(p),
				AK:       ak,
				SK:       sk,
				RegionID: r,
				Name:     ak,
			},
		},
	}
	if err := globalConfig.verify(); err != nil {
		log.Printf("I! no valid environment variables, skip")
		return err
	}
	log.Println("I! load env config success")
	return nil
}

// InitFileConfig load configuration from config file
func InitFileConfig(filePath ...string) (err error) {
	globalConfig, err = loadConfig(filePath...)
	if err != nil {
		log.Printf("E! load from config file error, %v", err)
		return err
	}
	return nil
}

func loadConfig(filePath ...string) (*Config, error) {
	var confPath string
	if len(filePath) == 0 {
		confPath = "conf/config.yaml"
	} else {
		confPath = filePath[0]
	}

	f, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err = yaml.Unmarshal(f, &config); err != nil {
		return nil, err
	}
	if err = config.verify(); err != nil {
		return nil, err
	}
	log.Println("I! load file config success")
	return &config, nil
}

// verify check the variables in the config
func (c Config) verify() error {
	if len(c.CloudAccounts) == 0 {
		return errors.New("cloud_accounts config is required")
	}
	for _, account := range c.CloudAccounts {
		if account.AK == "" || account.SK == "" || account.Provider == "" {
			return errors.New("cloud_account ak/sk/provider/name config is required")
		}
		if account.Provider.String() == cloud.Undefined {
			return fmt.Errorf("invalid provider")
		}
		if account.Name == "" {
			account.Name = account.AK
		}
	}
	return nil
}

func GetGlobalConfig() *Config {
	return globalConfig
}
