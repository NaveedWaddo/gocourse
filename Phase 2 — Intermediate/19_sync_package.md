# 2.7 — sync Package

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.6 — Select Statement

---

## 📖 Overview

The `sync` package provides low-level synchronization primitives for coordinating goroutines via **shared memory** (as opposed to channels). Use it when multiple goroutines need to safely access shared state.

The main types:
- `sync.WaitGroup` — wait for a group of goroutines to finish
- `sync.Mutex` — mutual exclusion lock
- `sync.RWMutex` — reader/writer lock
- `sync.Once` — run something exactly once
- `sync.Map` — concurrent-safe map
- `sync.Pool` — reusable object pool

---

## ⏳ `sync.WaitGroup` — Wait for Goroutines

A `WaitGroup` waits for a collection of goroutines to finish. The counter tracks how many goroutines are running.

```go
var wg sync.WaitGroup

wg.Add(1)    // increment counter — call BEFORE launching goroutine
go func() {
    defer wg.Done()   // decrement counter when goroutine exits
    doWork()
}()

wg.Wait()    // block until counter reaches 0
```

### Three methods

| Method | What it does |
|--------|-------------|
| `wg.Add(n)` | Increment counter by n |
| `wg.Done()` | Decrement counter by 1 (same as `Add(-1)`) |
| `wg.Wait()` | Block until counter is 0 |

### Full example

```go
func main() {
    var wg sync.WaitGroup
    results := make([]int, 5)

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            results[idx] = idx * idx   // safe — each goroutine writes its own index
        }(i)
    }

    wg.Wait()
    fmt.Println(results)   // [0 1 4 9 16]
}
```

### Rules

```go
// ✅ Add BEFORE launching goroutine
wg.Add(1)
go func() { defer wg.Done(); doWork() }()

// ❌ Adding inside the goroutine — race condition
go func() {
    wg.Add(1)   // might not run before wg.Wait() returns!
    defer wg.Done()
    doWork()
}()
```

> Never copy a `WaitGroup` — always pass a pointer: `func work(wg *sync.WaitGroup)`.

---

## 🔒 `sync.Mutex` — Mutual Exclusion

A `Mutex` ensures only **one goroutine** accesses a critical section at a time.

```go
var mu sync.Mutex
counter := 0

var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        mu.Lock()
        counter++     // protected — only one goroutine at a time
        mu.Unlock()
    }()
}
wg.Wait()
fmt.Println(counter)   // always 1000
```

### Always unlock — use `defer`

```go
func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()   // ✅ guaranteed to unlock even if panic
    c.count++
}
```

### Embedding mutex in a struct (idiomatic)

```go
type SafeMap struct {
    mu   sync.Mutex
    data map[string]int
}

func (m *SafeMap) Set(key string, val int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = val
}

func (m *SafeMap) Get(key string) (int, bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    v, ok := m.data[key]
    return v, ok
}

func NewSafeMap() *SafeMap {
    return &SafeMap{data: make(map[string]int)}
}
```

### Deadlock with Mutex

```go
// ❌ Deadlock — trying to lock a mutex you already hold
func bad() {
    mu.Lock()
    mu.Lock()   // blocks forever — same goroutine can't lock twice
}

// ❌ Deadlock — two goroutines each waiting for the other's lock
// goroutine 1: holds mu1, wants mu2
// goroutine 2: holds mu2, wants mu1
```

> Go's `sync.Mutex` is **not reentrant** — locking it twice in the same goroutine deadlocks.

---

## 📖 `sync.RWMutex` — Reader/Writer Lock

An `RWMutex` allows **multiple concurrent readers** OR **one exclusive writer** — more efficient than a plain Mutex when reads are frequent and writes are rare.

```go
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]string
}

// Read — multiple goroutines can hold RLock simultaneously
func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

// Write — exclusive lock, blocks all readers and writers
func (c *SafeCache) Set(key, val string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = val
}
```

| Lock type | Method | Who can hold simultaneously |
|-----------|--------|---------------------------|
| Read lock | `RLock()` / `RUnlock()` | Multiple goroutines |
| Write lock | `Lock()` / `Unlock()` | Only one goroutine |

Use `RWMutex` when: reads >> writes (e.g., cache, config, registry).

---

## 1️⃣ `sync.Once` — Run Exactly Once

`Once` ensures a function is executed **exactly once**, regardless of how many goroutines call it — perfect for lazy initialization and singleton patterns.

```go
var once sync.Once
var instance *Database

func GetDatabase() *Database {
    once.Do(func() {
        // This runs exactly once — even if called from 100 goroutines simultaneously
        instance = &Database{
            conn: connectToDB(),
        }
        fmt.Println("database initialized")
    })
    return instance
}
```

```go
// Multiple calls — function runs only once
once.Do(setup)   // runs
once.Do(setup)   // skipped
once.Do(setup)   // skipped
```

### Important: if `Do` panics, it's considered done

```go
// ❌ If fn panics, Once marks it done — future calls are skipped
once.Do(func() {
    panic("setup failed")
})
once.Do(func() { /* never runs */ })
```

---

## 🗺️ `sync.Map` — Concurrent-Safe Map

A built-in concurrent map — safe for simultaneous reads and writes from multiple goroutines without external locking.

```go
var m sync.Map

// Store
m.Store("key", "value")

// Load
val, ok := m.Load("key")
if ok {
    fmt.Println(val.(string))   // type assertion needed — returns interface{}
}

// LoadOrStore — atomic: load if exists, store if not
actual, loaded := m.LoadOrStore("key", "new-value")
// loaded=true if key existed, actual=existing value

// Delete
m.Delete("key")

// Range — iterate (order not guaranteed)
m.Range(func(key, value any) bool {
    fmt.Println(key, value)
    return true   // return false to stop iteration
})
```

### When to use `sync.Map` vs `Mutex + map`

| | `sync.Map` | `Mutex + map` |
|--|-----------|--------------|
| Read-heavy workload | ✅ Optimized | ❌ Lock contention |
| Write-heavy workload | ❌ Slower | ✅ Better |
| Type safety | ❌ `any` values | ✅ Typed |
| Many keys written once, read many times | ✅ Ideal | OK |
| General purpose | OK | ✅ Preferred |

> For most use cases, `Mutex + map` is simpler and type-safe. Use `sync.Map` specifically for the "write once, read many" pattern or when you need the atomic `LoadOrStore`.

---

## ♻️ `sync.Pool` — Object Reuse

`Pool` is a cache of reusable objects — reduces GC pressure by recycling allocations:

```go
var bufPool = sync.Pool{
    New: func() any {
        return make([]byte, 0, 1024)   // create new buffer if pool empty
    },
}

func processRequest(data []byte) {
    // Get a buffer from the pool
    buf := bufPool.Get().([]byte)
    buf = buf[:0]   // reset length, keep capacity
    defer bufPool.Put(buf)   // return to pool when done

    // Use buf...
    buf = append(buf, data...)
    process(buf)
}
```

Rules about `Pool`:
- Items may be **garbage collected at any time** — pool is a cache, not a store
- `New` function creates items when pool is empty — set it always
- Never store pointers to pool objects outside the pool's scope

---

## 🔢 `sync/atomic` — Atomic Operations

For simple numeric operations, `sync/atomic` is faster than a Mutex:

```go
import "sync/atomic"

var counter int64

// Atomic increment
atomic.AddInt64(&counter, 1)

// Atomic read
val := atomic.LoadInt64(&counter)

// Atomic write
atomic.StoreInt64(&counter, 0)

// Compare-and-swap — only swap if current value matches expected
swapped := atomic.CompareAndSwapInt64(&counter, 0, 1)
```

Use atomic for:
- Counters (hit counts, metrics)
- Flags (shutdown flag, initialized flag)
- Simple state transitions with CAS

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

// Thread-safe counter using Mutex
type Counter struct {
    mu    sync.Mutex
    value int64
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int64 {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// Read-heavy cache using RWMutex
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewCache() *Cache {
    return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *Cache) Set(key, val string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = val
}

// Singleton using Once
type Config struct{ DSN string }

var (
    configOnce sync.Once
    config     *Config
)

func GetConfig() *Config {
    configOnce.Do(func() {
        config = &Config{DSN: "postgres://localhost/mydb"}
        fmt.Println("config initialized (once)")
    })
    return config
}

func main() {
    // WaitGroup + Mutex counter
    fmt.Println("=== Mutex Counter ===")
    counter := &Counter{}
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }
    wg.Wait()
    fmt.Println("Counter:", counter.Value())   // always 1000

    // Atomic counter
    fmt.Println("\n=== Atomic Counter ===")
    var atomicCount int64
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            atomic.AddInt64(&atomicCount, 1)
        }()
    }
    wg.Wait()
    fmt.Println("Atomic:", atomic.LoadInt64(&atomicCount))   // always 1000

    // RWMutex cache
    fmt.Println("\n=== RWMutex Cache ===")
    cache := NewCache()
    cache.Set("name", "Alice")
    cache.Set("lang", "Go")

    var rwWg sync.WaitGroup
    for i := 0; i < 5; i++ {
        rwWg.Add(1)
        go func(id int) {
            defer rwWg.Done()
            name, _ := cache.Get("name")
            fmt.Printf("reader %d: name=%s\n", id, name)
        }(i)
    }
    rwWg.Wait()

    // Once
    fmt.Println("\n=== sync.Once ===")
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            cfg := GetConfig()
            _ = cfg
        }()
    }
    wg.Wait()
    // "config initialized (once)" prints only once
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Copying sync types — undefined behavior
```go
// ❌ Never copy a Mutex, WaitGroup, Once, etc.
mu := sync.Mutex{}
mu2 := mu   // copies internal state — undefined behavior!

// ✅ Always use pointers
func work(mu *sync.Mutex) { ... }
```

### 2. Forgetting to unlock — deadlock
```go
// ❌ If function returns early, Unlock never called
func bad() {
    mu.Lock()
    if err != nil {
        return   // Unlock skipped!
    }
    mu.Unlock()
}

// ✅ Always defer Unlock right after Lock
func good() {
    mu.Lock()
    defer mu.Unlock()   // always runs
    // ...
}
```

### 3. Holding lock too long
```go
// ❌ Lock held during slow I/O — all other goroutines blocked
mu.Lock()
result := callExternalAPI()   // slow!
mu.Unlock()

// ✅ Do slow work outside the lock
result := callExternalAPI()
mu.Lock()
cache[key] = result
mu.Unlock()
```

### 4. RLock when you need Lock
```go
// ❌ RLock is for reads only — writing under RLock is a race!
mu.RLock()
m[key] = value   // RACE — another reader might be reading!
mu.RUnlock()

// ✅
mu.Lock()
m[key] = value
mu.Unlock()
```

### 5. wg.Add inside the goroutine
```go
// ❌ Race — wg.Wait() might return before Add is called
go func() {
    wg.Add(1)   // too late!
    defer wg.Done()
}()
wg.Wait()

// ✅ Add before launching
wg.Add(1)
go func() { defer wg.Done() }()
```

---

## 🎯 Must-Know Things (Quick Reference)

- `WaitGroup`: `Add` before launch, `Done` in goroutine, `Wait` to block
- `Mutex`: `Lock` / `Unlock` — always `defer Unlock` right after `Lock`
- `RWMutex`: multiple concurrent `RLock` for reads, exclusive `Lock` for writes
- `Once`: runs function **exactly once** — perfect for singletons and init
- `sync.Map`: concurrent-safe but untyped — prefer `Mutex + map` for type safety
- `Pool`: recycle objects to reduce GC — items can be collected anytime
- **Never copy** sync types — always pass pointers
- Mutex is **not reentrant** — locking twice in same goroutine = deadlock
- Hold locks for the **minimum time** — move slow work outside the lock
- `sync/atomic` — faster than Mutex for simple int/pointer operations

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between `sync.Mutex` and `sync.RWMutex`?**

> A `Mutex` allows only one goroutine at a time — both reads and writes are exclusive. An `RWMutex` differentiates: multiple goroutines can hold the read lock (`RLock`) simultaneously, but the write lock (`Lock`) is exclusive and blocks all readers and writers. Use `RWMutex` when reads are far more frequent than writes — like a cache or configuration store. For write-heavy workloads, the overhead of `RWMutex` may outweigh the benefit and a plain `Mutex` is simpler.

---

**Q2: What is `sync.Once` and what problem does it solve?**

> `sync.Once` ensures a function runs exactly once, even when called concurrently from multiple goroutines. It solves the lazy initialization / singleton problem — you want to initialize an expensive resource (database connection, config) exactly once on first use, safely in a concurrent program. Without `Once`, you'd need a mutex and a boolean flag to achieve the same thing, which is error-prone. `Once` encapsulates that pattern cleanly.

---

**Q3: Why should you never copy `sync` types?**

> `sync` types like `Mutex`, `WaitGroup`, `Once`, and `Cond` contain internal state (lock state, counters, flags). Copying them copies that state at a point in time, creating two independent values that don't share the same lock — so locking one has no effect on the other. This leads to races and deadlocks. Always pass them by pointer. Go vet detects this: it warns about copying types with `sync.Locker` in their method set.

---

**Q4: What is the `sync.WaitGroup` Add-before-launch rule?**

> `wg.Add(n)` must be called before launching the goroutine, not inside it. If you call `Add` inside the goroutine, there's a race: the goroutine might not execute before `wg.Wait()` is called, causing `Wait` to return before the goroutine has even started, let alone finished. By calling `Add` synchronously before `go`, you guarantee the counter is incremented before `Wait` could possibly return.

---

**Q5: When would you use `sync.Map` vs a `Mutex` with a regular map?**

> Use `sync.Map` for the specific pattern of: keys written once and then read many times, or when many goroutines access disjoint sets of keys. It's optimized for these cases with lock-free reads. For general-purpose concurrent maps — especially write-heavy ones — `Mutex + map` is simpler, type-safe (no `any` type assertions), and often faster. `sync.Map` also has no `len` method, making it less ergonomic for general use.

---

**Q6: What is `sync.Pool` and when should you use it?**

> `sync.Pool` is a cache of reusable temporary objects. When an object is needed, `Get` returns one from the pool or creates a new one via the `New` function. When done, `Put` returns it. This reduces allocation and GC pressure for objects that are frequently created and discarded — like byte buffers, JSON encoders, or request objects. Important: the pool can drop items at any GC cycle, so it's a performance optimization only, not reliable storage.

---

**Q7: What is a deadlock and what are the common causes with `sync.Mutex`?**

> A deadlock occurs when goroutines are permanently blocked waiting for each other. Common `Mutex` causes: (1) forgetting to unlock — a goroutine locks but returns early without unlocking; fix with `defer mu.Unlock()`. (2) Double-locking — a goroutine tries to lock a mutex it already holds, since `sync.Mutex` is not reentrant. (3) Lock ordering — two goroutines each hold one lock and want the other's; fix by always acquiring locks in the same order. Go's race detector won't catch deadlocks, but `go test -timeout` can help detect them.

---

*Next up: **Topic 2.8 — Defer, Panic & Recover** 🚀*
