package leaf

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var CommonPool = NewPool(4)

type Runner interface {
	runnerId() uint
	run()
	shutdown()
	whenError(e error)
}

type Pool struct {
	size      int
	ch        chan Runner
	container map[uint]Runner
	lock      sync.RWMutex
}

func (p *Pool) submit(r Runner) {
	p.ch <- r
}
func (p *Pool) get(id uint) (Runner, bool) {
	p.lock.RLock()
	ctx, ok := p.container[id]
	p.lock.RUnlock()
	return ctx, ok
}

func (p *Pool) start() {
	for i := 0; i < p.size; i++ {
		go func() {
			for it := range p.ch {
				log.Println("Start to handle Runner :", it.runnerId())
				p.handleRunner(it)
			}
		}()
	}
}

func (p *Pool) handleRunner(run Runner) {
	defer func() {
		p.lock.Lock()
		delete(p.container, run.runnerId())
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
			log.Printf("Runner %d complete with error %v.\n", run.runnerId(), e)
		} else {
			log.Printf("Runner %d complete without error.\n", run.runnerId())
		}
	}()
	p.lock.Lock()
	p.container[run.runnerId()] = run
	p.lock.Unlock()
	run.run()
}

func NewPool(coreSize int) *Pool {
	p := &Pool{
		size:      coreSize,
		ch:        make(chan Runner, 100000),
		container: make(map[uint]Runner),
	}
	p.start()
	return p
}
