package bucket

import (
	"errors"
	"sync"
	"time"
)

const sleepTime = 10 * time.Second

var ErrRejected = errors.New("rejected")

type Bucket struct {
	mu      sync.Mutex
	key     string
	load    int
	maxLoad int
	lastUse time.Time
}

func NewBucket(key string, maxLoad int, delCh chan<- string) *Bucket {
	b := &Bucket{
		key:     key,
		maxLoad: maxLoad,
	}

	go func(b *Bucket) {
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
	}(b)

	return b
}

func (b *Bucket) check() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.load < b.maxLoad {
		b.load++

		return nil
	}

	return ErrRejected
}
