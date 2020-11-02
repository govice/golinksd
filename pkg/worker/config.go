package worker

type Config struct {
	Workers []*Worker `json:"workers"`
}

func (c *Config) Length() int {
	return len(c.Workers)
}
