package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if handled, code := runCLI(os.Args[1:]); handled {
		os.Exit(code)
	}

	app := NewApp()
	err := wails.Run(&options.App{
		Title:     "Varnqor",
		Width:     1180,
		Height:    760,
		MinWidth:  920,
		MinHeight: 620,
		AssetServer: &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 9, G: 10, B: 12, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
