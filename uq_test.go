package uniqueue_test

import (
	"fmt"
	"sync"
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

	uq := uniqueue.NewUQ[int](2)
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
	go func() {
		expected = 1
		for got = range uq.Front() {
			if got != expected {
				t.Errorf("Expected %d, got %d", expected, got)
			}
			expected += 1
		}
		wg.Done()
	}()
	uq.Back() <- 1
	uq.Back() <- 2
	uq.Back() <- 3
	close(uq.Back())
	wg.Wait()
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

func TestRemove(t *testing.T) {
	var got, expected int

	uq := uniqueue.NewUQ[int](2)
	go func() {
		uq.Back() <- 1
		uq.Back() <- 2
		uq.Back() <- 3
		uq.RemoveUnique(1)
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

func Example() {
	uq := uniqueue.NewUQ[int](2)

	go func() {
		uq.Back() <- 1
		uq.Back() <- 2
		uq.Back() <- 3

		// these values should not be added because 1 is already on the queue
		uq.Back() <- 1
		uq.Back() <- 1

		// remove 3 from the unique restriction so it can be added again, once it's added again the unique restriction applies again
		uq.RemoveUnique(3)
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
