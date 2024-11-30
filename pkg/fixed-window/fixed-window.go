package fixedwindow

import (
	"sync"
	"time"
)

type FixedWindow struct {
	mu      sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (l *FixedWindow) Allow(ip string) (bool, time.Duration) {
	l.mu.RLock()
	count, exists := l.clients[ip]
	l.mu.RUnlock()

	if !exists || count < l.limit {
		l.mu.Lock()
		if !exists {
			go l.resetCount(ip)
		}

		l.clients[ip]++
		l.mu.Unlock()
		return true, 0
	}

	return false, l.window
}

func (l *FixedWindow) resetCount(ip string) {
	time.Sleep(l.window)
	l.mu.Lock()
	delete(l.clients, ip)
	l.mu.Unlock()
}
