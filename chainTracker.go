package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type ChainTracker struct {
	daemon *daemon
}

func NewChainTracker(daemon *daemon) (*ChainTracker, error) {
	return &ChainTracker{
		daemon: daemon,
	}, nil
}

func (ct *ChainTracker) Execute(ctx context.Context) error {
	logln("starting chain tracker")
	if err := ct.initialize(); err != nil {
		return err
	}
	trackingPeriod := viper.GetInt("tracking_period")
	logln("tracking period:", trackingPeriod)
	syncTicker := time.NewTicker(time.Millisecond * time.Duration(trackingPeriod))

	for {
		select {
		case <-syncTicker.C:
			if err := ct.synchronize(); err != nil {
				errln("failed to synchronize chain", err)
			}
		case <-ctx.Done():
			logln("received termination on chain tracker context")
			return nil
		}
	}
}

func (ct *ChainTracker) initialize() error {
	os.Mkdir(ct.chainDir(), os.ModePerm)
	return nil
}

func (ct *ChainTracker) chainDir() string {
	return filepath.Join(ct.daemon.HomeDir(), "chain")
}

func (ct *ChainTracker) synchronize() error {
	remoteGCI, err := ct.daemon.golinksService.GetGCI()
}

// func (cs *ChainSyncer) getMasterGCI() (int, error) {
// 	daemon.blockchainService.
// }
