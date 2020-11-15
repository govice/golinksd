package main

import (
	"github.com/govice/golinks-daemon/cmd"
	"github.com/govice/golinks-daemon/pkg/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
