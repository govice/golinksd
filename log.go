package main

import (
	"log"
	"os"
)

var stdOutLogger = log.New(os.Stdout, "", log.Ltime)
var logLogger = log.New(os.Stderr, "log: ", log.Ltime)
var errorLogger = log.New(os.Stderr, "error: ", log.Ltime)

func println(v ...interface{}) {
	stdOutLogger.Println(v...)
}

func logln(v ...interface{}) {
	logLogger.Println(v...)
}

func logf(format string, v ...interface{}) {
	logLogger.Printf(format, v...)
}

func errln(v ...interface{}) {
	errorLogger.Println(v...)
}

func fatalln(v ...interface{}) {
	stdOutLogger.Println(v...)
	os.Exit(1)
}
