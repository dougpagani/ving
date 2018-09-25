package config

import (
	"os"

	"github.com/BurntSushi/toml"
	ports "github.com/yittg/ving/addons/port/types"
)

var searchDir = []string{".", os.Getenv("HOME")}

// Config custom
type Config struct {
	AddOns AddOnConfig `toml:"add-ons"`
}

// AddOnConfig add on configs
type AddOnConfig struct {
	Ports ports.PortsConfig
}

var customConfig Config

// GetConfig get custom config
func GetConfig() *Config {
	return &customConfig
}

func init() {
	for _, rcDir := range searchDir {
		rcFile := rcDir + "/.ving.rc"
		if _, err := os.Stat(rcFile); os.IsNotExist(err) {
			continue
		}
		if _, err := toml.DecodeFile(rcFile, &customConfig); err != nil {
			panic(err)
		}
		return
	}

}
