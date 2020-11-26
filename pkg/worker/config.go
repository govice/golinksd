package worker

type Config struct {
	Workers []*Worker `json:"workers"`
}

func (c *Config) Length() int {
	return len(c.Workers)
}

type ConfigReaderWriter interface {
	ConfigReader
	ConfigWriter
}

type ConfigReader interface {
	ReadConfig() (*Config, error)
}

type ConfigWriter interface {
	WriteConfig(cfg *Config) error
}
