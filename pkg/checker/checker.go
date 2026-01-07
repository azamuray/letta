// pkg/checker/checker.go
package checker

import (
	"net"
	"time"
)

type Checker struct {
	timeout time.Duration
	servers []string
	counter int
}

func New() *Checker {
	return &Checker{
		timeout: 1 * time.Second,
		servers: []string{
			"45.130.214.133:22",
		},
		counter: 0,
	}
}

func (c *Checker) IsConnected() bool {
	// Пробуем несколько серверов
	for attempt := 0; attempt < 3; attempt++ {
		c.counter = (c.counter + 1) % len(c.servers)
		server := c.servers[c.counter]

		conn, err := net.DialTimeout("tcp", server, c.timeout)
		if err == nil {
			conn.Close()
			return true
		}

		// Маленькая пауза между попытками
		if attempt < 2 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return false
}
