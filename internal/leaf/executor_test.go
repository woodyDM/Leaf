package leaf

import (
	"log"
	"testing"
	"time"
)

func TestSubmit(t *testing.T) {
	r1 := newRunner(1, 1)
	r2 := newRunner(2, 1)
	r3 := newRunner(3, 2)
	r4 := newRunner(4, 2)
	r5 := newRunner(5, 3)
	r6 := newRunner(6, 4)
	r7 := newRunner(7, 1)

	pool := NewPool(3)
	go func() {
		for i := 0; i < 20; i++ {
			log.Printf("%d %v\n", pool.QueuedTask(), pool.RunningTask())
			time.Sleep(230 * time.Millisecond)
		}
	}()
	pool.Submit(r1)
	pool.Submit(r2)
	pool.Submit(r3)
	pool.Submit(r4)
	pool.Submit(r5)
	pool.Submit(r6)
	pool.Submit(r7)

	time.Sleep(10 * time.Second)
	log.Println("OK")
}

func (m *mockRunner) runnerId() uint {
	return m.id
}

func (m *mockRunner) groupId() uint {
	return m.gid
}

func (m *mockRunner) run() {
	time.Sleep(1 * time.Second)
	m.frun = true
}

func (m *mockRunner) shutdown() {
	m.fshut = true
}

func (m *mockRunner) whenError(e error) {
	panic("implement me")
}

type mockRunner struct {
	id    uint
	gid   uint
	frun  bool
	fshut bool
}

func newRunner(id uint, gid uint) *mockRunner {
	return &mockRunner{
		id:  id,
		gid: gid,
	}
}
