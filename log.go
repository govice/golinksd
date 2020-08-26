// Copyright 2020 Kevin Gentile
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"
)

var stdOutLogger = log.New(os.Stdout, "", log.Ltime)
var logLogger = log.New(os.Stderr, "golinksd LOG: ", log.Ltime)
var errorLogger = log.New(os.Stderr, "golinksd ERROR: ", log.Ltime)

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
