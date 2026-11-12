package edge

import (
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ICoreWebView2EnvironmentOptions, implemented in Go so browser arguments can
// be passed to CreateCoreWebView2EnvironmentWithOptions. (Added for the
// Sommerfeld Academy browser; the embedded loader ignores the
// WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS environment variable.)

var (
	iidIUnknown                        = NewGUID("{00000000-0000-0000-C000-000000000046}")
	iidICoreWebView2EnvironmentOptions = NewGUID("{2FDE08A8-1E9A-4766-8C05-95A9CEB9D1C5}")

	ole32              = windows.NewLazySystemDLL("ole32.dll")
	procCoTaskMemAlloc = ole32.NewProc("CoTaskMemAlloc")
)

const (
	sOK          = 0
	eNoInterface = 0x80004002
	ePointer     = 0x80004003
	eOutOfMemory = 0x8007000E
)

type environmentOptionsVtbl struct {
	_IUnknownVtbl
	GetAdditionalBrowserArguments             ComProc
	PutAdditionalBrowserArguments             ComProc
	GetLanguage                               ComProc
	PutLanguage                               ComProc
	GetTargetCompatibleBrowserVersion         ComProc
	PutTargetCompatibleBrowserVersion         ComProc
	GetAllowSingleSignOnUsingOSPrimaryAccount ComProc
	PutAllowSingleSignOnUsingOSPrimaryAccount ComProc
}

type environmentOptions struct {
	vtbl *environmentOptionsVtbl
	refs int32

	args, language, version string
	sso                     int32
}

var environmentOptionsFn = environmentOptionsVtbl{
	_IUnknownVtbl{
		NewComProc(func(this *environmentOptions, refiid *GUID, object *uintptr) uintptr {
			if object == nil {
				return ePointer
			}
			if *refiid == *iidIUnknown || *refiid == *iidICoreWebView2EnvironmentOptions {
				*object = uintptr(unsafe.Pointer(this))
				atomic.AddInt32(&this.refs, 1)
				return sOK
			}
			*object = 0
			return eNoInterface
		}),
		NewComProc(func(this *environmentOptions) uintptr { return uintptr(atomic.AddInt32(&this.refs, 1)) }),
		NewComProc(func(this *environmentOptions) uintptr { return uintptr(atomic.AddInt32(&this.refs, -1)) }),
	},
	NewComProc(func(this *environmentOptions, out *uintptr) uintptr { return coTaskString(this.args, out) }),
	NewComProc(func(this *environmentOptions, v *uint16) uintptr { this.args = windows.UTF16PtrToString(v); return sOK }),
	NewComProc(func(this *environmentOptions, out *uintptr) uintptr { return coTaskString(this.language, out) }),
	NewComProc(func(this *environmentOptions, v *uint16) uintptr {
		this.language = windows.UTF16PtrToString(v)
		return sOK
	}),
	NewComProc(func(this *environmentOptions, out *uintptr) uintptr { return coTaskString(this.version, out) }),
	NewComProc(func(this *environmentOptions, v *uint16) uintptr {
		this.version = windows.UTF16PtrToString(v)
		return sOK
	}),
	NewComProc(func(this *environmentOptions, out *int32) uintptr {
		if out == nil {
			return ePointer
		}
		*out = this.sso
		return sOK
	}),
	NewComProc(func(this *environmentOptions, v int32) uintptr { this.sso = v; return sOK }),
}

func newEnvironmentOptions(args string) *environmentOptions {
	return &environmentOptions{
		vtbl: &environmentOptionsFn,
		refs: 1,
		args: args,
		// The oldest runtime that understands these options (WebView2 SDK 1.0.622).
		version: "86.0.616.0",
	}
}

// coTaskString returns s as a CoTaskMemAlloc'd UTF-16 string, as COM getters must.
func coTaskString(s string, out *uintptr) uintptr {
	if out == nil {
		return ePointer
	}
	u, err := windows.UTF16FromString(s)
	if err != nil {
		u = []uint16{0}
	}
	p, _, _ := procCoTaskMemAlloc.Call(uintptr(len(u) * 2))
	if p == 0 {
		*out = 0
		return eOutOfMemory
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u)), u)
	*out = p
	return sOK
}
