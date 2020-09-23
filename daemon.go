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
	primaryCancel     context.CancelFunc
	cancelFuncs       []context.CancelFunc
	service           service.Service
	logger            service.Logger
	errorGroup        errgroup.Group
	blockchainService *BlockchainService
	configService     *ConfigService
	golinksService    *GolinksService
	webserver         *Webserver
	workerManager     *WorkerManager
	chainTracker      *ChainTracker
	// gui               *GUI

	chainMutex sync.Mutex
}

func NewDaemon() (*daemon, error) {
	// SERVICES
	d := &daemon{}
	if err := d.initializeServices(); err != nil {
		return nil, err
	}

	if err := d.initializeBackgroundTasks(); err != nil {
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

func (d *daemon) initializeServices() error {
	cs, err := NewConfigService(d)
	if err != nil {
		errln("failed to initialize configuration service")
		return err
	}
	d.configService = cs

	gs, err := NewGolinksService(d)
	if err != nil {
		errln("failed to iniitalize golinks service")
		return err
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

	workerManager, err := NewWorkerManager(d)
	if err != nil {
		errln("failed to initialize worker")
		return err
	}
	d.workerManager = workerManager

	ct, err := NewChainTracker(d)
	if err != nil {
		errln("failed to initialize chain tracker")
		return err
	}
	d.chainTracker = ct

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
	d.primaryCancel = primaryCancel

	if viper.GetBool("development") {
		d.errorGroup.Go(func() error {
			return d.ExecuteFrontend(primaryContext)
		})
	}

	workerCtx, cancelDaemon := context.WithCancel(primaryContext)

	d.errorGroup.Go(func() error {
		return d.ExecuteWorkerManager(workerCtx)
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

func (d *daemon) ExecuteWorkerManager(ctx context.Context) error {
	return d.workerManager.Execute(ctx)
}

func (d *daemon) Start(s service.Service) error {
	go d.run()
	return nil
}

func (d *daemon) StopDaemon() error {
	return d.Stop(d.service)
}

func (d *daemon) Stop(s service.Service) error {
	d.primaryCancel()
	d.errorGroup.Wait()
	return nil
}

func (d *daemon) HomeDir() string {
	return d.configService.HomeDir()
}

func (d *daemon) ExecuteChainTracker(ctx context.Context) error {
	return d.chainTracker.Execute(ctx)
}

// WARN: this must be executed from the main thread
// func (d *daemon) RunGUI() {
// 	d.gui.ShowAndRun()
// }
