package tray

import (
	_ "embed"
	"log"

	"github.com/energye/systray"
)

//go:embed icon.ico
var iconData []byte

var (
	showWindowCallback func()
	quitCallback       func()
	running            bool
)

func SetCallbacks(onShow func(), onQuit func()) {
	showWindowCallback = onShow
	quitCallback = onQuit
}

func Start() {
	if running {
		return
	}
	running = true
	go systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Trae Switch")
	systray.SetTooltip("Trae Switch - 点击显示窗口")

	systray.SetOnClick(func(menu systray.IMenu) {
		if showWindowCallback != nil {
			showWindowCallback()
		}
	})

	systray.SetOnDClick(func(menu systray.IMenu) {
		if showWindowCallback != nil {
			showWindowCallback()
		}
	})

	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	mShow := systray.AddMenuItem("显示主窗口", "显示应用主窗口")
	mQuit := systray.AddMenuItem("退出程序", "退出 Trae Switch")

	mShow.Click(func() {
		if showWindowCallback != nil {
			showWindowCallback()
		}
	})

	mQuit.Click(func() {
		if quitCallback != nil {
			quitCallback()
		}
		systray.Quit()
	})

	log.Println("System tray ready")
}

func onExit() {
	running = false
	log.Println("System tray exited")
}

func Quit() {
	systray.Quit()
}