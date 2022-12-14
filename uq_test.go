package uniqueue_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/munnik/uniqueue"
)

func TestSinglePushAndSinglePop(t *testing.T) {
	uq := uniqueue.NewUQ[int](2)
	uq.Back() <- 1
	got := <-uq.Front()
	if got != 1 {
		t.Errorf("Expected 1, got %d", got)
	}
}

func TestMultiplePushAndSinglePop(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](2)
	uq.Back() <- 1
	uq.Back() <- 2

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}

func TestMultiplePushAndMultiplePop(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](3)
	uq.Back() <- 1
	uq.Back() <- 2
	uq.Back() <- 3

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	expected = 2
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	uq.Back() <- 4

	expected = 3
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	expected = 4
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}

func TestChannelClose(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](5)

	wg := sync.WaitGroup{}
	wg.Add(1)

	var sum int32
	go func() {
		expected = 1
		for got = range uq.Front() {
			if got != expected {
				t.Errorf("Expected %d, got %d", expected, got)
			}
			atomic.AddInt32(&sum, int32(expected))
			expected += 1
		}
		wg.Done()
	}()

	uq.Back() <- 1
	uq.Back() <- 2
	uq.Back() <- 3
	close(uq.Back())

	wg.Wait()

	if sum != 6 {
		t.Errorf("Expected the sum to be 6, got %d", got)
	}
}

func TestChannelUniqueness(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](2)
	go func() {
		uq.Back() <- 1
		uq.Back() <- 2
		uq.Back() <- 3
		uq.Back() <- 1
		uq.Back() <- 1
		uq.Back() <- 3
	}()

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 2
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 3
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	if len(uq.Front()) != 0 {
		t.Errorf("Expected empty channel")
	}
}

func TestIgnoreConstraint(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](2)
	go func() {
		uq.Back() <- 1
		uq.Back() <- 2
		uq.Back() <- 3
		uq.IgnoreConstraintFor(1)
		uq.Back() <- 1
		uq.Back() <- 1
		uq.Back() <- 3
	}()

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 2
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 3
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	if len(uq.Front()) != 0 {
		t.Errorf("Expected empty channel")
	}
}

func TestAutoRemoveConstraint(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](6)
	uq.AutoRemoveConstraint = true
	uq.Back() <- 1
	uq.Back() <- 2
	// this value should not be added because 1 is still on the queue
	uq.Back() <- 1

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	// now 1 can be added to the queue again because the value is not on the queue anymore
	uq.Back() <- 1

	expected = 2
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	expected = 1
	got = <-uq.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}

func Example() {
	uq := uniqueue.NewUQ[int](2)

	go func() {
		uq.Back() <- 1
		uq.Back() <- 2
		uq.Back() <- 3

		// these values should not be added because 1 is already on the queue
		uq.Back() <- 1
		uq.Back() <- 1

		// ignore 3 from the unique constraint so it can be added again, once it's added again the unique constraint is enabled again
		uq.IgnoreConstraintFor(3)
		uq.Back() <- 3
		uq.Back() <- 3

		// close the back channel of the queue, this will automatically close the front channel
		close(uq.Back())
	}()

	// read the unique values from the queue, when the queue is closed this loop will terminate
	for value := range uq.Front() {
		fmt.Println(value)
	}

	// the front channel should be closed when the back channel is closed
	if _, ok := <-uq.Front(); !ok {
		fmt.Println("The channel is closed")
	}

	// Output:
	// 1
	// 2
	// 3
	// 3
	// The channel is closed
}
