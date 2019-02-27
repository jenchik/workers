package workers

import (
	"context"
	"errors"
	"sync"
)

// ErrGroupStopped group error message already stopped
var ErrGroupStopped = errors.New("group is already stopped")

// Group of workers controlling background jobs execution
// allows graceful stop all running background jobs
type Group struct {
	add     chan Job
	done    chan struct{}
	running chan struct{}
	stop    context.CancelFunc
}

// NewGroup yield new workers group
func NewGroup(ctx context.Context) *Group {
	g := &Group{
		add:     make(chan Job),
		done:    make(chan struct{}),
		running: make(chan struct{}),
	}
	ctx, g.stop = context.WithCancel(ctx)
	go g.run(ctx)
	return g
}

func (g *Group) run(ctx context.Context) {
	defer close(g.done)
	wg := new(sync.WaitGroup)
	jobs := make([]Job, 0, 8)
	do := func(j Job) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			j(ctx)
		}()
	}
	for {
		select {
		case job := <-g.add:
			if jobs != nil {
				jobs = append(jobs, job)
				continue
			}
			do(job)
		case <-g.running:
			for _, job := range jobs {
				do(job)
			}
			jobs = nil
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

// Add workers to group, if group runned then start worker immediately
func (g *Group) Add(workers ...*Worker) error {
	for _, worker := range workers {
		if worker == nil || worker.job == nil {
			continue
		}
		select {
		case g.add <- worker.Run:
		case <-g.done:
			return ErrGroupStopped
		}
	}
	return nil
}

// OnDemand link worker with group and then run by on demand
func (g *Group) OnDemand(worker *Worker) *onDemand {
	return &onDemand{g, worker}
}

// Run starting each worker in separate goroutine with wait.Group control
func (g *Group) Run() {
	select {
	case g.running <- struct{}{}:
	case <-g.done:
	}
}

// Stop cancel workers context
func (g *Group) Stop() {
	g.stop()
}

// Wait until all runned workers was completed.
// Be careful! It can be deadlock if some worker hanging
func (g *Group) Wait(ctx context.Context) (err error) {
	if ctx == nil {
		<-g.done
		return
	}
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-g.done:
	}
	return
}

// AddGroup add groups to group as child
func (g *Group) AddGroup(groups ...*Group) error {
	w := New(func(ctx context.Context) {
		<-ctx.Done()
		for _, child := range groups {
			child.Stop()
		}
		for _, child := range groups {
			child.Wait(nil)
		}
	})
	return g.Add(w)
}
