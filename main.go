package main

import (
	"log"
	"os"

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

func main() {
	log.Println("DOCKER_MACHINE_IP: " + os.Getenv("DOCKER_MACHINE_IP"))
	log.Println("PORT: " + os.Getenv("PORT"))
	log.Println("AUTH_SERVER: " + os.Getenv("AUTH_SERVER"))
	if os.Getenv("AUTH_SERVER") == "" {
		log.Fatal("Enviroment AUTH_SERVER does not exist")
	}

	serviceConfig := &service.Config{
		Name:        "golinksDaemon",
		DisplayName: "GoLinks Daemon",
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
