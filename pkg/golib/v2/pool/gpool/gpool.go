package gpool

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrLackPoolFunc will be returned when invokers don't provide function for pool.
	ErrLackPoolFunc = errors.New("must provide function for pool")

	// ErrInvalidPoolExpiry will be returned when setting a negative number as the periodic duration to purge goroutines.
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("this pool has been closed")

	// ErrPoolOverload will be returned when the pool is full and no workers available.
	ErrPoolOverload = errors.New("too many goroutines blocked on submit or Nonblocking is set")

	// ErrInvalidPreAllocSize will be returned when trying to set up a negative capacity under PreAlloc mode.
	ErrInvalidPreAllocSize = errors.New("can not set up a negative capacity under PreAlloc mode")
)

const (
	// DefaultCleanIntervalTime is the interval time to clean up goroutines.
	DefaultCleanIntervalTime = time.Second
)

const (
	// OPENED represents that the pool is opened.
	OPENED = iota

	// CLOSED represents that the pool is closed.
	CLOSED
)

var (
	// workerChanCap determines whether the channel of a worker should be a buffered channel
	// to get the best performance. Inspired by fasthttp at
	// https://github.com/valyala/fasthttp/blob/master/workerpool.go#L139
	workerChanCap = func() int {
		// Use blocking channel if GOMAXPROCS=1.
		// This switches context from sender to receiver immediately,
		// which results in higher performance (under go1.5 at least).
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}

		// Use non-blocking workerChan if GOMAXPROCS>1,
		// since otherwise the sender might be dragged down if the receiver is CPU-bound.
		return 1
	}()
)

type GPool interface {
	Running() int
	Free() int
	Cap() int
	Tune(int) int
	IsClosed() bool
	incRunning()
	decRunning()

	// 需要子类实现的方法
	Release()
	Reboot()
}

type basicPool struct {
	// capacity of the pool, a negative value means that the capacity of pool is limitless, an infinite pool is used to
	// avoid potential issue of endless blocking caused by nested usage of a pool: submitting a task to pool
	// which submits a new task to the same pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// lock for protecting the worker queue.
	lock sync.Locker

	// state is used to notice the pool to closed itself.
	state int32

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// blockingNum is the number of the goroutines already been blocked on pool.Submit, protected by pool.lock
	blockingNum int

	options *Options
}

// Running returns the number of the currently running goroutines.
func (p *basicPool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns a available goroutines to work, -1 indicates this pool is unlimited.
func (p *basicPool) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

// Cap returns the capacity of this pool.
func (p *basicPool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *basicPool) Tune(size int) {
	if capacity := p.Cap(); capacity == -1 || size <= 0 || size == capacity || p.options.PreAlloc {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

func (p *basicPool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == CLOSED
}

// incRunning increases the number of the currently running goroutines.
func (p *basicPool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *basicPool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *basicPool) Release() {
}

func (p *basicPool) Reboot() {
}
