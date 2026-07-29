package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icon.png
var iconData []byte

func main() {
	app := NewApp()
	app.iconData = iconData

	err := wails.Run(&options.App{
		Title:     "FishTun",
		Width:     520,
		Height:    680,
		MinWidth:  480,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 30, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			app.DisconnectAll()
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
			Icon:             iconData,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
