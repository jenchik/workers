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

	w1 := workers.New(job1).ByCronSpec("@every 2s")
	w2 := workers.New(job2).ByCronSpec("@every 1s")
	w3 := workers.New(func(ctx context.Context) {
		log.Println("job3 start")
		<-ctx.Done()
		log.Println("job3 freezes for 5 seconds")
		time.Sleep(time.Second * 5)
	}).ByCronSpec("@every 1s")

	g := workers.NewGroup(context.Background())
	g.Add(w1, w2)

	g2 := workers.NewGroup(context.Background())
	g2.Add(w3)
	g.AddGroup(g2)

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
