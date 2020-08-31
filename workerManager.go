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
	"io/ioutil"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

type WorkerManager struct {
	daemon       *daemon
	errorGroup   errgroup.Group
	WorkerConfig *WorkerConfig
}

type WorkerConfig struct {
	Workers []*Worker `json:"workers"`
}

func NewWorkerManager(daemon *daemon) (*WorkerManager, error) {
	return &WorkerManager{daemon: daemon}, nil
}

func (w *WorkerManager) LoadWorkerConfig(daemon *daemon) (*WorkerConfig, error) {
	logln("loading worker config...")
	workerConfigPath := filepath.Join(daemon.configService.HomeDir(), "workers.json")
	configBytes, err := ioutil.ReadFile(workerConfigPath)
	if err != nil {
		return nil, err
	}

	workerConfig := &WorkerConfig{}
	if err := json.Unmarshal(configBytes, workerConfig); err != nil {
		return nil, err
	}

	for _, worker := range workerConfig.Workers {
		worker.daemon = daemon
	}

	return workerConfig, nil
}

func (w *WorkerManager) Execute(ctx context.Context) error {
	workerConfig, err := w.LoadWorkerConfig(w.daemon)
	if err != nil {
		errln("failed to load worker config", err)
		return err
	}
	w.WorkerConfig = workerConfig

	logln("starting workers...")
	for _, worker := range w.WorkerConfig.Workers {
		workerCtx, workerCancelFunc := context.WithCancel(ctx)
		worker.cancelFunc = workerCancelFunc
		w.errorGroup.Go(func() error { return worker.Execute(workerCtx) })
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
