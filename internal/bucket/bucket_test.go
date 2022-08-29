package bucket

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBucket(t *testing.T) {
	t.Run("tests delete channel", func(t *testing.T) {
		delCh := make(chan string)
		b := NewBucket("testLogin", 1000, delCh)

		time.Sleep(sleepTime)

		assert.Equal(t, b.key, <-delCh)
	})
}
