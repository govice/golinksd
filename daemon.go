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
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	pingNodeTicker <-chan time.Time
)

type daemon struct {
	cancelFuncs       []context.CancelFunc
	service           service.Service
	logger            service.Logger
	errorGroup        errgroup.Group
	blockchainService *BlockchainService
	configService     *ConfigService
	golinksService    *GolinksService
	webserver         *Webserver
	worker            *Worker
	chainTracker      *ChainTracker
	gui               *GUI

	chainMutex sync.Mutex
}

func NewDaemon() (*daemon, error) {
	// SERVICES
	d := &daemon{}
	if err := d.initializeServies(); err != nil {
		return nil, err
	}

	if err := d.initializeBackgroundTasks(); err != nil {
		return nil, err
	}

	if err := d.initializeGUI(); err != nil {
		return nil, err
	}

	// DAEMON CONFIG
	serviceConfig := &service.Config{
		Name:        "golinksd",
		DisplayName: "golinksd",
		Description: "golinks daemon",
	}

	s, err := service.New(d, serviceConfig)
	if err != nil {
		return nil, err
	}
	d.service = s

	d.logger, err = s.Logger(nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *daemon) initializeServies() error {
	cs, err := NewConfigService(d)
	if err != nil {
		errln("failed to initialize configuration service")
	}
	d.configService = cs

	bs, err := NewBlockchainService(d)
	if err != nil {
		errln("failed to initialize blockchain service")
		return err
	}
	d.blockchainService = bs

	gs, err := NewGolinksService(d)
	if err != nil {
		errln("failed to iniitalize golinks service")
	}
	d.golinksService = gs

	return nil
}

func (d *daemon) initializeBackgroundTasks() error {
	// WORKERS
	webserver, err := NewWebserver(d)
	if err != nil {
		errln("failed to initialize webserver")
		return err
	}
	d.webserver = webserver

	worker, err := NewWorker(d)
	if err != nil {
		errln("failed to initialize worker")
		return err
	}
	d.worker = worker

	ct, err := NewChainTracker(d)
	if err != nil {
		errln("failed to initialize chain tracker")
		return err
	}
	d.chainTracker = ct

	return nil
}

func (d *daemon) initializeGUI() error {

	gui, err := NewGUI(d)
	if err != nil {
		errln("failed to initialize GUI")
		return err
	}
	d.gui = gui
	return nil
}

func (d *daemon) Execute() error {
	if err := d.service.Run(); err != nil {
		return err
	}
	return nil
}

func (d *daemon) run() error {

	primaryContext, primaryCancel := context.WithCancel(context.Background())

	d.cancelFuncs = append(d.cancelFuncs, primaryCancel)

	if viper.GetBool("development") {
		d.errorGroup.Go(func() error {
			return d.ExecuteFrontend(primaryContext)
		})
	}

	workerCtx, cancelDaemon := context.WithCancel(primaryContext)

	d.errorGroup.Go(func() error {
		return d.ExecuteWorker(workerCtx)
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelDaemon)

	chainTrackerCtx, cancelChainTracker := context.WithCancel(primaryContext)
	d.errorGroup.Go(func() error {
		return d.ExecuteChainTracker(chainTrackerCtx)
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelChainTracker)

	<-primaryContext.Done()
	return nil
}

func (d *daemon) ExecuteFrontend(ctx context.Context) error {
	return d.webserver.Execute(ctx)
}

func (d *daemon) ExecuteWorker(ctx context.Context) error {
	return d.worker.Execute(ctx)
}

func (d *daemon) Start(s service.Service) error {
	go d.run()
	return nil
}

func (d *daemon) StopDaemon() error {
	return d.Stop(d.service)
}

func (d *daemon) Stop(s service.Service) error {
	//TODO hold primary context and use it to cancel children?
	for i, cancel := range d.cancelFuncs {
		logf("canceling context %d/%d\n", i+1, len(d.cancelFuncs))
		cancel()
	}
	d.errorGroup.Wait()
	return nil
}

func (d *daemon) HomeDir() string {
	return d.configService.HomeDir()
}

func (d *daemon) ExecuteChainTracker(ctx context.Context) error {
	return d.chainTracker.Execute(ctx)
}
