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
	// Get work day start time
	GetWorkDayStart() time.Time
	// Get work day end time
	GetWorkDayEnd(rest bool) time.Time
	// Get work day duration as duration
	GetWorkDayDuration() time.Duration
}

type config struct {
	WorkDayStart    string        `yaml:"start"`
	WorkDayDuration time.Duration `yaml:"duration"`
	RestDuration    time.Duration `yaml:"rest"`
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

func (c *config) String() string {
	return "Config{" +
		", workDayStart='" + c.WorkDayStart + "'" +
		", workDayDuration=" + c.WorkDayDuration.String() +
		"}"
}

func NewConfig(path string) Config {
	config := &config{}
	data, error := os.ReadFile(path)
	if error != nil {
		panic(error)
	}
	error = yaml.Unmarshal(data, config)
	if error != nil {
		panic(error)
	}
	return config
}
