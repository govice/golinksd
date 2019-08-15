package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

var ledger = Ledger{
	Nodes: []Node{
		Node{
			Address:   "http://" + os.Getenv("DOCKER_MACHINE_IP") + ":8080",
			Available: true,
		},
		Node{
			Address:   "http://" + os.Getenv("DOCKER_MACHINE_IP") + ":8081",
			Available: true,
		},
	},
}

func startWebserver() {
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

}
