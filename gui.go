package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"github.com/govice/golinks-daemon/scene"
)

type GUI struct {
	daemon *daemon
	app    fyne.App
}

func NewGUI(daemon *daemon) (*GUI, error) {
	return &GUI{
		daemon: daemon,
		app:    app.New(),
	}, nil
}

func (g *GUI) ShowAndRun() {
	scene.Primary(g.app)
	g.app.Run()
}
