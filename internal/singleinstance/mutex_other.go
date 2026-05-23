//go:build !windows

package singleinstance

func Acquire() bool  { return true }
func Release()       {}
