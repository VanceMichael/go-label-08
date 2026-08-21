package workerpool

import (
	"context"
	"sync"
)

type poolLifecycle struct {
	mu       sync.RWMutex
	stopOnce sync.Once
	workers  sync.WaitGroup
	doneCh   chan struct{}
	started  bool
	stopped  bool
	runCtx   context.Context
}

func newPoolLifecycle() *poolLifecycle {
	return &poolLifecycle{doneCh: make(chan struct{})}
}

func (life *poolLifecycle) start(parent context.Context) (context.Context, bool) {
	if parent == nil {
		parent = context.Background()
	}
	life.mu.Lock()
	defer life.mu.Unlock()
	if life.started || life.stopped {
		return nil, false
	}
	life.started = true
	life.runCtx = parent
	return life.runCtx, true
}

func (life *poolLifecycle) accepting() bool {
	life.mu.RLock()
	defer life.mu.RUnlock()
	return life.started && !life.stopped
}

func (life *poolLifecycle) done() <-chan struct{} {
	return life.doneCh
}

func (life *poolLifecycle) addWorker() {
	life.workers.Add(1)
}

func (life *poolLifecycle) workerDone() {
	life.workers.Done()
}

func (life *poolLifecycle) stopAndClose() {
	life.mu.Lock()
	life.stopped = true
	life.mu.Unlock()
	close(life.doneCh)
}

func (life *poolLifecycle) stop() {
	life.stopOnce.Do(life.stopAndClose)
	life.workers.Wait()
}
