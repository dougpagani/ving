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
	UI     UIConfig
}

// UIConfig for custom chart board
type UIConfig struct {
	MaxRow int `toml:"max-row"`
}

// AddOnConfig add on configs
type AddOnConfig struct {
	Ports ports.PortsConfig
}

var customConfig *Config

// GetConfig get custom config
func GetConfig() *Config {
	return customConfig
}

func init() {
	customConfig = &Config{
		UI: UIConfig{
			MaxRow: 4,
		},
	}
	for _, rcDir := range searchDir {
		rcFile := rcDir + "/.ving.toml"
		if _, err := os.Stat(rcFile); os.IsNotExist(err) {
			continue
		}
		if _, err := toml.DecodeFile(rcFile, customConfig); err != nil {
			panic(err)
		}
		return
	}

}
