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
	return &Config{
		Workers: []*Worker{
			{
				RootPath:         "/tmp/root",
				GenerationPeriod: 100,
				IgnorePaths:      []string{"/tmp/ignore"},
			},
		}}, nil
}

func (rw *testConfigManager) WriteConfig(cfg *Config) error {
	return nil
}

func TestNew(t *testing.T) {
	ts := &testServicer{}
	service, err := New(ts, &testConfigManager{})
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}

	config := service.WorkerConfig

	if len(config.Workers) != 1 {
		t.Error("expected worker length 1")
	}

	worker := config.Workers[0]
	if worker.RootPath != "/tmp/root" {
		t.Error("expected root path /tmp/root")
	}

	if worker.IgnorePaths[0] != "/tmp/ignore" {
		t.Error("expected ignored path /tmp/ignore")
	}

	if worker.GenerationPeriod != 100 {
		t.Error("expeted worker generation period of 100")
	}
}
