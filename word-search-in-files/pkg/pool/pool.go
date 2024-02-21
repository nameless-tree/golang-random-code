package pool

import (
	"fmt"
	"sync"
)

type Pool interface {
	Start()
	Stop()
	StopForce()
	AddWork(Task)
}

type Task interface {
	Execute() error
	OnFailure(error)
}

type WorkingPool struct {
	workersN int

	tasks     chan Task
	quit      chan struct{} // close to signal the workers to stop working
	quitforce chan struct{} // close to signal the workers to stop working immediately
	start     sync.Once     // ensure the pool can only be started once
	stop      sync.Once     // ensure the pool can only be stopped once
	stopforce sync.Once

	mStopped  *sync.RWMutex // check if pool was stopped before add some task
	isStopped bool          // check if pool was stopped before add some task
}

func NewPool(workersN int, chanS int) (Pool, error) {
	if workersN <= 0 {
		return nil, fmt.Errorf("worker pool cannot consist of less than 1 worker")
	}

	if chanS < 0 {
		return nil, fmt.Errorf("attempting to create worker pool with a negative channel size")
	}

	return &WorkingPool{
		workersN:  workersN,
		tasks:     make(chan Task, chanS),
		quit:      make(chan struct{}),
		quitforce: make(chan struct{}),

		start:     sync.Once{},
		stop:      sync.Once{},
		stopforce: sync.Once{},

		mStopped: &sync.RWMutex{},
	}, nil
}

func (p *WorkingPool) Start() {
	p.start.Do(func() {
		p.mStopped.Lock()

		p.isStopped = false

		p.mStopped.Unlock()

		p.startWorkers()
	})
}

func (p *WorkingPool) Stop() {
	p.stop.Do(func() {
		p.mStopped.Lock()

		p.isStopped = true

		p.mStopped.Unlock()

		close(p.quit)
	})
}

func (p *WorkingPool) StopForce() {
	p.stopforce.Do(func() {
		p.mStopped.Lock()

		p.isStopped = true

		p.mStopped.Unlock()

		close(p.quitforce)
	})
}

// AddWork adds work to the WorkingPool. If the channel buffer is full (or 0) and
// all workers are occupied, this will hang until work is consumed or Stop() is called.
func (p *WorkingPool) AddWork(t Task) {
	p.mStopped.RLock()

	if p.isStopped {
		p.mStopped.RUnlock()
		return
	}

	p.mStopped.RUnlock()

	select {
	case <-p.quit:
	case p.tasks <- t:
	}
}

// AddWorkNonBlocking adds work to the WorkingPool and returns immediately
// func (p *WorkingPool) AddWorkNonBlocking(t Task) {
// go p.AddWork(t)
// }

func (p *WorkingPool) startWorkers() {
	for i := 0; i < p.workersN; i++ {
		go func(workerNum int) {

			for {
				select {
				case <-p.quitforce:
					return

				case <-p.quit:
					for task := range p.tasks {
						if err := task.Execute(); err != nil {
							// log.Printf("W %d failed task\n", workerNum)
							task.OnFailure(err)
						}
					}

					return

				case task, ok := <-p.tasks:
					if !ok {
						return
					}

					if err := task.Execute(); err != nil {
						task.OnFailure(err)
					}

				}
			}
		}(i)
	}
}
