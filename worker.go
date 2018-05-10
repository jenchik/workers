package workers

import (
	"context"
	"time"
)

type (
	Job func(context.Context)

	Worker struct {
		job      Job
		done     func()
		locker   LockFunc
		schedule ScheduleFunc
	}
)

func New(job Job) *Worker {
	return &Worker{
		job: job,
	}
}

func (w *Worker) BySchedule(s ScheduleFunc) *Worker {
	w.schedule = s
	return w
}

func (w *Worker) ByTimer(period time.Duration) *Worker {
	w.schedule = ByTimer(period)
	return w
}

func (w *Worker) ByTicker(period time.Duration) *Worker {
	w.schedule = ByTicker(period)
	return w
}

func (w *Worker) ByCronSpec(spec string) *Worker {
	w.schedule = ByCronSchedule(spec)
	return w
}

func (w *Worker) WithDone(done func()) *Worker {
	w.done = done
	return w
}

func (w *Worker) WithLock(l Locker) *Worker {
	w.locker = WithLock(l)
	return w
}

func (w *Worker) Run(ctx context.Context) {
	if w.done != nil {
		defer w.done()
	}
	job := w.job

	if w.locker != nil {
		job = w.locker(ctx, job)
	}

	if w.schedule != nil {
		job = w.schedule(ctx, job)
	}

	job(ctx)
}

func (w *Worker) RunOnce(ctx context.Context) {
	if w.done != nil {
		defer w.done()
	}
	job := w.job

	if w.locker != nil {
		job = w.locker(ctx, job)
	}

	job(ctx)
}
