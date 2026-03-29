package main

import (
	_ "embed"

	"github.com/tonytangdev/clearpaste/internal/clipboard"
	"github.com/tonytangdev/clearpaste/internal/tray"
)

//go:embed icons/icon.png
var iconDefault []byte

//go:embed icons/icon_active.png
var iconActive []byte

//go:embed icons/icon_disabled.png
var iconDisabled []byte

func main() {
	// Inject icons into tray package
	tray.IconDefault = iconDefault
	tray.IconActive = iconActive
	tray.IconDisabled = iconDisabled

	cb := &clipboard.System{}
	monitor := clipboard.NewMonitor(cb, cb)

	stop := make(chan struct{})

	monitor.OnCleaned = func() {
		tray.FlashIcon()
	}

	// Start clipboard monitor in background
	go monitor.Start(stop)

	// Run tray (blocks until quit)
	tray.Run(tray.Callbacks{
		OnToggle: func(enabled bool) {
			monitor.SetEnabled(enabled)
		},
		OnUndo: func() bool {
			return monitor.Undo()
		},
		HasUndo: func() bool {
			return monitor.HasUndo()
		},
	})

	close(stop)
}
