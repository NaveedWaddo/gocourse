# 2.5 — Channels

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.4 — Goroutines

---

## 📖 What is a Channel?

A **channel** is a typed conduit through which goroutines communicate by sending and receiving values. Channels are Go's primary mechanism for safely passing data between goroutines — no locks needed.

```go
ch := make(chan int)   // create a channel that carries int values

go func() { ch <- 42 }()   // send 42 into the channel
value := <-ch               // receive from the channel
fmt.Println(value)          // 42
```

> **Go proverb:** *"Do not communicate by sharing memory; share memory by communicating."*

---

## 🏗️ Creating Channels

```go
// Unbuffered channel — synchronizes sender and receiver
ch := make(chan int)
ch := make(chan string)
ch := make(chan Person)

// Buffered channel — has a capacity queue
ch := make(chan int, 5)      // capacity of 5
ch := make(chan string, 100)
```

Zero value of a channel is `nil` — sending/receiving on nil channel blocks forever:

```go
var ch chan int   // nil channel
ch <- 1          // blocks forever
<-ch             // blocks forever
```

---

## ↔️ Unbuffered Channels — Synchronous

An unbuffered channel has **no storage** — the sender blocks until a receiver is ready, and vice versa. Both sides must be ready at the same time — this is called a **rendezvous**.

```go
ch := make(chan int)

go func() {
    fmt.Println("sending...")
    ch <- 99          // blocks here until main receives
    fmt.Println("sent!")
}()

time.Sleep(1 * time.Second)
fmt.Println("receiving...")
value := <-ch         // unblocks the goroutine above
fmt.Println("received:", value)
```

Output:
```
sending...
receiving...
sent!
received: 99
```

The sender and receiver **synchronize** at the channel — unbuffered channels are useful for signaling and guaranteeing handoff.

---

## 📦 Buffered Channels — Asynchronous

A buffered channel has an internal queue of fixed capacity. The sender only blocks when the buffer is **full**. The receiver only blocks when the buffer is **empty**.

```go
ch := make(chan int, 3)   // buffer capacity = 3

ch <- 1   // doesn't block — buffer has room
ch <- 2   // doesn't block
ch <- 3   // doesn't block
// ch <- 4   // would block — buffer full

fmt.Println(<-ch)   // 1
fmt.Println(<-ch)   // 2
fmt.Println(<-ch)   // 3
```

```
Buffer state visualization:

After 3 sends:  [ 1 | 2 | 3 ]   ← full, next send blocks
After 1 recv:   [ 2 | 3 |   ]   ← room for one more send
After 3 recvs:  [   |   |   ]   ← empty, next recv blocks
```

---

## ➡️ Directional Channels

You can restrict a channel to send-only or receive-only — used in function signatures to communicate intent and enforce safety:

```go
chan T       // bidirectional — can send and receive
chan<- T     // send-only
<-chan T     // receive-only
```

```go
// producer only sends — chan<- enforces this
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)
}

// consumer only receives — <-chan enforces this
func consumer(ch <-chan int) {
    for v := range ch {
        fmt.Println("received:", v)
    }
}

func main() {
    ch := make(chan int, 5)   // bidirectional — passed to both
    go producer(ch)
    consumer(ch)
}
```

Attempting to send on a receive-only channel (or vice versa) is a **compile error** — great for catching bugs early.

---

## 🔒 Closing Channels

Use `close(ch)` to signal that no more values will be sent:

```go
ch := make(chan int, 5)
ch <- 1
ch <- 2
ch <- 3
close(ch)

// Receive remaining values + detect close
for {
    v, ok := <-ch
    if !ok {
        fmt.Println("channel closed")
        break
    }
    fmt.Println(v)
}
```

`ok` is `false` when the channel is closed AND drained. Receiving from a closed, empty channel returns the zero value immediately.

### `for range` over a channel — idiomatic

```go
ch := make(chan int, 5)
go func() {
    for i := 1; i <= 5; i++ {
        ch <- i
    }
    close(ch)   // MUST close, or range loops forever
}()

for v := range ch {   // automatically stops when channel is closed
    fmt.Println(v)
}
```

> ⚠️ Only the **sender** should close a channel — never the receiver. Closing a channel twice panics.

### Rules about closing

```go
close(nil)       // ❌ PANIC — closing nil channel
close(closedCh)  // ❌ PANIC — closing already-closed channel
// Sending to a closed channel also panics:
closedCh <- 1    // ❌ PANIC
```

---

## 🔄 Channel as Signal (Done Pattern)

A common pattern — use an empty struct channel purely as a signal:

```go
done := make(chan struct{})   // struct{} takes 0 bytes

go func() {
    doWork()
    close(done)   // signal completion
}()

<-done   // wait for signal
fmt.Println("work complete")
```

`struct{}{}` as channel type — zero memory, pure signal. This is the idiomatic Go "done" channel.

---

## 🔁 Channel Direction in Pipelines

Channels connect goroutines into **pipelines** — data flows from one stage to the next:

```go
// Stage 1: generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: square each number
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func main() {
    // Pipeline: generate → square → print
    nums := generate(2, 3, 4, 5)
    squares := square(nums)

    for v := range squares {
        fmt.Println(v)   // 4, 9, 16, 25
    }
}
```

Each stage runs in its own goroutine — data flows through channels connecting them.

---

## 📊 Channel States Summary

| State | Send | Receive |
|-------|------|---------|
| nil channel | Blocks forever | Blocks forever |
| Open, empty (unbuffered) | Blocks until receiver ready | Blocks until sender ready |
| Open, empty (buffered) | Doesn't block (if cap > 0) | Blocks |
| Open, has data (buffered) | Blocks if full | Doesn't block |
| Closed, has data | PANIC | Returns data, ok=true |
| Closed, empty | PANIC | Returns zero, ok=false |

---

## 💻 Complete Example — Producer/Consumer Pipeline

```go
package main

import (
    "fmt"
    "sync"
)

// generate sends values into a channel
func generate(done <-chan struct{}, values ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, v := range values {
            select {
            case out <- v:
            case <-done:   // respect cancellation
                return
            }
        }
    }()
    return out
}

// double multiplies each value by 2
func double(done <-chan struct{}, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * 2:
            case <-done:
                return
            }
        }
    }()
    return out
}

// merge fans in multiple channels into one
func merge(done <-chan struct{}, channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    output := func(ch <-chan int) {
        defer wg.Done()
        for v := range ch {
            select {
            case out <- v:
            case <-done:
                return
            }
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go output(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
    done := make(chan struct{})
    defer close(done)

    // Pipeline: generate → double → print
    nums := generate(done, 1, 2, 3, 4, 5)
    doubled := double(done, nums)

    fmt.Println("Pipeline output:")
    for v := range doubled {
        fmt.Println(v)   // 2, 4, 6, 8, 10
    }

    // Buffered channel as a semaphore (limit concurrency)
    fmt.Println("\nSemaphore demo (max 2 concurrent):")
    sem := make(chan struct{}, 2)   // allow 2 concurrent operations
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            sem <- struct{}{}         // acquire
            defer func() { <-sem }() // release
            fmt.Printf("task %d running\n", id)
        }(i)
    }
    wg.Wait()
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Deadlock — all goroutines blocked
```go
// ❌ Deadlock — nobody receives from ch
ch := make(chan int)
ch <- 1   // blocks forever — fatal error: all goroutines are asleep
```

### 2. Forgetting to close causes range to hang
```go
ch := make(chan int)
go func() {
    ch <- 1
    ch <- 2
    // ❌ forgot close(ch) — range loops forever
}()
for v := range ch { fmt.Println(v) }
```

### 3. Closing from the receiver side
```go
// ❌ Only sender should close
go receiver(ch) {
    close(ch)   // dangerous — sender might send after this → panic
}
```

### 4. Sending to a closed channel panics
```go
close(ch)
ch <- 1   // ❌ PANIC: send on closed channel
```

### 5. Using channel where mutex is clearer
```go
// ❌ Overcomplicated — channel used just for a simple counter
ch := make(chan int, 1)
ch <- 0
count := <-ch
count++
ch <- count

// ✅ Just use atomic or mutex for simple shared state
var count int64
atomic.AddInt64(&count, 1)
```

---

## 🎯 Must-Know Things (Quick Reference)

- `make(chan T)` — unbuffered (synchronous, blocks both sides until ready)
- `make(chan T, n)` — buffered (async, blocks sender only when full)
- `chan<- T` send-only, `<-chan T` receive-only — use in function signatures
- `close(ch)` — signals no more values; only the **sender** should close
- `v, ok := <-ch` — `ok` is `false` when channel closed and drained
- `for v := range ch` — loops until channel closed; **must** close or it hangs
- nil channel blocks forever — both send and receive
- Sending to a closed channel **panics**
- Closing a nil or already-closed channel **panics**
- `chan struct{}` — zero-byte signal channel (done pattern)
- Buffered channel as semaphore — limit concurrency with `make(chan struct{}, n)`
- Deadlock — Go runtime detects and panics: "all goroutines are asleep"

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between a buffered and unbuffered channel?**

> An unbuffered channel has no storage — a send blocks until a receiver is ready and vice versa. Both sides synchronize at the channel (rendezvous). A buffered channel has an internal queue of fixed size — a send only blocks when the buffer is full, and a receive only blocks when the buffer is empty. Unbuffered channels are used for synchronization and guaranteed handoff; buffered channels are used for decoupling producer and consumer speeds or as semaphores.

---

**Q2: What happens when you send to or receive from a nil channel?**

> Both operations block forever. A nil channel is the zero value of a channel type — declared with `var ch chan int` without `make`. This behavior is actually useful in `select` statements: a nil channel case is never selected, so you can disable a case by setting it to nil.

---

**Q3: Who should close a channel and why?**

> Only the sender should close a channel, never the receiver. Sending to a closed channel panics, so the receiver has no way to know if the sender will send again. The sender knows it's done sending, so it's the only safe party to close. When multiple goroutines send to the same channel, use `sync.WaitGroup` to coordinate and have a separate goroutine close after all senders finish. Closing from the receiver side risks a panic if a sender sends after the close.

---

**Q4: What does `v, ok := <-ch` tell you?**

> The comma-ok idiom on a channel receive returns the value and a boolean. `ok` is `true` if the value was successfully received from an open channel. `ok` is `false` if the channel is closed and drained — in this case `v` is the zero value of the channel's type. This is how you distinguish "received zero value" from "channel is closed". `for range` uses this internally to stop iteration when the channel closes.

---

**Q5: What is a deadlock in Go and how does the runtime handle it?**

> A deadlock occurs when all goroutines are blocked — typically waiting on channels or mutexes that will never be unblocked. Go's runtime detects when ALL goroutines are asleep simultaneously and panics with "fatal error: all goroutines are asleep - deadlock!". Note it only detects total deadlock — partial deadlocks (some goroutines blocked, some running) are not detected and manifest as goroutine leaks instead.

---

**Q6: How do you use a buffered channel as a semaphore?**

> Create a buffered channel with capacity equal to the maximum concurrency you want: `sem := make(chan struct{}, n)`. To acquire, send into it: `sem <- struct{}{}` — this blocks if n goroutines are already running. To release, receive from it: `<-sem`. This limits the number of goroutines running a section simultaneously to n. It's a clean, idiomatic pattern for rate-limiting goroutines — for example, limiting parallel HTTP requests or database connections.

---

**Q7: What is the pipeline pattern with channels?**

> A pipeline is a series of stages connected by channels, where each stage is a goroutine that reads from an input channel, transforms the data, and sends to an output channel. Each stage function takes a `<-chan` input and returns a `<-chan` output, running its processing goroutine internally. The stages run concurrently — while stage 2 processes item N, stage 1 is already working on item N+1. This is a natural fit for stream processing, ETL, and data transformation workloads.

---

*Next up: **Topic 2.6 — Select Statement** 🚀*
