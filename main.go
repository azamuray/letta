package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Letta запущен")
	fmt.Println("Выполняется проверка подключения...")

	if checkInternet() {
		fmt.Println("✅ Интернет есть")
	} else {
		fmt.Println("❌ Интернет отсутствует")
	}

}

func checkInternet() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Head("http://clients3.google.com/generate_204")
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == 204
}
