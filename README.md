# Worker

[![GoDoc](http://godoc.org/github.com/jenchik/workers?status.png)](http://godoc.org/github.com/jenchik/workers)
[![Build Status](https://travis-ci.org/jenchik/workers.svg?branch=master)](https://travis-ci.org/jenchik/workers)
[![codecov](https://codecov.io/gh/jenchik/workers/branch/master/graph/badge.svg)](https://codecov.io/gh/jenchik/workers)
[![Go Report Card](https://goreportcard.com/badge/github.com/jenchik/workers?)](https://goreportcard.com/report/github.com/jenchik/workers)
[![codebeat badge](https://codebeat.co/badges/e7cc5c65-0017-48fb-a963-832f9f7b4f07)](https://codebeat.co/projects/github-com-jenchik-workers-master)

Package worker adding the abstraction layer around background jobs,
allows make a job periodically, observe execution time and to control concurrent execution.

Group of workers allows to control jobs start time and
wait until all runned workers finished when we need stop all jobs.

## Features

* Scheduling, use one from existing `workers.By*` schedule functions. Supporting cron schedule spec format by [robfig/cron](https://github.com/robfig/cron) parser.
* Graceful stop, wait until all running jobs was completed.

## Example

```go
wg := workers.NewGroup(context.Background())
wg.Add(
    workers.
        New(func(context.Context) {}).
        ByTicker(time.Second),

    workers.
        New(func(context.Context) {}).
        ByTimer(time.Second),

    workers.
        New(func(context.Context) {}).
        ByCronSpec("@every 1s"),
)
wg.Run()
```

See more examples [here](/examples)
