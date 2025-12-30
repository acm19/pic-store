package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Extract embedded binaries on startup
	exiftoolPath, jpegoptimPath, err := ExtractBinaries()
	if err != nil {
		log.Fatalf("Failed to extract binaries: %v", err)
	}

	// Create application instance with extracted binary paths
	app := NewApp(exiftoolPath, jpegoptimPath)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Pics - Media Organiser",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
