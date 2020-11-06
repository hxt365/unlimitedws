package gopool

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func BenchmarkPool(b *testing.B) {
	b.Run("do jobs without gopool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			func() {
				time.Sleep(10 * time.Millisecond)
			}()
		}
	})

	b.Run("do jobs with gopool", func(b *testing.B) {
		p, err := NewPool(10, 100)
		assert.NoError(b, err)

		for i := 0; i < b.N; i++ {
			p.Schedule(func() {
				time.Sleep(10 * time.Millisecond)
			})
		}
	})
}
