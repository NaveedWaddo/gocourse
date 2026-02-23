# 1.4 — Constants & `iota`

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.3 — Variables & Data Types

---

## 📖 What are Constants?

A **constant** is a value that is fixed at compile time and **cannot be changed** at runtime. If you try to reassign a constant, you get a compile error.

Use constants for values that should never change — mathematical values like `Pi`, configuration limits, status codes, days of the week, etc.

```go
const Pi = 3.14159
const MaxRetries = 3
const AppName = "MyApp"
```

---

## 🧩 Declaring Constants

### Single constant

```go
const Pi float64 = 3.14159265358979
const greeting = "Hello"   // type inferred as string
const maxSize = 100        // type inferred as int
```

### Grouped constants (idiomatic)

```go
const (
    StatusOK       = 200
    StatusNotFound = 404
    StatusError    = 500
)
```

---

## 🔑 Constants vs Variables — Key Differences

| Feature | `const` | `var` / `:=` |
|---------|---------|--------------|
| Value set at | Compile time | Runtime |
| Can be changed? | ❌ No | ✅ Yes |
| Can use function return? | ❌ No | ✅ Yes |
| Zero value? | ❌ No — must assign | ✅ Yes |
| Types allowed | Only basic types | Any type |

```go
const x = 10       // ✅
const y = len("hi") // ❌ compile error — len() is a runtime call... 
                    //    wait, actually len of a string literal IS allowed!
const y = len("hi") // ✅ this actually works — string literal length is compile-time

// But this is NOT allowed:
s := "hello"
const z = len(s)   // ❌ s is a variable — not a compile-time constant
```

> Constants must be computable at **compile time**. That means no function calls that depend on runtime state, no variables — only literals and other constants.

---

## 🧠 Typed vs Untyped Constants

This is one of Go's more subtle and powerful features.

### Untyped constants (preferred)

```go
const x = 42        // untyped integer constant
const y = 3.14      // untyped floating-point constant
const z = "hello"   // untyped string constant
```

Untyped constants are **more flexible** — they can be used wherever the type is compatible, without explicit conversion:

```go
const x = 42

var a int = x       // ✅ works
var b float64 = x   // ✅ works — x adapts!
var c int32 = x     // ✅ works
```

### Typed constants (less flexible)

```go
const x int = 42

var a int = x       // ✅
var b float64 = x   // ❌ cannot use x (int) as float64 — explicit conversion needed
```

> **Rule of thumb:** Prefer untyped constants unless you specifically need to restrict the type.

---

## 🔢 `iota` — The Auto-Incrementing Constant

`iota` is a **special identifier** available only inside a `const` block. It starts at `0` and increments by `1` for each constant in the block.

It's Go's elegant way to define **enumerated values** without repetition.

### Basic usage

```go
const (
    A = iota   // 0
    B = iota   // 1
    C = iota   // 2
)
```

But you don't need to repeat `= iota` — Go repeats the expression automatically:

```go
const (
    A = iota   // 0
    B          // 1
    C          // 2
    D          // 3
)
```

> `iota` resets to `0` at the start of each `const` block.

---

### Skipping values with `_`

Use the blank identifier `_` to skip a value:

```go
const (
    _  = iota   // 0 — skip
    Small       // 1
    Medium      // 2
    Large       // 3
)
```

This is a very common pattern when `0` has no meaningful value in your enum.

---

### Starting from a non-zero value

```go
const (
    First = iota + 1   // 1
    Second             // 2
    Third              // 3
)
```

---

### Powers of two using `iota`

This is one of the most powerful uses of `iota` — creating bit flags:

```go
const (
    Read    = 1 << iota   // 1 << 0 = 1   (001)
    Write                 // 1 << 1 = 2   (010)
    Execute               // 1 << 2 = 4   (100)
)

// Combine with bitwise OR
permissions := Read | Write   // 3 (011)

// Check with bitwise AND
if permissions & Read != 0 {
    fmt.Println("has read permission")
}
```

This pattern is used heavily in Go's standard library (e.g., `os.O_RDONLY`, `os.O_WRONLY`).

---

### Real-world example: Weekdays

```go
type Weekday int

const (
    Sunday Weekday = iota   // 0
    Monday                  // 1
    Tuesday                 // 2
    Wednesday               // 3
    Thursday                // 4
    Friday                  // 5
    Saturday                // 6
)

func main() {
    day := Wednesday
    fmt.Println(day)           // 3
    fmt.Println(day == Friday) // false
}
```

---

### Real-world example: Server status

```go
type Status int

const (
    StatusUnknown Status = iota   // 0
    StatusStarting                // 1
    StatusRunning                 // 2
    StatusStopping                // 3
    StatusStopped                 // 4
)
```

---

### `iota` resets per `const` block

```go
const (
    A = iota   // 0
    B          // 1
)

const (
    C = iota   // 0 — resets!
    D          // 1
)
```

---

## 💻 Complete Example

```go
package main

import "fmt"

// Typed constant
const Pi float64 = 3.14159

// Grouped constants
const (
    MaxConnections = 100
    Timeout        = 30
    AppVersion     = "1.0.0"
)

// Enum with iota
type Direction int

const (
    North Direction = iota   // 0
    East                     // 1
    South                    // 2
    West                     // 3
)

// Bit flags with iota
const (
    PermRead    = 1 << iota   // 1
    PermWrite                 // 2
    PermExecute               // 4
)

func main() {
    // Basic constants
    fmt.Println("Pi:", Pi)
    fmt.Println("Max connections:", MaxConnections)
    fmt.Println("Version:", AppVersion)

    // Direction enum
    dir := North
    fmt.Println("Direction:", dir)   // 0

    // Bit flags
    userPerms := PermRead | PermWrite   // 3
    fmt.Printf("Permissions: %b\n", userPerms)   // 11

    if userPerms&PermRead != 0 {
        fmt.Println("✅ Can read")
    }
    if userPerms&PermWrite != 0 {
        fmt.Println("✅ Can write")
    }
    if userPerms&PermExecute != 0 {
        fmt.Println("✅ Can execute")
    } else {
        fmt.Println("❌ Cannot execute")
    }
}
```

Output:
```
Pi: 3.14159
Max connections: 100
Version: 1.0.0
Direction: 0
Permissions: 11
✅ Can read
✅ Can write
❌ Cannot execute
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Constants can't hold runtime values
```go
// ❌ compile error — math.Sqrt is a runtime call
const root2 = math.Sqrt(2)

// ✅ use a variable instead
var root2 = math.Sqrt(2)
```

### 2. `iota` only works inside `const` blocks
```go
// ❌ iota outside const block is undefined
var x = iota   // compile error
```

### 3. `iota` value is the line position, not usage count
```go
const (
    A = 5      // iota = 0, but A = 5 (iota unused here)
    B = iota   // iota = 1, so B = 1
    C          // iota = 2, so C = 2
)
```

### 4. Can't use `:=` for constants
```go
// ❌ := is for variables only
Pi := 3.14   // this is a variable, not a constant — can be reassigned!

// ✅
const Pi = 3.14
```

### 5. Typed constants restrict where they can be used
```go
const x int32 = 100
var y int64 = x   // ❌ type mismatch — need explicit conversion
var y int64 = int64(x)  // ✅
```

---

## 🎯 Must-Know Things (Quick Reference)

- Constants are declared with `const` and are **immutable at runtime**
- Constants must be set to **compile-time values** — no function calls on variables
- **Untyped constants** are more flexible — they adapt to the context they're used in
- `iota` starts at `0` and increments by `1` for each line in a `const` block
- `iota` **resets to `0`** at the start of each new `const` block
- Use `_` to skip an `iota` value
- `1 << iota` is the classic pattern for **bit flags / bitmasks**
- Go doesn't have a built-in `enum` keyword — `iota` + a custom type IS Go's enum

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between a constant and a variable in Go?**

> A constant is a value fixed at compile time and cannot be reassigned. A variable is allocated at runtime and can be changed. Constants can only hold basic types (numbers, strings, booleans) and their values must be computable at compile time — no function calls on runtime values. Variables have no such restriction. Constants are declared with `const`, variables with `var` or `:=`.

---

**Q2: What is `iota` in Go?**

> `iota` is a special constant identifier built into Go that's only available inside `const` blocks. It starts at `0` and increments by `1` for each constant declaration in the block. It resets to `0` at the beginning of each new `const` block. `iota` is Go's idiomatic way to define enumerations — since Go has no `enum` keyword, you use a custom type and `iota` to achieve the same result.

---

**Q3: What is the difference between typed and untyped constants in Go?**

> An untyped constant (e.g., `const x = 42`) has a "default type" but can adapt to any compatible type without explicit conversion — you can assign it to `int`, `int32`, `float64`, etc. A typed constant (e.g., `const x int = 42`) is locked to that specific type and requires explicit conversion if used elsewhere. Untyped constants are more flexible and are the preferred choice in most cases.

---

**Q4: How does `iota` work with bit flags?**

> By combining `iota` with bit shifting (`1 << iota`), each constant becomes a unique power of two: 1, 2, 4, 8, etc. These can then be combined with bitwise OR (`|`) to represent multiple flags at once, and checked with bitwise AND (`&`). This is a common pattern for permission systems, option flags, and feature toggles — exactly how Go's `os` package defines file open flags like `O_RDONLY`, `O_WRONLY`, `O_RDWR`.

---

**Q5: Can a constant hold the result of a function call?**

> Only if the function is a built-in that can be evaluated at compile time. For example, `const n = len("hello")` is valid because the compiler can evaluate `len` on a string literal at compile time. But `const n = len(someVar)` is not valid because `someVar` is runtime data. Standard library functions like `math.Sqrt()` cannot be used in constants — you'd use a `var` for those.

---

**Q6: How does `iota` reset?**

> `iota` resets to `0` at the start of every new `const` block. It does NOT reset between individual constants within the same block — it increments with each line. So two separate `const (...)` blocks each start their own `iota` sequence from `0`.

---

**Q7: How do you skip a value with `iota`?**

> Use the blank identifier `_` on a line in the `const` block. The `iota` counter still increments, but the value is discarded. This is commonly used to skip `0` when `0` has no meaningful representation in an enumeration — for instance, to make the zero value "unknown" or "unset", you might skip it and start meaningful values at `1`.

---

*Next up: **Topic 1.5 — Operators** 🚀*
