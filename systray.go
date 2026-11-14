package main

import (
	"github.com/energye/systray"
)

// onTrayReady builds the tray menu. It is called by systray only after the
// internal window/menu exists, so menu items must be created here.
//
// Click callbacks run on systray's own message-pump goroutine. They must never
// block: every handler forwards to the Wails dispatcher with a non-blocking send.
func onTrayReady(iconData []byte, makeMore, showWindow, disconnectAll, quit chan<- struct{}) {
	systray.SetIcon(iconData)
	systray.SetTooltip("FishTun")

	notify := func(dst chan<- struct{}) {
		select {
		case dst <- struct{}{}:
		default:
		}
	}

	// Left click shows the window; right click opens the menu (Windows/macOS).
	systray.SetOnClick(func(systray.IMenu) { notify(showWindow) })
	systray.SetOnRClick(func(m systray.IMenu) { _ = m.ShowMenu() })

	makeMoreItem := systray.AddMenuItem("Make More Tunnels", "Open window to add another tunnel")
	makeMoreItem.Click(func() { notify(makeMore) })
	systray.AddSeparator()
	showWindowItem := systray.AddMenuItem("Show Window", "Show the application window")
	showWindowItem.Click(func() { notify(showWindow) })
	disconnectAllItem := systray.AddMenuItem("Disconnect All", "Disconnect all tunnels")
	disconnectAllItem.Click(func() { notify(disconnectAll) })
	quitItem := systray.AddMenuItem("Quit", "Stop all tunnels and quit application")
	quitItem.Click(func() { notify(quit) })
}
