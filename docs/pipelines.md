# Pipelines

Ther's no formal definition of a pipeline in Go. It's just another pattern to
pass data through multiple `goroutines`. Ussually, they have one goroutine at
the start and one at the end, that would read, and write the processed data
respectably. In the middle, we have multiple other `goroutines` that handles
the streamed data in some way. 

The first stage is usually called a _source_ or _producer_ while the last one is
known as a _sink_ or _consumer_.

## Example

First, let's write a _producer_ function that will take a number of `ints` as
arguments, that will be piped to a newly created `chan int`.

```go
// gen is a producer of int values.
func gen(nums ...int) <-chan int {
  out := make(chan int)
  go func(){
    for _, n := range nums {
      out <- n
    }
    close(out)
  }()
  return out
}
```

Then we create the `sq` function that will act as the _consumer_ of the `gen` 
producer. It will square each received `int` and pipe the result into a new
`chan int`.

```go
func sq(in <-chan int) <-chan int {
  out := make(chan int)
  go func() {
    for n := range in {
      out <- n * n
    }
    close(out)
  }()
  return out
}
```

The `main` function will connect each stage and use the resulting data.

```go
func main() {
  // Set up the pipeline
  c := gen(2, 3)
  out := sq(c)

  // Consume the output
  fmt.Println(<-out) // 4
  fmt.Println(<-out) // 9
}
```

Since the types of the `sq` function matches, we can combine them as many
times as we want. Also, we can rewrite the `main` function to use a `for 
range` as the other functions.

```go
func main() {
  for n := range sq(sq(gen(2, 3))) {
    fmt.Println(n) // 16, then 81
  }
}
```

Multiple functions can read from the same channel. This is a good way of
paralellising processes. Bare in mind that each value inside a channel will only
be consumed by one `goroutine` listening to the channel. Meaning, if we have two
functions listening to the same channel, they will receive the values one after
the other. Channels works as FIFO queues.

We can then gather the outputs from each process and combine them into another
channel. The order of the results of this new _merged_ channel will not 
necessarily be equal to the order of the _producer_ channel.

```go
func main() {
  in := gen(2, 3)

  // Distribute the sq work across two goroutines that both read from in.
  c1 := sq(in)
  c2 := sq(in)

  // Consume the merged output from c1 and c2.
  for n := range merge(c1, c2) {
    fmt.Println(n) // 4, then 9, or 9 then 4.
  }
}
```

```go
func merge(cs ...<-chan int) <- chan int {
  var wg sync.WaitGroup
  out := make(chan int)

  // Start an output goroutine for each input channel in cs. output copies
  // values froom c to out until c is closed, then calls wg.Done.
  output := func(c <-chan int) {
    for := range c {
      out <- n
    }
    wg.Done()
  }
  wg.Add(len(cs))
  for _, c := range cs {
    go output(c)
  }

  // Start a goroutine to close out once all the output goroutines are done.
  // This must start after the wg.Add call.
  go func() {
    wg.Wait()
    close(out)
  }()
  return out
}
```

Important:

1. All stages close their outbound channels when all the send operations are
done.
2. All stages keep receiving values from inbound channels until those channels
are closed.

But what if we need to exit from a stage earlier, say, because of an error. Our
current setup will keep blocking until the inbound channel stops sending all its
values.

**Goroutines are not garbage collected; they must exit on their own.**

If we know the number of values that should exist on the channel we can use
`buffers` instead. `buffers` send operation completes inmediately if ther is
space in the buffer.

We can rewrite our _producer_ function taking advantage of the fact that we know
in advanced the number of `ints` that should be piped.

```go
func gen(nums ...int) <-chan int {
  out := make(chan int, len(nums))
  for _, n := range nums {
    out <- n
  }
  close(out)
  return out
}
```

We need a way to tell an unkown and unbounded number of `goroutines` to stop
sending their values downstream. In Go, we can do this by closing a channel,
because _a receive operation on a closed channel can always proceed immediately,
yielding the element type's zero value_.

A recieve expression in an assignment or initialization of the special form:

```go
x, ok = <-ch
x, ok := <-ch
var x, ok = <-ch
var x, ok T = <-ch
```

yields an additional untyped boolean result reporting whether the communication
suceeded. The value of `ok` is `true` if the value received was delivered by a
successful send operaio to the channel, or `false` if it is a zero value
generated because the channel is closed and empty.

So, we can use `done` channel to close all the stages from the `main` function.

```go
func main() {
  // Set up a done channel that's shared by the whole pipeline, and close that 
  // channel when this pipeline exits, as a signal for all the goroutines we 
  // started to exit.
  done := make(chan struct{})
  defer close(done)

  in := gen(done, 2, 3)

  // Distribute the sq work across two goroutines that both read from in.
  c1 := sq(done, in)
  c2 := sq(done, in)
  
  // Consume the first value from output.
  out := merge(done, c1, c2)
  fmt.Println(<-out) // 4 or 9

  // done will be closed by the deferred call.
}
```

All the stage functions must be able to receive the `done` channel to close
their respective channels, and stop consuming values from upstream. And doing it
will not block the upstream stages, since they also should have received the
`done` message.

```go
func merge(done <-chan struct{}, cs ...<-chan int) <-chan int {
  var wg sync.WaitGroup
  out := make(chan int)

  // Start an output goroutine for each input channel in cs. output copies
  // values from c to out until c or done is closed, then calls wg.Done.
  output := func(c <-chan int) {
    defer wg.Done()
    for n := range c {
      select {
        case out <- n:
        case <- done:
          return
      }
    }
  }
  // The rest is unchanged.
}

func sq(done <-chan struct{}, in <-chan int) <-chan int {
  out := make(chan int)
  go func() {
    defer close(out)
    for n := range in {
      select {
        case out <- n * n:
        case <-done:
          return
      }
    }
  }()
  return out
}
```

Guidelines:

1. Stages close their outbound channels when all the send operations are done.
2. Stages keep receiving values from inbound channels until those channels are
closed or the senders are unblocked.

Pipelines unblock senders either by ensuring there's enough buffer for all the 
values that are sent or by explicitly signalling senders when the receiver may
abandon the channel.

















