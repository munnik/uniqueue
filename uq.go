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

func (u *UQ[T]) AddConstraint(value T) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.constraints[value]; !ok {
		u.constraints[value] = struct{}{}
		return nil
	}
	return errors.New("Constraint already exists")
}

func (u *UQ[T]) RemoveConstraint(value T) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.constraints, value)
}

func (u *UQ[T]) linkChannels() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go u.shiftToFront(wg)
	go u.readFromBack(wg)

	wg.Wait()
}

func (u *UQ[T]) shiftToFront(wg *sync.WaitGroup) {
	for value := range u.queue {
		u.front <- value
		if u.AutoRemoveConstraint {
			u.RemoveConstraint(value)
		}
	}

	close(u.front)

	wg.Done()
}

func (u *UQ[T]) readFromBack(wg *sync.WaitGroup) {
	for value := range u.back {
		if err := u.AddConstraint(value); err == nil {
			u.queue <- value
		}
	}

	close(u.queue)

	wg.Done()
}
