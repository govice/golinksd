package main

import (
	"github.com/govice/golinks-daemon/pkg/daemon"
	"github.com/govice/golinks-daemon/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	d, err := daemon.New()
	if err != nil {
		log.Fatalln(err)
	}

	log.Logln("PORT: " + viper.GetString("port"))
	log.Logln("AUTH_SERVER: " + viper.GetString("auth_server"))

	if err := d.Execute(); err != nil {
		log.Fatalln(err)
	}

	if err := d.StopDaemon(); err != nil {
		log.Fatalln(err)
	}
}
