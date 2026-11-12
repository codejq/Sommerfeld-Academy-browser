//go:build windows

// Sommerfeld Academy Browser: a WebView2 window locked to https://s-i-u.com/
// in which media may autoplay without a user gesture.
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

const (
	homeURL = "https://s-i-u.com/"
	appName = "Sommerfeld Academy"
)

var version = "dev"

// browserArgs are passed to the WebView2 browser process (through
// ICoreWebView2EnvironmentOptions, see third_party/go-webview2). The autoplay
// policy lets play() start audio/video before the user has interacted with the
// page (otherwise it rejects with NotAllowedError).
const browserArgs = "--autoplay-policy=no-user-gesture-required"

// initScript runs in every page before its own scripts: browser-style
// keyboard shortcuts, since the window has no toolbar.
const initScript = `
window.addEventListener('keydown', function (e) {
  if (e.key === 'F5' || (e.ctrlKey && e.key === 'r')) { e.preventDefault(); location.reload(); }
  else if (e.altKey && e.key === 'ArrowLeft') { e.preventDefault(); history.back(); }
  else if (e.altKey && e.key === 'ArrowRight') { e.preventDefault(); history.forward(); }
  else if (e.altKey && e.key === 'Home') { e.preventDefault(); location.href = '` + homeURL + `'; }
}, true);
`

// audioScript asks the app to come forward (sabAudioStarted, bound in main)
// when audible sound starts: an unmuted <audio>/<video> starts playing, or a
// Web Audio source starts. Calls are throttled to one per second.
const audioScript = `
(function () {
  var last = 0;
  function started() {
    var now = Date.now();
    if (now - last < 1000 || !window.sabAudioStarted) return;
    last = now;
    window.sabAudioStarted();
  }
  document.addEventListener('playing', function (e) {
    var m = e.target;
    if (m instanceof HTMLMediaElement && !m.muted && m.volume > 0) started();
  }, true);
  if (window.AudioScheduledSourceNode) {
    var start = AudioScheduledSourceNode.prototype.start;
    AudioScheduledSourceNode.prototype.start = function () {
      if (this.context && this.context.state === 'running') started();
      return start.apply(this, arguments);
    };
  }
})();
`

func main() {
	runtime.LockOSThread()
	if !singleInstance() {
		return // the running copy was brought to the front
	}
	if autostartEnabled() {
		_ = applyAutostart(true)
	}

	var w webview2.WebView
	quitting := false
	w = webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus:   true,
		DataPath:    dataDir(),
		BrowserArgs: browserArgs,
		// No internet or a server error: show the offline screen, which
		// retries the page every 10 seconds.
		NavigationFailed: func(url string, status int) { w.SetHtml(offlinePage(url)) },
		// Closing the window only hides it to the tray; "Exit" there quits.
		OnClose: func() bool {
			if quitting {
				return true
			}
			hideWindow(uintptr(w.Window()))
			return false
		},
		WindowOptions: webview2.WindowOptions{
			Title: appName, Width: 1280, Height: 820, IconId: 1, Center: true,
		},
	})
	if w == nil { // WebView2 runtime missing: fall back to the default browser
		openDefault(homeURL)
		return
	}
	defer w.Destroy()
	w.SetSize(800, 560, webview2.HintMin)
	w.Init(initScript)
	w.Init(audioScript)
	hwnd := uintptr(w.Window())
	w.Bind("sabAudioStarted", func() { w.Dispatch(func() { bringToFront(hwnd) }) })
	trayDone := runTray(w, hwnd, func() {
		quitting = true
		w.Destroy()
	})
	if launchedAtSignIn() {
		procShowWindow.Call(hwnd, swMinimize) // started with Windows: wait on the taskbar
	}
	w.Navigate(homeURL)
	w.Run()
	systray.Quit() // also when the window was closed another way (e.g. logoff)
	select {       // let the tray remove its icon, so no ghost icon is left
	case <-trayDone:
	case <-time.After(2 * time.Second):
	}
}

// dataDir keeps cookies, logins and cache per user.
func dataDir() string {
	base, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "SommerfeldAcademyBrowser", "webview2")
}

func openDefault(url string) {
	verb, _ := windows.UTF16PtrFromString("open")
	u, _ := windows.UTF16PtrFromString(url)
	_ = windows.ShellExecute(0, verb, u, nil, nil, windows.SW_SHOWNORMAL)
}
