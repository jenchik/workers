package workers_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jenchik/workers"
	. "github.com/smartystreets/goconvey/convey"
)

func TestByCustomSchedule(t *testing.T) {
	Convey("Given target int == 5, decrement job and custom schedule (until int > 0)", t, func() {
		var (
			i   int32 = 5
			res       = make(chan struct{})
			job       = func(ctx context.Context) {
				atomic.AddInt32(&i, -1)
			}
		)

		schedule := func(ctx context.Context, j workers.Job) workers.Job {
			return func(ctx context.Context) {
				for atomic.LoadInt32(&i) > 0 {
					j(ctx)
				}
				res <- struct{}{}
			}
		}

		Convey("When run worker", func() {
			go workers.New(job).
				BySchedule(schedule).
				Run(context.Background())

			Convey("Job should be executed 5 times", func() {
				So(readFromChannelWithTimeout(res), ShouldBeTrue)
				So(atomic.LoadInt32(&i), ShouldEqual, 0)
			})
		})
	})
}

func TestByTimer(t *testing.T) {

	Convey("Given job who send to result channel execution time and sleep for 1s", t, func() {
		res := make(chan time.Time)
		job := createWriterJob(time.Second, res)

		Convey("When create worker and run with 1s timer", func() {
			wrk := workers.
				New(job).
				ByTimer(time.Second)

			ctx, cancel := context.WithCancel(context.Background())
			go wrk.Run(ctx)
			expectedNextExecutionTime := time.Now().Add(time.Second)

			Convey("job should be executed after 1s from previous run", func() {
				timer := time.NewTimer(2 * time.Second)
				defer timer.Stop()

				for i := 0; i < 3; i++ {
					select {
					case r := <-res:
						So(int64(expectedNextExecutionTime.Sub(r).Seconds()), ShouldEqual, 0)
						expectedNextExecutionTime = r.Add(time.Second)
						timer.Reset(2 * time.Second)
					case <-timer.C:
						So(false, ShouldBeTrue)
					}
				}
			})

			Convey("When cancel context", func() {
				cancel()

				Convey("job execution should be stopped", func() {
					timer := time.NewTimer(2 * time.Second)
					defer timer.Stop()

					select {
					case <-res:
						So(false, ShouldBeTrue)
					case <-timer.C:
						So(true, ShouldBeTrue)
					}
				})
			})
		})
	})
}

func TestByTicker(t *testing.T) {

	Convey("Given job who send to result channel execution time and sleep for 1s", t, func() {
		res := make(chan time.Time)
		job := createWriterJob(time.Second, res)

		Convey("When create worker and run with 1s ticker", func() {
			wrk := workers.
				New(job).
				ByTicker(time.Second)

			ctx, cancel := context.WithCancel(context.Background())
			go wrk.Run(ctx)
			expectedNextExecutionTime := time.Now().Add(time.Second)

			Convey("job should be executed every 1s", func() {
				timer := time.NewTimer(2 * time.Second)
				defer timer.Stop()

				for i := 0; i < 3; i++ {
					select {
					case r := <-res:
						So(int64(expectedNextExecutionTime.Sub(r).Seconds()), ShouldEqual, 0)
						expectedNextExecutionTime = r.Add(time.Second)
						timer.Reset(2 * time.Second)
					case <-timer.C:
						So(false, ShouldBeTrue)
					}
				}
			})

			Convey("When cancel context", func() {
				cancel()

				Convey("job execution should be stopped", func() {
					timer := time.NewTimer(time.Second)
					defer timer.Stop()

					select {
					case <-res:
						So(false, ShouldBeTrue)
					case <-timer.C:
						So(true, ShouldBeTrue)
					}
				})
			})
		})
	})

	Convey("Given job who send to channels start/stop signals, blocking with context", t, func() {
		var (
			start    = make(chan struct{})
			stop     = make(chan struct{})
			complete = make(chan struct{})
		)

		job := func(ctx context.Context) {
			start <- struct{}{}
			<-ctx.Done()
			stop <- struct{}{}
		}

		Convey("When run with ticker", func() {
			wrk := workers.New(job).ByTicker(time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				wrk.Run(ctx)
				complete <- struct{}{}
			}()
			So(readFromChannelWithTimeout(start), ShouldBeTrue)

			Convey("Cancel context should stop job on next run (context check prioriity)", func() {
				cancel()
				So(readFromChannelWithTimeout(stop), ShouldBeTrue)
				So(readFromChannelWithTimeout(complete), ShouldBeTrue)
			})
		})
	})

	Convey("Given job which send stop start events to channels", t, func() {
		var (
			start    = make(chan struct{})
			stop     = make(chan struct{})
			complete = make(chan struct{})
		)

		job := func(ctx context.Context) {
			start <- struct{}{}
			stop <- struct{}{}
		}

		Convey("When start job with ticker, set immediiatly for run on start", func() {
			wrk := workers.
				New(job).
				ByTicker(time.Minute).
				SetImmediately(true)

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				wrk.Run(ctx)
				complete <- struct{}{}
			}()

			Convey("Job should execute, cancel context should stop worker", func() {
				So(readFromChannelWithTimeout(start), ShouldBeTrue)
				So(readFromChannelWithTimeout(stop), ShouldBeTrue)

				// skip context check priiority
				<-time.Tick(100 * time.Millisecond)
				cancel()
				So(readFromChannelWithTimeout(complete), ShouldBeTrue)
			})
		})
	})
}

func TestByCronSchedule(t *testing.T) {

	Convey("Given job who send to result channel execution time and sleep for 1s", t, func() {
		res := make(chan time.Time)
		job := createWriterJob(time.Microsecond, res)

		Convey("When create worker with incorrect cron spec should panic", func() {
			So(func() { workers.New(job).ByCronSpec("завтра") }, ShouldPanic)
			So(func() { workers.New(job).ByCronSpec("@today") }, ShouldPanic)
			So(func() { workers.New(job).ByCronSpec("*") }, ShouldPanic)
		})

		Convey("When create worker and run with 1s cron schedule", func() {
			wrk := workers.
				New(job).
				ByCronSpec("@every 1s")

			ctx, cancel := context.WithCancel(context.Background())
			go wrk.Run(ctx)
			expectedNextExecutionTime := time.Now().Add(time.Second)

			Convey("job should be executed every 1s", func() {
				timer := time.NewTimer(2 * time.Second)
				defer timer.Stop()

				for i := 0; i < 3; i++ {
					select {
					case r := <-res:
						So(int64(expectedNextExecutionTime.Sub(r).Seconds()), ShouldEqual, 0)
						expectedNextExecutionTime = r.Add(time.Second)
						timer.Reset(2 * time.Second)
					case <-timer.C:
						So(false, ShouldBeTrue)
					}
				}
			})

			Convey("When cancel context", func() {
				cancel()

				Convey("job execution should be stopped", func() {
					timer := time.NewTimer(time.Second)
					defer timer.Stop()

					select {
					case <-res:
						So(false, ShouldBeTrue)
					case <-timer.C:
						So(true, ShouldBeTrue)
					}
				})
			})
		})
	})
}
func createWriterJob(sleep time.Duration, ch chan time.Time) workers.Job {
	return func(ctx context.Context) {
		select {
		case ch <- time.Now():
		case <-ctx.Done():
		}
	}
}
