# Contexts

Carries deadlines, cancellation signals, and other request-scoped values across
API boundaries and between processes.

Contexts should be passed as arguments, and shouldn't be stored inside structs.
They can also be replaced with a derived context using `WithCancel`,
`WithTimeout`, `WithDeadline`, or `WithValue`. When a context is canceled, all 
Contexts derived from it are also canceled.

These functions take in a parent context, and return a child context, plus a 
`cancel` function. Calling the `cancel` function will:

1. Cancel all `children` contexts.
2. Remove the `children` reference from the parent.
3. Stop all associated timers.

## Best practises

1. Don't store contexts in structs, pass it as a variable:
  ```go
  func DoSomething(ctx context.Context, arg Arg) error {
    // use ctx
  }
  ```
2. Do not pass a `nil` context. Use the `context.TODO` context.
3. Use context values only for request-scoped data that transits processes and
APIs, not for passing optional parameters to functions.
4. The same context may be passed to functions running in different goroutines.

## Context package

The `context.Context` type definition looks like this:

```go
type Context interface {
  // Done returns a channel that is closed when this Context is canceled or
  // times out.
  Done() <-chan struct{}

  // Err indicates why this context was canceled, after the Done channel is
  // closed.
  Err () error

  // Deadline returns the time when this Context will be canceled, if any.
  Deadline() (deadline time.Time, ok bool)

  // Value returns the value associated with key or nil if none.
  Value(key interface{}) interface{}
}
```

Some interesting thoughts:

- The `Context` doesn't have a `Cancel` method because the function receiving
a cancelation signal is usually not the ones that sends the signal.
- The `Deadline` method allows other goroutines to check if they should start
any work at all.
- The data retrieved through the `Value` method should be safe to use by
multiple processes.

The `Done` method is provided to be used in `select` statements:

```go
func Stream(ctx context.Context, out chan<- Value) error {
  for {
    v, err := DoSomething(ctx)
    if err := nil {
      return err
    }
    select {
      case <- ctx.Done():
        return ctx.Err()
      case out <- v:
    }
  }
}
```

The `Err` message returns `nil` if the context has not been closed yet. If the
context was closed it returns a non-nil error explaining why the context was
closed:

- `Canceled` if the context was canceled.
- `DeadlineExeeded` if the context's deadline passed.

A key identifies a value inside a Context. Functions that wish to store values
in Context typically allocate a key in a global variable then use that key as
the argument to `context.WithValue` and `Context.Value`. 

Packages that define a Context key should provide type-safe accessors for the
values stored using that key:

```go
package user

import "context"

// User is the value stored in Context.
type User struct {...}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// userKey is the key for user.User values in Contexts. It is unexported;
// clients use user.NewContext and user.FromContext instead of using this key
// directly.
var userKey key

// NewContext returns a new Context that carries value u.
func NewContext(ctx context.Context, u *User) context.Context {
  return context.WithValue(ctx, userKey, u)
}

// FromContext returns the User value stored in ctx, if any.
func FromContext(ctx context.Context) (*User, bool) {
  u, ok := ctx.Value(userKey).(*User)
  return u, ok
}
```

## Derived Contexts

Derivded contexts are stored like a tree, where `context.Background` acts always
as the root. The `Background` context is never canceled, has no deadlines, and
has no values. 

```go
// WithCancel returns a copy of parent whose Done channel is closed as soon as
// parent.Done is closed or cancel is called.
// Canceled is an error exported by the context.Context package.
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
  if parent == nil {
    panic("cannot create context from nil parent")
  }
  c := newCancelCtx(parent)
  propagateCancel(parent, &c)
  return &c, func() { c.cancel(true, Canceled) }
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
  return cancelCtx{Context: parent}
}

// goroutines counts the number of goroutines ever created; for testing.
var goroutines int32

// propageteCancel arranges for child to be cancelled when parent is.
func propagateCancel(parent Context, child canceler) {
  done := parent.Done()
  if done == nil {
    return // parent is never canceled
  }
  
  // Check if the parent Context is already done.
  select {
    case <- done:
      // parent is already canceled
      child.cancel(false, parent.Err())
      return
    default:
  }

  if p, ok := parentCancelCtx(parent); ok {
    p.mu.Lock()
    if p.err != nil {
      // parent has already been canceled
      child.cancel(false, p.err)
    } else {
      if p.children == nil {
        p.children = make(map[canceler]struct{})
      }
      p.children[child] = struct{}{}
    }
    p.mu.Unlock()
  } else {
    atomic.AddInt32(&goroutines, +1)
    go func() {
      select {
        case <- parent.Done()
          child.cancel(false, parent.Err())
        case <- child.Done():
      }
    }()
  }
}

// &cancelCtxKey is the key that a cancelCtx returns itself for.
var cancelCtxKey int

// parentCancelCtx returns the underlying *cancelCtx for parent.
func parentCancelCtx(parent Context) (*cancelCtx, bool) {
  done := parent.Done()
  if done == closedchan || done == nil {
    return nil, false
  }
  p, ok := parent.Value(&cancelCtxKey).(*cancelCtx)
  if !ok {
    return nil, false
  }
  p.mu.Lock()
  ok = p.done == done
  p.mu.Unlock()
  if !ok {
    return nil, false
  }
  return p, true
}

// A canceler is a context type that can be canceled directly. The
// implementatations are *cancelCtx and *timerCtx.
type canceler interface {
  cancel(removeFromParent bool, err error)
  Done() <- chan struct{}
}

// closedchan is a reusable closed channel
var closedchan = make(chan struct{})

func init() {
  close(closedchan)
}

// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
type cancelCtx struct {
  Context

  mu        sync.Mutex            // protects following fields
  done      chan sturct{}         // created lazily, closed by first cancel call
  children  map[canceler]struct{} // set to nil by the first cancel call
  err       error                 // set to non-nil by the first cancel call
}

func (c *cancelCtx) Value(key interface{}) interface{} {
  if key == &cancelCtxKey {
    return c
  }
  return c.Context.Value(key)
}

func (c *cancelCtx) Done() <-chan struct{} {
  c.mu.Lock()
  if c.done == nil {
    c.done = make(chan struct{})
  }
  d := c.done
  c.mu.Unlock()
  return d
}

func (c *cancelCtx) Err() error {
  c.mu.Lock()
  err := c.err
  c.mu.Unlock()
  return err
}
```

## Errors

```go
// Canceled is the error returned by Context.Err when the context is canceled.
var Canceled = errors.New("context canceled")

// DeadlineExceeded is the error returned by Context.Err when the context's 
// deadline passes.
var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceedError struct {}

func (deadlineExceededError) Error() string { 
  return "context deadline exceeded" 
}
func (deadlineExceededError) Timeout() bool { return true }
func (deadlineExceededError) Temporary() bool { return true }
```






















