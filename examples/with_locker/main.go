package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/jenchik/workers"
)

func wrapLocker(store lock.RedisClient, key string, opts lock.Options, job workers.Job) workers.Job {
	return func(ctx context.Context) {
		l, err := lock.Obtain(store, key, &opts)
		if err != nil {
			return
		}
		defer l.Unlock()
		job(ctx)
	}
}

func wrapLockerWithError(store lock.RedisClient, key string, opts lock.Options, job func(context.Context, error)) workers.Job {
	return func(ctx context.Context) {
		l, err := lock.Obtain(store, key, &opts)
		if err != nil {
			job(ctx, err)
			return
		}
		defer l.Unlock()
		job(ctx, nil)
	}
}

func main() {
	store := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Start")
	w := workers.New(ctx)

	opts := lock.Options{
		LockTimeout: time.Second,
		RetryCount:  3,
		RetryDelay:  time.Millisecond * 200,
	}

	// sample worker #1
	t := time.NewTicker(time.Second * 2)
	w.AddJob(workers.Options{
		Handler: wrapLocker(store, "worker #1", opts, func(ctx context.Context) {
			fmt.Printf("[%s] #1 Do something\n", time.Now().Format(time.RFC3339))
		}),
		TickRun: t.C,
	})

	// sample worker #2
	t2 := time.NewTicker(time.Second)
	w.AddJob(workers.Options{
		Handler: wrapLockerWithError(store, "worker #2", opts, func(ctx context.Context, err error) {
			if err != nil {
				fmt.Printf("[%s] Error: %v\n", time.Now().Format(time.RFC3339))
				return
			}
			fmt.Printf("[%s] #2 Do something\n", time.Now().Format(time.RFC3339))
		}),
		TickRun: t2.C,
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
