package config

import (
	"fmt"

	"github.com/yittg/ving/errors"
)

// UIConfig for custom chart board
type UIConfig struct {
	MaxRow          int `toml:"max-row"`
	SparklineHeight int `toml:"sp-height"`
}

// Validate UIConfig
func (c *UIConfig) Validate() error {
	if c.MaxRow <= 0 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("max row of ui should be positive(max-row=%d)", c.MaxRow),
		}
	}
	if c.SparklineHeight < 0 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("single chart line height should not be negative(sp-height=%d)", c.SparklineHeight),
		}
	}
	return nil
}

// Default config of UIConfig
func Default() UIConfig {
	return UIConfig{
		MaxRow:          4,
		SparklineHeight: 3,
	}
}
