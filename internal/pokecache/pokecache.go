package internal

import (
	"sync";
	"time";
)

type cacheEntry struct {
    createdAt	time.Time
    val			[]byte
}

type Cache struct {
    entries		map[string]cacheEntry
    mu			*sync.Mutex
    interval	time.Duration
}

func (c Cache) Add(key string, val []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries[key] = cacheEntry {
		createdAt: 	time.Now(),
		val:		val,
    }
}

func (c Cache) Get(key string) ([]byte, bool) {
    entry, ok := c.entries[key]
    if !ok {
		return []byte{}, false
    }
    return entry.val, true
}

func (c Cache) reapLoop() {
    ticker := time.NewTicker(time.Second)
    isTimeToCheck := make(chan bool)
    go func() {
		time.Sleep(c.interval)
		isTimeToCheck <- true
    }()

    for ;; {
		select {
		case <-isTimeToCheck:
			if len(c.entries) > 0 {
				now := time.Now()
				for key, entry := range c.entries {
					if now.After(entry.createdAt.Add(c.interval)) {
						c.mu.Lock()
						delete(c.entries, key)
						c.mu.Unlock()
					}
				}
			}
			ticker.Reset(c.interval)
			break
		default:
			break
		}
    }

}

func NewCache(interval time.Duration) Cache {
    // Avoid panic
    if interval <= 0 {
	interval = 60 * time.Second
    }

    nc := Cache {
	entries:	make(map[string]cacheEntry),
	mu:		&sync.Mutex{},
	interval:	interval,
    }

    // start interval clock
    go nc.reapLoop()

    return nc
}

