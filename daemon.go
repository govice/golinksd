package main

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	pingNodeTicker <-chan time.Time
)

type daemon struct {
	cancelFuncs []context.CancelFunc
	service     service.Service
}

var (
	g errgroup.Group
)

func (d *daemon) run() error {

	router := gin.Default()
	blockchainService = &BlockchainService{
		mutex: sync.Mutex{},
	}
	//TODO load blockchain from file
	if viper.GetBool("genesis") {
		blockchainService.resetChain()
	}

	primaryContext, primaryCancel := context.WithCancel(context.Background())

	d.cancelFuncs = append(d.cancelFuncs, primaryCancel)

	if viper.GetBool("development") {
		g.Go(func() error {
			var frontendErr error
			go func() {
				if frontendErr = registerFrontendRoutes(router); frontendErr != nil {
					return
				}

				if frontendErr = router.Run(":" + viper.GetString("port")); frontendErr != nil {
					return
				} // listen and serve on PORT
			}()
			<-primaryContext.Done()
			return frontendErr
		})
	}

	if err := registerAPIRoutes(router); err != nil {
		return err
	}

	daemonCtx, cancelDaemon := context.WithCancel(primaryContext)

	g.Go(func() error {
		if err := startHost(daemonCtx); err != nil {
			return err
		}

		return nil
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelDaemon)

	<-primaryContext.Done()
	return nil
}

func pingNodes() {
	pingNodeTicker = time.Tick(15 * time.Second)
	for range pingNodeTicker {
		ledger.PingNodes()
	}
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
	g.Wait()
	return nil
}
