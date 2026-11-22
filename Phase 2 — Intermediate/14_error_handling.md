# 2.2 — Error Handling

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.1 — Interfaces

---

## 📖 Go's Philosophy on Errors

Go treats errors as **regular values** — not exceptions. There's no `try/catch/finally`. Errors are returned from functions just like any other value, and the caller is responsible for checking them.

```go
result, err := someFunction()
if err != nil {
    // handle it
}
```

This is intentional. Go's designers believed that exceptions make control flow opaque and encourage ignoring errors. With explicit error returns, error handling is visible, mandatory, and part of the function's contract.

---

## 🔑 The `error` Interface

`error` is a built-in interface with a single method:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method satisfies the `error` interface. This is why errors are flexible — they can carry any data you want.

---

## 🏗️ Creating Errors

### Method 1: `errors.New` — simple static error

```go
import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidInput
    }
    // ...
    return nil, ErrNotFound
}
```

### Method 2: `fmt.Errorf` — formatted error with context

```go
import "fmt"

func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("findUser: invalid id %d", id)
    }
    return nil, fmt.Errorf("findUser: user %d not found", id)
}
```

### Method 3: Custom error type — carries structured data

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Message: "must be non-negative"}
    }
    if age > 150 {
        return &ValidationError{Field: "age", Message: "unrealistically large"}
    }
    return nil
}
```

---

## 🎁 Error Wrapping — `%w`

Go 1.13 introduced error wrapping — adding context to an error while preserving the original for inspection:

```go
func readConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("readConfig: %w", err)   // wrap with %w
    }
    // ...
    return nil
}
```

The `%w` verb wraps the error — the returned error contains the original as its "cause". This builds an error chain:

```
readConfig: open config.json: no such file or directory
     ↑ added context          ↑ original OS error
```

---

## 🔍 `errors.Is` — Checking the Error Chain

`errors.Is(err, target)` checks if `err` OR ANY error in its chain equals `target`:

```go
var ErrNotFound = errors.New("not found")

func getUser(id int) error {
    return fmt.Errorf("getUser: %w", ErrNotFound)   // wrapped
}

err := getUser(1)

// ❌ Direct comparison misses wrapped errors
fmt.Println(err == ErrNotFound)          // false — it's wrapped!

// ✅ errors.Is unwraps the chain
fmt.Println(errors.Is(err, ErrNotFound)) // true
```

Use `errors.Is` for **sentinel error comparison** — never `==` on possibly-wrapped errors.

---

## 🔍 `errors.As` — Extracting Error Types from the Chain

`errors.As(err, &target)` checks if any error in the chain matches the target type and sets it:

```go
err := validateAge(-5)

var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Field:", valErr.Field)     // age
    fmt.Println("Message:", valErr.Message) // must be non-negative
}
```

This works even when the error is wrapped:

```go
wrapped := fmt.Errorf("outer: %w", &ValidationError{Field: "age", Message: "bad"})

var valErr *ValidationError
if errors.As(wrapped, &valErr) {   // ✅ unwraps and finds ValidationError
    fmt.Println(valErr.Field)      // age
}
```

---

## 📋 Sentinel Errors — Package-Level Error Variables

Sentinel errors are package-level `var` declarations used as known, checkable error values:

```go
// Define at package level
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrTimeout      = errors.New("timeout")
)

// Return them
func fetchData(id int) ([]byte, error) {
    if id < 0 {
        return nil, ErrNotFound
    }
    // ...
}

// Check them
data, err := fetchData(-1)
if errors.Is(err, ErrNotFound) {
    // handle not found specifically
}
```

Standard library examples: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`.

---

## 🔄 Error Handling Patterns

### Pattern 1: Early return (most common)

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("processFile: %w", err)
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return fmt.Errorf("processFile: read: %w", err)
    }

    if err := process(data); err != nil {
        return fmt.Errorf("processFile: process: %w", err)
    }

    return nil
}
```

### Pattern 2: Collecting multiple errors

```go
type MultiError struct {
    Errors []error
}

func (m *MultiError) Error() string {
    msgs := make([]string, len(m.Errors))
    for i, e := range m.Errors {
        msgs[i] = e.Error()
    }
    return strings.Join(msgs, "; ")
}

func validateUser(u User) error {
    var errs MultiError
    if u.Name == "" {
        errs.Errors = append(errs.Errors, errors.New("name required"))
    }
    if u.Age < 0 {
        errs.Errors = append(errs.Errors, errors.New("age must be non-negative"))
    }
    if len(errs.Errors) > 0 {
        return &errs
    }
    return nil
}
```

### Pattern 3: errors.Join (Go 1.20+)

```go
err1 := errors.New("first error")
err2 := errors.New("second error")

combined := errors.Join(err1, err2)
fmt.Println(combined)
// first error
// second error

errors.Is(combined, err1)   // true
errors.Is(combined, err2)   // true
```

### Pattern 4: Panic and recover for truly unrecoverable situations

```go
// panic is NOT for normal error handling — use it for programmer errors
func mustParseURL(raw string) *url.URL {
    u, err := url.Parse(raw)
    if err != nil {
        panic(fmt.Sprintf("mustParseURL: invalid URL %q: %v", raw, err))
    }
    return u
}

// recover in defer — used in frameworks to catch panics
func safeExecute(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    fn()
    return nil
}
```

---

## 🏷️ Implementing the `error` Interface

Any struct with `Error() string` is an error:

```go
type HTTPError struct {
    Code    int
    Message string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func fetchUser(id int) (*User, error) {
    if id == 0 {
        return nil, &HTTPError{Code: 400, Message: "id cannot be zero"}
    }
    return nil, &HTTPError{Code: 404, Message: "user not found"}
}

func main() {
    _, err := fetchUser(0)

    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        fmt.Println("Status code:", httpErr.Code)   // 400
        fmt.Println("Message:", httpErr.Message)    // id cannot be zero
    }
}
```

---

## 💻 Complete Example

```go
package main

import (
    "errors"
    "fmt"
    "strconv"
)

// Sentinel errors
var (
    ErrNotFound    = errors.New("not found")
    ErrPermission  = errors.New("permission denied")
)

// Custom error type
type ParseError struct {
    Input string
    Cause error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse error for input %q: %v", e.Input, e.Cause)
}

func (e *ParseError) Unwrap() error {
    return e.Cause   // allows errors.Is/As to look through it
}

// Function that wraps errors
func parseAge(s string) (int, error) {
    n, err := strconv.Atoi(s)
    if err != nil {
        return 0, &ParseError{Input: s, Cause: err}
    }
    if n < 0 || n > 150 {
        return 0, fmt.Errorf("parseAge: age %d out of range [0, 150]", n)
    }
    return n, nil
}

// Function that wraps with context
func loadUserAge(userID int, rawAge string) (int, error) {
    age, err := parseAge(rawAge)
    if err != nil {
        return 0, fmt.Errorf("loadUserAge(id=%d): %w", userID, err)
    }
    return age, nil
}

func main() {
    // Success case
    age, err := loadUserAge(1, "30")
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Age:", age)   // Age: 30
    }

    // Error case — bad input
    _, err = loadUserAge(2, "abc")
    if err != nil {
        fmt.Println("\nError:", err)
        // loadUserAge(id=2): parse error for input "abc": strconv.Atoi: ...

        // Check with errors.As — unwraps the chain
        var parseErr *ParseError
        if errors.As(err, &parseErr) {
            fmt.Println("Bad input was:", parseErr.Input)
        }
    }

    // Sentinel error check
    _, err = loadUserAge(3, "999")
    if err != nil {
        fmt.Println("\nError:", err)
        // Not a ParseError — it's a range error string
    }

    // Demonstrate error chain
    base := errors.New("base error")
    wrapped1 := fmt.Errorf("layer 1: %w", base)
    wrapped2 := fmt.Errorf("layer 2: %w", wrapped1)

    fmt.Println("\nError chain:")
    fmt.Println(wrapped2)
    fmt.Println("Is base?", errors.Is(wrapped2, base))   // true — traverses chain
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Returning typed nil as error
```go
func getError() error {
    var err *MyError = nil
    return err   // ❌ non-nil interface! err != nil is true for caller
}
// ✅ Fix
func getError() error {
    return nil
}
```

### 2. Using `==` instead of `errors.Is`
```go
err := fmt.Errorf("wrap: %w", ErrNotFound)
err == ErrNotFound          // ❌ false — wrapped!
errors.Is(err, ErrNotFound) // ✅ true
```

### 3. Using `errors.As` with non-pointer target
```go
var e ValidationError    // ❌ must be a pointer to pointer
errors.As(err, &e)       // wrong — e must be *ValidationError

var e *ValidationError   // ✅
errors.As(err, &e)
```

### 4. Ignoring errors
```go
os.Remove("file.txt")     // ❌ error silently ignored
_ = os.Remove("file.txt") // ⚠️  intentionally ignored — at least it's explicit
if err := os.Remove("file.txt"); err != nil { ... }  // ✅
```

### 5. Wrapping errors multiple times with the same message
```go
// ❌ Redundant — double-wrapping
return fmt.Errorf("processFile: %w", fmt.Errorf("processFile: %w", err))

// ✅ Add new context at each layer
return fmt.Errorf("processFile: %w", err)
```

---

## 🎯 Must-Know Things (Quick Reference)

- `error` is a built-in interface: `type error interface { Error() string }`
- Go returns errors as values — **no exceptions**, no `try/catch`
- `errors.New("msg")` — simple static error
- `fmt.Errorf("context: %w", err)` — wrap with context, preserves chain
- `errors.Is(err, target)` — checks entire error chain for a match
- `errors.As(err, &target)` — extracts a specific error type from the chain
- **Sentinel errors** — package-level `var Err... = errors.New(...)` for known conditions
- **Custom error types** — implement `Error() string`, add `Unwrap() error` for chain support
- Never `return typedNilPointer` as an `error` — always `return nil`
- Use `errors.Join` (Go 1.20+) to combine multiple errors
- `panic` is NOT for normal error handling — only for unrecoverable programmer errors

---

## ❓ Interview Questions & Answers

---

**Q1: Why does Go use error values instead of exceptions?**

> Go's designers believe exceptions make error handling invisible — you can call a function without knowing it might throw, and errors can bypass large swaths of code silently. With error values, every function that can fail says so in its signature, and the caller must explicitly decide what to do. This makes error handling visible, reviewable, and part of the API contract. The downside is verbosity — the `if err != nil` pattern repeats often — but the tradeoff is clarity.

---

**Q2: What is the difference between `errors.Is` and `errors.As`?**

> `errors.Is(err, target)` checks if `err` or any error in its unwrap chain is equal to `target`. It's used for sentinel error comparison — checking if a specific known error occurred. `errors.As(err, &target)` checks if any error in the chain has a specific type and if so, sets `target` to that error value. It's used when you need to access fields of a custom error type. Both traverse the error chain via `Unwrap()`.

---

**Q3: What does `fmt.Errorf` with `%w` do?**

> `%w` wraps an error — the resulting error contains the original as its cause, accessible via `errors.Unwrap()`. This allows adding context at each layer while preserving the original error for inspection. For example: `fmt.Errorf("readConfig: %w", err)` adds "readConfig:" context while keeping the underlying OS error intact. `errors.Is` and `errors.As` traverse this chain automatically.

---

**Q4: What are sentinel errors and when should you use them?**

> Sentinel errors are package-level error variables — `var ErrNotFound = errors.New("not found")`. They represent known, expected error conditions that callers might want to handle specifically. Use them when callers need to distinguish between different error cases and take different actions. Examples from the stdlib: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`. Callers check with `errors.Is(err, ErrNotFound)` — never `==` in case the error is wrapped.

---

**Q5: How do you create a custom error type that supports the error chain?**

> Define a struct implementing `Error() string`, and add an `Unwrap() error` method that returns the wrapped cause. The `Unwrap` method is what `errors.Is` and `errors.As` use to traverse the chain. For example:
> ```go
> type MyError struct { Msg string; Cause error }
> func (e *MyError) Error() string { return e.Msg }
> func (e *MyError) Unwrap() error { return e.Cause }
> ```
> Now wrapping and unwrapping work automatically with the standard library functions.

---

**Q6: When should you use `panic` vs returning an error?**

> Return an error for expected, recoverable failure conditions — file not found, invalid input, network timeout. Use `panic` only for programmer errors — conditions that represent a bug in the program itself, like violating an invariant or a truly impossible state. Library code should almost never panic. The `Must...` naming convention (e.g., `template.Must`, `regexp.MustCompile`) signals functions that panic on error — they're meant for initialization-time use where panicking is acceptable.

---

**Q7: What is the typed nil error gotcha?**

> When a function returns a typed nil pointer as an `error` interface, the interface value is NOT nil — its type field is set even though the data is nil. So `err != nil` evaluates to true even though you intended to return "no error". This happens when you have `var err *MyError = nil; return err`. The fix is always `return nil` directly when there's no error — never return a typed nil through an interface.

---

*Next up: **Topic 2.3 — Packages & Modules** 🚀*
