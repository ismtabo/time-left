package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	duration, error := time.ParseDuration(value.Value)
	if error != nil {
		return error
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

type Config interface {
	// Reload config
	Reload() error
	// Get work day start time
	GetWorkDayStart() time.Time
	// Get work day end time
	GetWorkDayEnd(rest bool) time.Time
	// Get work day duration as duration
	GetWorkDayDuration() time.Duration
	// Get refresh interval
	GetRefreshInterval() time.Duration
	// Get truncate duration
	GetTruncateDuration() time.Duration
}

type config struct {
	path             string
	WorkDayStart     string        `yaml:"start"`
	WorkDayDuration  time.Duration `yaml:"duration"`
	RestDuration     time.Duration `yaml:"rest"`
	RefreshInterval  time.Duration `yaml:"refresh,omitempty"`
	TruncateInterval time.Duration `yaml:"truncate,omitempty"`
}

func NewConfig(path string) Config {
	config := &config{
		path: path,
	}
	if err := config.Reload(); err != nil {
		panic(err)
	}
	return config
}

func (c *config) Reload() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *config) GetWorkDayStart() time.Time {
	startHour, err := time.Parse("15:04", c.WorkDayStart)
	if err != nil {
		panic(fmt.Sprintf("Invalid start time: %s", c.WorkDayStart))
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), startHour.Hour(), startHour.Minute(), startHour.Second(), startHour.Nanosecond(), now.Location())
}

func (c *config) GetWorkDayEnd(rest bool) time.Time {
	if rest {
		return c.GetWorkDayStart().Add(c.WorkDayDuration + c.RestDuration)
	}
	return c.GetWorkDayStart().Add(c.WorkDayDuration)
}

func (c *config) GetWorkDayDuration() time.Duration {
	return c.WorkDayDuration
}

func (c *config) GetRefreshInterval() time.Duration {
	if c.RefreshInterval == 0 {
		return 30 * time.Second
	}
	return c.RefreshInterval
}

func (c *config) GetTruncateDuration() time.Duration {
	if c.TruncateInterval == 0 {
		return 1 * time.Minute
	}
	return c.TruncateInterval
}

func (c *config) String() string {
	return "Config{" +
		"workDayStart='" + c.WorkDayStart + "'" +
		", workDayDuration='" + c.WorkDayDuration.String() + "'" +
		"}"
}
