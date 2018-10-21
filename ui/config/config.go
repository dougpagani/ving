package config

import (
	"fmt"

	"github.com/yittg/ving/errors"
)

// UIConfig for custom chart board
type UIConfig struct {
	MaxRow          int `toml:"max-chart-row"`
	SparklineHeight int `toml:"chart-height"`
}

// Validate UIConfig
func (c *UIConfig) Validate() error {
	if c.MaxRow <= 0 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("max row of chart should be positive(max-chart-row=%d)", c.MaxRow),
		}
	}
	if c.SparklineHeight < 0 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("single chart line height should not be negative(chart-height=%d)", c.SparklineHeight),
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
