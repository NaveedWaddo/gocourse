# 3.3 — Advanced Concurrency Patterns

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.2 — Context Package

---

## 📖 Overview

This topic builds on goroutines, channels, select, and context to cover the **patterns** that real Go services use. These patterns compose the primitives into reliable, cancellable, production-grade concurrent systems.

Patterns covered:
- **Pipeline** — stages connected by channels
- **Fan-out** — distribute work across multiple goroutines
- **Fan-in** (merge) — combine multiple channels into one
- **Worker Pool** — bounded concurrency
- **Semaphore** — limit concurrent access
- **Errgroup** — goroutines with error propagation
- **Rate Limiter** — control throughput
- **Bounded Pipeline with Cancellation** — the full production pattern

---

## 🔗 Pattern 1 — Pipeline

A pipeline connects stages via channels. Each stage reads from an input channel, transforms data, and sends to an output channel. Stages run concurrently.

```go
// Stage: generate values
func generate(ctx context.Context, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage: square each value
func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage: filter even values
func filterEven(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            if n%2 == 0 {
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Pipeline: generate → square → filterEven → print
    nums   := generate(ctx, 1, 2, 3, 4, 5)
    squares := square(ctx, nums)
    evens   := filterEven(ctx, squares)

    for v := range evens {
        fmt.Println(v)   // 4, 16 (squares of 2, 4)
    }
}
```

**Key pipeline rules:**
- Each stage owns its output channel and closes it when done
- Stages exit cleanly on `ctx.Done()`
- `for range` on a channel stops when the channel is closed

---

## 📤 Pattern 2 — Fan-Out

Fan-out distributes work from one channel to multiple goroutines running in parallel:

```go
func fanOut(ctx context.Context, in <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = processWorker(ctx, in)
    }
    return channels
}

func processWorker(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * 2:   // do some work
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

All workers read from the **same** input channel — Go's channel semantics ensure each item goes to exactly one worker.

---

## 📥 Pattern 3 — Fan-In (Merge)

Fan-in merges multiple channels into one — collect results from all workers:

```go
func merge(ctx context.Context, channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    // Launch one goroutine per input channel
    forward := func(ch <-chan int) {
        defer wg.Done()
        for n := range ch {
            select {
            case merged <- n:
            case <-ctx.Done():
                return
            }
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go forward(ch)
    }

    // Close merged once all forwarders finish
    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

---

## 🏊 Pattern 4 — Worker Pool

A worker pool is the most common production pattern — a fixed number of goroutines process an unbounded stream of jobs:

```go
type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID  int
    Output string
    Err    error
}

func workerPool(
    ctx    context.Context,
    jobs   <-chan Job,
    nWorkers int,
) <-chan Result {
    results := make(chan Result)

    var wg sync.WaitGroup
    wg.Add(nWorkers)

    for i := 0; i < nWorkers; i++ {
        workerID := i
        go func() {
            defer wg.Done()
            for {
                select {
                case job, ok := <-jobs:
                    if !ok {
                        return   // jobs channel closed
                    }
                    output, err := processJob(ctx, job)
                    select {
                    case results <- Result{JobID: job.ID, Output: output, Err: err}:
                    case <-ctx.Done():
                        return
                    }
                case <-ctx.Done():
                    return
                }
            }
        }()
        _ = workerID
    }

    // Close results after all workers finish
    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}

func processJob(ctx context.Context, job Job) (string, error) {
    // Simulate work
    select {
    case <-time.After(10 * time.Millisecond):
        return "processed: " + job.Data, nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

### Sending jobs to the pool

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    jobs := make(chan Job, 100)
    results := workerPool(ctx, jobs, 5)   // 5 workers

    // Send jobs (in goroutine so we don't block)
    go func() {
        defer close(jobs)
        for i := 1; i <= 20; i++ {
            select {
            case jobs <- Job{ID: i, Data: fmt.Sprintf("item-%d", i)}:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Collect results
    for r := range results {
        if r.Err != nil {
            fmt.Printf("job %d error: %v\n", r.JobID, r.Err)
            continue
        }
        fmt.Printf("job %d: %s\n", r.JobID, r.Output)
    }
}
```

---

## 🚦 Pattern 5 — Semaphore

A semaphore limits the number of goroutines doing something simultaneously — using a buffered channel:

```go
type Semaphore chan struct{}

func NewSemaphore(n int) Semaphore {
    return make(Semaphore, n)
}

func (s Semaphore) Acquire() { s <- struct{}{} }
func (s Semaphore) Release() { <-s }

// Context-aware acquire
func (s Semaphore) AcquireCtx(ctx context.Context) error {
    select {
    case s <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Usage — limit to 3 concurrent HTTP requests
sem := NewSemaphore(3)
var wg sync.WaitGroup

for _, url := range urls {
    wg.Add(1)
    url := url
    go func() {
        defer wg.Done()
        if err := sem.AcquireCtx(ctx); err != nil {
            return
        }
        defer sem.Release()
        fetch(ctx, url)
    }()
}
wg.Wait()
```

---

## ⚡ Pattern 6 — errgroup

`golang.org/x/sync/errgroup` is the idiomatic way to run goroutines and collect the first error:

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) ([]string, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([]string, len(urls))

    for i, url := range urls {
        i, url := i, url   // capture loop vars
        g.Go(func() error {
            body, err := fetchURL(ctx, url)
            if err != nil {
                return fmt.Errorf("fetch %s: %w", url, err)
            }
            results[i] = body
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err   // first error from any goroutine
    }
    return results, nil
}
```

`errgroup.WithContext` returns a derived context that is **automatically cancelled when any goroutine returns an error**. All goroutines should respect this context.

### With concurrency limit

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(5)   // max 5 goroutines at once (Go 1.20+)

for _, url := range urls {
    url := url
    g.Go(func() error {
        return process(ctx, url)
    })
}
err := g.Wait()
```

---

## ⏱️ Pattern 7 — Rate Limiter

Control how many operations happen per unit of time using `time.Ticker`:

```go
// Simple rate limiter — N operations per second
func rateLimit(ctx context.Context, n int) <-chan struct{} {
    ticker := time.NewTicker(time.Second / time.Duration(n))
    tokens := make(chan struct{})

    go func() {
        defer ticker.Stop()
        defer close(tokens)
        for {
            select {
            case <-ticker.C:
                select {
                case tokens <- struct{}{}:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    return tokens
}

// Usage — 10 requests/second
limiter := rateLimit(ctx, 10)
for _, req := range requests {
    <-limiter   // wait for token
    go sendRequest(ctx, req)
}
```

For production, use `golang.org/x/time/rate`:

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 5)   // 10/sec, burst of 5

for _, req := range requests {
    if err := limiter.Wait(ctx); err != nil {
        return err   // context cancelled
    }
    go sendRequest(ctx, req)
}
```

---

## 💻 Complete Example — Full Pipeline with Fan-Out/In

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// Job and Result types
type WorkItem struct{ ID int; Value int }
type WorkResult struct{ ID int; Result int; Err error }

// Stage 1: generate work items
func generateWork(ctx context.Context, items []WorkItem) <-chan WorkItem {
    out := make(chan WorkItem)
    go func() {
        defer close(out)
        for _, item := range items {
            select {
            case out <- item:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 2 (worker): process one item
func processItem(ctx context.Context, item WorkItem) WorkResult {
    select {
    case <-time.After(time.Duration(10) * time.Millisecond):
        return WorkResult{ID: item.ID, Result: item.Value * item.Value}
    case <-ctx.Done():
        return WorkResult{ID: item.ID, Err: ctx.Err()}
    }
}

// Stage 2 (pool): fan-out to N workers, fan-in results
func processPool(ctx context.Context, in <-chan WorkItem, n int) <-chan WorkResult {
    results := make(chan WorkResult)
    var wg sync.WaitGroup
    wg.Add(n)

    for i := 0; i < n; i++ {
        go func() {
            defer wg.Done()
            for {
                select {
                case item, ok := <-in:
                    if !ok { return }
                    r := processItem(ctx, item)
                    select {
                    case results <- r:
                    case <-ctx.Done():
                        return
                    }
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()
    return results
}

// Stage 3: aggregate results
func aggregate(results <-chan WorkResult) (int, []error) {
    total := 0
    var errs []error
    for r := range results {
        if r.Err != nil {
            errs = append(errs, r.Err)
            continue
        }
        total += r.Result
    }
    return total, errs
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Build work items
    items := make([]WorkItem, 10)
    for i := range items {
        items[i] = WorkItem{ID: i + 1, Value: i + 1}
    }

    // Pipeline: generate → pool(4 workers) → aggregate
    start := time.Now()
    work    := generateWork(ctx, items)
    results := processPool(ctx, work, 4)
    total, errs := aggregate(results)

    fmt.Printf("Processed %d items in %v\n", len(items), time.Since(start))
    fmt.Printf("Sum of squares 1..10: %d\n", total)   // 385
    if len(errs) > 0 {
        fmt.Println("errors:", errs)
    }

    // Show cancellation in action
    fmt.Println("\n--- Cancellation demo ---")
    ctx2, cancel2 := context.WithTimeout(context.Background(), 25*time.Millisecond)
    defer cancel2()

    bigItems := make([]WorkItem, 100)
    for i := range bigItems {
        bigItems[i] = WorkItem{ID: i, Value: i}
    }
    work2    := generateWork(ctx2, bigItems)
    results2 := processPool(ctx2, work2, 4)
    total2, errs2 := aggregate(results2)
    fmt.Printf("Partial result: %d, errors: %d\n", total2, len(errs2))
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Forgetting to close job channels — workers hang forever
```go
// ❌ Workers loop forever waiting on jobs
jobs := make(chan Job)
go sendJobs(jobs)   // never closes jobs
// workers block on <-jobs forever after sendJobs returns

// ✅ Always close the jobs channel when done sending
go func() {
    defer close(jobs)
    for _, j := range allJobs { jobs <- j }
}()
```

### 2. Closing results channel too early in fan-in
```go
// ❌ Closing results before all workers are done
go func() {
    close(results)   // some workers still running!
}()

// ✅ Use WaitGroup — close only after ALL workers done
go func() {
    wg.Wait()
    close(results)
}()
```

### 3. Not capturing loop variables in goroutines
```go
for i, url := range urls {
    // ❌ i and url captured by reference — all goroutines see final value
    go func() { fetch(url) }()

    // ✅ Pass as arguments
    go func(u string) { fetch(u) }(url)
}
```

### 4. Unbounded goroutine spawning
```go
// ❌ Spawns one goroutine per item — may create millions
for _, item := range hugeList {
    go process(item)
}

// ✅ Use a worker pool with bounded N goroutines
results := workerPool(ctx, jobsCh, 20)
```

### 5. Ignoring context in pipeline stages
```go
// ❌ Stage runs to completion even after context is cancelled
func badStage(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * 2   // blocks if nobody reads — goroutine leak!
        }
    }()
    return out
}

// ✅ Always select on ctx.Done()
func goodStage(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * 2:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- **Pipeline** — stages connected by channels; each stage owns and closes its output channel
- **Fan-out** — multiple workers reading from the same input channel
- **Fan-in (merge)** — multiple channels merged into one with WaitGroup + goroutine
- **Worker pool** — N fixed goroutines processing jobs; close jobs channel to signal done
- **Semaphore** — buffered channel `make(chan struct{}, n)` to limit concurrency
- **errgroup** — run goroutines, collect first error, auto-cancel on failure
- Always `defer close(out)` in goroutine that owns the output channel
- Always select on `ctx.Done()` in every blocking operation in a stage
- Use `WaitGroup` to know when all workers are done before closing the results channel
- Never spawn unbounded goroutines — always use a pool
- Loop variable capture — always pass loop vars as function arguments
- `errgroup.SetLimit(n)` (Go 1.20+) — bounded concurrency with errgroup

---

## ❓ Interview Questions & Answers

---

**Q1: What is the pipeline pattern and what are its rules?**

> A pipeline is a series of processing stages connected by channels, where each stage runs in its own goroutine. Data flows from source → stage1 → stage2 → sink. The rules are: (1) each stage receives from an upstream channel and sends to a downstream channel; (2) each stage owns its output channel and closes it when it's done producing; (3) stages exit cleanly when their input channel is closed or their context is cancelled; (4) every blocking send should `select` on `ctx.Done()` to avoid goroutine leaks if a downstream consumer stops reading.

---

**Q2: What is the difference between fan-out and fan-in?**

> Fan-out distributes work from one channel to multiple goroutines — all workers read from the same input channel and Go's channel semantics ensure each item goes to exactly one worker. Fan-in (merge) is the reverse — multiple channels are combined into one output channel, typically using one forwarding goroutine per input and a `WaitGroup` to know when all are done so the output can be closed. Together they implement the scatter-gather pattern: fan-out to parallelize work, fan-in to collect results.

---

**Q3: How does a worker pool work and what closes the results channel?**

> A worker pool launches N goroutines that all read from a shared jobs channel. The sender closes the jobs channel when all jobs are sent — this causes workers' `range` loops to exit. A `WaitGroup` tracks when all N workers have finished; a separate goroutine calls `wg.Wait()` and then closes the results channel. This is the critical ordering: close jobs → workers finish → wg hits zero → results is closed → collector's range loop ends. Getting this order wrong causes deadlocks or goroutine leaks.

---

**Q4: What is a semaphore and how do you implement one in Go?**

> A semaphore limits the number of goroutines executing a section simultaneously. In Go, it's implemented with a buffered channel of capacity N: `sem := make(chan struct{}, N)`. Acquiring means sending into it — if N goroutines already hold the semaphore, the send blocks. Releasing means receiving from it. For context-aware acquisition, use `select` with `ctx.Done()`. This pattern is used for rate-limiting concurrent HTTP requests, database connections, or any resource with a limited pool.

---

**Q5: What is `errgroup` and how does it differ from `sync.WaitGroup`?**

> `sync.WaitGroup` waits for goroutines to finish but has no error handling — goroutines can't return errors. `errgroup` (from `golang.org/x/sync`) adds error propagation: each goroutine passed to `g.Go()` returns an error, and `g.Wait()` returns the first non-nil error from any goroutine. `errgroup.WithContext` also returns a context that is automatically cancelled when any goroutine errors, allowing others to bail out early. It's the standard pattern for parallel work where any failure should cancel the rest.

---

**Q6: How do you prevent goroutine leaks in a pipeline with cancellation?**

> Every blocking channel send in a stage must `select` on `ctx.Done()`. Without this, if a downstream stage stops reading (because it was cancelled), upstream stages block forever on their sends — a goroutine leak. The pattern is: `select { case out <- value: case <-ctx.Done(): return }`. Similarly, goroutines reading from channels must handle both the channel closing (normal completion) and context cancellation. Using `context.WithCancel` or `context.WithTimeout` and passing it through every stage makes the whole pipeline cancellable with a single `cancel()` call.

---

*Next up: **Topic 3.4 — Reflection** 🚀*
