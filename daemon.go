package main

import (
	"os"
	"time"

	"github.com/kardianos/service"
)

var pingNodeTicker <-chan time.Time

func (d *daemon) Run(s service.Service) error {
	go startPeer()
	go startWebserver()
	// go pingNodes()
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
