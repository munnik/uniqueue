# uniqueue
Create a queue (go chan) that only contains unique elements. This can be useful when this queue is feeding workers and you want to make sure that the workers don't end up doing double work.

## Example
```go
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
```
