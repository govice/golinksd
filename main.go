package main

import (
	"log"

	"github.com/spf13/viper"
)

func main() {
	if err := setupConfig(); err != nil {
		fatalln(err)
	}

	if err := checkLogin(); err != nil {
		fatalln(err)
	}

	log.Println("PORT: " + viper.GetString("port"))
	log.Println("AUTH_SERVER: " + viper.GetString("auth_server"))

	d, err := NewDaemon()
	if err != nil {
		fatalln(err)
	}

	if err := d.Execute(); err != nil {
		fatalln(err)
	}
}
