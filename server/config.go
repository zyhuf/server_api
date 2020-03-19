package server

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const CONFIG_FILE_PATH = "/opt/reyzar/server-api/config.yaml"

type Config struct {
	ServerAddr    string
	DhcpAgentAddr []string
	DbAddr        []string
	KeyPrefix     string
}

func GetConfig() (conf *Config, err error) {
	buff, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buff, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
