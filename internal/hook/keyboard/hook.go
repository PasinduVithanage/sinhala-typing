//go:build windows

package keyboard

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procSendInput           = user32.NewProc("SendInput")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	HC_ACTION      = 0
	LLKHF_INJECTED = 0x10

	VK_BACK   = 0x08
	VK_SHIFT  = 0x10
	VK_CTRL   = 0x11
	VK_MENU   = 0x12
	VK_SPACE  = 0x20
	VK_RETURN = 0x0D
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type KeyEvent struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Shift     bool
	Ctrl      bool
	Alt       bool
	IsKeyDown bool
	Swallowed bool // true when the hook suppressed this keydown from reaching the app
}

type Hook struct {
	handle             windows.Handle
	events             chan KeyEvent
	active             atomic.Bool
	interceptBackspace atomic.Bool // set by Processor when engine has pending state
	mu                 sync.Mutex
	hookProc           uintptr
	stopCh             chan struct{}
}

func New() *Hook {
	return &Hook{
		events: make(chan KeyEvent, 256),
		stopCh: make(chan struct{}),
	}
}

var globalHook *Hook

func (h *Hook) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	globalHook = h
	cb := windows.NewCallback(lowLevelKeyboardProc)
	h.hookProc = cb

	mod, _, _ := procGetModuleHandleW.Call(0)
	handle, _, err := procSetWindowsHookExW.Call(WH_KEYBOARD_LL, cb, mod, 0)
	if handle == 0 {
		return err
	}
	h.handle = windows.Handle(handle)
	h.active.Store(true)
	return nil
}

func (h *Hook) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handle != 0 {
		procUnhookWindowsHookEx.Call(uintptr(h.handle))
		h.handle = 0
		h.active.Store(false)
	}
}

func (h *Hook) IsActive() bool          { return h.active.Load() }
func (h *Hook) SetActive(v bool)        { h.active.Store(v) }
func (h *Hook) Events() <-chan KeyEvent { return h.events }

// SetInterceptBackspace controls whether VK_BACK is swallowed by the hook.
// The Processor calls this after each event to reflect the engine's pending state.
func (h *Hook) SetInterceptBackspace(v bool) { h.interceptBackspace.Store(v) }

// shouldIntercept returns true for keydown events the hook should swallow and
// route through the transliteration engine.  Modifier combos (Ctrl/Alt) are
// always passed through so hotkeys work normally.
func shouldIntercept(vk uint32, ctrl, alt bool) bool {
	if ctrl || alt {
		return false
	}
	return (vk >= 0x41 && vk <= 0x5A) || // A–Z
		vk == VK_SPACE
}

func lowLevelKeyboardProc(nCode int, wParam, lParam uintptr) uintptr {
	if nCode != HC_ACTION || globalHook == nil {
		return callNext(nCode, wParam, lParam)
	}

	kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

	// Skip events we injected ourselves (prevents infinite loops).
	if kbd.Flags&LLKHF_INJECTED != 0 {
		return callNext(nCode, wParam, lParam)
	}

	if !globalHook.active.Load() {
		return callNext(nCode, wParam, lParam)
	}

	ctrl := isKeyPressed(VK_CTRL)
	alt := isKeyPressed(VK_MENU)
	isDown := wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN

	event := KeyEvent{
		VkCode:    kbd.VkCode,
		ScanCode:  kbd.ScanCode,
		Flags:     kbd.Flags,
		IsKeyDown: isDown,
		Shift:     isKeyPressed(VK_SHIFT),
		Ctrl:      ctrl,
		Alt:       alt,
	}

	swallow := isDown && (
		shouldIntercept(kbd.VkCode, ctrl, alt) ||
			(kbd.VkCode == VK_BACK && globalHook.interceptBackspace.Load()))

	event.Swallowed = swallow

	select {
	case globalHook.events <- event:
		if swallow {
			return 1
		}
	default:
		// Channel full — pass through to avoid blocking the hook thread.
	}

	return callNext(nCode, wParam, lParam)
}

func callNext(nCode int, wParam, lParam uintptr) uintptr {
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func isKeyPressed(vk uint32) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

func (h *Hook) RunMessageLoop() {
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
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			return
		}
	}
}
