package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

var ledger Ledger

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
