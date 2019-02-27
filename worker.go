package workers

import (
	"context"
	"time"
)

type (
	// Job is target background job
	Job func(context.Context)

	// Worker is builder for job with optional schedule and exclusive control
	Worker struct {
		job         Job
		done        func()
		locker      LockFunc
		schedule    ScheduleFunc
		immediately bool
	}
)

// New returns new worker with target job
func New(job Job) *Worker {
	return &Worker{
		job: job,
	}
}

// BySchedule set schedule wrapper func for job
func (w *Worker) BySchedule(s ScheduleFunc) *Worker {
	w.schedule = s
	return w
}

// ByTimer set schedule timer job wrapper with period
func (w *Worker) ByTimer(period time.Duration) *Worker {
	w.schedule = ByTimer(period)
	return w
}

// ByTicker set schedule ticker job wrapper with period
func (w *Worker) ByTicker(period time.Duration) *Worker {
	w.schedule = ByTicker(period)
	return w
}

// ByCronSpec set schedule job wrapper by cron spec
func (w *Worker) ByCronSpec(spec string) *Worker {
	w.schedule = ByCronSchedule(spec)
	return w
}

// SetImmediately set execute job on Run setting
func (w *Worker) SetImmediately(executeOnRun bool) *Worker {
	w.immediately = executeOnRun
	return w
}

// WithDone set job with defer custom function
func (w *Worker) WithDone(done func()) *Worker {
	w.done = done
	return w
}

// WithLock set job lock wrapper
func (w *Worker) WithLock(l Locker) *Worker {
	w.locker = WithLock(l)
	return w
}

// Run job, wrap job to lock and schedule wrappers
func (w *Worker) Run(ctx context.Context) {
	if w.done != nil {
		defer w.done()
	}
	job := w.job

	if w.locker != nil {
		job = w.locker(ctx, job)
	}

	if w.immediately {
		job(ctx)

		if w.schedule == nil {
			return
		}

		// check context before run immediately job again
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	if w.schedule != nil {
		job = w.schedule(ctx, job)
	}

	job(ctx)
}

// RunOnce job, wrap job to lock
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
