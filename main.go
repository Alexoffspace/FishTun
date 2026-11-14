package main

import (
	"context"
	"embed"
	"log"
	"runtime"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icon.png
var iconPNG []byte

//go:embed icon.ico
var iconICO []byte

// trayIcon returns the icon bytes in a format the platform can render. Windows
// (Win32 LoadImage/CreateIconFromResource) requires a real ICO container; other
// platforms accept PNG directly.
func trayIcon() []byte {
	if runtime.GOOS == "windows" {
		return iconICO
	}
	return iconPNG
}

func main() {
	// Enforce a single running instance: a second launch just surfaces the
	// window of the instance that is already running, then exits.
	primary, err := acquireSingleInstance()
	if err != nil {
		log.Printf("single-instance check failed: %v", err)
	}
	if !primary {
		return
	}
	defer releaseInstance()

	app := NewApp()
	app.iconData = trayIcon()
	app.startTray()

	err = wails.Run(&options.App{
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
		OnBeforeClose:    app.onBeforeClose,
		OnShutdown: func(ctx context.Context) {
			systray.Quit()
			app.DisconnectAll()
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
			Icon:             iconPNG,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
