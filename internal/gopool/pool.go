package gopool

import (
	"fmt"
	"time"
)

var ErrScheduleTimeout = fmt.Errorf("schedule timeout")

type Pool struct {
	job chan func()
}

func NewPool(size, queue int) (*Pool, error) {
	if size == 0 {
		return nil, fmt.Errorf("no worker")
	}

	if queue == 0 {
		return nil, fmt.Errorf("no job in queue")
	}

	p := &Pool{
		job: make(chan func(), queue),
	}

	for i := 0; i < size; i++ {
		go p.worker()
	}

	return p, nil
}

func (p *Pool) Schedule(job func()) {
	_ = p.schedule(job, nil)
}

func (p *Pool) ScheduleTimeout(timeout time.Duration, job func()) error {
	return p.schedule(job, time.After(timeout))
}

func (p *Pool) schedule(job func(), timeout <-chan time.Time) error {
	select {
	case <-timeout:
		return ErrScheduleTimeout
	case p.job <- job:
		return nil
	}
}

func (p *Pool) worker() {
	for task := range p.job {
		task()
	}
}
