package main

import (
	"context"
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"trae-switch/internal/config"
	"trae-switch/internal/elevate"
	"trae-switch/internal/tray"
	"trae-switch/internal/truststore"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if !truststore.IsRunningAsAdmin() {
		if err := elevate.Elevate(); err == nil {
			os.Exit(0)
		}
	}

	app := NewApp()

	tray.SetCallbacks(
		func() { app.ShowWindow() },
		func() { app.QuitApp() },
	)

	tray.Start()

	if _, err := config.Load(); err != nil {
		println("Failed to load config: " + err.Error())
	}

	err := wails.Run(&options.App{
		Title:            "Trae Switch",
		Width:            500,
		Height:           700,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose: func(ctx context.Context) bool {
			if config.GetTrayMode() {
				app.HideWindowWithNotification()
				return true
			}
			return false
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
			Theme:                             windows.SystemDefault,
		},
	})

	if err != nil {
		println("Error: " + err.Error())
	}

	tray.Quit()
}
