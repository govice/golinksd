package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerConsoleHandlers(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/console")
	})

	router.GET("/console", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "GoLinks | Home",
			"cards": consoleCards,
		})
	})

	router.GET("/console/addBlock", func(c *gin.Context) {
		c.HTML(http.StatusOK, "addBlock.html", gin.H{
			"title": "GoLinks | Add Block",
		})
	})

	router.POST("/console/addBlock", func(c *gin.Context) {
		formContent := c.PostForm("blockContentTextArea")
		log.Println(formContent)
		if len(formContent) > 0 {
			if _, err := blockchainService.addBlock([]byte(formContent)); err != nil {
				c.Redirect(http.StatusSeeOther, "/error")
				return
			}
		}
		c.Redirect(http.StatusSeeOther, "/console")
	})

	router.GET("/console/getChain", func(c *gin.Context) {
		c.JSON(http.StatusOK, blockchainService.chain)
	})

	router.GET("console/deleteChain", func(c *gin.Context) {
		c.HTML(http.StatusOK, "deleteChain.html", gin.H{
			"title":   "GoLinks | Delete Chain",
			"heading": "Delete Chain?",
		})
	})

	router.POST("console/deleteChain", func(c *gin.Context) {
		blockchainService.resetChain()
		c.Redirect(http.StatusSeeOther, "/console")
	})

}

var consoleCards = []ConsoleCard{
	ConsoleCard{
		Title: "Block Chainer",
		Options: []CardOption{
			CardOption{
				Label: "Add Block",
				URL:   "/console/addBlock",
			},
			CardOption{
				Label: "Get Chain",
				URL:   "/console/getChain",
			},
			CardOption{
				Label: "Delete Chain",
				URL:   "/console/deleteChain",
			},
		},
	},
}
