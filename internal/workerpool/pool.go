package workerpool

import (
	"context"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
)

type Job func(context.Context) error

type Pool struct {
	workers int
	jobs    chan Job
	life    *poolLifecycle
}

func New(workers, queue int) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queue < workers {
		queue = workers
	}
	return &Pool{
		workers: workers,
		jobs:    make(chan Job, queue),
		life:    newPoolLifecycle(),
	}
}

func (p *Pool) Start(ctx context.Context) {
	workerContext, ok := p.life.start(ctx)
	if !ok {
		return
	}
	for i := 0; i < p.workers; i++ {
		p.life.addWorker()
		go func() {
			defer p.life.workerDone()
			for {
				select {
				case <-workerContext.Done():
					return
				case <-p.life.done():
					return
				case job := <-p.jobs:
					if job != nil {
						_ = job(workerContext)
					}
				}
			}
		}()
	}
}
func (p *Pool) Submit(ctx context.Context, job Job) error {
	if job == nil {
		return domain.ErrInvalid
	}
	if !p.life.accepting() {
		return domain.ErrState
	}
	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.life.done():
		return domain.ErrState
	}
}

func (p *Pool) Stop() {
	p.life.stop()
}
