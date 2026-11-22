# 3.6 — Testing in Go

> **Phase:** 3 — Advanced | **Difficulty:** Intermediate–Advanced  
> **Prerequisites:** Topic 3.5 — `unsafe` Package

---

## 📖 Overview

Go has first-class testing built into the toolchain. No third-party framework is needed for basic testing — just the `testing` package and `go test`.

Core testing features:
- **Unit tests** — `TestXxx` functions
- **Table-driven tests** — idiomatic Go test style
- **Subtests** — `t.Run` for structured grouping
- **Benchmarks** — `BenchmarkXxx` for performance
- **Examples** — `ExampleXxx` for documentation + testing
- **Fuzz testing** — `FuzzXxx` for property-based testing (Go 1.18+)
- **Test helpers** — `t.Helper()`, `testify`, `gomock`

---

## 🧪 Basic Tests

Test files end in `_test.go` and are excluded from normal builds:

```go
// math.go
package math

func Add(a, b int) int { return a + b }
func Divide(a, b float64) (float64, error) {
    if b == 0 { return 0, errors.New("division by zero") }
    return a / b, nil
}
```

```go
// math_test.go
package math   // same package — access unexported identifiers

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

### Test functions — the `testing.T` API

```go
t.Error(args...)     // log failure, continue test
t.Errorf(fmt, ...)   // log formatted failure, continue test
t.Fatal(args...)     // log failure, STOP test immediately
t.Fatalf(fmt, ...)   // log formatted failure, STOP test immediately
t.Log(args...)       // log message (only shown on failure or -v)
t.Logf(fmt, ...)     // log formatted message
t.Skip(args...)      // skip this test
t.Skipf(fmt, ...)    // skip with message
t.Parallel()         // run this test in parallel with other parallel tests
t.Helper()           // mark caller as a test helper (improves error line numbers)
```

---

## 📋 Table-Driven Tests — The Idiomatic Go Way

Instead of one function per case, define a table of inputs/outputs:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
        {"mixed", -5, 10, 5},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := Add(tc.a, tc.b)
            if got != tc.want {
                t.Errorf("Add(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
            }
        })
    }
}

func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"fractional", 7, 2, 3.5, false},
        {"by zero", 10, 0, 0, true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Divide(tc.a, tc.b)
            if (err != nil) != tc.wantErr {
                t.Fatalf("Divide(%v, %v) error = %v; wantErr %v",
                    tc.a, tc.b, err, tc.wantErr)
            }
            if !tc.wantErr && got != tc.want {
                t.Errorf("Divide(%v, %v) = %v; want %v",
                    tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

`t.Run` creates a **subtest** — runnable individually: `go test -run TestDivide/by_zero`.

---

## 🔧 Test Helpers

Mark helper functions with `t.Helper()` so error lines point to the call site:

```go
// assertEqual is a reusable helper
func assertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()   // ← without this, errors point to this line, not the caller
    if got != want {
        t.Errorf("got %v; want %v", got, want)
    }
}

func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

// Usage
func TestSomething(t *testing.T) {
    result, err := SomeFunc()
    assertNoError(t, err)          // error line points HERE, not to assertNoError
    assertEqual(t, result, 42)
}
```

---

## 🏗️ Test Setup and Teardown

### `TestMain` — package-level setup/teardown

```go
func TestMain(m *testing.M) {
    // Setup — runs before any test in the package
    db := setupTestDB()

    // Run all tests
    code := m.Run()

    // Teardown — runs after all tests
    db.Close()

    os.Exit(code)
}
```

### `t.Cleanup` — per-test teardown

```go
func TestWithCleanup(t *testing.T) {
    // t.Cleanup runs when the test (and all subtests) finish
    tmpDir := t.TempDir()   // automatically cleaned up
    t.Cleanup(func() {
        fmt.Println("cleaning up after test")
    })

    // Use tmpDir...
}
```

### Useful `testing.T` utilities

```go
t.TempDir()         // creates a temp dir, cleaned up automatically
t.Setenv(k, v)      // sets env var, restored after test
t.Chdir(dir)        // changes working dir, restored after test (Go 1.24+)
```

---

## ⚡ Benchmarks

Benchmarks measure performance. They run with `go test -bench=.`:

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {   // b.N is adjusted by the framework
        Add(2, 3)
    }
}

func BenchmarkDivide(b *testing.B) {
    b.ResetTimer()   // exclude setup time
    for i := 0; i < b.N; i++ {
        Divide(10, 3)
    }
}
```

### Running benchmarks

```bash
go test -bench=.               # run all benchmarks
go test -bench=BenchmarkAdd    # run specific benchmark
go test -bench=. -benchmem     # include memory allocation stats
go test -bench=. -count=5      # run each benchmark 5 times
go test -bench=. -benchtime=5s # run for 5 seconds instead of default 1s
```

### Benchmark output

```
BenchmarkAdd-8          1000000000    0.2390 ns/op
BenchmarkDivide-8        500000000    2.123  ns/op    0 B/op    0 allocs/op
```

- `BenchmarkAdd-8` — ran on 8 CPUs
- `1000000000` — b.N iterations
- `0.2390 ns/op` — nanoseconds per operation
- `0 B/op` — bytes allocated per operation (with `-benchmem`)
- `0 allocs/op` — heap allocations per operation

### Benchmark best practices

```go
func BenchmarkStringConcat(b *testing.B) {
    // Setup outside the loop
    prefix := "hello"
    b.ResetTimer()   // don't count setup

    for i := 0; i < b.N; i++ {
        _ = prefix + " world"   // assign to _ to prevent dead-code elimination
    }
}

// Parallel benchmark
func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Add(2, 3)
        }
    })
}
```

---

## 📖 Examples

Example functions serve as both documentation and tests:

```go
func ExampleAdd() {
    fmt.Println(Add(2, 3))
    // Output:
    // 5
}

func ExampleDivide() {
    result, err := Divide(10, 2)
    fmt.Println(result, err)
    // Output:
    // 5 <nil>
}
```

- Run with `go test`
- Shown in `go doc` / `pkg.go.dev`
- The `// Output:` comment is the expected output — test fails if output differs

---

## 🎲 Fuzz Testing (Go 1.18+)

Fuzz tests generate random inputs to find edge cases:

```go
func FuzzDivide(f *testing.F) {
    // Seed corpus — known inputs to start from
    f.Add(10.0, 2.0)
    f.Add(-5.0, 3.0)
    f.Add(0.0, 1.0)

    f.Fuzz(func(t *testing.T, a, b float64) {
        result, err := Divide(a, b)
        if err != nil {
            return   // division by zero is expected
        }
        // Invariant: result * b should equal a (approximately)
        if b != 0 && abs(result*b-a) > 1e-9 {
            t.Errorf("Divide(%v, %v) = %v; result*b = %v", a, b, result, result*b)
        }
    })
}
```

```bash
go test -fuzz=FuzzDivide          # run fuzzer (until failure or Ctrl-C)
go test -fuzz=FuzzDivide -fuzztime=30s  # fuzz for 30 seconds
go test                           # runs seed corpus only (regression)
```

Failed fuzz inputs are saved in `testdata/fuzz/` and replayed on future `go test` runs.

---

## 🔌 Interfaces for Testability

Design code against interfaces so you can substitute fakes in tests:

```go
// Production code depends on interface, not concrete type
type UserStore interface {
    GetUser(ctx context.Context, id int) (*User, error)
    SaveUser(ctx context.Context, u *User) error
}

type UserService struct {
    store UserStore
}

func NewUserService(store UserStore) *UserService {
    return &UserService{store: store}
}

// Test fake — implements UserStore
type fakeStore struct {
    users map[int]*User
}

func (f *fakeStore) GetUser(_ context.Context, id int) (*User, error) {
    u, ok := f.users[id]
    if !ok { return nil, fmt.Errorf("user %d not found", id) }
    return u, nil
}

func (f *fakeStore) SaveUser(_ context.Context, u *User) error {
    f.users[u.ID] = u
    return nil
}

func TestUserService(t *testing.T) {
    store := &fakeStore{users: map[int]*User{
        1: {ID: 1, Name: "Alice"},
    }}
    svc := NewUserService(store)
    user, err := svc.GetUser(context.Background(), 1)
    if err != nil { t.Fatal(err) }
    if user.Name != "Alice" {
        t.Errorf("got %q; want %q", user.Name, "Alice")
    }
}
```

---

## 🧪 `testify` — The Popular Assertion Library

While not in the standard library, `testify` is used in the vast majority of Go projects:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWithTestify(t *testing.T) {
    result, err := Divide(10, 2)

    require.NoError(t, err)          // like t.Fatal — stops test on failure
    assert.Equal(t, 5.0, result)     // like t.Error — continues on failure
    assert.InDelta(t, 5.0, result, 0.001)

    // Collections
    assert.Contains(t, []int{1, 2, 3}, 2)
    assert.Len(t, []string{"a", "b"}, 2)
    assert.Empty(t, []int{})

    // Error checking
    _, err = Divide(10, 0)
    assert.Error(t, err)
    assert.EqualError(t, err, "division by zero")
}
```

`require` vs `assert`:
- `require` — test stops immediately on failure (like `t.Fatal`)
- `assert` — test continues on failure (like `t.Error`)

---

## 💻 Complete Example

```go
// calculator.go
package calculator

import (
    "errors"
    "fmt"
    "math"
)

var ErrDivisionByZero = errors.New("division by zero")
var ErrNegativeSqrt   = errors.New("cannot sqrt negative number")

func Add(a, b float64) float64 { return a + b }

func Divide(a, b float64) (float64, error) {
    if b == 0 { return 0, ErrDivisionByZero }
    return a / b, nil
}

func Sqrt(x float64) (float64, error) {
    if x < 0 { return 0, fmt.Errorf("Sqrt(%v): %w", x, ErrNegativeSqrt) }
    return math.Sqrt(x), nil
}
```

```go
// calculator_test.go
package calculator

import (
    "errors"
    "math"
    "testing"
)

// --- Table-driven tests ---
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr error
    }{
        {"10/2", 10, 2, 5, nil},
        {"7/2", 7, 2, 3.5, nil},
        {"div zero", 10, 0, 0, ErrDivisionByZero},
        {"negative/positive", -6, 3, -2, nil},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Divide(tc.a, tc.b)
            if !errors.Is(err, tc.wantErr) {
                t.Fatalf("error = %v; wantErr = %v", err, tc.wantErr)
            }
            if tc.wantErr == nil && math.Abs(got-tc.want) > 1e-9 {
                t.Errorf("got %v; want %v", got, tc.want)
            }
        })
    }
}

func TestSqrt(t *testing.T) {
    tests := []struct {
        name    string
        x       float64
        want    float64
        wantErr bool
    }{
        {"perfect square", 4, 2, false},
        {"irrational", 2, math.Sqrt(2), false},
        {"zero", 0, 0, false},
        {"negative", -1, 0, true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Sqrt(tc.x)
            if (err != nil) != tc.wantErr {
                t.Fatalf("Sqrt(%v) error = %v; wantErr = %v", tc.x, err, tc.wantErr)
            }
            if !tc.wantErr && math.Abs(got-tc.want) > 1e-9 {
                t.Errorf("Sqrt(%v) = %v; want %v", tc.x, got, tc.want)
            }
        })
    }
}

// --- Benchmark ---
func BenchmarkDivide(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Divide(10, 3)
    }
}

func BenchmarkSqrt(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        Sqrt(2.0)
    }
}

// --- Example (also serves as documentation) ---
func ExampleDivide() {
    result, err := Divide(10, 2)
    fmt.Println(result, err)
    // Output:
    // 5 <nil>
}

// --- Fuzz ---
func FuzzSqrt(f *testing.F) {
    f.Add(0.0)
    f.Add(4.0)
    f.Add(100.0)

    f.Fuzz(func(t *testing.T, x float64) {
        result, err := Sqrt(x)
        if x < 0 {
            if err == nil { t.Errorf("Sqrt(%v): expected error for negative input", x) }
            return
        }
        if err != nil { t.Fatalf("Sqrt(%v): unexpected error: %v", x, err) }
        // Invariant: result^2 ≈ x
        if x > 0 && math.Abs(result*result-x)/x > 1e-9 {
            t.Errorf("Sqrt(%v) = %v; result^2 = %v", x, result, result*result)
        }
    })
}
```

### Running tests

```bash
go test ./...                    # run all tests in all packages
go test -v ./...                 # verbose — show each test name
go test -run TestDivide          # run tests matching pattern
go test -run TestDivide/div_zero # run specific subtest
go test -race ./...              # run with race detector
go test -cover ./...             # show coverage summary
go test -coverprofile=cov.out && go tool cover -html=cov.out  # HTML coverage report
go test -bench=. -benchmem       # run benchmarks with memory stats
go test -fuzz=FuzzSqrt           # run fuzz testing
go test -count=1                 # disable test caching
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Using `t.Error` when you should use `t.Fatal`
```go
// ❌ Test continues after nil pointer — next line panics
user, err := GetUser(id)
t.Error("expected no error:", err)   // test continues!
fmt.Println(user.Name)               // PANIC — user is nil

// ✅ Use t.Fatal for conditions that make further testing impossible
user, err := GetUser(id)
if err != nil { t.Fatalf("unexpected error: %v", err) }
fmt.Println(user.Name)   // safe
```

### 2. Forgetting `t.Parallel()` race conditions
```go
for _, tc := range tests {
    tc := tc   // ✅ capture loop var — REQUIRED with t.Parallel()
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()
        // use tc here
    })
}
```

### 3. Test pollution — shared global state
```go
// ❌ Tests modify global state — order-dependent failures
var db *Database
func init() { db = connectDB() }

// ✅ Each test creates its own state
func TestSomething(t *testing.T) {
    db := newTestDB(t)
    t.Cleanup(func() { db.Close() })
}
```

### 4. Comparing floats with ==
```go
// ❌ Floating point imprecision
if got == 3.14 { ... }

// ✅ Use tolerance
if math.Abs(got - 3.14) > 1e-9 { t.Errorf(...) }
```

### 5. Benchmark not preventing compiler optimizations
```go
var result float64

func BenchmarkSqrt(b *testing.B) {
    var r float64
    for i := 0; i < b.N; i++ {
        r = math.Sqrt(2.0)   // compiler may optimize away if r is unused
    }
    result = r   // ✅ assign to package-level var to prevent elimination
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- Test files: `_test.go` suffix; functions: `TestXxx(t *testing.T)`
- `t.Error`/`t.Errorf` — fail and continue; `t.Fatal`/`t.Fatalf` — fail and stop
- **Table-driven tests** with `t.Run(name, func)` — idiomatic Go testing
- `t.Helper()` — marks helper functions so error lines point to call site
- `t.Parallel()` — run test in parallel; always capture loop var with `tc := tc`
- `t.Cleanup(fn)` — teardown after test; `t.TempDir()` — auto-cleaned temp dir
- `TestMain(m *testing.M)` — package-level setup/teardown; call `m.Run()`
- `BenchmarkXxx(b *testing.B)` — loop `b.N` times; `b.ResetTimer()` after setup
- `ExampleXxx()` — documentation + test via `// Output:` comment
- `FuzzXxx(f *testing.F)` — fuzz testing; seed with `f.Add()`
- `go test -race` — **always run with race detector in CI**
- `go test -cover` — coverage; `-coverprofile` for HTML report
- `-run TestXxx/subtest` — run specific subtests with `/` separator
- Design code with interfaces so dependencies can be replaced with fakes

---

## ❓ Interview Questions & Answers

---

**Q1: What is the table-driven test pattern and why is it preferred in Go?**

> Table-driven testing defines test cases as a slice of structs — each with a name, inputs, and expected outputs — then iterates with `t.Run` to create a subtest for each case. It's preferred because: adding a new test case is a single line, all cases share the same test logic (no duplication), subtests are independently runnable and identifiable by name, and failures show exactly which case failed. The Go standard library itself uses this pattern extensively. It makes tests read like a specification of the function's behavior.

---

**Q2: What is the difference between `t.Error` and `t.Fatal`?**

> Both mark a test as failed, but `t.Fatal` also stops the test function immediately (like `return`), while `t.Error` lets the test continue. Use `t.Fatal` when a failure makes subsequent code meaningless or dangerous — like when a nil pointer would cause a panic on the next line. Use `t.Error` when you want to report multiple independent failures in one run — like checking several fields of a struct. `t.Fatalf` and `t.Errorf` are the formatted variants.

---

**Q3: What does `t.Helper()` do and why should test helpers use it?**

> `t.Helper()` marks the current function as a test helper. When a test fails inside a helper, Go normally reports the error at the line inside the helper — not at the call site, making it hard to find which test case failed. After calling `t.Helper()`, failure lines are reported at the call site instead. Any reusable assertion function should call `t.Helper()` as its first line. Libraries like `testify` do this internally.

---

**Q4: How do benchmarks work in Go?**

> Benchmark functions are named `BenchmarkXxx` and receive `*testing.B`. The framework runs the loop body `b.N` times, adjusting `b.N` until results are statistically stable. You run them with `go test -bench=.`. Use `b.ResetTimer()` to exclude setup time. Use `b.ReportAllocs()` or `-benchmem` to include memory allocation counts. Assign results to a package-level variable to prevent the compiler from optimizing away the benchmarked code. Benchmarks are useful for measuring the impact of code changes and comparing implementation strategies.

---

**Q5: What is fuzz testing and when does it help?**

> Fuzz testing automatically generates random inputs to find edge cases that crash or violate invariants. You define a `FuzzXxx` function, provide seed inputs with `f.Add()`, and write a fuzz target function that defines invariants to check. Go's fuzzer mutates the seeds to explore the input space. It's particularly valuable for: parsers and decoders (find inputs that cause panics), encoding/decoding roundtrips (marshal then unmarshal should produce the original), mathematical functions (check properties like commutativity), and any function handling untrusted external input. Failed inputs are saved and become regression tests.

---

**Q6: How do you make code testable in Go?**

> The primary technique is dependency injection via interfaces — instead of instantiating concrete types (DB connections, HTTP clients, file systems) inside a function, accept them as interface parameters. Tests substitute fakes or mocks that implement the same interface. Other practices: avoid package-level global state that tests can pollute (use `t.Setenv`, `t.Cleanup`, or initialize in each test); design pure functions where possible (same input → same output, no side effects); use `t.TempDir()` for file system tests; accept `io.Reader`/`io.Writer` instead of `*os.File` so tests use `strings.NewReader` and `bytes.Buffer`.

---

**Q7: What is `TestMain` and when do you use it?**

> `TestMain(m *testing.M)` is a special function that, if present, controls the test binary's execution. It runs before any test and you call `m.Run()` to execute the tests, capturing the exit code. Use it for package-level setup and teardown that's too expensive to repeat per test — starting a test database, spinning up a local server, or initializing a shared resource. Unlike `init()`, `TestMain` has explicit control over the exit code and can perform cleanup after all tests complete. Call `os.Exit(m.Run())` to ensure proper exit codes.

---

*Next up: **Topic 3.7 — HTTP Servers** 🚀*
