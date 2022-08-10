package uniqueue_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/munnik/uniqueue"
)

func TestSinglePushAndSinglePop(t *testing.T) {
	queue := uniqueue.NewUQ[int](2)
	queue.Back() <- 1
	got := <-queue.Front()
	if got != 1 {
		t.Errorf("Expected 1, got %d", got)
	}
}

func TestMultiplePushAndSinglePop(t *testing.T) {
	var got, expected int

	queue := uniqueue.NewUQ[int](2)
	queue.Back() <- 1
	queue.Back() <- 2

	expected = 1
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}

func TestMultiplePushAndMultiplePop(t *testing.T) {
	var got, expected int

	queue := uniqueue.NewUQ[int](2)
	queue.Back() <- 1
	queue.Back() <- 2
	queue.Back() <- 3

	expected = 1
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	expected = 2
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	queue.Back() <- 4

	expected = 3
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	expected = 4
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}

func TestChannelClose(t *testing.T) {
	var got, expected int

	queue := uniqueue.NewUQ[int](5)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		expected = 1
		for got = range queue.Front() {
			if got != expected {
				t.Errorf("Expected %d, got %d", expected, got)
			}
			expected += 1
		}
		wg.Done()
	}()
	queue.Back() <- 1
	queue.Back() <- 2
	queue.Back() <- 3
	close(queue.Back())
	wg.Wait()
}

func TestChannelUniqueness(t *testing.T) {
	var got, expected int

	queue := uniqueue.NewUQ[int](2)
	go func() {
		queue.Back() <- 1
		queue.Back() <- 2
		queue.Back() <- 3
		queue.Back() <- 1
		queue.Back() <- 1
		queue.Back() <- 3
	}()

	expected = 1
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 2
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 3
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	if len(queue.Front()) != 0 {
		t.Errorf("Expected empty channel")
	}
}

func TestRemove(t *testing.T) {
	var got, expected int

	queue := uniqueue.NewUQ[int](2)
	go func() {
		queue.Back() <- 1
		queue.Back() <- 2
		queue.Back() <- 3
		queue.RemoveUnique(1)
		queue.Back() <- 1
		queue.Back() <- 1
		queue.Back() <- 3
	}()

	expected = 1
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 2
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 3
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
	expected = 1
	got = <-queue.Front()
	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}

	if len(queue.Front()) != 0 {
		t.Errorf("Expected empty channel")
	}
}

func Example() {
	queue := uniqueue.NewUQ[int](2)

	go func() {
		queue.Back() <- 1
		queue.Back() <- 2
		queue.Back() <- 3

		// these values should not be added because 1 is already on the queue
		queue.Back() <- 1
		queue.Back() <- 1

		// remove 3 from the unique restriction so it can be added again, once it's added again the unique restriction applies again
		queue.RemoveUnique(3)
		queue.Back() <- 3
		queue.Back() <- 3

		// close the back channel of the queue, this will automatically close the front channel
		close(queue.Back())
	}()

	// read the unique values from the queue, when the queue is closed this loop will terminate
	for value := range queue.Front() {
		fmt.Println(value)
	}

	// the front channel should be closed when the back channel is closed
	if _, ok := <-queue.Front(); !ok {
		fmt.Println("The channel is closed")
	}

	// Output:
	// 1
	// 2
	// 3
	// 3
	// The channel is closed
}
