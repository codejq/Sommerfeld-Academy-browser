//go:build windows

package main

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// Start with Windows: a value under the user's Run key launches the exe with
// autostartFlag at sign-in (no admin needed). It is on by default and written
// on every start, so it follows the exe if it moves; unticking "Start with
// Windows" in the tray removes it and records that choice.
const (
	autostartFlag = "--autostart"
	runKey        = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue      = "Sommerfeld Academy"
	settingsKey   = `Software\SommerfeldAcademyBrowser`
	disabledValue = "AutostartDisabled"
)

// launchedAtSignIn reports whether Windows started the app at sign-in.
func launchedAtSignIn() bool {
	for _, a := range os.Args[1:] {
		if a == autostartFlag {
			return true
		}
	}
	return false
}

// autostartEnabled reports whether the user wants the app to start with Windows.
func autostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, settingsKey, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue(disabledValue)
	return err != nil || v == 0
}

// setAutostart adds or removes the Run entry and remembers the choice.
func setAutostart(on bool) error {
	s, _, err := registry.CreateKey(registry.CURRENT_USER, settingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer s.Close()
	disabled := uint32(1)
	if on {
		disabled = 0
	}
	if err := s.SetDWordValue(disabledValue, disabled); err != nil {
		return err
	}
	return applyAutostart(on)
}

// applyAutostart writes (on) or deletes (off) the Run entry.
func applyAutostart(on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		if err := k.DeleteValue(runValue); err != nil && err != registry.ErrNotExist {
			return err
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	return k.SetStringValue(runValue, `"`+exe+`" `+autostartFlag)
}
