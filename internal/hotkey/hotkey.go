//go:build windows

package hotkey

import (
	"runtime"
	"strings"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	modAlt     uint32 = 0x0001
	modControl uint32 = 0x0002
	modShift   uint32 = 0x0004
	modWin     uint32 = 0x0008

	wmHotkey = 0x0312
	wmNull   = 0x0000
	hotkeyID = 1
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessageHK     = user32.NewProc("GetMessageW")
	procPostThreadMsg    = user32.NewProc("PostThreadMessageW")
	procGetCurrentThread = kernel32.NewProc("GetCurrentThreadId")
)

// Hotkey listens for a system-wide hotkey and calls onTrigger when pressed.
type Hotkey struct {
	onTrigger func()
	stopCh    chan struct{}
	threadID  atomic.Uint32
}

func New(onTrigger func()) *Hotkey {
	return &Hotkey{
		onTrigger: onTrigger,
		stopCh:    make(chan struct{}),
	}
}

// Run registers combo (e.g. "Ctrl+Alt+S") and blocks in a message loop.
// Must be called in a dedicated goroutine.
func (h *Hotkey) Run(combo string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tid, _, _ := procGetCurrentThread.Call()
	h.threadID.Store(uint32(tid))

	mods, vk := parseCombo(combo)
	ret, _, _ := procRegisterHotKey.Call(0, hotkeyID, uintptr(mods), uintptr(vk))
	if ret == 0 {
		return
	}
	defer procUnregisterHotKey.Call(0, hotkeyID)

	type MSG struct {
		Hwnd    uintptr
		Message uint32
		WParam  uintptr
		LParam  uintptr
		Time    uint32
		Pt      [2]int32
	}
	var msg MSG
	for {
		select {
		case <-h.stopCh:
			return
		default:
		}
		ret, _, _ := procGetMessageHK.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			return
		}
		if msg.Message == wmHotkey && int(msg.WParam) == hotkeyID {
			h.onTrigger()
		}
	}
}

// Stop signals the hotkey loop to exit.
func (h *Hotkey) Stop() {
	select {
	case <-h.stopCh:
	default:
		close(h.stopCh)
	}
	// Wake GetMessageW so the goroutine can check stopCh.
	if tid := h.threadID.Load(); tid != 0 {
		procPostThreadMsg.Call(uintptr(tid), wmNull, 0, 0)
	}
}

// parseCombo converts "Ctrl+Alt+S" into Windows modifier flags and a VK code.
func parseCombo(combo string) (mods uint32, vk uint32) {
	parts := strings.Split(combo, "+")
	for i, p := range parts {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case "ctrl", "control":
			mods |= modControl
		case "alt":
			mods |= modAlt
		case "shift":
			mods |= modShift
		case "win":
			mods |= modWin
		default:
			if i == len(parts)-1 {
				p = strings.TrimSpace(p)
				if len(p) == 1 {
					vk = uint32(strings.ToUpper(p)[0])
				}
			}
		}
	}
	return
}
