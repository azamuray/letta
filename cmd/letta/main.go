// cmd/lettawidget/main.go
package main

import (
	"letta/pkg/checker"
	"time"

	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("🌐")
	systray.SetTooltip("Letta - проверка интернета")

	mStatus := systray.AddMenuItem("Проверяем...", "")
	mStatus.Disable()

	mQuit := systray.AddMenuItem("Выйти", "")

	chk := checker.New()

	go func() {
		for {
			time.Sleep(1 * time.Second)

			if chk.IsConnected() {
				systray.SetTitle("✅")
				systray.SetTooltip("Интернет подключен")
				mStatus.SetTitle("✅ Онлайн")
			} else {
				systray.SetTitle("❌")
				systray.SetTooltip("Нет интернета")
				mStatus.SetTitle("❌ Оффлайн")
			}
		}
	}()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {}
