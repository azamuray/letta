package main

import (
	"fmt"
	"time"

	"letta/pkg/checker"
)

func main() {
	fmt.Println("🌐 Letta запущен")

	checker := checker.New()
	var prevStatus bool
	firstCheck := true

	for {
		status := checker.IsConnected()

		if status != prevStatus || firstCheck {
			printStatus(status)
			prevStatus = false
		}

		time.Sleep(5 * time.Second)
	}

}

func printStatus(connected bool) {
	timestamp := time.Now().Format("15:04:05")

	if connected {
		fmt.Printf("[%s] ✅ Интернет подключен\n", timestamp)
	} else {
		fmt.Printf("[%s] ❌ Нет подключения\n", timestamp)
	}
}
