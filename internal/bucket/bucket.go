package bucket

import (
	"errors"
	"sync"
	"time"
)

const sleepTime = 10 * time.Second

var ErrRejected = errors.New("rejected")

type bucket struct {
	mu      sync.Mutex
	key     string
	load    int
	maxLoad int
	lastUse time.Time
}

func NewBucket(key string, maxLoad int, delCh chan<- string) *bucket {
	b := &bucket{
		key:     key,
		maxLoad: maxLoad,
	}

	go func() {
		ticker := time.NewTicker(time.Duration(int(time.Minute) / b.maxLoad))
		defer ticker.Stop()

		for {
			<-ticker.C

			b.mu.Lock()
			if b.load > 0 {
				b.load--
				b.mu.Unlock()
				continue
			}
			b.mu.Unlock()

			if b.lastUse.IsZero() {
				b.lastUse = time.Now()
			} else if time.Since(b.lastUse) > sleepTime {
				delCh <- b.key
				break
			}
		}
	}()

	return b
}

func (b *bucket) check() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.load < b.maxLoad {
		b.load++

		return nil
	}

	return ErrRejected
}
