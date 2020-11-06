package gopool

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

//func TestPoolBenchmark(t *test.B) {
//	t.Run("run with worker gopool", func(b *test.B) {
//		gopool := x509.NewCertPool()
//	})
//}

func TestPool(t *testing.T) {
	t.Run("create a worker gopool with 10 workers and 10 jobs in queue", func(t *testing.T) {
		_, err := NewPool(10, 10)
		assert.NoError(t, err)
	})

	t.Run("create a worker gopool with 0 workers", func(t *testing.T) {
		_, err := NewPool(0, 3)
		assert.Error(t, err)
	})

	t.Run("create a worker gopool with 0 job queue", func(t *testing.T) {
		_, err := NewPool(5, 0)
		assert.Error(t, err)
	})

	t.Run("schedule some jobs", func(t *testing.T) {
		numJobs := 5
		expected := 5
		count := 0
		mu := sync.Mutex{}
		job := func() {
			mu.Lock()
			count++
			mu.Unlock()
		}

		p, err := NewPool(5, 5)
		assert.NoError(t, err)

		for i := 0; i < numJobs; i++ {
			p.Schedule(job)
		}

		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, expected, count)
	})

	t.Run("schedule some jobs with timeout and expect error", func(t *testing.T) {
		p, err := NewPool(1, 1)
		assert.NoError(t, err)

		p.Schedule(func() {
			time.Sleep(20 * time.Millisecond)
		})
		p.Schedule(func() {})
		err = p.ScheduleTimeout(10*time.Millisecond, func() {})

		assert.Error(t, err)
	})

	t.Run("schedule some jobs with timeout and expect no error", func(t *testing.T) {
		p, err := NewPool(1, 1)
		assert.NoError(t, err)

		p.Schedule(func() {
			time.Sleep(10 * time.Millisecond)
		})
		p.Schedule(func() {})
		err = p.ScheduleTimeout(30*time.Millisecond, func() {})

		assert.NoError(t, err)
	})
}
