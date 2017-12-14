package times

import (
	"container/ring"
	"sync"
	"time"

	"github.com/cuigh/auxo/data"
)

type entry struct {
	action func() bool
	next   *entry
}

// Wheel is a time wheel like a clock.
type Wheel struct {
	lock     sync.Mutex
	r        *ring.Ring
	interval time.Duration
	stopper  data.Chan
}

// NewWheel creates a Wheel.
func NewWheel(interval time.Duration, count int) *Wheel {
	w := &Wheel{
		r:        ring.New(count),
		interval: interval,
		stopper:  make(data.Chan, 1),
	}
	go w.start()
	return w
}

func (w *Wheel) start() {
	t := time.NewTicker(w.interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if w.r = w.r.Next(); w.r.Value != nil {
				go w.fire(w.r)
			}
		case <-w.stopper:
			return
		}
	}
}

// Stop prevents the wheel from firing.
func (w *Wheel) Stop() {
	w.stopper.TrySend()
}

// Add adds an action to tail. If action return true, it will be fired again at next cycle.
func (w *Wheel) Add(action func() bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	e := &entry{action: action}
	r := w.r.Prev()
	if v := r.Value; v != nil {
		e.next = v.(*entry)
	}
	r.Value = e
}

func (w *Wheel) fire(r *ring.Ring) {
	w.lock.Lock()
	defer w.lock.Unlock()

	var head, tail *entry
	for e := r.Value.(*entry); e != nil; e = e.next {
		if e.action() {
			if head == nil {
				head, tail = e, e
			} else {
				tail.next = e
				tail = tail.next
			}
		}
	}
	if head == nil {
		r.Value = nil
	} else {
		tail.next = nil
		r.Value = head
	}
}
