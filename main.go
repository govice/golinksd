package main

import (
	"log"
	"os"
	"sync"

	"github.com/govice/golinks/blockchain"
	maddr "github.com/multiformats/go-multiaddr"

	"github.com/kardianos/service"
)

// var logger service.Logger

type daemon struct{}

type Config struct {
	RendezvousString string
	BootstrapPeers   []maddr.Multiaddr
	ListenAddresses  []maddr.Multiaddr
	ProtocolID       string
}

var chainMutex sync.Mutex
var chain blockchain.Blockchain

func main() {
	log.Println("DOCKER_MACHINE_IP: " + os.Getenv("DOCKER_MACHINE_IP"))
	log.Println("PORT: " + os.Getenv("PORT"))

	serviceConfig := &service.Config{
		Name:        "golinksDaemon",
		DisplayName: "GoLinks Daemon",
	}

	//TODO load blockchain from file
	if os.Getenv("GENESIS") == "true" {
		chainReset()
	}

	daemon := &daemon{}
	s, err := service.New(daemon, serviceConfig)
	if err != nil {
		panic(err)
	}

	logger, err := s.Logger(nil)
	if err != nil {
		panic(err)
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
