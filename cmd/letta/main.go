// cmd/lettawidget/main.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/enescakir/emoji"
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("🌐")

	go func() {
		for {
			client := &http.Client{
				Timeout: 1 * time.Second,
				Transport: &http.Transport{
					DisableKeepAlives: true,
				},
			}

			resp, err := client.Get("http://45.130.214.133:8080")

			if err != nil {
				// Интернета нет
				systray.SetTitle("❌")
			} else {
				// Интернет есть
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var info struct {
					CountryCode string `json:"countryCode"`
				}
				json.Unmarshal(body, &info)

				if info.CountryCode != "" {
					flag, _ := emoji.CountryFlag(info.CountryCode)
					systray.SetTitle(string(flag))
				} else {
					systray.SetTitle("✅")
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	mQuit := systray.AddMenuItem("Выйти", "")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {}
