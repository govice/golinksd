package log

import (
	"log"
	"os"
)

var stdOutLogger = log.New(os.Stdout, "", log.Ltime)
var logLogger = log.New(os.Stderr, "golinksd LOG: ", log.Ltime)
var errorLogger = log.New(os.Stderr, "golinksd ERROR: ", log.Ltime)
var warningLogger = log.New(os.Stderr, "golinksd WARNING: ", log.Ltime)

func Println(v ...interface{}) {
	stdOutLogger.Println(v...)
}

func Logln(v ...interface{}) {
	logLogger.Println(v...)
}

func Logf(format string, v ...interface{}) {
	logLogger.Printf(format, v...)
}

func Errln(v ...interface{}) {
	errorLogger.Println(v...)
}

func Fatalln(v ...interface{}) {
	stdOutLogger.Println(v...)
	os.Exit(1)
}

func Warnln(v ...interface{}) {
	warningLogger.Println(v...)
}
