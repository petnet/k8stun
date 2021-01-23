package k8stun

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Tunnels []Tunnel `yaml:"tunnels"`
}

// LoadConfig ...
func LoadConfig(configFile string) (Config, error) {
	var cfg Config
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, fmt.Errorf(
			"unable to read config file '%s' %v",
			configFile, err)
	}

	type unmarshaler func([]byte, interface{}) error
	var parse unmarshaler = json.Unmarshal
	if strings.HasSuffix(strings.ToLower(configFile), "yaml") ||
		strings.HasSuffix(strings.ToLower(configFile), "yml") {
		parse = yaml.Unmarshal
	}

	err = parse(file, &cfg)
	if err != nil {
		return cfg, fmt.Errorf(
			"unable to parse config file: %s -- %s",
			configFile, err)
	}

	return cfg, nil
}
