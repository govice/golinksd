package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kardianos/service"
)

var logger service.Logger

type daemon struct{}

var router *mux.Router

func (d *daemon) Start(s service.Service) error {
	go d.Run(s)
	return nil
}

func (d *daemon) Run(s service.Service) error {
	router = mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/index.html")
	})

	go func() {
		http.ListenAndServe("localhost:8080", router)
	}()
	return nil
}

func (d *daemon) Stop(s service.Service) error {
	os.Exit(0)
	return nil
}

func main() {
	serviceConfig := &service.Config{
		Name:        "GolinksDaemon",
		DisplayName: "GoLinks Daemon",
	}

	daemon := &daemon{}
	s, err := service.New(daemon, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
