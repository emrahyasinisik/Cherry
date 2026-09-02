package llm

import (
	"context"
	"sync"

	"github.com/icerde/api/internal/store"
)

type Occupancy struct {
	BusyA  bool
	BusyB  bool
	Queued int
	Last   store.LlmSlot
}

type lease struct {
	Slot    store.LlmSlot
	Version store.LlmVersion
	q       *queue
}

func (l lease) Release() {
	if l.q != nil && l.Slot != "" {
		l.q.release(l.Slot)
	}
}

type waiter struct {
	ch chan store.LlmSlot
}

type queue struct {
	mu      sync.Mutex
	busyA   bool
	busyB   bool
	last    store.LlmSlot
	waiters []*waiter
}

func newQueue() *queue {
	return &queue{}
}

func (q *queue) snapshot() Occupancy {
	q.mu.Lock()
	defer q.mu.Unlock()
	last := q.last
	if last == "" {
		last = store.SlotA
	}
	return Occupancy{
		BusyA:  q.busyA,
		BusyB:  q.busyB,
		Queued: len(q.waiters),
		Last:   last,
	}
}

func (q *queue) acquire(ctx context.Context) (store.LlmSlot, error) {
	q.mu.Lock()
	if !q.busyA {
		q.busyA = true
		q.last = store.SlotA
		q.mu.Unlock()
		return store.SlotA, nil
	}
	if !q.busyB {
		q.busyB = true
		q.last = store.SlotB
		q.mu.Unlock()
		return store.SlotB, nil
	}
	w := &waiter{ch: make(chan store.LlmSlot, 1)}
	q.waiters = append(q.waiters, w)
	q.mu.Unlock()

	select {
	case slot := <-w.ch:
		return slot, nil
	case <-ctx.Done():
		q.cancel(w)
		select {
		case slot := <-w.ch:
			q.release(slot)
		default:
		}
		return "", ctx.Err()
	}
}

func (q *queue) cancel(target *waiter) {
	q.mu.Lock()
	defer q.mu.Unlock()
	kept := q.waiters[:0]
	for _, w := range q.waiters {
		if w != target {
			kept = append(kept, w)
		}
	}
	q.waiters = kept
}

func (q *queue) release(slot store.LlmSlot) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.waiters) > 0 {
		w := q.waiters[0]
		q.waiters = q.waiters[1:]
		q.last = slot
		w.ch <- slot
		return
	}
	switch slot {
	case store.SlotA:
		q.busyA = false
	case store.SlotB:
		q.busyB = false
	}
}
