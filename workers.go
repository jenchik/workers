package workers

import (
	"context"
	"sync"
	"time"
)

type (
	Job func(context.Context)

	Options struct {
		Handler  Job
		Canceler func()
		TickRun  <-chan time.Time
	}

	Workers struct {
		add  chan Options
		done chan struct{}
	}
)

func New(ctx context.Context) *Workers {
	w := &Workers{
		add:  make(chan Options, 1),
		done: make(chan struct{}),
	}
	go w.run(ctx)
	return w
}

func (w *Workers) run(ctx context.Context) {
	defer close(w.done)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	for {
		select {
		case opts := <-w.add:
			wg.Add(1)
			go func(o Options) {
				defer func() {
					if o.Canceler != nil {
						o.Canceler()
					}
					wg.Done()
				}()
				for {
					select {
					case _, ok := <-o.TickRun:
						if !ok {
							return
						}
						o.Handler(ctx)
					case <-ctx.Done():
						return
					}
				}
			}(opts)
		case <-ctx.Done():
			wg.Done()
			wg.Wait()
			return
		}
	}
}

func (w *Workers) AddJob(opts Options) {
	w.add <- opts
}

func (w *Workers) Wait(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-w.done:
	}
	return
}
