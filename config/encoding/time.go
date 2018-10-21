package encoding

import "time"

// Duration type for config encoding
type Duration struct {
	Value time.Duration
}

// UnmarshalText for toml encoding
func (d *Duration) UnmarshalText(text []byte) (err error) {
	d.Value, err = time.ParseDuration(string(text))
	return err
}
