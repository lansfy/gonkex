package yaml_file

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Duration wraps time.Duration to support custom unmarshalling
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Duration
// It handles both integer (seconds) and duration strings
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Attempt to unmarshal into an integer first
	var seconds float32
	if err := unmarshal(&seconds); err == nil {
		d.Duration = time.Duration(int(seconds*1000)) * time.Millisecond
		return d.checkPositive()
	}

	// Attempt to unmarshal into a string (duration format)
	var durationStr string
	if err := unmarshal(&durationStr); err == nil {
		parsedDuration, parseErr := time.ParseDuration(durationStr)
		if parseErr != nil {
			return fmt.Errorf("invalid duration string: %w", parseErr)
		}
		d.Duration = parsedDuration
		return d.checkPositive()
	}

	return errors.New("invalid duration value: must be an integer (seconds) or a duration string")
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	unmarshal := func(v interface{}) error {
		return json.Unmarshal(data, v)
	}
	return d.UnmarshalYAML(unmarshal)
}

func (d *Duration) checkPositive() error {
	if d.Duration >= 0 {
		return nil
	}
	d.Duration = 0
	return errors.New("invalid duration value: cannot be negative")
}
