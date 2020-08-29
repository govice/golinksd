package scene

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func HelloWorld(app fyne.App) {
	w := app.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(widget.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.Show()
}

func Primary(app fyne.App) {
	w := app.NewWindow("golinks daemon")

	menuItem := fyne.NewMenuItem("Item1", func() { fmt.Println("menu item 1") })

	preferencesItem := fyne.NewMenuItem("Preferences", func() { fmt.Println("settings") })

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File", menuItem),
		fyne.NewMenu("Edit", preferencesItem),
	)

	w.SetMainMenu(mainMenu)

	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Home", theme.HomeIcon(), homeScreen(app)),
	)

	tabs.SetTabLocation(widget.TabLocationLeading)
	tabs.SelectTabIndex(0)
	w.SetContent(tabs)
	w.Resize(fyne.NewSize(800, 600))

	w.Show()
}

func homeScreen(app fyne.App) fyne.CanvasObject {
	vbox := widget.NewVBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Home screen", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewHBox(layout.NewSpacer()),
	)
	return vbox
}
