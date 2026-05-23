package ipc

import (
	"context"
	"sync/atomic"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sinhala-assistant/internal/config"
	"sinhala-assistant/internal/engine/normalization"
	"sinhala-assistant/internal/engine/transliteration"
	"sinhala-assistant/internal/hook/clipboard"
	"sinhala-assistant/internal/hook/keyboard"
	"sinhala-assistant/internal/hotkey"
	"sinhala-assistant/internal/startup"
	"sinhala-assistant/internal/tray"
)

// App is the Wails application struct — all exported methods are IPC-callable.
type App struct {
	ctx         context.Context
	cfg         *config.Config
	engine      *transliteration.Engine
	hook        *keyboard.Hook
	processor   *keyboard.Processor
	clipMonitor *clipboard.Monitor
	trayIcon    *tray.Tray
	hk          *hotkey.Hotkey
	engineRunning atomic.Bool // guards against double-starting the processor
}

func NewApp() *App { return &App{} }

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = config.Load()

	// Sync OS startup entry with the saved preference.
	_ = startup.Set(a.cfg.RunAtStartup)

	a.engine = transliteration.NewEngine(a.cfg.MappingProfile, a.cfg.UseZWJClusters)
	a.hook = keyboard.New()
	a.processor = keyboard.NewProcessor(a.hook, a.engine)
	a.clipMonitor = clipboard.NewMonitor(a.onClipboardChanged)

	a.trayIcon = tray.New(tray.Config{
		OnToggle:   func() { a.ToggleEngine() },
		OnSettings: a.ShowSettings,
		OnClipFix:  func() { a.FixClipboard() },
		OnQuit:     a.Quit,
	})
	go a.trayIcon.Run()

	// Global hotkey (default: Ctrl+Alt+S).
	a.hk = hotkey.New(func() { a.ToggleEngine() })
	go a.hk.Run(a.cfg.ToggleHotkey)

	if a.cfg.AutoStart {
		a.startEngine()
	}
}

func (a *App) DOMReady(ctx context.Context) {
	runtime.EventsEmit(ctx, "engine:state", a.GetEngineState())
}

func (a *App) Shutdown(ctx context.Context) {
	a.hook.Stop()
	a.clipMonitor.Stop()
	a.trayIcon.Stop()
	if a.hk != nil {
		a.hk.Stop()
	}
}

func (a *App) ToggleEngine() bool {
	if a.hook.IsActive() {
		a.hook.SetActive(false)
		a.processor.Stop()
		a.engineRunning.Store(false)
		runtime.EventsEmit(a.ctx, "engine:stopped", nil)
		a.trayIcon.SetIcon(tray.IconInactive)
		return false
	}
	a.startEngine()
	return true
}

func (a *App) startEngine() {
	if !a.engineRunning.CompareAndSwap(false, true) {
		return // already running
	}
	// Fresh processor with a new done channel for each start cycle.
	a.processor = keyboard.NewProcessor(a.hook, a.engine)
	if err := a.hook.Start(); err != nil {
		a.engineRunning.Store(false)
		runtime.EventsEmit(a.ctx, "engine:error", err.Error())
		return
	}
	go a.processor.Run()
	go a.hook.RunMessageLoop()
	a.trayIcon.SetIcon(tray.IconActive)
	runtime.EventsEmit(a.ctx, "engine:started", nil)
}

func (a *App) FixClipboard() map[string]interface{} {
	text, err := clipboard.GetClipboardText()
	if err != nil || text == "" {
		return map[string]interface{}{"fixed": false, "error": "empty clipboard"}
	}
	repaired := normalization.RepairString(text)
	if repaired == text {
		return map[string]interface{}{"fixed": false, "message": "no issues found"}
	}
	if err := clipboard.SetClipboardText(repaired); err != nil {
		return map[string]interface{}{"fixed": false, "error": err.Error()}
	}
	return map[string]interface{}{"fixed": true, "original": text, "repaired": repaired}
}

func (a *App) AnalyzeText(text string) []normalization.Detection {
	return normalization.Analyze(text)
}

func (a *App) RepairText(text string) string {
	return normalization.RepairString(text)
}

func (a *App) GetEngineState() map[string]interface{} {
	return map[string]interface{}{
		"active":  a.hook != nil && a.hook.IsActive(),
		"mapping": a.cfg.MappingProfile,
		"version": "1.0.0",
	}
}

func (a *App) GetConfig() *config.Config { return a.cfg }

func (a *App) SaveConfig(cfg config.Config) error {
	prevHotkey := a.cfg.ToggleHotkey
	a.cfg = &cfg

	if cfg.ToggleHotkey != prevHotkey {
		if a.hk != nil {
			a.hk.Stop()
		}
		a.hk = hotkey.New(func() { a.ToggleEngine() })
		go a.hk.Run(cfg.ToggleHotkey)
	}

	// Keep OS startup entry in sync with the setting.
	_ = startup.Set(cfg.RunAtStartup)

	return config.Save(&cfg)
}

func (a *App) ShowSettings() {
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "ui:show-settings", nil)
}
func (a *App) Quit() { runtime.Quit(a.ctx) }

func (a *App) onClipboardChanged(text string) {
	if !a.cfg.AutoFixClipboard {
		return
	}
	issues := normalization.Analyze(text)
	if len(issues) > 0 {
		runtime.EventsEmit(a.ctx, "clipboard:issues", map[string]interface{}{
			"count": len(issues),
			"text":  text,
		})
	}
}
