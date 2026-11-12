# Sommerfeld Academy Browser

**Download the latest version:** [SommerfeldAcademy.exe](https://github.com/codejq/Sommerfeld-Academy-browser/releases/latest/download/SommerfeldAcademy.exe)

A small Windows app that opens https://s-i-u.com/ in its own window (Microsoft Edge WebView2).
Audio and video on the site may start without a click first: the WebView2 runtime is started with
`--autoplay-policy=no-user-gesture-required`, so `play()` no longer rejects with `NotAllowedError`.

Stack: Go (CGO off) + `github.com/jchv/go-webview2`, the same as SonarBridge. The library is vendored
in `third_party/go-webview2` (via a `replace` in go.mod) with one addition,
`pkg/edge/environmentoptions.go`: a Go implementation of `ICoreWebView2EnvironmentOptions`, exposed as
`WebViewOptions.BrowserArgs`. It is needed because the WebView2 loader embedded in the library ignores
the `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS` environment variable.

## Build

```powershell
./build.ps1            # -> dist\SommerfeldAcademy.exe
./build.ps1 -Version 1.2.0
```

Needs Go 1.27+. The icon is cut from `build/winres/logo.png` by `tools/genicon` (delete `icon.png` to regenerate). The exe is a single file; it needs the WebView2 runtime (built into Windows 11 and
current Windows 10). Without it the site opens in the default browser instead.

## Behaviour

- **Autoplay:** media may start without a click (see above).
- **Sound brings the window forward:** when the page starts playing audible sound (an unmuted
  `<audio>`/`<video>`, or Web Audio) while the app is minimized or behind other windows, the window
  is restored, brought to the front and focused (`focus.go`). If Windows
  refuses the focus change, the taskbar button flashes instead.
- **Offline screen:** if a page fails to load (no internet, DNS or server error), an offline screen
  (`offline.go`) is shown. It retries the page every 10 seconds, at once when Windows reports the
  network is back, or when the user presses "Retry now".

The vendored library also gained `WebViewOptions.NavigationFailed` (`pkg/edge/navigation.go`).

- **Tray:** closing the window hides it to the notification area; the site keeps running (sound
  still brings the window forward). Click the tray icon or "Open" to show it, "Exit" to quit.
  Starting the app again while it runs shows the existing window instead of a second copy.

- **Starts with Windows:** the app adds itself to the user's startup programs
  (`HKCU\...\CurrentVersion\Run`, no admin needed) and then opens minimized to the taskbar at sign-in.
  Untick "Start with Windows" in the tray menu to stop that; the choice is remembered.

## Use

| Key | Action |
|---|---|
| F5 / Ctrl+R | Reload |
| Alt+← / Alt+→ | Back / forward |
| Alt+Home | Home page |

Cookies and logins are kept per user in `%LOCALAPPDATA%\SommerfeldAcademyBrowser\webview2`.
