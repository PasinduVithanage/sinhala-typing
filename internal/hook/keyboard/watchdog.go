//go:build windows

package keyboard

import (
	"sync"
	"time"
)

// InjectionGuard rate-limits SendInput calls to prevent runaway injection.
type InjectionGuard struct {
	lastInject time.Time
	count      int
	mu         sync.Mutex
}

const maxInjectionsPerSecond = 50

func (g *InjectionGuard) Allow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	if now.Sub(g.lastInject) > time.Second {
		g.count = 0
		g.lastInject = now
	}
	if g.count >= maxInjectionsPerSecond {
		return false
	}
	g.count++
	return true
}

// WatchdogRoutine reinstalls the hook if Windows silently drops it.
func (h *Hook) WatchdogRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if !h.active.Load() {
			continue
		}
		if h.handle == 0 {
			h.Stop()
			_ = h.Start()
		}
	}
}
