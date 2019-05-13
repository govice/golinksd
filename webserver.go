package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/govice/golinks-daemon/net"
)

var router *gin.Engine

var ledger = net.Ledger{
	Nodes: []net.Node{
		net.Node{
			Address:   "http://" + os.Getenv("DOCKER_MACHINE_IP") + ":8080",
			Available: true,
		},
		net.Node{
			Address:   "http://" + os.Getenv("DOCKER_MACHINE_IP") + ":8081",
			Available: true,
		},
	},
}

func startWebserver() {
	router = gin.Default()
	templatesHome := os.Getenv("TEMPLATES_HOME")
	log.Println("Templates Home: " + templatesHome)
	if templatesHome != "" {
		router.LoadHTMLGlob(templatesHome + "/*")
	} else {
		router.LoadHTMLGlob("./templates/*")
	}
	router.GET("/error", func(c *gin.Context) {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title": "GoLinks | Error",
		})
	})

	registerConsoleHandlers(router)
	router.Run(":" + os.Getenv("PORT")) // listen and serve on PORT
}
