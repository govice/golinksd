package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Webserver struct {
	router            *gin.Engine
	daemon            *daemon
	blockchainService *BlockchainService
	workerService     *WorkerService
}

func NewWebserver(daemon *daemon) (*Webserver, error) {
	w := &Webserver{
		router: gin.Default(),
		daemon: daemon,
	}

	bs, err := NewBlockchainService(daemon)
	if err != nil {
		errln("failed to initialize blockchain service")
		return nil, err
	}
	w.blockchainService = bs

	ws, err := NewWorkerService(daemon)
	if err != nil {
		errln("failed to initializie worker service")
		return nil, err
	}
	w.workerService = ws

	//TODO remove with load ledger
	if viper.GetBool("genesis") {
		bs.resetChain()
	}

	templateResourceHome := viper.GetString("templates_home")
	logln("looking for templates in", templateResourceHome)
	_, err = os.Stat(templateResourceHome)
	// TODO in dev mode this should force to cwd
	if os.IsNotExist(err) {
		logln("defaulting to templates in cwd")
		w.router.LoadHTMLGlob("./templates/*")
	} else {
		logln("loading templates from", templateResourceHome)
		w.router.LoadHTMLGlob(templateResourceHome + "/*")
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
