package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
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

func (g *GUI) Show() {
	w := g.app.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(widget.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.Show()
}
