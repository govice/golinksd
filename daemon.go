package main

import (
	"context"
	"fmt"
	"os"
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
}

var (
	g errgroup.Group
)

func (d *daemon) Run(s service.Service) error {
	router := gin.Default()
	blockchainService = &BlockchainService{
		mutex: sync.Mutex{},
	}
	//TODO load blockchain from file
	if viper.GetBool("genesis") {
		blockchainService.resetChain()
	}

	primaryContext, primaryCancel := context.WithCancel(context.Background())
	defer func() {
		primaryCancel()
	}()
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

	p2pCtx, cancelP2P := context.WithCancel(primaryContext)

	g.Go(func() error {
		if err := startPeer(p2pCtx); err != nil {
			return err
		}

		return nil
	})
	d.cancelFuncs = append(d.cancelFuncs, cancelP2P)

	return nil
}

func pingNodes() {
	pingNodeTicker = time.Tick(15 * time.Second)
	for range pingNodeTicker {
		ledger.PingNodes()
	}
}

func (d *daemon) Start(s service.Service) error {
	d.Run(s)
	return nil
}

func (d *daemon) Stop(s service.Service) error {
	for i, cancel := range d.cancelFuncs {
		fmt.Printf("canceling context %d/%d\n", i+1, len(d.cancelFuncs))
		cancel()
	}
	g.Wait()
	os.Exit(0)
	return nil
}
