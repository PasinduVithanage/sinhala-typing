package main

import (
	"embed"
	"log"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	winopts "github.com/wailsapp/wails/v2/pkg/options/windows"

	"sinhala-assistant/internal/ipc"
	"sinhala-assistant/internal/singleinstance"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	runtime.LockOSThread() // required for Windows message loops

	if !singleinstance.Acquire() {
		// Another instance is already running in the system tray — exit silently.
		os.Exit(0)
	}
	defer singleinstance.Release()

	app := ipc.NewApp()

	err := wails.Run(&options.App{
		Title:             "Sinhala Assistant",
		Width:             420,
		Height:            600,
		Frameless:         true,
		StartHidden:       true,
		HideWindowOnClose: true,

		AssetServer: &assetserver.Options{Assets: assets},

		Windows: &winopts.Options{
			WindowIsTranslucent: true,
			Theme:               winopts.Dark,
			CustomTheme: &winopts.ThemeSettings{
				DarkModeTitleBar:  winopts.RGB(18, 18, 20),
				DarkModeTitleText: winopts.RGB(255, 255, 255),
				DarkModeBorder:    winopts.RGB(40, 40, 45),
			},
		},

		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		OnDomReady: app.DOMReady,
		Bind:       []interface{}{app},
	})

	if err != nil {
		log.Fatal(err)
	}
}
