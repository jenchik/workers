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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job1 := incrementJobFunc("job1", &r, 2)
	job2 := incrementJobFunc("job2", &r, -1)

	w1 := workers.New(job1).ByCronSpec("@every 2s")
	w2 := workers.New(job2).ByCronSpec("@every 1s")

	w3 := workers.New(func(context.Context) {
		log.Println("job3 start, send command stop")
		cancel()
	}).ByCronSpec("@every 10s")

	w4 := workers.New(func(ctx context.Context) {
		log.Println("job4 start")
		<-ctx.Done()
		log.Println("job4 freezes for minute")
		time.Sleep(time.Minute)
		log.Println("job4 exit") // will not be printed
	}).WithDone(func() {
		log.Println("Worker #4 was runned") // will not be printed
	}).ByCronSpec("@every 1s")

	g := workers.NewGroup(context.Background())
	g.Add(w1, w2, w3, w4)

	<-grace.ShutdownContext(ctx).Done()

	log.Println("Stopping...")
	g.Stop()

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel2()
	if err := g.Wait(ctx2); err != nil {
		log.Println("Error while stopping workers:", err)
	}

	log.Println("Stopped")
}

func incrementJobFunc(name string, target *int32, delta int32) func(context.Context) {
	return func(ctx context.Context) {
		log.Printf("%s start, int before: %d after %d", name, atomic.LoadInt32(target), atomic.AddInt32(target, delta))
	}
}
