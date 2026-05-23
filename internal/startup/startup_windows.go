//go:build windows

package startup

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName = "SinhalaAssistant"
)

// Set writes or removes the Windows startup registry value for the current user.
func Set(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if !enable {
		return k.DeleteValue(appName)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(appName, `"`+exe+`"`)
}

// IsEnabled reports whether the startup registry value exists for the current user.
func IsEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	return err == nil
}
