# Project Context — Sommerfeld Academy Browser

## What it is
A single-file Windows desktop app (`SommerfeldAcademy.exe`) that opens **https://s-i-u.com/**
(Sommerfeld International Academy) in its own window, using Microsoft Edge **WebView2**. It is a
kiosk-style "site wrapper": no address bar, locked to the academy site, tuned so lesson audio/video
works without user interaction.

## Why it exists
- The site plays audio/video that browsers block until the user clicks (`NotAllowedError`). The app
  starts WebView2 with `--autoplay-policy=no-user-gesture-required`, so `play()` just works.
- When sound starts while the window is minimized/hidden, the app comes to the front so the student
  notices.
- It stays running in the tray and starts with Windows, so it is always available.

## Tech stack
| Area | Choice |
|---|---|
| Language | Go 1.27+, `CGO_ENABLED=0`, `GOOS=windows GOARCH=amd64` |
| Web view | `github.com/jchv/go-webview2`, **vendored and patched** in `third_party/go-webview2` (via `replace` in `go.mod`) |
| Tray | `fyne.io/systray` |
| Win32 / registry | `golang.org/x/sys/windows` |
| Icon / exe resources | `tools/genicon` (makes `icon.png`/`icon.ico` from the logo) + `go-winres` (manifest, version info, icon → `rsrc_windows_amd64.syso`) |

## Source layout
| File | Purpose |
|---|---|
| `main.go` | Entry point: single instance, WebView2 window, init scripts (keyboard shortcuts, audio detection), navigation |
| `focus.go` | Win32 helpers: bring window to front / restore / flash taskbar; single-instance check |
| `offline.go` | Bilingual (Arabic/English) offline screen with auto-retry every 10 s |
| `tray.go` | Tray icon and menu (Open, Start with Windows, Exit) |
| `autostart.go` | `HKCU\...\CurrentVersion\Run` entry; `--autostart` flag starts minimized |
| `tools/genicon/` | Icon generator (cuts round medallion from `build/winres/logo.png`) |
| `build/winres/` | `winres.json` (manifest + version resource), `logo.png`, `icon.png`, `icon.ico` |
| `third_party/go-webview2/` | Patched library — adds `WebViewOptions.BrowserArgs` (`pkg/edge/environmentoptions.go`) and `NavigationFailed` (`pkg/edge/navigation.go`) |
| `build.ps1` | Local/CI build script → `dist\SommerfeldAcademy.exe` |
| `.github/workflows/build.yml` | CI: builds on every push to `main` and publishes a tagged release |

## Key behaviours
- **Autoplay** without a user gesture.
- **Sound brings the window forward** (JS hook on `playing` events and Web Audio `start()` → bound Go
  function `sabAudioStarted` → `bringToFront`).
- **Offline screen** on navigation failure; retries every 10 s, on network-back, or "Retry now".
- **Close = hide to tray**; tray "Exit" quits. A second launch shows the running window.
- **Starts with Windows** (on by default, can be unticked in the tray; choice stored in
  `HKCU\Software\SommerfeldAcademyBrowser`).
- Shortcuts: F5/Ctrl+R reload, Alt+←/→ back/forward, Alt+Home home.
- User data (cookies, logins): `%LOCALAPPDATA%\SommerfeldAcademyBrowser\webview2`.
- If the WebView2 runtime is missing, the site opens in the default browser instead.

## Build
```powershell
./build.ps1                 # -> dist\SommerfeldAcademy.exe (version 1.0.0)
./build.ps1 -Version 1.2.3  # version is stamped into the exe resources and main.version
```
`dist/` and `*.syso` are git-ignored (generated).

## CI / Releases
- Every push to `main` runs `.github/workflows/build.yml` on `windows-latest`:
  1. Computes the next version `v1.0.<run number>` and builds with `build.ps1 -Version`.
  2. Creates tag `v1.0.N` on the pushed commit and a GitHub Release with `SommerfeldAcademy.exe`
     attached, marked as **latest**.
- **Permanent download link (always the newest build):**
  `https://github.com/codejq/Sommerfeld-Academy-browser/releases/latest/download/SommerfeldAcademy.exe`
  This works because each release is marked "latest" and the asset name never changes. Keep the
  exe name `SommerfeldAcademy.exe` or the link breaks.
- To bump the major/minor part, edit `VERSION_PREFIX` in the workflow.

## Conventions / gotchas
- Windows-only code: every Go file has `//go:build windows`.
- Do not replace the vendored WebView2 library with upstream — upstream ignores
  `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`, so autoplay would break.
- UI text for end users is bilingual Arabic · English.
- Brand colours: navy `#14234b`, gold `#f2b705`.
