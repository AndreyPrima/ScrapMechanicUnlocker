package main

import (
	"fyne.io/fyne/v2/app"

	"smunlocker/theme"
	"smunlocker/ui"
)

func main() {
	a := app.NewWithID("smunlocker")
	a.Settings().SetTheme(theme.Black{})
	w := ui.NewWindow(a)
	w.ShowAndRun()
}
