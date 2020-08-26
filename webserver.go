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
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Webserver struct {
	router *gin.Engine
	daemon *daemon
}

func NewWebserver(daemon *daemon) (*Webserver, error) {
	w := &Webserver{
		router: gin.Default(),
		daemon: daemon,
	}

	if err := w.registerFrontendRoutes(); err != nil {
		errln("failed to initialize frontend routes")
		return nil, err
	}

	if err := w.registerAPIRoutes(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Webserver) registerFrontendRoutes() error {
	templatesHome := viper.GetString("templates_home")
	log.Println("Templates Home: " + templatesHome)
	if templatesHome != "" {
		w.router.LoadHTMLGlob(templatesHome + "/*")
	} else {
		w.router.LoadHTMLGlob("./templates/*")
	}
	w.router.GET("/error", func(c *gin.Context) {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title": "GoLinks | Error",
		})
	})

	return w.registerConsoleHandlers()
}

func (w *Webserver) registerAPIRoutes() error {
	apiGroup := w.router.Group("/api")
	apiGroup.Use(w.externalAuthenticator())
	{
		apiGroup.POST("/chain", w.postBlockEndpoint)
		apiGroup.GET("/chain", w.getChainEndpoint)
		apiGroup.POST("/chain/find", w.findBlockEndpoint)
	}

	return nil
}

func (w *Webserver) Execute(ctx context.Context) error {
	var frontendErr error
	go func() {
		if err := w.router.Run(":" + viper.GetString("port")); err != nil {
			frontendErr = err
		} // listen and serve on PORT
	}()
	<-ctx.Done()
	logln("received termination on webserver context")

	return frontendErr
}
