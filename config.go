package k8stun

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Tunnels []Tunnel
}

// LoadConfig ...
func LoadConfig(configFile string) (Config, error) {
	var cfg Config
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, fmt.Errorf(
			"Unable to read config file '%s' %v\n",
			configFile, err)
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, fmt.Errorf(
			"Unable to parse config file: %s -- %s\n",
			configFile, err)
	}

	return cfg, nil
}
