package workers

import (
	"context"
	"errors"
	"sync"
)

var ErrGroupStopped = errors.New("group is already stopped")

type Group struct {
	add  chan Job
	done chan struct{}
	stop context.CancelFunc
}

func NewGroup(ctx context.Context) *Group {
	g := &Group{
		add:  make(chan Job, 1),
		done: make(chan struct{}),
	}
	ctx, g.stop = context.WithCancel(ctx)
	go g.run(ctx)
	return g
}

func (g *Group) run(ctx context.Context) {
	defer close(g.done)
	wg := new(sync.WaitGroup)
	for {
		select {
		case job := <-g.add:
			wg.Add(1)
			go func(j Job) {
				defer wg.Done()
				j(ctx)
			}(job)
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

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

func (g *Group) OnDemand(worker *Worker) *onDemand {
	return &onDemand{g, worker}
}

func (g *Group) Stop() {
	g.stop()
}

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
