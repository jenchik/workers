package main

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/jenchik/grace"
	"github.com/jenchik/workers"
)

func main() {
	var r int32 = 1
	log.Println("Start")

	job1 := incrementJobFunc("job1", &r, 2)
	job2 := incrementJobFunc("job2", &r, -1)

	w1 := workers.New(job1).ByTicker(time.Second * 2)
	w2 := workers.New(job2).ByCronSpec("@every 1s").WithDone(func() {
		// do something for cancel
		// example close resources or recover
		log.Println("Worker #2: closed")
	})
	w3 := workers.New(func(context.Context) { panic("test") }).ByTimer(time.Second * 2).WithDone(func() {
		e := recover()
		log.Println("Worker #3: recover:", e)
	})

	g := workers.NewGroup(context.Background())
	g.Add(w1, w2, w3)

	<-grace.ShutdownContext(context.Background()).Done()

	log.Println("Stopping...")
	g.Stop()
	g.Wait(nil)

	log.Println("Stopped")
}

func incrementJobFunc(name string, target *int32, delta int32) func(context.Context) {
	return func(ctx context.Context) {
		log.Printf("%s start, int before: %d after %d", name, atomic.LoadInt32(target), atomic.AddInt32(target, delta))
	}
}
