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

type testConfigManager struct {
	ConfigReads  int
	ConfigWrites int
	Config       *Config
}

func newTestConfigManager(initial *Config) *testConfigManager {
	return &testConfigManager{
		Config: initial,
	}
}

func (rw *testConfigManager) ReadConfig() (*Config, error) {
	rw.ConfigReads++
	return rw.Config, nil
}

func (rw *testConfigManager) WriteConfig(cfg *Config) error {
	rw.ConfigWrites++
	rw.Config = cfg
	return nil
}

func TestNew(t *testing.T) {
	ts := &testServicer{}
	cm := newTestConfigManager(&Config{
		Workers: []*Worker{
			{
				RootPath:         "/tmp/root",
				GenerationPeriod: 100,
				IgnorePaths:      []string{"/tmp/ignore"},
			},
		}})
	service, err := New(ts, cm)
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

	if cm.ConfigReads != 1 {
		t.Error("expected 1 config reads. got", cm.ConfigReads)
	}

	if cm.ConfigWrites != 0 {
		t.Error("expected 0 config writes. got", cm.ConfigWrites)
	}
}

func TestAddWorker(t *testing.T) {
	ts := &testServicer{}
	initial := &Config{
		Workers: []*Worker{
			{
				RootPath:         "/tmp/root",
				GenerationPeriod: 100,
				IgnorePaths:      []string{"/tmp/ignore"},
			},
		}}
	cm := newTestConfigManager(initial)
	service, err := New(ts, cm)
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}

	worker, err := service.addWorker("/tmp/root2", 100, []string{"/tmp/ingnore2"})
	if err != nil {
		t.Error("expected successful worker add.", err)
	}

	if worker.RootPath != cm.Config.Workers[1].RootPath {
		t.Error("expected root path", cm.Config.Workers[1], ". got ", worker.RootPath)
	}

	if worker.GenerationPeriod != cm.Config.Workers[1].GenerationPeriod {
		t.Error("expected generation period", cm.Config.Workers[1], ". got ", worker.GenerationPeriod)
	}

	if worker.IgnorePaths[0] != cm.Config.Workers[1].IgnorePaths[0] {
		t.Error("expected ignore path", cm.Config.Workers[1].IgnorePaths[0], ". got", worker.IgnorePaths[0])
	}

	if cm.ConfigWrites != 1 {
		t.Error("expected 1 config write on worker add. got", cm.ConfigWrites)
	}

	if cm.ConfigReads != 1 {
		t.Error("expected 1 config read on New. got", cm.ConfigReads)
	}
}
