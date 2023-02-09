package main

import "sync"

type queue[T any] struct {
	mu    *sync.Mutex
	cond  *sync.Cond
	items []T
}

func newQueueWithCapacity[T any](size int) queue[T] {
	q := queue[T]{
		mu:    &sync.Mutex{},
		items: make([]T, 0, size),
	}
	q.cond = sync.NewCond(q.mu)
	return q
}

func (q *queue[T]) pushBack(v T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, v)
	if len(q.items) == 1 {
		q.cond.Broadcast()
	}
}

func (q *queue[T]) popFront() T {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		q.cond.Wait()
	}
	r := q.items[0]
	copy(q.items, q.items[1:])
	q.items = q.items[:len(q.items)-1]
	return r
}

type sharedValue[T any] struct {
	v  T
	rw *sync.RWMutex
}

func newSharedValue[T any](init T) sharedValue[T] {
	return sharedValue[T]{
		v:  init,
		rw: &sync.RWMutex{},
	}
}

func (sv *sharedValue[T]) get() T {
	sv.rw.RLock()
	defer sv.rw.RUnlock()
	return sv.v
}

func (sv *sharedValue[T]) set(newVal T) {
	sv.rw.Lock()
	sv.v = newVal
	sv.rw.Unlock()
}
