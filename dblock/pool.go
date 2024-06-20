package dblock

import "sync"

type Task func()

type Pool struct {
	wg          *sync.WaitGroup
	stopC       chan struct{}
	WorkerCount int
}

func (p *Pool) Start() chan<- Task {
	queue := make(chan Task)
	p.stopC = make(chan struct{})
	p.wg = &sync.WaitGroup{}

	for i := 0; i < p.WorkerCount; i++ {
		go p.worker(queue)
	}

	return queue
}
func (p *Pool) Stop() {
	close(p.stopC)
	p.wg.Wait()
}
func (p *Pool) worker(queue <-chan Task) {
	for {
		select {
		case task := <-queue:
			func() {
				p.wg.Add(1)
				defer func() {
					recover()
					p.wg.Done()
				}()
				task()
			}()
		case <-p.stopC:
			return
		}
	}
}
