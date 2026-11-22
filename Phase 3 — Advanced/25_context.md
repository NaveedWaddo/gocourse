# 3.2 — Context Package

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.1 — Generics

---

## 📖 What is `context`?

The `context` package carries **deadlines**, **cancellation signals**, and **request-scoped values** across API boundaries and between goroutines. It's the standard way in Go to:

- Cancel in-flight operations when the caller no longer needs the result
- Set timeouts on operations (HTTP requests, DB queries, etc.)
- Pass request-scoped data (trace IDs, auth tokens) without polluting function signatures

```go
import "context"
```

Every long-running or I/O operation in idiomatic Go accepts a `context.Context` as its **first parameter**:

```go
func DoWork(ctx context.Context, ...) error { ... }
```

---

## 🌳 The Context Tree

Contexts form a **tree** — child contexts inherit from parents. Cancelling a parent cancels all its children automatically.

```
context.Background()         ← root (never cancelled)
    └── WithCancel()         ← cancellable
        ├── WithTimeout()    ← auto-cancels after duration
        │   └── WithValue()  ← carries a value
        └── WithDeadline()   ← auto-cancels at a time
```

---

## 🌱 Creating Contexts

### Root contexts — two options

```go
ctx := context.Background()   // top-level, never cancelled — use at main/init/test
ctx := context.TODO()         // placeholder — use when you plan to add context later
```

Both are empty contexts — `Background` is the real root; `TODO` signals "I know I need context here but haven't decided what yet".

### Derived contexts

```go
// WithCancel — manual cancellation
ctx, cancel := context.WithCancel(parent)
defer cancel()   // always call cancel to release resources

// WithTimeout — auto-cancel after duration
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// WithDeadline — auto-cancel at specific time
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()

// WithValue — attach a value
ctx = context.WithValue(parent, key, value)
```

> Always `defer cancel()` immediately after creating a cancellable context — even for timeouts/deadlines. This ensures resources are freed if the operation completes before the timeout.

---

## ❌ Cancellation — `WithCancel`

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        time.Sleep(2 * time.Second)
        cancel()   // signal cancellation after 2s
    }()

    // Block until context is cancelled
    <-ctx.Done()
    fmt.Println("cancelled:", ctx.Err())   // context canceled
}
```

### Checking for cancellation in a goroutine

```go
func worker(ctx context.Context, jobs <-chan int) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()   // context.Canceled or context.DeadlineExceeded
        case job, ok := <-jobs:
            if !ok {
                return nil   // jobs channel closed
            }
            if err := processJob(ctx, job); err != nil {
                return err
            }
        }
    }
}
```

### `ctx.Done()` — the cancellation channel

`ctx.Done()` returns a `<-chan struct{}` that is closed when the context is cancelled or times out. You use it in `select` statements to detect cancellation.

```go
// Three context error values
context.Canceled          // context was cancelled via cancel()
context.DeadlineExceeded  // context timed out

// Check which one
if err := ctx.Err(); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("timed out")
    } else {
        fmt.Println("cancelled")
    }
}
```

---

## ⏱️ Timeout — `WithTimeout`

The most common pattern — every outbound call should have a timeout:

```go
func fetchUser(ctx context.Context, id int) (*User, error) {
    // Create a child context with 3 second timeout
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("https://api.example.com/users/%d", id), nil)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("fetchUser: timed out after 3s")
        }
        return nil, fmt.Errorf("fetchUser: %w", err)
    }
    defer resp.Body.Close()
    // decode response...
    return user, nil
}
```

### HTTP server — per-request timeout

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // r.Context() already carries the request's context
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    result, err := db.QueryContext(ctx, "SELECT ...")
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "request timed out", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    _ = result
}
```

---

## 🔑 Values — `WithValue`

Pass request-scoped data through the call stack without changing every function signature:

```go
// Define a key type — always use unexported type to avoid collisions
type contextKey string

const (
    keyRequestID contextKey = "requestID"
    keyUserID    contextKey = "userID"
)

// Store a value
ctx = context.WithValue(ctx, keyRequestID, "req-123-abc")

// Retrieve a value — returns interface{}, must type-assert
if id, ok := ctx.Value(keyRequestID).(string); ok {
    fmt.Println("request ID:", id)
}
```

### Common uses for context values

```go
// Logging middleware — attach trace/request ID
func withRequestID(ctx context.Context, r *http.Request) context.Context {
    id := r.Header.Get("X-Request-ID")
    if id == "" {
        id = uuid.New().String()
    }
    return context.WithValue(ctx, keyRequestID, id)
}

// Auth middleware — attach authenticated user
func withUser(ctx context.Context, user *User) context.Context {
    return context.WithValue(ctx, keyUserID, user)
}

// Retrieve in handler
func getRequestID(ctx context.Context) string {
    id, _ := ctx.Value(keyRequestID).(string)
    return id
}
```

> **Use context values sparingly** — only for request-scoped data that truly spans multiple layers (trace IDs, auth info). Never use context to pass optional function parameters — use explicit function arguments instead.

---

## 🏗️ Propagating Context — Best Practices

```go
// ✅ Accept context as FIRST parameter, always named ctx
func DoSomething(ctx context.Context, arg1 string, arg2 int) error {
    // Pass ctx to every downstream call
    if err := db.QueryContext(ctx, query); err != nil {
        return err
    }
    if err := callService(ctx, data); err != nil {
        return err
    }
    return nil
}

// ✅ HTTP handler — use r.Context()
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    result, err := doWork(ctx)
    // ...
}

// ❌ Never store context in a struct
type Service struct {
    ctx context.Context   // wrong — context should flow through functions
}

// ✅ Pass context as a parameter
type Service struct{}
func (s *Service) DoWork(ctx context.Context) error { ... }
```

---

## 🌐 Context in the Standard Library

Many standard library functions accept context — always pass it:

```go
// net/http
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)

// database/sql
rows, err := db.QueryContext(ctx, query, args...)
err = db.ExecContext(ctx, query, args...)

// os/exec
cmd := exec.CommandContext(ctx, "sleep", "10")
// cmd is killed if ctx is cancelled

// grpc
client.SomeRPC(ctx, request)
```

---

## 💻 Complete Example

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

type contextKey string
const traceKey contextKey = "traceID"

// Simulate a slow DB query that respects context
func queryDB(ctx context.Context, query string) (string, error) {
    resultCh := make(chan string, 1)

    go func() {
        time.Sleep(200 * time.Millisecond)   // simulate DB latency
        resultCh <- "result: " + query
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case <-ctx.Done():
        return "", fmt.Errorf("queryDB: %w", ctx.Err())
    }
}

// Service layer — propagates context
func getUser(ctx context.Context, id int) (string, error) {
    traceID, _ := ctx.Value(traceKey).(string)
    fmt.Printf("[%s] getUser(%d)\n", traceID, id)

    result, err := queryDB(ctx, fmt.Sprintf("SELECT * FROM users WHERE id=%d", id))
    if err != nil {
        return "", fmt.Errorf("getUser: %w", err)
    }
    return result, nil
}

func main() {
    // --- WithValue + WithTimeout ---
    fmt.Println("=== Timeout + Value ===")
    rootCtx := context.Background()
    ctx := context.WithValue(rootCtx, traceKey, "trace-abc-123")

    ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
    defer cancel()

    user, err := getUser(ctx, 42)
    if err != nil {
        fmt.Println("error:", err)
    } else {
        fmt.Println("success:", user)
    }

    // --- Timeout exceeded ---
    fmt.Println("\n=== Timeout Exceeded ===")
    ctx2, cancel2 := context.WithTimeout(rootCtx, 100*time.Millisecond)
    defer cancel2()

    _, err = getUser(ctx2, 99)
    fmt.Println("error:", err)
    fmt.Println("is DeadlineExceeded:", errors.Is(err, context.DeadlineExceeded))

    // --- Manual cancellation ---
    fmt.Println("\n=== Manual Cancel ===")
    ctx3, cancel3 := context.WithCancel(rootCtx)

    go func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Println("cancelling...")
        cancel3()
    }()

    select {
    case <-ctx3.Done():
        fmt.Println("ctx3 done:", ctx3.Err())
    case <-time.After(1 * time.Second):
        fmt.Println("timeout waiting")
    }

    // --- Cancelling parent cancels children ---
    fmt.Println("\n=== Parent Cancels Children ===")
    parent, parentCancel := context.WithCancel(rootCtx)
    child1, _ := context.WithCancel(parent)
    child2, _ := context.WithTimeout(parent, 10*time.Second)
    defer child2.(*struct{ context.Context })(nil) // just illustrating

    parentCancel()   // cancel parent

    fmt.Println("parent done:", parent.Err())   // context canceled
    fmt.Println("child1 done:", child1.Err())   // context canceled — inherited!
    _ = child2
}
```

Output:
```
=== Timeout + Value ===
[trace-abc-123] getUser(42)
success: result: SELECT * FROM users WHERE id=42

=== Timeout Exceeded ===
[trace-abc-123] getUser(99) -- wait, rootCtx has no traceKey
error: getUser: queryDB: context deadline exceeded
is DeadlineExceeded: true

=== Manual Cancel ===
cancelling...
ctx3 done: context canceled

=== Parent Cancels Children ===
parent done: context canceled
child1 done: context canceled
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Storing context in a struct
```go
// ❌ Context should not be stored in a struct
type Handler struct {
    ctx context.Context
}

// ✅ Pass context as a parameter
func (h *Handler) Handle(ctx context.Context) error { ... }
```

### 2. Not calling cancel — goroutine/resource leak
```go
// ❌ Cancel never called if function returns early
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
result, err := doWork(ctx)
if err != nil { return err }   // cancel() never called!
cancel()

// ✅ Always defer cancel immediately
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
```

### 3. Using context values for required parameters
```go
// ❌ Wrong — db is a required dependency, not request-scoped data
ctx = context.WithValue(ctx, "db", db)

// ✅ Pass db explicitly
func DoWork(ctx context.Context, db *sql.DB) error { ... }
```

### 4. Using built-in types as context keys — collision risk
```go
// ❌ String key — any package can accidentally overwrite it
ctx = context.WithValue(ctx, "userID", 42)

// ✅ Unexported custom type — guaranteed unique
type contextKey string
const keyUserID contextKey = "userID"
ctx = context.WithValue(ctx, keyUserID, 42)
```

### 5. Ignoring context cancellation in long loops
```go
// ❌ Never checks if context is cancelled
for _, item := range hugeList {
    expensiveProcess(item)   // ignores ctx
}

// ✅ Check ctx.Done() periodically
for _, item := range hugeList {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    expensiveProcess(item)
}
```

### 6. Creating context.Background() in middleware instead of using r.Context()
```go
// ❌ Loses the request's existing deadline/cancellation
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()   // wrong — ignores request context!
    doWork(ctx)
}

// ✅ Always start from r.Context()
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    doWork(ctx)
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- `context.Background()` — root context, use at main/test/top of request
- `context.TODO()` — placeholder when you haven't decided the context yet
- `WithCancel` — manual cancel; `WithTimeout` — auto-cancel after duration; `WithDeadline` — auto-cancel at time
- **Always `defer cancel()`** immediately after `WithCancel`/`WithTimeout`/`WithDeadline`
- `ctx.Done()` — channel closed on cancellation; `ctx.Err()` — the reason
- `context.Canceled` — manually cancelled; `context.DeadlineExceeded` — timed out
- Cancelling a parent context cancels **all its children** automatically
- Context is the **first parameter** of any function that does I/O or long work
- **Never store context in a struct** — always pass as parameter
- `WithValue` keys must be **unexported custom types** — avoid collisions
- Use context values **only** for request-scoped data (trace IDs, auth) — not for function parameters
- `http.NewRequestWithContext(ctx, ...)` and `db.QueryContext(ctx, ...)` — always pass ctx to stdlib

---

## ❓ Interview Questions & Answers

---

**Q1: What is the `context` package used for?**

> The `context` package provides a standard way to carry cancellation signals, deadlines, and request-scoped values across API boundaries. It solves three problems: (1) cancellation — when a caller no longer needs a result (user cancelled, upstream timed out), all downstream work should stop; (2) deadlines — operations should automatically stop after a set duration to prevent resource exhaustion; (3) request-scoped values — data like trace IDs and auth info that should flow through the call stack without polluting every function signature.

---

**Q2: What is the difference between `context.Background()` and `context.TODO()`?**

> Both return empty, non-cancellable contexts, but they communicate intent. `context.Background()` is the real root — use it at the start of `main`, at the top of HTTP handlers (when there's no incoming context), and in tests. `context.TODO()` is a placeholder — use it when you know a function should accept a context but haven't wired it up yet, or when you're refactoring code to add context support. Both are functionally identical; the distinction is purely semantic for code readers.

---

**Q3: Why must you always call the cancel function returned by `WithCancel` / `WithTimeout`?**

> Not calling `cancel` causes a resource leak — the context and its associated goroutine (for timeout-based contexts) stay alive until the parent is cancelled or the program exits. Even if the operation finishes before the timeout, the timer goroutine keeps running. The pattern is always `ctx, cancel := context.WithTimeout(parent, d); defer cancel()` — deferring immediately ensures cancel is always called regardless of how the function returns.

---

**Q4: How do you check if a context was cancelled vs timed out?**

> `ctx.Err()` returns the reason: `context.Canceled` if the context was manually cancelled via `cancel()`, or `context.DeadlineExceeded` if a timeout or deadline was exceeded. You check with `errors.Is(err, context.DeadlineExceeded)`. Many standard library functions wrap context errors in their own errors, so always use `errors.Is` for the check rather than direct equality, since the error may be wrapped with `%w`.

---

**Q5: What is the rule for context value keys and why?**

> Context value keys should always be unexported custom types — never built-in types like `string` or `int`. If two packages both use the string `"userID"` as a key, they'd accidentally overwrite each other's values. Using an unexported type like `type contextKey string` in each package guarantees uniqueness — because unexported types from different packages are distinct types even if they have the same underlying value. This is a convention enforced by good practice, not the compiler.

---

**Q6: What is the rule "never store context in a struct" about?**

> Contexts are request-scoped — they carry cancellation and deadlines specific to one operation. Storing a context in a struct ties the struct's lifetime to a specific request, making it hard to reuse the struct for different requests with different deadlines. It also makes it unclear where cancellation can occur. The Go documentation explicitly states: "Do not store Contexts inside a struct type; instead, pass a Context explicitly to each function that needs it." This keeps data flow explicit and functions composable.

---

**Q7: How does context cancellation propagate through a tree of goroutines?**

> Contexts form a tree — child contexts derived from a parent inherit its cancellation. When a parent context is cancelled (either manually or by timeout), all children are cancelled simultaneously and automatically. This means if you have a request that spawns multiple goroutines, each with a child context derived from the request context, cancelling the request cancels all goroutines at once. Each goroutine is responsible for checking `ctx.Done()` and returning when cancelled — the runtime doesn't kill goroutines; it signals them via the Done channel.

---

*Next up: **Topic 3.3 — Advanced Concurrency Patterns** 🚀*
