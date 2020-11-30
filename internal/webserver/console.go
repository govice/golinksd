package webserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/govice/golinksd/pkg/log"
	"github.com/govice/golinksd/pkg/worker"
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
		workersCard := &ConsoleCard{
			Title: "Workers",
			Buttons: []*CardButton{
				{
					Label: "Add Worker",
					Class: "btn btn-block btn-success",
					URL:   "/console/worker/add",
				},
			},
		}

		for index, worker := range w.servicer.WorkerService().WorkerConfig.Workers {
			option := &CardOption{
				Label: worker.RootPath,
				URL:   "/console/worker/view/" + strconv.Itoa(index),
			}
			workersCard.Options = append(workersCard.Options, option)
		}
		consoleCards := []*ConsoleCard{
			{
				Title: "Block Chainer",
				Options: []*CardOption{
					{
						Label: "Add Block",
						URL:   "/console/addBlock",
					},
					{
						Label: "Get Chain",
						URL:   "/console/getChain",
					},
					{
						Label: "Delete Chain",
						URL:   "/console/deleteChain",
					},
				},
			},
			workersCard,
		}

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
			if _, err := w.servicer.BlockchainService().AddBlock([]byte(formContent)); err != nil {
				c.Redirect(http.StatusSeeOther, "/error")
				return
			}
		}
		c.Redirect(http.StatusSeeOther, "/console")
	})

	router.GET("/console/getChain", func(c *gin.Context) {
		c.JSON(http.StatusOK, w.servicer.BlockchainService().Chain())
	})

	router.GET("console/deleteChain", func(c *gin.Context) {
		c.HTML(http.StatusOK, "deleteChain.html", gin.H{
			"title":   "GoLinks | Delete Chain",
			"heading": "Delete Chain?",
		})
	})

	router.POST("console/deleteChain", func(c *gin.Context) {
		w.servicer.BlockchainService().ResetChain()
		c.Redirect(http.StatusSeeOther, "/console")
	})

	router.GET("console/worker/view/:id", func(c *gin.Context) {
		idStr, ok := c.Params.Get("id")
		if !ok {
			log.Logln("invalid worker id:", idStr)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		worker, err := w.servicer.WorkerService().GetWorkerByIndex(id)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.HTML(http.StatusOK, "worker.tmpl.html", gin.H{
			"Root":          worker.RootPath,
			"Index":         id,
			"RefreshPeriod": worker.GenerationPeriod,
		})

	})

	router.POST("console/worker/delete/:id", func(c *gin.Context) {
		idStr, ok := c.Params.Get("id")
		if !ok {
			log.Logln("invalid worker id:", idStr)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := w.servicer.WorkerService().DeleteWorkerByIndex(id); err != nil {
			log.Logln(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Redirect(http.StatusSeeOther, "/console")
	})

	router.GET("console/worker/add", func(c *gin.Context) {
		c.HTML(http.StatusOK, "workerAdd.tmpl.html", nil)
	})

	router.POST("console/worker/add", func(c *gin.Context) {
		rootPath, ok := c.GetPostForm("workerRoot")
		if !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		generationPeriodStr, ok := c.GetPostForm("generationPeriod")
		if !ok {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		generationPeriod, err := strconv.Atoi(generationPeriodStr)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		//TODO IGNORE PATHS
		if err := w.servicer.WorkerService().AddWorker(&worker.NewWorkerConfig{
			RootPath:         rootPath,
			GenerationPeriod: generationPeriod,
		}); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusSeeOther, "/console")
	})

	return nil

}

//CardOption is used in templating to display options in a ConsoleCard
type CardOption struct {
	Label string `json:"label"`
	URL   string `json:"URL"`
}

type CardButton struct {
	Label string
	URL   string
	Class string
}

//ConsoleCard is in templating to display a card on the console
type ConsoleCard struct {
	Title   string        `json:"title"`
	Options []*CardOption `json:"options"`
	Buttons []*CardButton `json:"buttons"`
}
