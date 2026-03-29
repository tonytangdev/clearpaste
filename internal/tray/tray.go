package tray

import (
	"time"

	"fyne.io/systray"
)

// Callbacks for tray actions.
type Callbacks struct {
	OnToggle func(enabled bool)
	OnUndo   func() bool
	HasUndo  func() bool
}

// Run starts the system tray. Blocks until quit.
func Run(cb Callbacks) {
	systray.Run(func() { onReady(cb) }, onExit)
}

func onReady(cb Callbacks) {
	systray.SetIcon(IconDefault)
	systray.SetTitle("")
	systray.SetTooltip("ClearPaste — clipboard cleaner")

	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Toggle clipboard monitoring", true)
	mUndo := systray.AddMenuItem("Undo last clean", "Restore original clipboard text")
	mUndo.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit ClearPaste")

	go func() {
		for {
			select {
			case <-mEnabled.ClickedCh:
				if mEnabled.Checked() {
					mEnabled.Uncheck()
					systray.SetIcon(IconDisabled)
					cb.OnToggle(false)
				} else {
					mEnabled.Check()
					systray.SetIcon(IconDefault)
					cb.OnToggle(true)
				}

			case <-mUndo.ClickedCh:
				if cb.OnUndo != nil {
					cb.OnUndo()
				}

			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if cb.HasUndo != nil && cb.HasUndo() {
				mUndo.Enable()
			} else {
				mUndo.Disable()
			}
		}
	}()
}

func onExit() {
	// Cleanup if needed
}

// FlashIcon briefly shows the active icon then reverts to default.
func FlashIcon() {
	systray.SetIcon(IconActive)
	go func() {
		time.Sleep(2 * time.Second)
		systray.SetIcon(IconDefault)
	}()
}
