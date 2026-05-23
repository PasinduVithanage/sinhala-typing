//go:build windows

package keyboard

import (
	"time"
	"unicode/utf16"
	"unsafe"

	clipboardpkg "sinhala-assistant/internal/hook/clipboard"
)

const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_UNICODE = 0x0004
	KEYEVENTF_KEYUP   = 0x0002
	VK_V              = 0x56
)

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	_    [4]byte
	Ki   KEYBDINPUT
	_    [8]byte
}

// InjectUnicode sends a Unicode string as keyboard input events.
func InjectUnicode(text string) error {
	runes := []rune(text)
	u16 := utf16.Encode(runes)

	inputs := make([]INPUT, 0, len(u16)*2)
	for _, u := range u16 {
		inputs = append(inputs,
			INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WScan: u, DwFlags: KEYEVENTF_UNICODE}},
			INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WScan: u, DwFlags: KEYEVENTF_UNICODE | KEYEVENTF_KEYUP}},
		)
	}
	if len(inputs) == 0 {
		return nil
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// InjectBackspaces sends N backspace key events.
func InjectBackspaces(n int) error {
	if n <= 0 {
		return nil
	}
	inputs := make([]INPUT, 0, n*2)
	for i := 0; i < n; i++ {
		inputs = append(inputs,
			INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_BACK}},
			INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_BACK, DwFlags: KEYEVENTF_KEYUP}},
		)
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// ReplaceWithSinhala erases deleteCount chars then injects the Sinhala string.
func ReplaceWithSinhala(sinhalaText string, deleteCount int) error {
	if err := InjectBackspaces(deleteCount); err != nil {
		return err
	}
	return InjectUnicode(sinhalaText)
}

// InjectViaClipboard erases deleteCount chars then pastes text via Ctrl+V.
// Used for apps where SendInput Unicode injection is unreliable (e.g. Photoshop).
func InjectViaClipboard(text string, deleteCount int) error {
	if deleteCount > 0 {
		if err := InjectBackspaces(deleteCount); err != nil {
			return err
		}
	}
	prev, _ := clipboardpkg.GetClipboardText()
	if err := clipboardpkg.SetClipboardText(text); err != nil {
		return err
	}
	if err := injectCtrlV(); err != nil {
		return err
	}
	// Restore previous clipboard content after the paste has been processed.
	if prev != "" {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = clipboardpkg.SetClipboardText(prev)
		}()
	}
	return nil
}

func injectCtrlV() error {
	inputs := []INPUT{
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_CTRL}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_V}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_V, DwFlags: KEYEVENTF_KEYUP}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_CTRL, DwFlags: KEYEVENTF_KEYUP}},
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if ret == 0 {
		return err
	}
	return nil
}
