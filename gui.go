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
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type GUI struct {
	daemon     *daemon
	app        fyne.App
	mainWindow *fyne.Window
}

func NewGUI(daemon *daemon) (*GUI, error) {
	a := app.New()
	return &GUI{
		daemon: daemon,
		app:    a,
	}, nil
}

func (g *GUI) ShowAndRun() {
	g.showPrimaryScene()
	g.app.Run()
}

func (g *GUI) showPrimaryScene() {
	logln("showing primary scene")
	menuItem := fyne.NewMenuItem("Item1", func() { fmt.Println("menu item 1") })

	preferencesItem := fyne.NewMenuItem("Preferences", func() { fmt.Println("settings") })

	mainWindow := g.app.NewWindow("golinks daemon")
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File", menuItem),
		fyne.NewMenu("Edit", preferencesItem),
	)

	mainWindow.SetMainMenu(mainMenu)

	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Home", theme.HomeIcon(), g.homeScene()),
		widget.NewTabItemWithIcon("Workers", theme.ComputerIcon(), g.workersScene()),
	)

	tabs.SetTabLocation(widget.TabLocationLeading)
	tabs.SelectTabIndex(0)
	mainWindow.SetContent(tabs)
	mainWindow.Resize(fyne.NewSize(800, 600))

	g.mainWindow = &mainWindow
	mainWindow.Show()
}

func (g *GUI) homeScene() fyne.CanvasObject {
	vbox := widget.NewVBox(
		widget.NewLabelWithStyle("Home screen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewHBox(layout.NewSpacer()),
	)
	return vbox
}

func (g *GUI) workersScene() fyne.CanvasObject {
	var workerAccordionItems []*widget.AccordionItem
	for _, worker := range g.daemon.workerManager.WorkerConfig.Workers {
		workerAccordionItems = append(workerAccordionItems, widget.NewAccordionItem(worker.RootPath, g.makeWorkerForm(worker)))

	}
	return widget.NewVBox(
		widget.NewLabelWithStyle("Workers", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewAccordionContainer(workerAccordionItems...),
	)
}

func (g *GUI) makeWorkerForm(worker *Worker) fyne.CanvasObject {
	browserButton := widget.NewButton("Open", func() {
		dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
			if err != nil {
				errln("worker root file opener error", err)
				return
			}

			if closer != nil {
				logln("closer Name", closer.Name())
			}
		}, *g.mainWindow)
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Root", Widget: browserButton},
		},
	}

	return form
}
