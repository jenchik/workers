package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jenchik/workers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Start")
	w := workers.New(ctx)

	t := time.NewTicker(time.Second * 2)
	w.AddJob(workers.Options{
		Handler: func(ctx context.Context) {
			fmt.Printf("[%s] Do something\n", time.Now().Format(time.RFC3339))
		},
		Canceler: func() {
			t.Stop()
			// do something for cancel
			// example close resources or recover
			fmt.Println("Worker #1: closed")
		},
		TickRun: t.C,
	})

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
