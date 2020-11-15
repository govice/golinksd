package main

import (
	"github.com/govice/golinksd/cmd"
	"github.com/govice/golinksd/pkg/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
