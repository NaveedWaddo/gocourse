# 2.8 — Defer, Panic & Recover

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.7 — sync Package

---

## 📖 Overview

Three closely related keywords that handle cleanup, unexpected failures, and recovery:

- **`defer`** — schedule a function call to run when the surrounding function returns
- **`panic`** — stop normal execution and unwind the stack
- **`recover`** — catch a panic and resume normal execution

---

## ⏸️ `defer` — Scheduled Cleanup

`defer` schedules a function call to execute **when the surrounding function returns** — regardless of how it returns (normal return, early return, or panic).

```go
func main() {
    defer fmt.Println("world")   // runs last
    fmt.Println("hello")         // runs first
}
// Output:
// hello
// world
```

### The most important use — resource cleanup

```go
func readFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()   // ✅ guaranteed to run, no matter what happens below

    data, err := io.ReadAll(f)
    if err != nil {
        return "", err   // f.Close() still runs!
    }
    return string(data), nil
}
```

Without `defer`, you'd need `f.Close()` before every `return` — easy to forget.

Other common cleanup patterns:
```go
mu.Lock()
defer mu.Unlock()   // always unlock

rows, _ := db.Query(...)
defer rows.Close()  // always close DB rows

wg.Add(1)
defer wg.Done()     // always signal done
```

---

## 📚 Defer is LIFO — Last In, First Out

Multiple defers execute in **reverse order** — like a stack:

```go
func main() {
    defer fmt.Println("first deferred — runs last")
    defer fmt.Println("second deferred — runs second")
    defer fmt.Println("third deferred — runs first")
    fmt.Println("function body")
}
// Output:
// function body
// third deferred — runs first
// second deferred — runs second
// first deferred — runs last
```

This mirrors resource acquisition order — if you open A then B, you close B then A.

---

## 📸 Defer Arguments Are Evaluated Immediately

The arguments to a deferred function are evaluated **when the defer statement is evaluated**, not when it executes:

```go
func main() {
    x := 10
    defer fmt.Println("deferred x:", x)   // x captured as 10 NOW
    x = 20
    fmt.Println("current x:", x)
}
// Output:
// current x: 20
// deferred x: 10   ← captured value at defer time, not at execution time
```

But if you defer a **closure** (anonymous function), it captures variables by reference:

```go
func main() {
    x := 10
    defer func() {
        fmt.Println("closure x:", x)   // reads x at execution time
    }()
    x = 20
    fmt.Println("current x:", x)
}
// Output:
// current x: 20
// closure x: 20   ← sees the updated value
```

---

## 🔁 Defer with Named Return Values

A deferred closure can read and **modify** named return values:

```go
func divide(a, b float64) (result float64, err error) {
    defer func() {
        if err != nil {
            // wrap the error with context
            err = fmt.Errorf("divide: %w", err)
        }
    }()

    if b == 0 {
        err = errors.New("division by zero")
        return
    }
    result = a / b
    return
}
```

This is a powerful pattern for wrapping errors in cleanup code.

---

## 💥 `panic` — Stopping Execution

`panic` immediately stops the current function, runs all deferred functions in the call stack, then crashes the program with a message.

```go
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("expected positive, got %d", n))
    }
    return n
}

func main() {
    mustPositive(-1)
    // Output:
    // goroutine 1 [running]:
    // main.mustPositive(...)
    // panic: expected positive, got -1
}
```

### When to use panic

```go
// ✅ Programmer errors — things that should never happen
func (s *Stack) Pop() int {
    if len(s.items) == 0 {
        panic("pop from empty stack")   // caller violated contract
    }
    // ...
}

// ✅ Init-time failures — unrecoverable setup errors
func mustParseTemplate(text string) *template.Template {
    t, err := template.New("").Parse(text)
    if err != nil {
        panic(err)   // template syntax error — program can't start correctly
    }
    return t
}

// ✅ The Must pattern — common in standard library
var tmpl = template.Must(template.ParseFiles("index.html"))
var re   = regexp.MustCompile(`^\d+$`)
```

### When NOT to use panic

```go
// ❌ Expected errors — use error return instead
func readFile(path string) {
    f, err := os.Open(path)
    if err != nil {
        panic(err)   // file not found is not a programmer error!
    }
}

// ✅ Return the error
func readFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}
```

> **Rule:** Use `panic` for programmer errors (invariant violations, impossible states). Use `error` for expected, recoverable failures (I/O, validation, network).

---

## 🛟 `recover` — Catching a Panic

`recover` stops a panic and returns the value passed to `panic`. It only works **inside a deferred function**:

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered panic: %v", r)
        }
    }()

    return a / b, nil   // panics if b == 0 (integer division by zero)
}

func main() {
    result, err := safeDiv(10, 2)
    fmt.Println(result, err)   // 5 <nil>

    result, err = safeDiv(10, 0)
    fmt.Println(result, err)   // 0 recovered panic: runtime error: integer divide by zero
}
```

### Rules about `recover`

```go
// ✅ recover works only inside a deferred function
defer func() {
    if r := recover(); r != nil { ... }
}()

// ❌ recover outside defer does nothing (returns nil)
r := recover()   // always nil here — not inside defer
```

### recover returns the panic value

```go
panic("something went wrong")   // recover returns "something went wrong" (string)
panic(42)                       // recover returns 42 (int)
panic(errors.New("oops"))       // recover returns an error value
```

---

## 🏗️ Panic Propagation — Unwinding the Stack

When a panic occurs, Go unwinds the call stack, running deferred functions at each level. If no `recover` is encountered, the program crashes:

```go
func c() {
    panic("disaster!")
}

func b() {
    defer fmt.Println("b's defer")   // runs during unwind
    c()
}

func a() {
    defer fmt.Println("a's defer")   // runs during unwind
    b()
}

func main() {
    defer fmt.Println("main's defer")   // runs during unwind
    a()
}
// Output:
// b's defer
// a's defer
// main's defer
// goroutine 1 [running]:
// panic: disaster!
```

---

## 🌐 Real-World Pattern — HTTP Middleware Recovery

The most common real-world use of `recover` — catching panics in HTTP handlers so one bad request doesn't crash the whole server:

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log the panic with stack trace
                log.Printf("panic recovered: %v\n%s", err, debug.Stack())
                // Return 500 to client instead of crashing
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## 💻 Complete Example

```go
package main

import (
    "errors"
    "fmt"
)

// --- defer for cleanup ---
func withCleanup() {
    fmt.Println("acquiring resources")
    defer fmt.Println("resource 1 released")   // runs third
    defer fmt.Println("resource 2 released")   // runs second
    defer fmt.Println("resource 3 released")   // runs first
    fmt.Println("doing work")
}

// --- defer modifying named returns ---
func safeSqrt(x float64) (result float64, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("safeSqrt panicked: %v", r)
            result = 0
        }
    }()

    if x < 0 {
        panic(fmt.Sprintf("cannot sqrt negative number: %v", x))
    }
    // simplified sqrt
    result = x * x
    return
}

// --- Must pattern ---
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("must be positive, got %d", n))
    }
    return n
}

// --- safe wrapper ---
func safeExecute(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            switch v := r.(type) {
            case error:
                err = v
            default:
                err = fmt.Errorf("panic: %v", v)
            }
        }
    }()
    fn()
    return nil
}

func main() {
    // Defer LIFO
    fmt.Println("=== Defer Cleanup ===")
    withCleanup()

    // recover from panic
    fmt.Println("\n=== Safe Sqrt ===")
    r1, err := safeSqrt(4)
    fmt.Printf("sqrt(4) = %.1f, err = %v\n", r1, err)

    r2, err := safeSqrt(-1)
    fmt.Printf("sqrt(-1) = %.1f, err = %v\n", r2, err)

    // safeExecute wrapper
    fmt.Println("\n=== Safe Execute ===")
    err = safeExecute(func() {
        fmt.Println("running safe function")
    })
    fmt.Println("err:", err)   // <nil>

    err = safeExecute(func() {
        panic(errors.New("something broke"))
    })
    fmt.Println("err:", err)   // panic: something broke

    // defer argument evaluation
    fmt.Println("\n=== Defer Argument Evaluation ===")
    x := 1
    defer fmt.Println("deferred x:", x)   // captures x=1 now
    x = 100
    fmt.Println("current x:", x)          // prints 100
    // deferred prints 1
}
```

Output:
```
=== Defer Cleanup ===
acquiring resources
doing work
resource 3 released
resource 2 released
resource 1 released

=== Safe Sqrt ===
sqrt(4) = 16.0, err = <nil>
sqrt(-1) = 0.0, err = safeSqrt panicked: cannot sqrt negative number: -1

=== Safe Execute ===
running safe function
err: <nil>
err: panic: something broke

=== Defer Argument Evaluation ===
current x: 100
deferred x: 1
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Defer in a loop — defers pile up
```go
// ❌ All files stay open until function returns
for _, path := range paths {
    f, _ := os.Open(path)
    defer f.Close()   // all defers run at function exit, not loop iteration!
}

// ✅ Use a closure to scope each defer
for _, path := range paths {
    func() {
        f, _ := os.Open(path)
        defer f.Close()   // runs when inner function returns (each iteration)
    }()
}
```

### 2. Ignoring defer return values
```go
// ❌ Close() can return an error — silently ignored with defer
defer f.Close()

// ✅ If you care about Close errors (e.g., writing to file):
defer func() {
    if err := f.Close(); err != nil {
        log.Println("close error:", err)
    }
}()
```

### 3. recover only works in direct defer
```go
// ❌ recover called from non-deferred function — doesn't work
func helper() {
    recover()   // returns nil, doesn't stop panic
}

defer helper()   // ❌ still doesn't work — recover must be DIRECTLY in deferred func

// ✅
defer func() {
    recover()   // works
}()
```

### 4. Using panic for normal errors
```go
// ❌ Wrong — file not found is expected
func load(path string) []byte {
    data, err := os.ReadFile(path)
    if err != nil { panic(err) }
    return data
}

// ✅ Return error
func load(path string) ([]byte, error) {
    return os.ReadFile(path)
}
```

### 5. Defer has a small performance cost in hot paths
```go
// In extremely hot loops, defer adds ~20-50ns overhead
// For most code this is irrelevant, but in tight inner loops:
func hotPath() {
    mu.Lock()
    // Don't defer here if this runs millions of times per second
    count++
    mu.Unlock()
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- `defer` runs when the **surrounding function returns** — normal, early, or panic
- Multiple defers execute in **LIFO order** (last in, first out)
- Defer arguments are evaluated **immediately** — closures capture by reference
- `defer` + named returns = can modify return values in cleanup
- `panic` unwinds the stack running deferred functions at each level
- Use `panic` for **programmer errors** only — never for expected failures
- `recover` must be called **directly inside a deferred function** — nowhere else
- `recover` returns the panic value — check for `nil` to know if panic occurred
- The **Must pattern** — `Must(fn())` panics on error, for init-time use
- Defer in loops defers to **function exit**, not loop iteration — wrap in closure

---

## ❓ Interview Questions & Answers

---

**Q1: What does `defer` do and when does it execute?**

> `defer` schedules a function call to execute when the surrounding function returns — whether that's a normal return, an early return, or a panic. The deferred function runs after the return value is set but before the function actually returns to its caller. Multiple defers in one function execute in LIFO order. `defer` is most commonly used for cleanup — closing files, unlocking mutexes, closing database connections — guaranteeing cleanup happens regardless of how the function exits.

---

**Q2: When are defer's arguments evaluated?**

> Arguments to a deferred function are evaluated immediately when the `defer` statement executes — not when the deferred function runs. So `defer fmt.Println(x)` captures the value of `x` at the defer line. However, if you defer an anonymous function (closure), it captures variables by reference and sees their values at execution time. This distinction matters: use closures when you need the latest values, use direct calls when you want values captured at defer time.

---

**Q3: What is the difference between `panic` and returning an error?**

> Return an `error` for expected, recoverable failure conditions — file not found, invalid input, network timeout. These are situations the caller should anticipate and handle. Use `panic` for programmer errors — violated invariants, impossible states, broken contracts that indicate a bug. Panics propagate up the stack and crash the program if not recovered. The rule: if a reasonable caller could be expected to handle it, return an error; if it represents a bug that should never happen, panic.

---

**Q4: How does `recover` work and where must it be called?**

> `recover` stops a panic and returns the value passed to `panic`. It must be called **directly inside a deferred function** — not in a function called by a deferred function, and not outside defer. If called outside a deferred context, it returns nil and has no effect. The typical pattern is `defer func() { if r := recover(); r != nil { /* handle */ } }()`. After `recover`, execution does not resume where the panic occurred — the function that deferred the recovery returns normally.

---

**Q5: What is the defer-in-loop gotcha?**

> Defers inside a loop don't execute at each iteration — they all execute when the enclosing function returns. So if you open files in a loop and defer their Close calls, all files remain open until the function exits, potentially exhausting file descriptors. The fix is to wrap the loop body in an immediately-invoked anonymous function, so each defer is scoped to that inner function and runs at the end of each iteration.

---

**Q6: What is the Must pattern in Go?**

> The Must pattern is a convention for wrapping functions that return `(T, error)` with a version that panics on error instead. For example, `regexp.MustCompile` panics if the regex is invalid. This is used for initialization-time operations where an error represents a bug (invalid hardcoded regex, malformed template) and recovery isn't meaningful. The name `Must` signals to callers that the function will panic on failure, making it appropriate only at package init or program startup, never in request-handling code.

---

**Q7: What happens to deferred functions during a panic?**

> When a panic occurs, Go stops executing the current function and begins unwinding the call stack. At each stack frame, deferred functions are executed in LIFO order before the frame is discarded. This unwind continues up the stack until either a `recover` is encountered in a deferred function (which stops the panic) or the goroutine's stack is fully unwound (which crashes the program with a panic message and stack trace). This is why `defer` provides cleanup guarantees even in the presence of panics.

---

*Next up: **Topic 2.9 — Stringer & fmt Interface** 🚀*
