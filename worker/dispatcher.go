package worker

import (
	"sync"
)

type Dispatcher struct {
	WorkerPool chan chan Job
	MaxWorker  int
	quit       chan bool
}

func NewDispater(maxWorker int) *Dispatcher {
	return &Dispatcher{
		MaxWorker:  maxWorker,
		WorkerPool: make(chan chan Job, maxWorker),
		quit:       make(chan bool),
	}
}

func (d *Dispatcher) Run(jobQueue chan Job, errChan chan error, wg *sync.WaitGroup) {
	for i := 0; i < d.MaxWorker; i++ {
		w := NewWorker(d.WorkerPool, errChan, wg)
		go w.Run()
	}
	d.dispatch(jobQueue)
}

func (d *Dispatcher) dispatch(jobQueue chan Job) {
	for {
		select {
		case job := <-jobQueue:
			jobChan := <-d.WorkerPool
			go func(jobchan chan Job, j Job) {
				jobchan <- j
			}(jobChan, job)
		case <-d.quit:
			return
		}
	}
}

func (d *Dispatcher) Stop() {
	d.quit <- true
}
