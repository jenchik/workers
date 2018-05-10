package workers

import (
	"context"
)

type LockFunc func(context.Context, Job) Job

// Locker interface
type Locker interface {
	Lock() error
	Unlock()
}

// WithLock returns func with call Worker in lock
func WithLock(l Locker) LockFunc {
	return func(ctx context.Context, j Job) Job {
		return func(ctx context.Context) {
			if err := l.Lock(); err != nil {
				return
			}
			defer l.Unlock()
			j(ctx)
		}
	}
}
