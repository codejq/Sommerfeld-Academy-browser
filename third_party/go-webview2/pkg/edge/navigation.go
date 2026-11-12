package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Navigation results. (Added for the Sommerfeld Academy browser.)

// COREWEBVIEW2_WEB_ERROR_STATUS values used by callers.
const (
	WebErrorStatusOperationCanceled = 14
)

var procCoTaskMemFree = ole32.NewProc("CoTaskMemFree")

// IsSuccess reports whether the navigation reached its page (false for an
// error page: no network, DNS failure, HTTP 4xx/5xx, ...).
func (i *ICoreWebView2NavigationCompletedEventArgs) IsSuccess() bool {
	var ok int32
	i.vtbl.GetIsSuccess.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&ok)))
	return ok != 0
}

// WebErrorStatus is the COREWEBVIEW2_WEB_ERROR_STATUS of a failed navigation.
func (i *ICoreWebView2NavigationCompletedEventArgs) WebErrorStatus() int {
	var st int32
	i.vtbl.GetWebErrorStatus.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&st)))
	return int(st)
}

// GetSource returns the URL of the current (or last attempted) page.
func (i *ICoreWebView2) GetSource() string {
	var p *uint16
	i.vtbl.GetSource.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&p)))
	if p == nil {
		return ""
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(p)))
	return windows.UTF16PtrToString(p)
}
