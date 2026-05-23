//go:build windows

package tray

import "github.com/getlantern/systray"

// IconState represents the visual state of the tray icon.
type IconState int

const (
	IconActive   IconState = iota
	IconInactive
	IconError
)

// Config holds the callbacks for tray menu actions.
type Config struct {
	OnToggle   func()
	OnSettings func()
	OnClipFix  func()
	OnQuit     func()
}

// Tray manages the system tray icon and menu.
type Tray struct {
	cfg        Config
	stopCh     chan struct{}
	iconCh     chan IconState
	menuToggle *systray.MenuItem
}

func New(cfg Config) *Tray {
	return &Tray{
		cfg:    cfg,
		stopCh: make(chan struct{}),
		iconCh: make(chan IconState, 1),
	}
}

func (t *Tray) Run()  { systray.Run(t.onReady, t.onExit) }
func (t *Tray) Stop() { systray.Quit() }

func (t *Tray) onReady() {
	systray.SetIcon(iconBytes(IconInactive))
	systray.SetTooltip("Sinhala Assistant — Inactive")
	systray.SetTitle("Sinhala Assistant")

	t.menuToggle = systray.AddMenuItem("Enable Sinhala Typing", "Toggle the keyboard engine")
	mSettings := systray.AddMenuItem("Settings", "Open settings panel")
	mClipFix := systray.AddMenuItem("Fix Clipboard", "Repair Sinhala in clipboard")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Sinhala Assistant")

	go func() {
		for {
			select {
			case <-t.menuToggle.ClickedCh:
				t.cfg.OnToggle()
			case <-mSettings.ClickedCh:
				t.cfg.OnSettings()
			case <-mClipFix.ClickedCh:
				t.cfg.OnClipFix()
			case <-mQuit.ClickedCh:
				t.cfg.OnQuit()
			case state := <-t.iconCh:
				t.applyIcon(state)
			case <-t.stopCh:
				return
			}
		}
	}()
}

func (t *Tray) SetIcon(state IconState) {
	select {
	case t.iconCh <- state:
	default:
	}
}

func (t *Tray) applyIcon(state IconState) {
	systray.SetIcon(iconBytes(state))
	switch state {
	case IconActive:
		systray.SetTooltip("Sinhala Assistant — Active")
		t.menuToggle.SetTitle("Disable Sinhala Typing")
	case IconInactive:
		systray.SetTooltip("Sinhala Assistant — Inactive")
		t.menuToggle.SetTitle("Enable Sinhala Typing")
	case IconError:
		systray.SetTooltip("Sinhala Assistant — Error")
	}
}

func (t *Tray) onExit() {}

