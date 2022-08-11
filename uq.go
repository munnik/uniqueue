package uniqueue

import (
	"errors"
	"sync"
)

// UQ is a uniqueue queue. It guarantees that a value is only once in the queue. The queue is thread safe.
// The unique constraint can be temporarily disabled to add multiple instances of the same value to the queue.
type UQ[T comparable] struct {
	back                 chan T
	queue                chan T
	front                chan T
	constraints          map[T]*constraint
	mu                   sync.Mutex
	AutoRemoveConstraint bool // if true, the constraint will be removed when the value is popped from the queue.
}

type constraint struct {
	count    uint // number of elements in the queue
	disabled bool
}

func NewUQ[T comparable](size uint) *UQ[T] {
	u := &UQ[T]{
		back:        make(chan T),
		queue:       make(chan T, size),
		front:       make(chan T),
		constraints: map[T]*constraint{},
	}

	go u.linkChannels()

	return u
}

// Get the back of the queue, this channel can be used to write values to.
func (u *UQ[T]) Back() chan<- T {
	return u.back
}

// Get the front of the queue, this channel can be used to read values from.
func (u *UQ[T]) Front() <-chan T {
	return u.front
}

// Ignores the constraint for a value v once, when the value is added to the queue again, the constraint is enabled again.
func (u *UQ[T]) IgnoreConstraintFor(v T) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.constraints[v]; !ok {
		u.constraints[v] = &constraint{}
	}

	u.constraints[v].disabled = true
}

// Manually add a constraint to the queue, only use in special cases when you want to prevent certain values to enter the queue.
func (u *UQ[T]) AddConstraint(v T) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.constraints[v]; !ok {
		u.constraints[v] = &constraint{
			count:    1,
			disabled: false,
		}
		return nil
	} else {
		if u.constraints[v].disabled {
			u.constraints[v].count += 1
			u.constraints[v].disabled = false
			return nil
		}
	}
	return errors.New("Already existing constraint prevents adding new constraint")
}

// Manually remove a constraint from the queue, this needs to be called when AutoRemoveConstraint is set to false. Useful when you want to remove the constraint only when a worker using the queue is finished processing the value.
func (u *UQ[T]) RemoveConstraint(v T) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.constraints[v]; ok {
		u.constraints[v].count -= 1
		if u.constraints[v].count == 0 {
			delete(u.constraints, v)
		}
	}
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
