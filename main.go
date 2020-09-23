package main

import (
	"github.com/spf13/viper"
)

func main() {
	d, err := NewDaemon()
	if err != nil {
		fatalln(err)
	}

	logln("PORT: " + viper.GetString("port"))
	logln("AUTH_SERVER: " + viper.GetString("auth_server"))

	if err := d.Execute(); err != nil {
		fatalln(err)
	}

	if err := d.StopDaemon(); err != nil {
		fatalln(err)
	}
}
