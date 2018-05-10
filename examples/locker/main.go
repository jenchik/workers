package main

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/jenchik/grace"
	"github.com/jenchik/workers"
)

func main() {
	var (
		r1 int32 = 1
		r2 int32 = 2
		r3 int32 = 3
		r4 int32 = 4
	)
	log.Println("Start")

	job1 := incrementJobFunc("job1", &r1, -1)
	job2 := incrementJobFunc("job2", &r2, -1)
	job3 := incrementJobFunc("job3", &r3, -1)
	job4 := incrementJobFunc("job4", &r4, -1)

	// custom schedule, until 0
	scheduleFunc := func(target *int32) func(ctx context.Context, j workers.Job) workers.Job {
		return func(ctx context.Context, j workers.Job) workers.Job {
			return func(ctx context.Context) {
				for atomic.LoadInt32(target) > 0 {
					j(ctx)
				}
			}
		}
	}

	customLocker := &customLocker{}

	w1 := workers.New(job1).BySchedule(scheduleFunc(&r1))
	w2 := workers.New(job2).BySchedule(scheduleFunc(&r2)).WithLock(customLocker)
	w3 := workers.New(job3).BySchedule(scheduleFunc(&r3)).WithLock(customLocker)
	w4 := workers.New(job4).BySchedule(scheduleFunc(&r4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g := workers.NewGroup(ctx)
	g.Add(w1, w2, w3, w4)

	<-grace.ShutdownContext(context.Background()).Done()

	log.Println("Stopping...")
	cancel()
	g.Wait(nil)

	log.Println("Stopped")
}

type customLocker struct {
	locked int32
}

func (c *customLocker) Lock() error {
	if atomic.CompareAndSwapInt32(&c.locked, 0, 1) {
		return nil
	}
	return errors.New("locked")
}

func (c *customLocker) Unlock() {
	atomic.StoreInt32(&c.locked, 0)
}

func incrementJobFunc(name string, target *int32, delta int32) func(context.Context) {
	return func(ctx context.Context) {
		log.Printf("%s start, int before: %d after %d", name, atomic.LoadInt32(target), atomic.AddInt32(target, delta))
	}
}
