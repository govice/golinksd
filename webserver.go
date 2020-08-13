package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var ledger Ledger

func registerFrontendRoutes(router *gin.Engine) error {
	templatesHome := viper.GetString("templates_home")
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

	if err := registerConsoleHandlers(router); err != nil {
		return err
	}

	return nil

}
