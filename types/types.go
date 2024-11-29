package types

// Common types

type Config struct {
	Sheet Sheet `yaml:"sheet"`
	Slack Slack `yaml:"slack"`
}

type RGB struct {
	Blue  float64 `yaml:"blue"`
	Green float64 `yaml:"green"`
	Red   float64 `yaml:"red"`
}