package webserver

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/govice/golinks-daemon/pkg/authentication"
	"github.com/govice/golinks-daemon/pkg/blockchain"
	"github.com/govice/golinks-daemon/pkg/log"
	"github.com/govice/golinks-daemon/pkg/worker"
	"github.com/spf13/viper"
)

type Webserver struct {
	router   *gin.Engine
	servicer Servicer
}

type BlockchainServicer interface {
	BlockchainService() *blockchain.Service
}

type WorkerServicer interface {
	WorkerService() *worker.Service
}

type AuthenticationServicer interface {
	AuthenticationService() *authentication.Service
}

type Servicer interface {
	BlockchainServicer
	WorkerServicer
	AuthenticationServicer
}

func New(servicer Servicer) (*Webserver, error) {
	return &Webserver{
		router:   gin.Default(),
		servicer: servicer,
	}, nil
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
	log.Logln("received termination on webserver context")

	return frontendErr
}
