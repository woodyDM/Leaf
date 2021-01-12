package leaf

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var CommonPool = NewPool(4)
var DefaultQueueSize = 1000000

type worker struct {
	ch chan Runner
	no int
}

func (p *Pool) RunningTask() []uint {
	p.lock.RLock()
	defer p.lock.RUnlock()
	r := make([]uint, 0)
	for k, _ := range p.container {
		r = append(r, k)
	}
	return r
}

func (p *Pool) QueuedTask() int {
	count := len(p.ch)
	for _, w := range p.workers {
		count += len(w.ch)
	}
	return count
}

func newWorker(no int) *worker {
	return &worker{
		ch: make(chan Runner, DefaultQueueSize),
		no: no,
	}
}

func (w *worker) size() int {
	return len(w.ch)
}

type Runner interface {
	runnerId() uint
	groupId() uint
	run()
	shutdown()
	whenError(e error)
}

type runnerWithWorker struct {
	run     Runner
	w       *worker
	counter int
}

func (r *runnerWithWorker) increment(i int) {
	r.counter += i
}

type Pool struct {
	size           int
	ch             chan Runner
	workers        []*worker
	container      map[uint]Runner
	groupContainer map[uint]*runnerWithWorker
	lock           sync.RWMutex
}

func (p *Pool) decideWorker(r Runner) {
	p.lock.Lock()
	defer p.lock.Unlock()
	runWrapper, ok := p.groupContainer[r.groupId()]

	if ok {
		//if group already running , use same channel
		runWrapper.increment(1)
		runWrapper.w.ch <- r
	} else {
		//find min size channel
		idx := 0
		num := p.workers[0].size()
		for i, win := range p.workers {
			currentLen := win.size()
			if currentLen < num {
				num = currentLen
				idx = i
			}
		}
		w := p.workers[idx]
		w.ch <- r
		wrapper := &runnerWithWorker{
			run:     r,
			w:       w,
			counter: 1,
		}
		p.groupContainer[r.groupId()] = wrapper
	}
}

func (p *Pool) Submit(r Runner) {
	if r != nil {
		p.ch <- r
	}
}

func start(pool *Pool, w *worker) {
	for it := range w.ch {
		log.Printf("[Worker %d Runner %d]: start to handle \n", w.no, it.runnerId())
		handleRunner(it, pool, w)
	}
}

func (p *Pool) get(runnerId uint) (Runner, bool) {
	p.lock.RLock()
	ctx, ok := p.container[runnerId]
	defer p.lock.RUnlock()
	return ctx, ok
}

func (p *Pool) start() {
	for i := 0; i < p.size; i++ {
		p.workers[i] = newWorker(i)
		go start(p, p.workers[i])
	}
	go func() {
		for r := range p.ch {
			p.decideWorker(r)
		}
	}()
}

func handleRunner(run Runner, p *Pool, w *worker) {
	defer func() {
		p.lock.Lock()
		delete(p.container, run.runnerId())
		wrapper := p.groupContainer[run.groupId()]
		if wrapper.counter == 1 {
			delete(p.groupContainer, run.groupId())
		} else {
			wrapper.increment(-1)
		}
		p.lock.Unlock()
		if p := recover(); p != nil {
			var e error
			if err, ok := p.(error); ok {
				e = err
			} else {
				msg := fmt.Sprintf("%v", p)
				e = errors.New(msg)
			}
			run.whenError(e)
			log.Printf("[Worker %d Runner %d] exit with error %v.\n", w.no, run.runnerId(), e)
		} else {
			log.Printf("[Worker %d Runner %d] exit without error.\n", w.no, run.runnerId())
		}
	}()
	p.lock.Lock()
	p.container[run.runnerId()] = run
	p.lock.Unlock()
	run.run()
}

func NewPool(coreSize int) *Pool {
	p := &Pool{
		size:           coreSize,
		workers:        make([]*worker, coreSize),
		container:      make(map[uint]Runner),
		groupContainer: make(map[uint]*runnerWithWorker),
		ch:             make(chan Runner, DefaultQueueSize),
	}
	p.start()
	return p
}
