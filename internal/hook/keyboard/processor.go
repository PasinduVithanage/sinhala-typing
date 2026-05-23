//go:build windows

package keyboard

import (
	"sinhala-assistant/internal/engine/transliteration"
	"sinhala-assistant/internal/window"
)

// Processor reads from the Hook's event channel, feeds keystrokes through the
// transliteration engine, and injects the resulting Sinhala Unicode text.
type Processor struct {
	hook   *Hook
	engine *transliteration.Engine
	guard  InjectionGuard
	done   chan struct{}
}

func NewProcessor(hook *Hook, engine *transliteration.Engine) *Processor {
	return &Processor{hook: hook, engine: engine, done: make(chan struct{})}
}

// Stop signals Run() to exit. Safe to call multiple times.
func (p *Processor) Stop() {
	select {
	case <-p.done: // already closed
	default:
		close(p.done)
	}
}

func (p *Processor) Run() {
	for {
		select {
		case event, ok := <-p.hook.Events():
			if !ok {
				return
			}
			if !event.IsKeyDown {
				continue
			}
			p.process(event)
		case <-p.done:
			return
		}
	}
}

func (p *Processor) process(event KeyEvent) {
	mode := foregroundCompatMode()

	switch {
	case event.VkCode >= 0x41 && event.VkCode <= 0x5A && !event.Ctrl && !event.Alt:
		char := KeyEventToRaw(event)
		if char == "" {
			break
		}
		result, ok := p.engine.Feed(char)
		if ok {
			p.inject(result, mode)
		}

	case event.VkCode == VK_BACK && event.Swallowed:
		dc, passthrough := p.engine.HandleBackspace()
		if passthrough {
			_ = InjectBackspaces(1)
		} else if dc > 0 {
			if p.guard.Allow() {
				_ = InjectBackspaces(dc)
			}
		}

	case event.VkCode == VK_SPACE && event.Swallowed:
		result, ok := p.engine.Flush()
		if ok {
			p.inject(result, mode)
		}
		_ = InjectUnicode(" ")
	}

	p.hook.SetInterceptBackspace(p.engine.HasPendingState())
}

func (p *Processor) inject(result transliteration.TransliterationResult, mode window.CompatMode) {
	if result.Output == "" && result.DeleteCount == 0 {
		return
	}
	if !p.guard.Allow() {
		return
	}
	switch mode {
	case window.CompatModeClipboard:
		_ = InjectViaClipboard(result.Output, result.DeleteCount)
	case window.CompatModeBlocked:
		// Do not inject into blocked apps.
	default:
		_ = ReplaceWithSinhala(result.Output, result.DeleteCount)
	}
}

// foregroundCompatMode returns the injection mode for the currently active window.
func foregroundCompatMode() window.CompatMode {
	info, err := window.GetForegroundWindowInfo()
	if err != nil || info == nil {
		return window.CompatModeKeyboard
	}
	return window.GetCompatMode(info.Exe)
}
