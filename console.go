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
	"net/http"

	"github.com/gin-gonic/gin"
)

func (w *Webserver) registerConsoleHandlers() error {
	router := w.router
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

	return nil
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
