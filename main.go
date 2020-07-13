package main

import (
	"log"
	"os"

	maddr "github.com/multiformats/go-multiaddr"

	"github.com/kardianos/service"
)

var daemonLogger service.Logger

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
		Description: "golinks daemon",
	}

	d := &daemon{}
	s, err := service.New(d, serviceConfig)
	if err != nil {
		fatalln(err)
	}

	daemonLogger, err = s.Logger(nil)
	if err != nil {
		fatalln(err)
	}

	err = s.Run()
	if err != nil {
		daemonLogger.Error(err)
	}
}
