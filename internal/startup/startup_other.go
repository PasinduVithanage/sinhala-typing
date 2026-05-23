//go:build !windows

package startup

func Set(_ bool) error { return nil }
func IsEnabled() bool  { return false }
