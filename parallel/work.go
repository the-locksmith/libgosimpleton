package parallel

import(
	"runtime"

	"sync"
	"time"
)

/*
 * Work has locks and can utilize wait groups.
 * Locks are a feature for the working in the goroutines. Internally
 * locking is only used for Wait() to be thread safe.
 * Wait groups are used with Start() and it's wrapper Do() to ensure
 * all work is done. With Run() wait groups are not used and a call
 * to Wait() will return immidiatly. This however is not enforced to 
 * make it possible to use Start() and Run() at the same time for
 * different purposes.
 * 
 * See work_test.go for examples.
*/
type Work struct {
	sync.Mutex

	// worker of numbers
	numberOfWorkers uint

	// wait!
	waitGroup	*sync.WaitGroup

	// final function. hint: use it to close channels
	finallyFunc	func()

	// used for initialise
	once		sync.Once

	// ensure waitGroup.Wait() only gets called once.
	waitOnce	sync.Once

	// required to reset work.waitOnce
	reset		sync.Once
}

// New York
func NewWork(anumberOfWorkers uint) (work *Work) {
	work = &Work{ numberOfWorkers: anumberOfWorkers }
	work.Initialise()
	return
}

// New York Finally
func NewWorkFinally(anumberOfWorkers uint, afinally func()) (work *Work) {
	work = &Work{ numberOfWorkers: anumberOfWorkers,
			finallyFunc: afinally,
			}
	work.Initialise()
	return
}

// i n i t i a l i z e
func (work *Work) Initialise() {
	work.once.Do(func() {
		if runtime.GOMAXPROCS(0) <= runtime.NumCPU() {
			runtime.GOMAXPROCS(runtime.NumCPU() * 2)
		}
		work.waitGroup = &sync.WaitGroup{}
	})
}

// start numberOfWorkers - 1 in separate goroutines and run one
// worker in this goroutine. the worker should block till the work is done.
// it is not guaranteed that all work will be processed nor that the last
// worker is really the last finishing (which can mean, that workers still
// processing work can be interrupted and shut down by program termination).
func (work *Work) Run(worker func()) {
	go func () {
		for i := uint(0); i < work.numberOfWorkers -1; i++ {
			go worker()
		}
	}()

	worker()
}

// do some work; this is equivalent to calling Start() and then Wait().
// calls finally func afterwards. guarantees all work is done. it blocks until
// all workers finish. should not be called with the go command.
func (work *Work) Do(worker func()) {
	work.start(worker)
	work.finally()
}

// start workers; a worker should read from a channel and from that do work
// does not call finally as it doesn't record when or if a goroutine finishes.
// can guarantee all work is done, if work.Wait is called. does not block.
func (work *Work) Start(worker func()) {
	work.waitGroup.Add(1)
	go func() {
		work.start(worker)
		work.waitGroup.Done()
	}()
}

// start a bunch of workers with waitGroup
func (work *Work) start(worker func()) {
	for i := uint(0); i < work.numberOfWorkers; i++ {
		work.waitGroup.Add(1)
		go func() {
			worker()
			work.waitGroup.Done()
		}()
	}
}

// feed workers; yummy; loop over something and push each value on a channel.
// it is good practise to first start feeding and then to open the workers
// via Do(), Start() or Run().
func (work *Work) Feed(feeder func()) {
	go func() {
		defer work.RecoverSendOnClosedChannel()
		feeder()
	}()
}

// feed workers by ticks; the tick function must not loop indefinitly, but return as quickly as possible.
// returns a channel that must be closed or a boolean value send to, when not needed anymore
// otherwise the tick will run forever.
func (work *Work) Tick(duration time.Duration, tick func()) (cancel chan bool) {
	cancel = make(chan bool, 1)

	go func() {
		defer work.RecoverSendOnClosedChannel()
	infinite:
		for {
			select {
			case <-time.Tick(duration):
				tick()
			// if anything happens on the cancel channel, abort.
			case <-cancel:
				break infinite
			}
		}
	}()

	return
}

// recover the panic 'send on closed channel' and ignore it. otherwise panic some more.
func (work *Work) RecoverSendOnClosedChannel() {
	// be safe for the future
	recovered := recover(); if recovered != nil {
		if e, ok := recovered.(error); ok {
			if e.Error() != "send on closed channel" {
				panic(e)
			}
		}
	}

	if e, ok := recovered.(string); ok {
		if e != "send on closed channel" {
			panic(e)
		}
	}
}

// ensure waitGroup.Wait() and work.finally are only called once then reset.
func (work *Work) finally() {
	work.Wait()

	if work.finallyFunc == nil {
		return
	}

	work.finallyFunc()
}

// a wrapper around the internally used *sync.WaitGroup.Wait(); thread safe
func (work *Work) Wait() {
	work.waitOnce.Do(func() {
		work.waitGroup.Wait()
		// allow resetting of work.waitGroup.Wait()
		work.reset = sync.Once{}
	})

	// reset work.WaitOnce
	work.reset.Do(func() {
		// reset of work.waitGroup.Wait()
		work.waitOnce = sync.Once{}
	})
}

// cancel a work. this should be preferably done by closing a channel.
// alternativly use a second, bool typed, channel to send a cancel signal.
func (work *Work) Timeout(duration time.Duration, cancel func()) {
	go func() {
		select {
			case <-time.After(duration):
				cancel()
		}
	}()
}

/* convinience functions */

// suggest a number of workers.
func SuggestNumberOfWorkers() uint {
	return uint(runtime.NumCPU() * 2)
}