package worker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/govice/golinks-daemon/pkg/chaintracker"
	"github.com/govice/golinks-daemon/pkg/config"
	"github.com/govice/golinks-daemon/pkg/golinks"
	"github.com/govice/golinks-daemon/pkg/log"
	"github.com/govice/golinks-daemon/pkg/scheduler"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	errorGroup   errgroup.Group
	WorkerConfig *Config
	ctx          context.Context
	mu           *sync.Mutex
	scheduler    *scheduler.Scheduler
	servicer     Servicer
}

type ConfigServicer interface {
	ConfigService() *config.Service
}

type GolinksServicer interface {
	GolinksService() *golinks.Service
}

type ChainTrackerServicer interface {
	ChainTrackerService() *chaintracker.Service
}

type WorkerServicer interface {
	WorkerService() *Service
}

type Servicer interface {
	ConfigServicer
	GolinksServicer
	ChainTrackerServicer
	WorkerServicer
}

func New(servicer Servicer) (*Service, error) {
	s, err := scheduler.New(viper.GetInt("concurrent_task_limit"))
	if err != nil {
		return nil, err
	}
	m := &Service{
		mu:        &sync.Mutex{},
		scheduler: s,
		servicer:  servicer,
	}

	workerConfig, err := m.loadWorkerConfig()
	if err != nil {
		log.Errln("failed to load worker config", err)
		return nil, err
	}
	m.WorkerConfig = workerConfig
	return m, nil
}

func (w *Service) loadWorkerConfig() (*Config, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	log.Logln("loading worker config...")
	workerConfigPath := filepath.Join(w.servicer.ConfigService().HomeDir(), "workers.json")

	_, err := os.Stat(workerConfigPath)
	if os.IsNotExist(err) {
		log.Logln("no worker configuration defined, initializing with empty config")
		return &Config{}, nil
	}

	configBytes, err := ioutil.ReadFile(workerConfigPath)
	if err != nil {
		return nil, err
	}

	workerConfig := &Config{}
	if err := json.Unmarshal(configBytes, workerConfig); err != nil {
		return nil, err
	}

	// reinitialize with initialized worker
	configOut := &Config{}
	for _, worker := range workerConfig.Workers {
		w, err := NewWorker(w.servicer, worker.RootPath, worker.GenerationPeriod, worker.IgnorePaths)
		if err != nil {
			log.Errln("failed to create new worker", err)
			return nil, err
		}
		configOut.Workers = append(configOut.Workers, w)
	}

	return configOut, nil
}

func (w *Service) saveWorkerConfig() error {
	log.Logln("saving worker config...")
	workerConfigPath := filepath.Join(w.servicer.ConfigService().HomeDir(), "workers.json")
	configBytes, err := json.MarshalIndent(w.WorkerConfig, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(workerConfigPath, configBytes, 0666)
}

func (w *Service) Execute(ctx context.Context) error {
	w.ctx = ctx
	log.Logln("starting workers...")
	if err := w.startWorkers(ctx); err != nil {
		return err
	}

	go w.scheduler.Run(w.ctx)

	for {
		select {
		case <-ctx.Done():
			log.Logln("worker manager terminating...")
			for _, worker := range w.WorkerConfig.Workers {
				worker.cancelFunc()
			}
			return nil
		}
	}
}

func (w *Service) startWorkers(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, worker := range w.WorkerConfig.Workers {
		if worker.running {
			continue
		}
		workerCtx, workerCancelFunc := context.WithCancel(ctx)
		worker.AddCancelFunc(func() {
			workerCancelFunc()
			worker.running = false
		})
		w.errorGroup.Go(func() error { return worker.Execute(workerCtx) })
		worker.running = true
	}
	return nil
}

var ErrWorkerManagerNotStarted = errors.New("worker cannot be restarted without an existing context")

func (w *Service) startNewWorkers() error {
	if w.ctx == nil {
		return ErrWorkerManagerNotStarted
	}
	return w.startWorkers(w.ctx)
}

func (w *Service) removeWorker(index int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	worker := w.WorkerConfig.Workers[index]
	worker.cancelFunc()
	w.WorkerConfig.Workers = append(w.WorkerConfig.Workers[:index], w.WorkerConfig.Workers[index+1:]...)
	return w.saveWorkerConfig()
}

func (w *Service) getWorker(index int) *Worker {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.WorkerConfig.Workers[index]
}

func (w *Service) addWorker(rootPath string, generationPeriod int, ignorePaths []string) (*Worker, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	worker, err := NewWorker(w.servicer, rootPath, generationPeriod, ignorePaths)
	if err != nil {
		return nil, err
	}

	w.WorkerConfig.Workers = append(w.WorkerConfig.Workers, worker)

	if err := w.saveWorkerConfig(); err != nil {
		log.Errln("failed to save worker config after adding worker")
		return nil, err
	}
	return worker, nil
}

func (w *Service) ScheduleWork(workerID string, task func() error) error {
	t := &WorkerTask{
		id:   workerID,
		work: task,
	}

	return w.scheduler.Schedule(t)
}

var ErrWorkerIndexOutOfBonds = errors.New("worker index out of bounds")

func (w *Service) GetWorkerByIndex(index int) (*Worker, error) {
	if index < 0 || index > w.WorkerConfig.Length()-1 {
		return nil, ErrWorkerIndexOutOfBonds
	}

	return w.WorkerConfig.Workers[index], nil
}

func (w *Service) DeleteWorkerByIndex(index int) error {
	if index < 0 || index > w.WorkerConfig.Length()-1 {
		return ErrWorkerIndexOutOfBonds
	}
	return w.removeWorker(index)
}

func (w *Service) AddWorker(rootPath string, generationPeriod int, ignorePaths []string) error {
	if _, err := w.addWorker(rootPath, generationPeriod, ignorePaths); err != nil {
		return err
	}
	return w.startNewWorkers()
}
