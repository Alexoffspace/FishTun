//go:build bindings

// Package main provides the entry point used only for Wails binding generation.
//
// Wails builds a temporary binary with the "bindings" tag and runs it to extract
// the bound method signatures; with that tag wails.Run() is replaced by a stub
// that only writes frontend/wailsjs and never starts the GUI. This file must
// therefore still call wails.Run with the same Bind list as the real main, but
// it must not start the system tray: the tray talks to the session D-Bus, which
// does not exist in headless build environments (CI containers).
package main

import (
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	app := NewApp()

	if err := wails.Run(&options.App{
		Title: "FishTun",
		Bind:  []interface{}{app},
	}); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
