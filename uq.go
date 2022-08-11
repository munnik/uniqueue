package uniqueue

import (
	"errors"
	"sync"
)

type UQ[T comparable] struct {
	back                 chan T
	queue                chan T
	front                chan T
	constraints          map[T]struct{}
	mu                   sync.Mutex
	AutoRemoveConstraint bool
}

func NewUQ[T comparable](size uint) *UQ[T] {
	u := &UQ[T]{
		back:        make(chan T),
		queue:       make(chan T, size),
		front:       make(chan T),
		constraints: map[T]struct{}{},
	}

	go u.linkChannels()

	return u
}

func (u *UQ[T]) Back() chan<- T {
	return u.back
}

func (u *UQ[T]) Front() <-chan T {
	return u.front
}

func (u *UQ[T]) AddConstraint(v T) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.constraints[v]; !ok {
		u.constraints[v] = struct{}{}
		return nil
	}
	return errors.New("Constraint already exists")
}

func (u *UQ[T]) RemoveConstraint(v T) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.constraints, v)
}

func (u *UQ[T]) linkChannels() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go u.shiftToFront(wg)
	go u.readFromBack(wg)

	wg.Wait()
}

func (u *UQ[T]) shiftToFront(wg *sync.WaitGroup) {
	for v := range u.queue {
		u.front <- v
		if u.AutoRemoveConstraint {
			u.RemoveConstraint(v)
		}
	}

	close(u.front)

	wg.Done()
}

func (u *UQ[T]) readFromBack(wg *sync.WaitGroup) {
	for v := range u.back {
		if err := u.AddConstraint(v); err == nil {
			u.queue <- v
		}
	}

	close(u.queue)

	wg.Done()
}
