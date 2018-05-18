package pool

import "errors"

type Pool struct {
	work chan func()
	sem  chan struct{}
	size int
}

func New(size int) *Pool {
	return &Pool{
		work: make(chan func()),
		sem:  make(chan struct{}, size),
		size: size,
	}
}

func (p *Pool) Schedule(task func()) error {
	if len(p.work) > 10 && len(p.sem) >= p.size-10 {
		return errors.New("queue full")
	}
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.worker(task)
	}

	return nil
}

func (p *Pool) worker(task func()) {
	defer func() {
		<-p.sem
	}()

	for {
		task()
		task = <-p.work
	}
}
