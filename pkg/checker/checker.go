package checker

import (
	"net/http"
	"time"
)

type Checker struct {
	timeout time.Duration
	urls    []string
}

func New() *Checker {
	return &Checker{
		timeout: 3 * time.Second,
		urls: []string{
			"http://clients3.google.com/generate_204",
			"http://connectivitycheck.gstatic.com/generate_204",
		},
	}
}

func (c *Checker) IsConnected() bool {
	client := http.Client{Timeout: c.timeout}

	for _, url := range c.urls {
		resp, err := client.Head(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
	}

	return false
}
