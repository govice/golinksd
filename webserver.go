package main

import (
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
	apiGroup.Use(externalAuthenticator())
	{
		apiGroup.POST("/chain", postBlockEndpoint)
		apiGroup.GET("/chain", getChainEndpoint)
		apiGroup.POST("/chain/find", findBlockEndpoint)
	}

	return nil
}

func (w *Webserver) Execute() error {
	if err := w.router.Run(":" + viper.GetString("port")); err != nil {
		return err
	} // listen and serve on PORT

	return nil
}
