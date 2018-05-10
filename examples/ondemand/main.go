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
	job3 := incrementJobFunc("job3", &r, 10)

	w1 := workers.New(job1).ByTicker(time.Second * 2)
	w2 := workers.New(job2).ByCronSpec("@every 1s")
	w3 := workers.New(job3).WithDone(func() {
		log.Println("Worker #3 was runned")
	})

	g := workers.NewGroup(context.Background())
	g.Add(w1, w2)

	d := g.OnDemand(w3)
	time.AfterFunc(time.Second*5, func() { d.Run() })

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
