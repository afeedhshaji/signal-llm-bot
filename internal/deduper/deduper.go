package deduper

import (
	"sync"
	"time"
)

type Deduper struct {
	mu   sync.Mutex
	data map[string]time.Time
	ttl  time.Duration
	done chan struct{}
}

func New(ttl time.Duration) *Deduper {
	d := &Deduper{data: make(map[string]time.Time), ttl: ttl, done: make(chan struct{})}
	go d.cleanupLoop()
	return d
}

func (d *Deduper) Seen(hash string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.data[hash]; ok && time.Since(t) <= d.ttl {
		return true
	}
	d.data[hash] = time.Now()
	return false
}

func (d *Deduper) cleanupLoop() {
	ticker := time.NewTicker(d.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			d.mu.Lock()
			for k, ts := range d.data {
				if now.Sub(ts) > d.ttl {
					delete(d.data, k)
				}
			}
			d.mu.Unlock()
		case <-d.done:
			return
		}
	}
}

func (d *Deduper) Stop() { close(d.done) }
