package main

import (
	"context"
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
	webserver         *Webserver
	worker            *Worker
}

func NewDaemon() (*daemon, error) {
	d := &daemon{}
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

	webserver, err := NewWebserver(d)
	if err != nil {
		errln("failed to initialize webserver")
		return nil, err
	}
	d.webserver = webserver

	worker, err := NewWorker(d)
	if err != nil {
		errln("failed to initialize worker")
		return nil, err
	}
	d.worker = worker

	bs, err := NewBlockchainService(d)
	if err != nil {
		errln("failed to initialize blockchain service")
		return nil, err
	}
	d.blockchainService = bs

	return d, nil
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

	<-primaryContext.Done()
	return nil
}

func (d *daemon) ExecuteFrontend(ctx context.Context) error {
	var frontendErr error
	go func() {
		if err := d.webserver.Execute(); err != nil {
			frontendErr = err
			return
		}
	}()
	<-ctx.Done()
	return frontendErr
}

func (d *daemon) ExecuteWorker(ctx context.Context) error {
	return d.worker.Execute(ctx)
}

func (d *daemon) Start(s service.Service) error {
	go d.run()
	return nil
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
