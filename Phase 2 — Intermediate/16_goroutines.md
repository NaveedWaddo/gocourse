# 2.4 — Goroutines

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.3 — Packages & Modules

---

## 📖 What is a Goroutine?

A **goroutine** is a lightweight, independently executing function managed by the Go runtime. It's Go's fundamental unit of concurrency — think of it as an extremely cheap thread.

```go
go someFunction()   // launch someFunction as a goroutine
```

The `go` keyword before a function call is all you need. That function now runs **concurrently** with the rest of your program.

---

## ⚡ Goroutines vs OS Threads

| | OS Thread | Goroutine |
|--|-----------|-----------|
| Stack size | ~1–8 MB (fixed) | ~2 KB (grows dynamically) |
| Creation cost | Expensive (kernel call) | Very cheap (Go runtime) |
| Switching cost | Expensive (context switch) | Cheap (cooperative + preemptive) |
| How many practical | Thousands | Millions |
| Managed by | OS | Go runtime scheduler |

You can run **millions of goroutines** simultaneously — something impossible with OS threads. This is what makes Go excellent for highly concurrent workloads like web servers.

---

## 🚀 Launching Goroutines

### Basic launch

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    fmt.Println("Hello,", name)
}

func main() {
    go sayHello("Alice")   // runs concurrently
    go sayHello("Bob")
    go sayHello("Carol")

    time.Sleep(100 * time.Millisecond)   // wait for goroutines to finish
    fmt.Println("main done")
}
```

> ⚠️ `time.Sleep` is a crude way to wait — we'll use `sync.WaitGroup` (Topic 2.7) and channels (Topic 2.5) properly. Without any waiting, `main` exits and kills all goroutines immediately.

### With an anonymous function

```go
go func() {
    fmt.Println("I'm an anonymous goroutine")
}()   // note the () — immediately invoked

// With arguments — always pass loop variables as arguments
for i := 0; i < 5; i++ {
    go func(n int) {      // n is a copy of i at this point
        fmt.Println(n)
    }(i)                  // pass i as argument here
}
```

---

## 🧠 The Go Scheduler — How It Works

Go uses an **M:N scheduler** — it maps M goroutines onto N OS threads:

```
Goroutines (millions)
    G G G G G G G G G G
         ↓ scheduler
    P   P   P   P          ← logical processors (GOMAXPROCS)
    |   |   |   |
    M   M   M   M          ← OS threads
```

Key concepts:

- **G** — Goroutine
- **M** — Machine (OS thread)
- **P** — Processor (logical, holds a run queue of goroutines)
- **GOMAXPROCS** — number of P's = number of OS threads running Go code in parallel (defaults to number of CPU cores)

```go
import "runtime"

fmt.Println(runtime.GOMAXPROCS(0))   // print current value
runtime.GOMAXPROCS(4)                // set to 4 (usually leave as default)
```

### Cooperative + preemptive scheduling

- Goroutines yield at function calls, channel operations, system calls
- Since Go 1.14, the scheduler is also **preemptive** — long-running loops can be interrupted
- Goroutines are NOT pinned to a thread — the scheduler can move them around

---

## 🔄 Goroutine Lifecycle

```
Created (go func())
    → Runnable (waiting for a P)
        → Running (on an M/P)
            → Blocked (waiting on channel, I/O, mutex)
            → Runnable again (when unblocked)
        → Dead (function returned)
```

A goroutine ends when its function returns. There is no way to kill a goroutine from the outside — you signal it via a channel or `context.Context` (Topic 3.2).

---

## 🔁 Concurrency vs Parallelism

```
Concurrency:  Dealing with multiple things at once (structure)
Parallelism:  Doing multiple things at once (execution)
```

- **Concurrency** — goroutines switching between tasks on one core
- **Parallelism** — goroutines running simultaneously on multiple cores

Go gives you both. Concurrency is built into the language; parallelism happens automatically when `GOMAXPROCS > 1`.

> **"Concurrency is not parallelism."** — Rob Pike

---

## ⚠️ The Main Goroutine Problem

When `main()` returns, **all goroutines are killed immediately** — even ones still running:

```go
func main() {
    go func() {
        time.Sleep(1 * time.Second)
        fmt.Println("This may never print!")
    }()
    // main returns immediately — goroutine is killed
}
```

Solutions (in order of preference):
1. **`sync.WaitGroup`** — wait for a known number of goroutines (Topic 2.7)
2. **Channels** — coordinate via communication (Topic 2.5)
3. **`time.Sleep`** — only for demos/tests, never production

---

## 🏁 sync.WaitGroup — The Right Way to Wait

Even though `WaitGroup` is covered fully in Topic 2.7, it's so essential to goroutines that we introduce it here:

```go
import "sync"

var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)          // tell WaitGroup: one more goroutine starting
    go func(n int) {
        defer wg.Done()   // tell WaitGroup: this goroutine is done
        fmt.Println("worker", n)
    }(i)
}

wg.Wait()   // block until all goroutines call Done()
fmt.Println("all workers done")
```

---

## 🏗️ Common Goroutine Patterns

### Pattern 1: Fire and forget

```go
// Launch a background task — don't need the result
go func() {
    sendEmailNotification(user)
}()
```

### Pattern 2: Worker pool (basic)

```go
func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for j := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, j)
    }
}

func main() {
    jobs := make(chan int, 100)
    var wg sync.WaitGroup

    // Launch 3 workers
    for w := 1; w <= 3; w++ {
        wg.Add(1)
        go worker(w, jobs, &wg)
    }

    // Send 9 jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)   // signal no more jobs

    wg.Wait()
}
```

### Pattern 3: Goroutine per request (web server style)

```go
// Go's net/http does this internally for every HTTP request
for {
    conn, err := listener.Accept()
    if err != nil {
        break
    }
    go handleConnection(conn)   // each connection gets its own goroutine
}
```

---

## 🔍 Goroutine Leaks — A Critical Problem

A goroutine leak happens when a goroutine is launched but **never terminates** — it stays blocked forever, consuming memory.

```go
// ❌ Goroutine leak — blocked forever if nobody reads from ch
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch   // blocks forever — nobody sends!
        fmt.Println(val)
    }()
    // function returns, ch goes out of scope, but goroutine is still blocked
}
```

Common causes:
- Goroutine waiting on a channel that's never written to or closed
- Goroutine waiting on a mutex that's never unlocked
- Infinite loop with no exit condition

Prevention:
- Always provide a way to signal goroutines to stop (channels, `context.Context`)
- Use `context.WithCancel` or `context.WithTimeout` for goroutines doing I/O
- Use tools like `goleak` in tests to detect leaks

```go
// ✅ Stoppable goroutine with done channel
func startWorker(done <-chan struct{}) {
    go func() {
        for {
            select {
            case <-done:
                fmt.Println("worker stopping")
                return
            default:
                doWork()
            }
        }
    }()
}

done := make(chan struct{})
startWorker(done)
// ... later ...
close(done)   // signals goroutine to stop
```

---

## 🏃 Race Conditions — Shared State is Dangerous

When multiple goroutines access shared data without synchronization, you get a **data race** — undefined behavior:

```go
// ❌ DATA RACE — both goroutines read/write counter simultaneously
counter := 0
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter++   // NOT safe — read-modify-write is not atomic
    }()
}
wg.Wait()
fmt.Println(counter)   // could be anything less than 1000!
```

### Detecting races with the race detector

```bash
go run -race main.go
go test -race ./...
```

Always run tests with `-race` in CI. The race detector adds ~2-20x overhead but catches real bugs.

### Fix options:
1. **`sync.Mutex`** — lock around shared access (Topic 2.7)
2. **`sync/atomic`** — atomic operations for simple counters
3. **Channels** — share by communicating, not by sharing memory

```go
// ✅ Fix with atomic
import "sync/atomic"
var counter int64
atomic.AddInt64(&counter, 1)   // atomic increment — safe

// ✅ Fix with mutex
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()
```

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// Simulate work with variable duration
func processTask(id int, duration time.Duration, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Task %d started\n", id)
    time.Sleep(duration)
    fmt.Printf("Task %d completed (took %v)\n", id, duration)
}

// Safe counter using atomic
func atomicCounterDemo() {
    var counter int64
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            atomic.AddInt64(&counter, 1)
        }()
    }
    wg.Wait()
    fmt.Println("Final counter:", counter)   // always 1000
}

func main() {
    fmt.Println("=== Concurrent Tasks ===")
    var wg sync.WaitGroup

    tasks := []time.Duration{
        300 * time.Millisecond,
        100 * time.Millisecond,
        200 * time.Millisecond,
        150 * time.Millisecond,
        250 * time.Millisecond,
    }

    start := time.Now()

    for i, d := range tasks {
        wg.Add(1)
        go processTask(i+1, d, &wg)
    }

    wg.Wait()
    fmt.Printf("All tasks done in %v (sequential would take %v)\n",
        time.Since(start),
        300+100+200+150+250)
    // concurrent: ~300ms vs sequential: 1000ms

    fmt.Println("\n=== Atomic Counter ===")
    atomicCounterDemo()
}
```

Output (order varies):
```
=== Concurrent Tasks ===
Task 1 started
Task 2 started
Task 3 started
Task 4 started
Task 5 started
Task 2 completed (took 100ms)
Task 4 completed (took 150ms)
Task 3 completed (took 200ms)
Task 5 completed (took 250ms)
Task 1 completed (took 300ms)
All tasks done in 301ms (sequential would take 1000ms)

=== Atomic Counter ===
Final counter: 1000
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Closure captures loop variable by reference
```go
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i)   // ❌ all goroutines see final value of i (5)
    }()
}

// ✅ Fix — pass as argument
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n)   // each goroutine gets its own copy
    }(i)
}
```

### 2. main() exits before goroutines finish
```go
func main() {
    go doWork()
    // ❌ main returns, goroutine killed
}

// ✅ Use WaitGroup or channel to wait
```

### 3. Goroutine leaks from blocked channels
```go
// ❌ goroutine blocked forever
go func() { result := <-neverSentChan }()
```

### 4. Data race on shared variables
```go
// ❌ Race condition
go func() { counter++ }()
go func() { counter++ }()

// ✅ Use atomic or mutex
```

### 5. Launching goroutines in init()
```go
// ❌ Hard to control lifecycle, hard to test
func init() {
    go backgroundWorker()
}
// ✅ Launch goroutines in main() or with explicit Start() method
```

---

## 🎯 Must-Know Things (Quick Reference)

- `go funcName()` — launches a goroutine, returns immediately
- Goroutines are **~2KB** stack, can have **millions** running simultaneously
- Go runtime scheduler maps goroutines onto OS threads (M:N model)
- `GOMAXPROCS` controls parallelism — defaults to number of CPU cores
- **main() exiting kills all goroutines** — always wait with WaitGroup or channels
- **Goroutine leaks** — goroutines blocked forever, consuming memory — always provide stop signal
- **Race conditions** — shared state without sync → use `-race` to detect
- Loop variable closure bug — always pass loop variable as function argument
- **"Do not communicate by sharing memory; share memory by communicating"** — Go proverb
- Use `context.Context` for goroutine cancellation in production (Topic 3.2)

---

## ❓ Interview Questions & Answers

---

**Q1: What is a goroutine and how does it differ from an OS thread?**

> A goroutine is a lightweight function managed by the Go runtime scheduler. Unlike OS threads which start with 1–8MB stacks and require kernel context switches, goroutines start with ~2KB stacks that grow dynamically and are scheduled cooperatively+preemptively in user space. This makes them extremely cheap to create and switch between — you can run millions of goroutines where thousands of OS threads would exhaust memory. The Go scheduler multiplexes goroutines onto a small number of OS threads using the M:N model.

---

**Q2: What is GOMAXPROCS and what does it control?**

> `GOMAXPROCS` is the number of logical processors (P's) the Go scheduler uses — it controls how many goroutines can run in true parallel simultaneously. It defaults to the number of available CPU cores. Setting it to 1 means goroutines run concurrently but not in parallel (one at a time). Setting it higher allows genuine parallelism. You rarely need to change it — the default is correct for most programs. It can be set via `runtime.GOMAXPROCS(n)` or the `GOMAXPROCS` environment variable.

---

**Q3: What is a goroutine leak and how do you prevent it?**

> A goroutine leak occurs when a goroutine is launched but never terminates — typically because it's blocked waiting on a channel that's never written to, or a mutex that's never unlocked. Leaked goroutines accumulate over time, consuming memory and potentially other resources. Prevention: always provide a cancellation mechanism — a `done` channel to close or a `context.Context` to cancel. Use the `goleak` library in tests to detect leaks. Tools like `runtime.NumGoroutine()` can also help monitor goroutine counts.

---

**Q4: What is a data race and how do you detect one in Go?**

> A data race occurs when two or more goroutines access the same memory location concurrently and at least one access is a write, with no synchronization between them. The result is undefined behavior — you might get wrong values, crashes, or intermittent bugs. Go has a built-in race detector: run `go run -race` or `go test -race` and it instruments memory accesses to detect races at runtime, printing exactly which goroutines are involved. Always run tests with `-race` in CI.

---

**Q5: What is the closure-in-loop goroutine bug?**

> When a goroutine's anonymous function closes over a loop variable, all goroutines share the same variable — by the time they run, the loop has finished and the variable holds its final value. For example, launching goroutines in a `for i` loop where all print `i` results in all printing the loop's final value. The fix is to pass the loop variable as a function argument: `go func(n int) { fmt.Println(n) }(i)` — this creates a copy of `i` for each goroutine.

---

**Q6: How do you stop a goroutine in Go?**

> Go has no built-in way to kill a goroutine from outside — you must signal it to stop itself. The two idiomatic patterns are: (1) a `done` channel — the goroutine listens with `select` and returns when the channel is closed; (2) `context.Context` — pass a context and check `ctx.Done()` or use it with blocking operations that accept a context. The goroutine is responsible for checking and honoring the stop signal. This design encourages clean resource cleanup via `defer`.

---

**Q7: What does "don't communicate by sharing memory; share memory by communicating" mean?**

> This Go proverb captures the preferred concurrency approach. Traditional concurrency uses shared memory protected by locks — threads read and write the same variables, using mutexes to prevent races. Go's preferred approach is to pass data between goroutines via channels — instead of having two goroutines access the same variable, one sends the value to the other. The receiver "owns" the data while it has it, so no synchronization is needed. This makes data flow explicit and eliminates many classes of race conditions. Both approaches are valid in Go, but channels are idiomatic for coordination.

---

*Next up: **Topic 2.5 — Channels** 🚀*
