# uniqueue

Create a queue (go chan) that only contains unique elements (no duplicates). This can be useful when this queue is feeding workers and you want to make sure that the workers don't end up doing double work. It is a very simple queue, with 1 constructor 1 property and 4 methods.

`uq := uniqueue.NewUQ[T](size int)` constructs an uniqueue of type `T` with a buffer of size `size`. `T` must be `comparable`. This function returns a pointer to the uniqueue.

`uq.AutoRemoveConstraint` is a `bool` property that when enabled automatically removes the constraint for a value `v` when it is popped from the queue.

`uq.Back()` returns a _write only_ channel that can be used to push new values on the queue.

`uq.Front()` returns a _read only_ channel that can be used to pop values from the queue.

`uq.AddConstraint(v T)` adds the unique constraint for the value `v` from the queue, this can be used to make sure that certain values never get on the queue.

`uq.RemoveConstraint(v T)` removes the unique constraint for the value `v` from the queue, if the value `v` is on the queue it will remain on the queue until it is popped. A new value `w`, where `v == w`, can be added to the queue. After that the unique constraint is applied again. If you want to add the value `x`, where `x == w` you need to call this method again.

## Example

```go
uq := uniqueue.NewUQ[int](2)

go func() {
  uq.Back() <- 1
  uq.Back() <- 2
  uq.Back() <- 3

  // these values should not be added because 1 is already on the queue
  uq.Back() <- 1
  uq.Back() <- 1

  // remove 3 from the unique constraint so it can be added again, once it's added again the unique constraint applies again
  uq.RemoveConstraint(3)
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
```
