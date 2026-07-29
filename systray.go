package main

import (
	"github.com/getlantern/systray"
)

func runTray(iconData []byte, makeMore, showWindow, disconnectAll, quit chan<- struct{}, tooltip <-chan string) {
	systray.Run(func() {
		systray.SetIcon(iconData)
		systray.SetTitle("FishTun")
		systray.SetTooltip("FishTun")

		makeMoreItem := systray.AddMenuItem("Make More Tunnels", "Open window to add another tunnel")
		showWindowItem := systray.AddMenuItem("Show Window", "Show the application window")
		systray.AddSeparator()
		disconnectAllItem := systray.AddMenuItem("Disconnect All", "Disconnect all tunnels")
		quitItem := systray.AddMenuItem("Quit", "Stop all tunnels and quit application")

		pipe := func(src <-chan struct{}, dst chan<- struct{}) {
			go func() {
				for range src {
					dst <- struct{}{}
				}
			}()
		}

		pipe(makeMoreItem.ClickedCh, makeMore)
		pipe(showWindowItem.ClickedCh, showWindow)
		pipe(disconnectAllItem.ClickedCh, disconnectAll)
		pipe(quitItem.ClickedCh, quit)

		go func() {
			for msg := range tooltip {
				systray.SetTooltip(msg)
			}
		}()
	}, func() {})
}
