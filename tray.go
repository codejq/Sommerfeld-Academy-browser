//go:build windows

package main

import (
	_ "embed"
	"runtime"

	"fyne.io/systray"
	"github.com/jchv/go-webview2"
)

//go:embed build/winres/icon.ico
var trayIcon []byte

// runTray shows the tray icon on its own OS thread: a left click (or "Open")
// shows the window, "Exit" quits the app. quit is called on the UI thread.
// The returned channel is closed once the icon has been removed.
func runTray(w webview2.WebView, hwnd uintptr, quit func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread() // systray runs its own message loop
		systray.Run(func() {
			systray.SetIcon(trayIcon)
			systray.SetTitle(appName)
			systray.SetTooltip(appName)
			show := func() { w.Dispatch(func() { bringToFront(hwnd) }) }
			systray.SetOnTapped(show)
			mOpen := systray.AddMenuItem("فتح · Open", "")
			mAuto := systray.AddMenuItemCheckbox("تشغيل مع Windows · Start with Windows", "", autostartEnabled())
			systray.AddSeparator()
			mExit := systray.AddMenuItem("خروج · Exit", "")
			go func() {
				for {
					select {
					case <-mOpen.ClickedCh:
						show()
					case <-mAuto.ClickedCh:
						on := !mAuto.Checked()
						if setAutostart(on) == nil {
							if on {
								mAuto.Check()
							} else {
								mAuto.Uncheck()
							}
						}
					case <-mExit.ClickedCh:
						w.Dispatch(quit)
						return
					}
				}
			}()
		}, func() { close(done) })
	}()
	return done
}
