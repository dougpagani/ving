package config

import (
	"fmt"

	"github.com/yittg/ving/addons/port/types"
	"github.com/yittg/ving/errors"
)

// PortsConfig for custom
type PortsConfig struct {
	Extra            []types.PortDesc
	ProbeConcurrency int `toml:"probe-concurrency"`
}

// Validate ports config
func (c *PortsConfig) Validate() error {
	if c.ProbeConcurrency <= 0 || c.ProbeConcurrency >= 1024 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("ports probe concurrency should in range [1,1023], (probe-concurrency=%d)", c.ProbeConcurrency),
		}
	}
	return nil
}

// Default config of ports add-on
func Default() PortsConfig {
	return PortsConfig{
		ProbeConcurrency: 1023,
	}
}
