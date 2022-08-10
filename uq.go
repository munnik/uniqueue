package uniqueue

import "sync"

type UQ[T comparable] struct {
	back         chan T
	front        chan T
	uniqueValues map[T]struct{}
	mu           sync.Mutex
}

func NewUQ[T comparable](size int) *UQ[T] {
	u := &UQ[T]{
		back:         make(chan T, size),
		front:        make(chan T, size),
		uniqueValues: map[T]struct{}{},
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

func (u *UQ[T]) RemoveUnique(value T) {
	u.mu.Lock()
	delete(u.uniqueValues, value)
	u.mu.Unlock()
}

func (u *UQ[T]) linkChannels() {

	for value := range u.back {
		u.mu.Lock()
		if _, ok := u.uniqueValues[value]; !ok {
			u.uniqueValues[value] = struct{}{}
			u.front <- value
		}
		u.mu.Unlock()
	}

	close(u.front)
}
