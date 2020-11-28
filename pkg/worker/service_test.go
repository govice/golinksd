package worker

import (
	"context"
	"errors"
	"testing"
	"time"

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
	service, err := NewDefault(ts, cm)
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
	service, err := NewDefault(ts, cm)
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

func TestRemoveWorker(t *testing.T) {
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
	service, err := NewDefault(ts, cm)
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}

	if len(cm.Config.Workers) != 1 {
		t.Error("expected 1 worker. got", cm.Config.Workers)
	}

	service.removeWorker(0)

	if len(cm.Config.Workers) != 0 {
		t.Error("expected empty workers. got", len(cm.Config.Workers))
	}

	if cm.ConfigReads != 1 {
		t.Error("expected 1 config read on New. got", cm.ConfigReads)
	}

	if cm.ConfigWrites != 1 {
		t.Error("expected 1 config write on worker remove. got", cm.ConfigWrites)
	}
}

func TestGetWorkerByIndex(t *testing.T) {
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
	service, err := NewDefault(ts, cm)
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}

	if len(cm.Config.Workers) != 1 {
		t.Error("expected 1 worker. got", cm.Config.Workers)
	}

	_, err = service.GetWorkerByIndex(0)
	if err != nil {
		t.Error(err)
	}

	_, err = service.GetWorkerByIndex(1)
	if err == nil || !errors.Is(err, ErrWorkerIndexOutOfBonds) {
		t.Error("expected error", ErrWorkerIndexOutOfBonds)
	}

	_, err = service.GetWorkerByIndex(-1)
	if err == nil || !errors.Is(err, ErrWorkerIndexOutOfBonds) {
		t.Error("expected error", ErrWorkerIndexOutOfBonds)
	}
}

func TestScheduleWork(t *testing.T) {
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
	service, err := New(ts, cm, 1)
	if err != nil {
		t.Error("failed to instantiate new service", err)
	}

	if len(cm.Config.Workers) != 1 {
		t.Error("expected 1 worker. got", cm.Config.Workers)
	}

	var i int

	service.ScheduleWork("1234", func() error {
		i++
		return nil
	})

	if i != 0 {
		t.Error("work executed with no running worker")
	}

	primary, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5)*time.Second))

	// this execution occurs as part of Execute
	go service.scheduler.Run(primary)

	service.ScheduleWork("5555", func() error {
		cancelFunc()
		return nil
	})

	<-primary.Done()
	if i != 1 {
		t.Error("deadline exceeded. expected scheduled work execution")
	}
}
