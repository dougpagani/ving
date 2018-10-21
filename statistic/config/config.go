package config

import (
	"fmt"
	"time"

	c "github.com/yittg/ving/config/encoding"
	"github.com/yittg/ving/errors"
)

// Config of statistic
type Config struct {
	ErrorRateThresh []float64 `toml:"error-rate-thresh"`
	Window          c.Duration
}

// Validate the statistic config
func (c *Config) Validate() error {
	if len(c.ErrorRateThresh) == 0 {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("empty error rate threshold array, (error-rate-thresh=%+v)", c.ErrorRateThresh),
		}
	}
	r := 0.0
	for _, thresh := range c.ErrorRateThresh {
		if r-thresh >= 0.00001 {
			return &errors.ConfigError{
				Msg: fmt.Sprintf("unordered error rate threshold array, (error-rate-thresh=%v)", c.ErrorRateThresh),
			}
		}
		r = thresh
	}

	if c.Window.Value < time.Second {
		return &errors.ConfigError{
			Msg: fmt.Sprintf("invalid statistic window, should longer than 1s, (window=%v)", c.Window),
		}
	}
	return nil
}

// Default config of statistic
func Default() Config {
	return Config{
		ErrorRateThresh: []float64{0.01, 0.1, 1},
		Window: c.Duration{
			Value: 10 * time.Second,
		},
	}
}
