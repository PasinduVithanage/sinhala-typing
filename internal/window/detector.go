//go:build windows

package window

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procQueryFullProcessImageName = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryFullProcessImageNameW")
)

// WindowInfo holds metadata about a foreground window.
type WindowInfo struct {
	Handle    windows.HWND
	Title     string
	ClassName string
	ProcessID uint32
	Exe       string
}

// CompatMode describes how to inject text into a target application.
type CompatMode int

const (
	CompatModeKeyboard  CompatMode = iota
	CompatModeClipboard
	CompatModeUIA
	CompatModeBlocked
)

// KnownApps maps executable names to their preferred injection mode.
var KnownApps = map[string]CompatMode{
	"chrome.exe":      CompatModeKeyboard,
	"firefox.exe":     CompatModeKeyboard,
	"msedge.exe":      CompatModeKeyboard,
	"WINWORD.EXE":     CompatModeKeyboard,
	"notepad.exe":     CompatModeKeyboard,
	"notepad++.exe":   CompatModeKeyboard,
	"Code.exe":        CompatModeKeyboard,
	"Slack.exe":       CompatModeKeyboard,
	"Discord.exe":     CompatModeKeyboard,
	"WhatsApp.exe":    CompatModeKeyboard,
	"EXCEL.EXE":       CompatModeKeyboard,
	"POWERPNT.EXE":    CompatModeKeyboard,
	"photoshop.exe":   CompatModeClipboard,
	"illustrator.exe": CompatModeClipboard,
	"java.exe":        CompatModeClipboard,
}

func GetForegroundWindowInfo() (*WindowInfo, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return nil, nil
	}

	info := &WindowInfo{Handle: windows.HWND(hwnd)}
	buf := make([]uint16, 512)

	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	info.Title = syscall.UTF16ToString(buf)

	procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	info.ClassName = syscall.UTF16ToString(buf)

	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	info.ProcessID = pid
	info.Exe = getProcessExe(pid)

	return info, nil
}

func getProcessExe(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)

	buf := make([]uint16, 512)
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageName.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		return ""
	}
	full := syscall.UTF16ToString(buf[:size])
	// Return just the filename
	for i := len(full) - 1; i >= 0; i-- {
		if full[i] == '\\' || full[i] == '/' {
			return full[i+1:]
		}
	}
	return fmt.Sprintf("%s", full)
}

func GetCompatMode(exe string) CompatMode {
	if mode, ok := KnownApps[exe]; ok {
		return mode
	}
	return CompatModeKeyboard
}
