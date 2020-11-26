package worker

import (
	"testing"

	"github.com/govice/golinksd/pkg/chaintracker"
	"github.com/govice/golinksd/pkg/config"
	"github.com/govice/golinksd/pkg/golinks"
)

type testServicer struct{}

func (s *testServicer) ConfigService() *config.Service {
	return &config.Service{}
}

func (s *testServicer) GolinksService() *golinks.Service {
	return &golinks.Service{}
}

func (s *testServicer) ChainTrackerService() *chaintracker.Service {
	return &chaintracker.Service{}
}

func (s *testServicer) WorkerService() *Service {
	return &Service{}
}

type testConfigManager struct{}

func (rw *testConfigManager) ReadConfig() (*Config, error) {
	return &Config{}, nil
}

func (rw *testConfigManager) WriteConfig(cfg *Config) error {
	return nil
}

func TestNew(t *testing.T) {
	ts := &testServicer{}
	_, err := New(ts, &testConfigManager{})
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}
}
