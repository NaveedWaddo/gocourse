# 2.6 — Select Statement

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.5 — Channels

---

## 📖 What is `select`?

`select` is like a `switch` but for **channel operations**. It waits on multiple channel sends/receives simultaneously and executes whichever one is ready first. If multiple are ready, it picks one **at random**.

```go
select {
case v := <-ch1:
    fmt.Println("received from ch1:", v)
case v := <-ch2:
    fmt.Println("received from ch2:", v)
case ch3 <- 99:
    fmt.Println("sent to ch3")
}
```

Without `select`, you'd have to pick one channel to wait on — blocking all others. `select` lets you wait on all of them at once.

---

## 🧩 Basic Syntax

```go
select {
case <-ch1:
    // ch1 had a value
case msg := <-ch2:
    // received msg from ch2
case ch3 <- value:
    // sent value to ch3
case ch4 <- compute():
    // compute() called only if this case is selected
default:
    // runs immediately if no channel is ready
}
```

Key rules:
- Each `case` must be a channel send or receive
- If **no case is ready** and there's no `default`, `select` **blocks**
- If **multiple cases are ready**, one is chosen **uniformly at random**
- If there's a **`default`**, it runs immediately when no other case is ready

---

## 🔀 Random Selection When Multiple Cases Ready

```go
ch1 := make(chan string, 1)
ch2 := make(chan string, 1)

ch1 <- "one"
ch2 <- "two"

// Both ready — Go picks randomly
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
}
// Could print either "ch1: one" or "ch2: two"
```

This randomness is **intentional** — it prevents starvation where one channel always wins.

---

## ⚡ `default` — Non-Blocking Operations

Adding `default` makes `select` non-blocking — it doesn't wait at all:

```go
ch := make(chan int, 1)

// Non-blocking receive
select {
case v := <-ch:
    fmt.Println("received:", v)
default:
    fmt.Println("no value ready")
}

// Non-blocking send
select {
case ch <- 42:
    fmt.Println("sent")
default:
    fmt.Println("channel full, dropped")
}
```

### Non-blocking check pattern

```go
func tryReceive(ch <-chan int) (int, bool) {
    select {
    case v := <-ch:
        return v, true
    default:
        return 0, false
    }
}
```

---

## ⏱️ Timeout Pattern

One of the most important `select` patterns — add a timeout to any blocking operation:

```go
import "time"

ch := make(chan string)

select {
case result := <-ch:
    fmt.Println("got result:", result)
case <-time.After(2 * time.Second):
    fmt.Println("timed out after 2s")
}
```

`time.After(d)` returns a `<-chan Time` that receives a value after duration `d`. When the timeout fires, its case is selected.

### Timeout on the whole operation

```go
func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
    resultCh := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := fetch(url)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-time.After(timeout):
        return "", fmt.Errorf("fetch timed out after %v", timeout)
    }
}
```

---

## 🛑 Done Channel — Cancellation Pattern

Using `select` to respect a cancellation signal while doing work:

```go
func worker(done <-chan struct{}, jobs <-chan int) {
    for {
        select {
        case <-done:
            fmt.Println("worker cancelled")
            return
        case job, ok := <-jobs:
            if !ok {
                fmt.Println("jobs channel closed")
                return
            }
            processJob(job)
        }
    }
}

func main() {
    done := make(chan struct{})
    jobs := make(chan int, 10)

    go worker(done, jobs)

    for i := 0; i < 5; i++ {
        jobs <- i
    }

    time.Sleep(100 * time.Millisecond)
    close(done)   // cancel the worker
}
```

---

## 🔁 `select` in a Loop — Event Loop Pattern

Combine `select` with `for` to handle multiple channels continuously:

```go
func eventLoop(
    ch1 <-chan string,
    ch2 <-chan int,
    done <-chan struct{},
) {
    for {
        select {
        case msg := <-ch1:
            fmt.Println("string message:", msg)

        case num := <-ch2:
            fmt.Println("number:", num)

        case <-done:
            fmt.Println("shutting down event loop")
            return
        }
    }
}
```

This is the backbone of many Go servers, event processors, and state machines.

---

## 🚦 Ticker — Repeated Timeouts

`time.Ticker` sends on a channel repeatedly at a fixed interval:

```go
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()   // always stop to free resources

done := make(chan struct{})
go func() {
    time.Sleep(2 * time.Second)
    close(done)
}()

for {
    select {
    case t := <-ticker.C:
        fmt.Println("tick at", t.Format("15:04:05.000"))
    case <-done:
        fmt.Println("ticker stopped")
        return
    }
}
```

> Always call `ticker.Stop()` — unlike `time.After`, a ticker keeps running and its goroutine leaks if not stopped.

---

## 🔕 Disabling a Case with nil

Setting a channel to `nil` disables that case in `select` — a nil channel blocks forever so it's never selected:

```go
ch1 := make(chan int, 1)
ch2 := make(chan int, 1)

ch1 <- 1
ch2 <- 2

// Disable ch1 after first receive
for i := 0; i < 2; i++ {
    select {
    case v, ok := <-ch1:
        if !ok {
            ch1 = nil   // disable this case
        } else {
            fmt.Println("ch1:", v)
        }
    case v := <-ch2:
        fmt.Println("ch2:", v)
    }
}
```

This pattern is useful in fan-in scenarios where you want to stop reading from a channel once it's exhausted.

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

// Simulates a slow service
func slowService(name string, delay time.Duration) <-chan string {
    ch := make(chan string, 1)
    go func() {
        time.Sleep(delay)
        ch <- fmt.Sprintf("response from %s", name)
    }()
    return ch
}

// First response wins — others are discarded
func fastest(services ...<-chan string) string {
    result := make(chan string, len(services))
    for _, svc := range services {
        svc := svc
        go func() {
            result <- <-svc
        }()
    }
    return <-result
}

func main() {
    // --- Timeout demo ---
    fmt.Println("=== Timeout Demo ===")
    slow := slowService("slow-api", 3*time.Second)
    fast := slowService("fast-api", 100*time.Millisecond)

    select {
    case r := <-fast:
        fmt.Println("fast:", r)
    case r := <-slow:
        fmt.Println("slow:", r)
    case <-time.After(500 * time.Millisecond):
        fmt.Println("timed out!")
    }

    // --- Ticker with done ---
    fmt.Println("\n=== Ticker Demo ===")
    ticker := time.NewTicker(200 * time.Millisecond)
    done := make(chan struct{})

    go func() {
        time.Sleep(700 * time.Millisecond)
        close(done)
    }()

    for {
        select {
        case <-ticker.C:
            fmt.Println("tick:", time.Now().Format("15:04:05.000"))
        case <-done:
            ticker.Stop()
            fmt.Println("done")
            goto next
        }
    }

next:
    // --- Non-blocking send ---
    fmt.Println("\n=== Non-blocking Send ===")
    ch := make(chan int, 2)
    for i := 1; i <= 4; i++ {
        select {
        case ch <- i:
            fmt.Printf("sent %d\n", i)
        default:
            fmt.Printf("dropped %d (channel full)\n", i)
        }
    }

    // --- Random selection ---
    fmt.Println("\n=== Random Selection ===")
    c1 := make(chan string, 1)
    c2 := make(chan string, 1)
    c1 <- "alpha"
    c2 <- "beta"

    rand.Seed(time.Now().UnixNano())
    for i := 0; i < 4; i++ {
        c1 <- "alpha"
        c2 <- "beta"
        select {
        case v := <-c1:
            fmt.Println("c1:", v)
        case v := <-c2:
            fmt.Println("c2:", v)
        }
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. `select` with no cases blocks forever
```go
select {}   // blocks forever — useful for keeping main alive, but document it!
```

### 2. `time.After` leaks in tight loops
```go
// ❌ Each iteration creates a new timer that's never GC'd until it fires
for {
    select {
    case <-time.After(1 * time.Second):
        doWork()
    }
}

// ✅ Use time.NewTicker or create timer once outside the loop
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()
for {
    select {
    case <-ticker.C:
        doWork()
    }
}
```

### 3. Assuming case order matters
```go
// ❌ Don't assume ch1 is always checked first
select {
case v := <-ch1:   // NOT guaranteed priority
case v := <-ch2:
}

// ✅ If priority matters, check explicitly first
select {
case v := <-highPriority:
    handle(v)
default:
}
select {
case v := <-highPriority:
    handle(v)
case v := <-lowPriority:
    handle(v)
}
```

### 4. Forgetting `default` in a polling loop causes high CPU
```go
// ❌ Without default, this is not polling — it blocks!
for {
    select {
    case v := <-ch:
        process(v)
    }
    // This is just a for-range equivalent, not polling
}

// ✅ With default, this actively polls (but burns CPU — add sleep)
for {
    select {
    case v := <-ch:
        process(v)
    default:
        time.Sleep(10 * time.Millisecond)
    }
}
```

### 5. nil channel in select blocks that case — not the whole select
```go
var ch chan int = nil
select {
case v := <-ch:        // never selected — nil blocks forever
    fmt.Println(v)
case <-time.After(1 * time.Second):
    fmt.Println("timeout")   // this runs
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- `select` waits on multiple channel operations simultaneously
- If **multiple cases ready** — one chosen **at random** (prevents starvation)
- If **no case ready and no default** — `select` **blocks**
- `default` makes `select` non-blocking — runs immediately if no channel ready
- `time.After(d)` — one-shot timeout channel, use for single timeouts
- `time.NewTicker(d)` — repeated ticks, always call `Stop()` to prevent leak
- `select {}` (empty) — blocks forever, use to keep main alive
- Setting a channel to `nil` **disables** that case in select
- `time.After` in a tight loop **leaks** — use `time.NewTicker` instead
- Combine `for` + `select` + `done` channel = idiomatic event loop

---

## ❓ Interview Questions & Answers

---

**Q1: What is the `select` statement and how does it differ from `switch`?**

> `select` is like `switch` but specifically for channel operations. Each case must be a channel send or receive. `select` blocks until at least one case is ready, then executes it. Unlike `switch`, which evaluates conditions top-to-bottom and takes the first match, `select` evaluates all cases simultaneously and if multiple are ready, picks one uniformly at random. `switch` is for values and conditions; `select` is for concurrency coordination.

---

**Q2: What happens when multiple cases in a `select` are ready simultaneously?**

> Go picks one of the ready cases uniformly at random. This is intentional design — it prevents starvation where one channel always wins because it happens to be listed first. The randomness ensures fair scheduling across channels over time. If you need priority (always check one channel before another), you must implement it explicitly — typically with a nested select or by checking the high-priority channel first with a non-blocking select.

---

**Q3: What does adding a `default` case to `select` do?**

> Adding `default` makes the `select` non-blocking. If no channel case is immediately ready, the `default` branch executes right away rather than blocking. This is used for: non-blocking sends/receives (try to send, drop if full), polling patterns, and checking channel state without committing to wait. Without `default`, `select` always blocks until at least one case is ready.

---

**Q4: How do you implement a timeout with `select`?**

> Use `time.After(duration)` which returns a `<-chan time.Time` that receives after the duration. Add it as a case in `select` alongside your real operation: `case <-time.After(2*time.Second): return errors.New("timed out")`. When the timeout fires before the other cases, the timeout case is selected. For repeated use in loops, prefer `time.NewTicker` or `time.NewTimer` to avoid the timer leak that `time.After` causes in tight loops.

---

**Q5: What is the `time.After` leak and how do you fix it?**

> `time.After(d)` creates a new `time.Timer` internally each time it's called. The timer and its goroutine persist in memory until it fires — even if the `select` case is never reached. In a tight loop, each iteration creates a new timer that lives for the full duration, causing unbounded memory growth. The fix is to create the timer once outside the loop using `time.NewTicker` for repeated intervals or `time.NewTimer` for one-shot use, and call `Stop()` when done.

---

**Q6: How do you disable a channel case in `select`?**

> Set the channel variable to `nil`. A nil channel blocks forever, so its case is never selected. This is useful in fan-in patterns: when one of several input channels closes, set it to nil to stop reading from it while still reading from the others. For example: `if !ok { ch1 = nil }` inside a select case effectively removes that case from future iterations.

---

**Q7: What does an empty `select {}` do and when would you use it?**

> An empty `select {}` blocks forever — it has no cases to evaluate so it never unblocks. It's occasionally used in `main()` to keep the program alive when all real work happens in goroutines: after launching background goroutines, `select {}` prevents main from returning and killing them. In production, you'd more typically block on an OS signal channel for graceful shutdown: `signal.Notify(sigCh, os.Interrupt); <-sigCh`.

---

*Next up: **Topic 2.7 — sync Package** 🚀*
