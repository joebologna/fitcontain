package main

import (
	"embed"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed media/*.png
var media embed.FS

func main() {
	os.Setenv("FYNE_THEME", "dark")

	myApp := app.New()
	myWindow := myApp.NewWindow("Fit Contain")

	padding := float32(8)
	scale := float32(2)
	// the image was originally 880x720, it's larger now - but the aspect ratio is the same, so let's just use it
	screenSize := fyne.NewSize(880/scale, 720/scale).AddWidthHeight(padding, padding)
	myWindow.Resize(screenSize)

	stuff := FitContain(media, screenSize, 1)

	myWindow.SetContent(stuff)

	myWindow.ShowAndRun()
}
