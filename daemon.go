package main

import (
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/spf13/viper"
)

var (
	pingNodeTicker <-chan time.Time
)

type daemon struct{}

func (d *daemon) Run(s service.Service) error {
	router = gin.Default()
	blockchainService = &BlockchainService{
		mutex: sync.Mutex{},
	}
	//TODO load blockchain from file
	if viper.GetBool("genesis") {
		blockchainService.resetChain()
	}
	go startPeer()
	go startWebserver()

	apiService = &APIService{
		router: router,
	}
	go apiService.startAPI()

	// go pingNodes()
	router.Run(":" + viper.GetString("port")) // listen and serve on PORT
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
	os.Exit(0)
	return nil
}
