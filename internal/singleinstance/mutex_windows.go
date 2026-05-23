//go:build windows

package singleinstance

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
)

var mutexHandle windows.Handle

// Acquire creates a named system-wide mutex.  Returns true if this process
// is the first instance, false if another instance already holds the mutex.
func Acquire() bool {
	name, _ := windows.UTF16PtrFromString("Global\\SinhalaAssistant_v1_SingleInstance")
	// Call directly so we capture GetLastError() in the same syscall frame.
	r1, _, lastErr := procCreateMutexW.Call(
		0,
		0,
		uintptr(unsafe.Pointer(name)),
	)
	handle := windows.Handle(r1)
	if handle == 0 {
		return false // CreateMutex itself failed — treat as "can't determine", let startup proceed
	}
	if e, ok := lastErr.(syscall.Errno); ok && e == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(handle)
		return false
	}
	mutexHandle = handle
	return true
}

// Release frees the mutex so the OS can clean up.  Call on app shutdown.
func Release() {
	if mutexHandle != 0 {
		windows.CloseHandle(mutexHandle)
		mutexHandle = 0
	}
}
