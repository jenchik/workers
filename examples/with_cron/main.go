package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jenchik/workers"
	"github.com/robfig/cron"
)

func wrapCron(c *cron.Cron, spec string, job workers.Job) (o workers.Options, err error) {
	t := make(chan time.Time)
	err = c.AddFunc(spec, func() {
		select {
		case t <- time.Now():
		default:
		}
	})
	if err != nil {
		return o, err
	}

	o.Handler = job
	o.TickRun = t
	return
}

func main() {
	schedule := cron.New()
	schedule.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Start")
	w := workers.New(ctx)

	opts, _ := wrapCron(schedule, "@every 2s", func(ctx context.Context) {
		fmt.Printf("[%s] Do something\n", time.Now().Format(time.RFC3339))
	})
	w.AddJob(opts)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-ch

	fmt.Println("Stoping...")
	cancel()

	gctx, gcancel := context.WithTimeout(context.Background(), time.Second*5)
	defer gcancel()
	w.Wait(gctx)

	fmt.Println("Stoped")
}
