package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/enescakir/emoji"
	"github.com/getlantern/systray"
)

var (
	// Храним последний известный IP и состояние интерфейсов
	lastIP          string
	lastInterfaces  string // хеш интерфейсов для отслеживания изменений
	lastIPMutex     sync.Mutex
	checkInProgress bool
	checkMutex      sync.Mutex
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("🌐")
	log.Println("🚀 Приложение запущено, начинаю мониторинг сети...")

	// Делаем первую проверку при запуске
	log.Println("📡 Первоначальная проверка сети...")
	checkAndUpdateIP()

	// Мониторим изменения сетевых интерфейсов
	// Проверяем каждые 0.5 секунды, но делаем запрос к бэкенду только при изменениях
	log.Println("👀 Мониторинг изменений сети запущен (проверка каждые 0.5 секунды)")
	go monitorNetworkChanges()

	// Добавляем меню выхода
	mQuit := systray.AddMenuItem("Выйти", "")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Держим программу живой
	select {}
}

// monitorNetworkChanges отслеживает изменения сетевых интерфейсов
// Делает запрос к бэкенду только когда интерфейсы изменились
func monitorNetworkChanges() {
	ticker := time.NewTicker(500 * time.Millisecond) // 0.5 секунды
	defer ticker.Stop()

	for range ticker.C {
		// Получаем текущее состояние интерфейсов
		currentInterfaces := getInterfacesHash()

		lastIPMutex.Lock()
		interfacesChanged := currentInterfaces != lastInterfaces
		lastIPMutex.Unlock()

		// Если интерфейсы изменились (подключился VPN, изменилась сеть и т.д.)
		if interfacesChanged {
			log.Println("🔔 Обнаружено изменение сети!")
			log.Printf("   Старое состояние: %s", lastInterfaces)
			log.Printf("   Новое состояние: %s", currentInterfaces)

			lastIPMutex.Lock()
			lastInterfaces = currentInterfaces
			lastIPMutex.Unlock()

			// Делаем запрос к бэкенду только при изменении сети
			log.Println("📡 Отправляю запрос к бэкенду...")
			checkAndUpdateIP()
		}
	}
}

// getInterfacesHash возвращает строку-идентификатор текущих сетевых интерфейсов
// Отслеживает только важные интерфейсы: основные сетевые (en*) и VPN (utun*)
func getInterfacesHash() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	var parts []string
	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Отслеживаем только важные интерфейсы:
		// - en* (основные сетевые интерфейсы: Ethernet, Wi-Fi)
		// - utun* (VPN интерфейсы)
		name := iface.Name
		isImportant := strings.HasPrefix(name, "en") || strings.HasPrefix(name, "utun")
		if !isImportant {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Собираем информацию об интерфейсе: имя + реальные IP адреса
		interfaceInfo := name
		hasRealIP := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP
				// Игнорируем link-local адреса IPv6 (fe80::) - они не важны
				// Игнорируем loopback (127.0.0.1)
				if ip.IsLoopback() {
					continue
				}
				if ip.IsLinkLocalUnicast() {
					continue
				}
				// Берем только реальные IP адреса (IPv4 или глобальные IPv6)
				interfaceInfo += ":" + ip.String()
				hasRealIP = true
			}
		}

		// Добавляем интерфейс, даже если у него нет IP (например, utun без IP - это тоже важно)
		// Это поможет отследить появление/исчезновение VPN интерфейсов
		if hasRealIP || strings.HasPrefix(name, "utun") {
			parts = append(parts, interfaceInfo)
		}
	}

	return strings.Join(parts, "|")
}

// checkAndUpdateIP проверяет текущий IP и обновляет иконку
func checkAndUpdateIP() {
	// Предотвращаем параллельные запросы
	checkMutex.Lock()
	if checkInProgress {
		checkMutex.Unlock()
		return
	}
	checkInProgress = true
	checkMutex.Unlock()

	defer func() {
		checkMutex.Lock()
		checkInProgress = false
		checkMutex.Unlock()
	}()

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	resp, err := client.Get("http://45.130.214.133:8080")
	if err != nil {
		// Интернета нет
		log.Printf("❌ Ошибка подключения к бэкенду: %v", err)

		lastIPMutex.Lock()
		hadIP := lastIP != ""
		lastIP = ""
		lastIPMutex.Unlock()

		// Всегда обновляем иконку при ошибке, даже если уже была "❌"
		systray.SetTitle("❌")
		if hadIP {
			log.Println("🔄 Интернет пропал, иконка обновлена на ❌")
		}
		return
	}

	// Интернет есть, читаем ответ
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Printf("❌ Ошибка чтения ответа: %v", err)
		systray.SetTitle("❌")
		return
	}

	var info struct {
		IP          string `json:"ip"`
		CountryCode string `json:"countryCode"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		log.Printf("⚠️  Ошибка парсинга JSON: %v", err)
		systray.SetTitle("✅")
		return
	}

	log.Printf("✅ Получен ответ от бэкенда: IP=%s, Страна=%s", info.IP, info.CountryCode)

	// Проверяем, изменился ли IP
	lastIPMutex.Lock()
	ipChanged := info.IP != lastIP
	oldIP := lastIP
	if ipChanged {
		lastIP = info.IP
	}
	lastIPMutex.Unlock()

	// Обновляем иконку всегда при успешном ответе
	// Это гарантирует, что флаг обновится даже если IP не изменился, но страна могла измениться
	if ipChanged {
		log.Printf("🔄 IP изменился: %s -> %s", oldIP, info.IP)
	}

	if info.CountryCode != "" {
		flag, _ := emoji.CountryFlag(info.CountryCode)
		systray.SetTitle(string(flag))
		if ipChanged {
			log.Printf("🏳️  Иконка обновлена: флаг страны %s (IP изменился)", info.CountryCode)
		} else {
			log.Printf("🏳️  Иконка обновлена: флаг страны %s", info.CountryCode)
		}
	} else {
		systray.SetTitle("✅")
		log.Println("✅ Иконка обновлена: интернет подключен")
	}
}

func onExit() {}
