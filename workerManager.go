// Copyright 2020 Kevin Gentile
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
)

type WorkerManager struct {
	daemon       *daemon
	errorGroup   errgroup.Group
	WorkerConfig *WorkerConfig
	ctx          context.Context
	mu           *sync.Mutex
}

type WorkerConfig struct {
	Workers []*Worker `json:"workers"`
}

func (wc *WorkerConfig) Length() int {
	return len(wc.Workers)
}

func NewWorkerManager(daemon *daemon) (*WorkerManager, error) {
	m := &WorkerManager{daemon: daemon,
		mu: &sync.Mutex{},
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
	configBytes, err := ioutil.ReadFile(workerConfigPath)
	if err != nil {
		return nil, err
	}

	workerConfig := &WorkerConfig{}
	if err := json.Unmarshal(configBytes, workerConfig); err != nil {
		return nil, err
	}

	for _, worker := range workerConfig.Workers {
		worker.daemon = w.daemon
	}

	return workerConfig, nil
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

	for {
		select {
		case <-ctx.Done():
			logln("worker manager terminating...")
			//TODO call canecl funcs?
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
		worker.cancelFunc = func() {
			workerCancelFunc()
			worker.running = false
		}
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
