package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type WorkerManager struct {
	daemon       *daemon
	errorGroup   errgroup.Group
	WorkerConfig *WorkerConfig
	ctx          context.Context
	mu           *sync.Mutex
	scheduler    *Scheduler
}

type WorkerConfig struct {
	Workers []*Worker `json:"workers"`
}

func (wc *WorkerConfig) Length() int {
	return len(wc.Workers)
}

func NewWorkerManager(daemon *daemon) (*WorkerManager, error) {
	s, err := NewScheduler(viper.GetInt("concurrent_task_limit"))
	if err != nil {
		return nil, err
	}
	m := &WorkerManager{daemon: daemon,
		mu:        &sync.Mutex{},
		scheduler: s,
	}

	workerConfig, err := m.loadWorkerConfig()
	if err != nil {
		errln("failed to load worker config", err)
		return nil, err
	}
	m.WorkerConfig = workerConfig
	return m, nil
}

func (w *WorkerManager) loadWorkerConfig() (*WorkerConfig, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	logln("loading worker config...")
	workerConfigPath := filepath.Join(w.daemon.configService.HomeDir(), "workers.json")

	_, err := os.Stat(workerConfigPath)
	if os.IsNotExist(err) {
		logln("no worker configuration defined, initializing with empty config")
		return &WorkerConfig{}, nil
	}

	configBytes, err := ioutil.ReadFile(workerConfigPath)
	if err != nil {
		return nil, err
	}

	workerConfig := &WorkerConfig{}
	if err := json.Unmarshal(configBytes, workerConfig); err != nil {
		return nil, err
	}

	// reinitialize with initialized worker
	configOut := &WorkerConfig{}
	for _, worker := range workerConfig.Workers {
		w, err := NewWorker(w.daemon, worker.RootPath, worker.GenerationPeriod, worker.IgnorePaths)
		if err != nil {
			errln("failed to create new worker", err)
			return nil, err
		}
		configOut.Workers = append(configOut.Workers, w)
		worker.daemon = w.daemon
	}

	return configOut, nil
}

func (w *WorkerManager) saveWorkerConfig() error {
	logln("saving worker config...")
	workerConfigPath := filepath.Join(w.daemon.configService.HomeDir(), "workers.json")
	configBytes, err := json.MarshalIndent(w.WorkerConfig, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(workerConfigPath, configBytes, 0666)
}

func (w *WorkerManager) Execute(ctx context.Context) error {
	w.ctx = ctx
	logln("starting workers...")
	if err := w.startWorkers(ctx); err != nil {
		return err
	}

	go w.scheduler.Run(w.ctx)

	for {
		select {
		case <-ctx.Done():
			logln("worker manager terminating...")
			for _, worker := range w.WorkerConfig.Workers {
				worker.cancelFunc()
			}
			return nil
		}
	}
}

func (w *WorkerManager) startWorkers(ctx context.Context) error {
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

func (w *WorkerManager) startNewWorkers() error {
	if w.ctx == nil {
		return ErrWorkerManagerNotStarted
	}
	return w.startWorkers(w.ctx)
}

func (w *WorkerManager) removeWorker(index int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	worker := w.WorkerConfig.Workers[index]
	worker.cancelFunc()
	w.WorkerConfig.Workers = append(w.WorkerConfig.Workers[:index], w.WorkerConfig.Workers[index+1:]...)
	return w.saveWorkerConfig()
}

func (w *WorkerManager) getWorker(index int) *Worker {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.WorkerConfig.Workers[index]
}

func (w *WorkerManager) addWorker(rootPath string, generationPeriod int, ignorePaths []string) (*Worker, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	worker, err := NewWorker(w.daemon, rootPath, generationPeriod, ignorePaths)
	if err != nil {
		return nil, err
	}

	w.WorkerConfig.Workers = append(w.WorkerConfig.Workers, worker)

	if err := w.saveWorkerConfig(); err != nil {
		errln("failed to save worker config after adding worker")
		return nil, err
	}
	return worker, nil
}

func (w *WorkerManager) scheduleWork(workerID string, task func() error) error {
	return w.scheduler.Schedule(workerID, task)
}
