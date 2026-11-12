//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procIsIconic                 = user32.NewProc("IsIconic")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procFindWindowW              = user32.NewProc("FindWindowW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procBringWindowToTop         = user32.NewProc("BringWindowToTop")
	procFlashWindowEx            = user32.NewProc("FlashWindowEx")
)

const (
	swHide     = 0
	swShow     = 5
	swMinimize = 6
	swRestore  = 9
)

// bringToFront shows the window if it is hidden in the tray, restores it if
// it is minimized, brings it in front of other windows and gives it the
// focus. Windows refuses SetForegroundWindow from a background app, so it
// briefly attaches to the input queue of the current foreground window; if
// that still fails the taskbar button flashes.
func bringToFront(hwnd uintptr) {
	if visible, _, _ := procIsWindowVisible.Call(hwnd); visible == 0 {
		procShowWindow.Call(hwnd, swShow)
	}
	fg, _, _ := procGetForegroundWindow.Call()
	if fg == hwnd {
		return // already in front
	}
	if iconic, _, _ := procIsIconic.Call(hwnd); iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
	}

	self := uintptr(windows.GetCurrentThreadId())
	other, _, _ := procGetWindowThreadProcessId.Call(fg, 0)
	if fg != 0 && other != 0 && other != self {
		procAttachThreadInput.Call(self, other, 1)
		defer procAttachThreadInput.Call(self, other, 0)
	}
	procBringWindowToTop.Call(hwnd)
	if ok, _, _ := procSetForegroundWindow.Call(hwnd); ok == 0 {
		flash(hwnd)
	}
}

// flash blinks the taskbar button until the window is activated.
func flash(hwnd uintptr) {
	type flashInfo struct {
		size    uint32
		hwnd    uintptr
		flags   uint32
		count   uint32
		timeout uint32
	}
	const flashAll, flashTimerNoFG = 0x3, 0xC
	fi := flashInfo{hwnd: hwnd, flags: flashAll | flashTimerNoFG}
	fi.size = uint32(unsafe.Sizeof(fi))
	procFlashWindowEx.Call(uintptr(unsafe.Pointer(&fi)))
}

// hideWindow hides the window to the tray; the page keeps running (and
// playing sound).
func hideWindow(hwnd uintptr) { procShowWindow.Call(hwnd, swHide) }

// singleInstance reports whether this is the only running copy. If another
// copy runs (perhaps hidden in the tray), its window is brought to the front
// and false is returned.
func singleInstance() bool {
	name, _ := windows.UTF16PtrFromString(`Local\SommerfeldAcademyBrowser`)
	if _, err := windows.CreateMutex(nil, false, name); err != windows.ERROR_ALREADY_EXISTS {
		return true // the handle is kept open until the process exits
	}
	class, _ := windows.UTF16PtrFromString("webview")
	title, _ := windows.UTF16PtrFromString(appName)
	if hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title))); hwnd != 0 {
		bringToFront(hwnd)
	}
	return false
}
