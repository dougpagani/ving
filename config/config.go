package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	ports "github.com/yittg/ving/addons/port/types"
	ui "github.com/yittg/ving/ui/config"
)

var searchDir = []string{".", os.Getenv("HOME")}

// Config custom
type Config struct {
	AddOns AddOnConfig `toml:"add-ons"`
	UI     ui.UIConfig
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

func validate() error {
	if err := customConfig.UI.Validate(); err != nil {
		return err
	}
	return nil
}

func init() {
	customConfig = &Config{
		UI: ui.Default(),
	}
	for _, rcDir := range searchDir {
		rcFile := rcDir + "/.ving.toml"
		if _, err := os.Stat(rcFile); os.IsNotExist(err) {
			continue
		}
		if _, err := toml.DecodeFile(rcFile, customConfig); err != nil {
			panic(err)
		}
		if err := validate(); err != nil {
			fmt.Printf("Invalid custom configuration file: %s\n%s\n", rcFile, err)
			os.Exit(1)
		}
		return
	}
}
