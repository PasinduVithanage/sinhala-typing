//go:build windows

package clipboard

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalSize       = kernel32.NewProc("GlobalSize")
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
)

// Monitor polls the clipboard for Sinhala text changes.
type Monitor struct {
	onChanged func(text string)
	stopCh    chan struct{}
}

func NewMonitor(onChanged func(text string)) *Monitor {
	return &Monitor{onChanged: onChanged, stopCh: make(chan struct{})}
}

func GetClipboardText() (string, error) {
	ret, _, err := procOpenClipboard.Call(0)
	if ret == 0 {
		return "", err
	}
	defer procCloseClipboard.Call()

	handle, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if handle == 0 {
		return "", nil
	}

	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		return "", nil
	}
	defer procGlobalUnlock.Call(handle)

	size, _, _ := procGlobalSize.Call(handle)
	u16 := (*[1 << 20]uint16)(unsafe.Pointer(ptr))[:size/2]
	n := 0
	for n < len(u16) && u16[n] != 0 {
		n++
	}
	return windows.UTF16ToString(u16[:n]), nil
}

func SetClipboardText(text string) error {
	u16, err := windows.UTF16FromString(text)
	if err != nil {
		return err
	}
	size := uintptr(len(u16) * 2)

	h, _, err := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if h == 0 {
		return err
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return nil
	}
	dst := (*[1 << 20]uint16)(unsafe.Pointer(ptr))
	copy(dst[:], u16)
	procGlobalUnlock.Call(h)

	ret, _, err := procOpenClipboard.Call(0)
	if ret == 0 {
		return err
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()
	procSetClipboardData.Call(CF_UNICODETEXT, h)
	return nil
}

// Start polls clipboard every 500ms and calls onChanged when content changes.
func (m *Monitor) Start() {
	go func() {
		var lastText string
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				text, err := GetClipboardText()
				if err == nil && text != lastText && text != "" {
					lastText = text
					m.onChanged(text)
				}
			}
		}
	}()
}

func (m *Monitor) Stop() { close(m.stopCh) }
