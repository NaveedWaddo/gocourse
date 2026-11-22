# 2.9 — Stringer & fmt Interface

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.8 — Defer, Panic & Recover

---

## 📖 Overview

Go's `fmt` package uses a set of **interfaces** to determine how to print values. By implementing these interfaces on your types, you control exactly how they appear in logs, debug output, and formatted strings. This is Go's idiomatic way of customizing type representation.

The key interfaces:
- `fmt.Stringer` — human-readable string (`String() string`)
- `fmt.GoStringer` — Go syntax representation (`GoString() string`)
- `error` — error string (`Error() string`)
- `fmt.Formatter` — full custom formatting control

---

## 🖨️ `fmt.Stringer` — The Most Important One

```go
type Stringer interface {
    String() string
}
```

When `fmt` prints a value and it implements `Stringer`, it calls `String()` automatically:

```go
type Point struct {
    X, Y int
}

// Without Stringer
p := Point{3, 4}
fmt.Println(p)   // {3 4}  ← default struct format

// With Stringer
func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

fmt.Println(p)         // (3, 4)
fmt.Printf("%v\n", p) // (3, 4)
fmt.Printf("%s\n", p) // (3, 4)
s := fmt.Sprint(p)    // "(3, 4)"
```

`fmt.Println`, `fmt.Printf` with `%v` or `%s`, `fmt.Sprint`, `log.Println` — all call `String()` automatically.

---

## 🔤 Implementing Stringer — Real Examples

### Enum with Stringer

```go
type Direction int

const (
    North Direction = iota
    East
    South
    West
)

func (d Direction) String() string {
    switch d {
    case North:
        return "North"
    case East:
        return "East"
    case South:
        return "South"
    case West:
        return "West"
    default:
        return fmt.Sprintf("Direction(%d)", int(d))
    }
}

func main() {
    d := North
    fmt.Println(d)          // North  (not 0)
    fmt.Printf("Going %v\n", East)   // Going East
}
```

### Currency type

```go
type Money struct {
    Amount   int64  // in cents
    Currency string
}

func (m Money) String() string {
    return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Amount)/100)
}

price := Money{Amount: 1999, Currency: "USD"}
fmt.Println(price)   // USD 19.99
```

### IP Address

```go
type IP [4]byte

func (ip IP) String() string {
    return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

addr := IP{192, 168, 1, 1}
fmt.Println(addr)   // 192.168.1.1
```

---

## 🔍 `fmt.GoStringer` — Go Syntax Representation

```go
type GoStringer interface {
    GoString() string
}
```

Used with `%#v` format verb — produces Go syntax representation:

```go
type Point struct{ X, Y int }

func (p Point) GoString() string {
    return fmt.Sprintf("Point{X: %d, Y: %d}", p.X, p.Y)
}

p := Point{3, 4}
fmt.Printf("%v\n",  p)    // {3 4}        — default
fmt.Printf("%#v\n", p)    // Point{X: 3, Y: 4}  — GoString()
```

Useful for debugging — produces output you could paste back into Go code.

---

## 📋 Format Verbs Reference

Understanding all format verbs is essential for effective use of `fmt`:

### General

| Verb | Meaning |
|------|---------|
| `%v` | Default format — calls `String()` if available |
| `%+v` | Struct with field names: `{X:3 Y:4}` |
| `%#v` | Go syntax: calls `GoString()` if available |
| `%T` | Type: `main.Point` |

### Integer

| Verb | Meaning | Example |
|------|---------|---------|
| `%d` | Decimal | `42` |
| `%b` | Binary | `101010` |
| `%o` | Octal | `52` |
| `%x` | Hex lowercase | `2a` |
| `%X` | Hex uppercase | `2A` |
| `%05d` | Zero-padded | `00042` |
| `%-5d` | Left-aligned | `42   ` |

### Float

| Verb | Meaning | Example |
|------|---------|---------|
| `%f` | Decimal point | `3.141593` |
| `%.2f` | 2 decimal places | `3.14` |
| `%e` | Scientific notation | `3.141593e+00` |
| `%g` | Compact: `%e` or `%f` | `3.14159` |

### String & Byte

| Verb | Meaning | Example |
|------|---------|---------|
| `%s` | String | `hello` |
| `%q` | Quoted string | `"hello"` |
| `%x` | Hex of bytes | `68656c6c6f` |

### Other

| Verb | Meaning |
|------|---------|
| `%t` | Boolean: `true` / `false` |
| `%p` | Pointer address |
| `%w` | Wrap error (only in `fmt.Errorf`) |

---

## 📝 `fmt.Sprintf` — Building Strings

```go
name := "Alice"
age := 30

// Basic string building
s := fmt.Sprintf("Name: %s, Age: %d", name, age)
// "Name: Alice, Age: 30"

// Padding and alignment
fmt.Printf("|%-10s|%10s|\n", "left", "right")
// |left      |     right|

// Formatting numbers
fmt.Printf("%.2f\n", 3.14159)   // 3.14
fmt.Printf("%08.2f\n", 3.14)    // 00003.14
fmt.Printf("%+d\n", 42)         // +42
fmt.Printf("%+d\n", -42)        // -42
```

---

## 🖊️ Writing to Destinations — Print Variants

```go
// To stdout
fmt.Print("no newline")
fmt.Println("with newline")
fmt.Printf("formatted %s\n", "string")

// To a string
s := fmt.Sprint("a", "b", "c")          // "abc"
s = fmt.Sprintln("a", "b", "c")         // "a b c\n"
s = fmt.Sprintf("hello %s", "world")    // "hello world"

// To any io.Writer
fmt.Fprint(os.Stderr, "error output")
fmt.Fprintln(w, "response body")
fmt.Fprintf(file, "log entry: %s\n", msg)

// Errors
err := fmt.Errorf("failed to open %s: %w", path, originalErr)
```

---

## 📥 `fmt.Scan` — Reading Input

```go
// Read from stdin
var name string
var age int
fmt.Scan(&name, &age)          // reads space-separated values
fmt.Scanln(&name, &age)        // reads until newline
fmt.Scanf("%s %d", &name, &age) // reads with format

// Read from a string
fmt.Sscan("Alice 30", &name, &age)
fmt.Sscanf("Alice 30", "%s %d", &name, &age)

// Read from any io.Reader
fmt.Fscan(reader, &name)
```

---

## 🎨 `fmt.Formatter` — Full Custom Control

For complete control over formatting, implement `fmt.Formatter`:

```go
type Formatter interface {
    Format(f fmt.State, verb rune)
}
```

```go
type Matrix struct {
    data [][]float64
    rows, cols int
}

func (m Matrix) Format(f fmt.State, verb rune) {
    switch verb {
    case 'v', 's':
        for _, row := range m.data {
            for j, val := range row {
                if j > 0 {
                    fmt.Fprint(f, "  ")
                }
                fmt.Fprintf(f, "%6.2f", val)
            }
            fmt.Fprintln(f)
        }
    case 'f':
        // compact single-line format
        fmt.Fprintf(f, "Matrix(%dx%d)", m.rows, m.cols)
    }
}
```

---

## 🔗 Stringer and the `error` Interface Together

Custom types often implement both:

```go
type HTTPError struct {
    Code    int
    Message string
}

// Implements error
func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// Implements Stringer (same output here, but can differ)
func (e *HTTPError) String() string {
    return e.Error()
}

err := &HTTPError{Code: 404, Message: "not found"}
fmt.Println(err)          // HTTP 404: not found  ← uses String()
fmt.Printf("%v\n", err)   // HTTP 404: not found
fmt.Printf("%s\n", err)   // HTTP 404: not found
```

---

## 💻 Complete Example

```go
package main

import "fmt"

// --- Direction enum with Stringer ---
type Weekday int

const (
    Monday Weekday = iota + 1
    Tuesday
    Wednesday
    Thursday
    Friday
    Saturday
    Sunday
)

func (d Weekday) String() string {
    names := [...]string{
        "Monday", "Tuesday", "Wednesday",
        "Thursday", "Friday", "Saturday", "Sunday",
    }
    if d < Monday || d > Sunday {
        return fmt.Sprintf("Weekday(%d)", int(d))
    }
    return names[d-1]
}

func (d Weekday) IsWeekend() bool {
    return d == Saturday || d == Sunday
}

// --- Temperature with GoStringer ---
type Temperature struct {
    Celsius float64
}

func (t Temperature) String() string {
    return fmt.Sprintf("%.1f°C", t.Celsius)
}

func (t Temperature) GoString() string {
    return fmt.Sprintf("Temperature{Celsius: %.1f}", t.Celsius)
}

func (t Temperature) Fahrenheit() float64 {
    return t.Celsius*9/5 + 32
}

// --- Color with Formatter ---
type Color struct{ R, G, B uint8 }

func (c Color) String() string {
    return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

func (c Color) GoString() string {
    return fmt.Sprintf("Color{R: %d, G: %d, B: %d}", c.R, c.G, c.B)
}

func (c Color) Hex() string {
    return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func main() {
    // Weekday Stringer
    fmt.Println("=== Weekday ===")
    for d := Monday; d <= Sunday; d++ {
        if d.IsWeekend() {
            fmt.Printf("%v → weekend\n", d)
        } else {
            fmt.Printf("%v → workday\n", d)
        }
    }

    // Temperature
    fmt.Println("\n=== Temperature ===")
    temps := []Temperature{{0}, {100}, {-40}, {37}}
    for _, t := range temps {
        fmt.Printf("%v = %.1f°F\n", t, t.Fahrenheit())
    }
    // GoString with %#v
    fmt.Printf("%#v\n", temps[0])   // Temperature{Celsius: 0.0}

    // Color
    fmt.Println("\n=== Color ===")
    red := Color{255, 0, 0}
    fmt.Printf("String:   %v\n", red)    // rgb(255, 0, 0)
    fmt.Printf("GoString: %#v\n", red)   // Color{R: 255, G: 0, B: 0}
    fmt.Printf("Hex:      %s\n", red.Hex()) // #FF0000

    // Format verbs
    fmt.Println("\n=== Format Verbs ===")
    n := 255
    fmt.Printf("decimal: %d\n", n)    // 255
    fmt.Printf("binary:  %b\n", n)    // 11111111
    fmt.Printf("hex:     %x\n", n)    // ff
    fmt.Printf("octal:   %o\n", n)    // 377

    fmt.Printf("padded:  |%8d|\n", n)    // |     255|
    fmt.Printf("left:    |%-8d|\n", n)   // |255     |
    fmt.Printf("zeros:   |%08d|\n", n)   // |00000255|

    pi := 3.14159265
    fmt.Printf("float:   %f\n", pi)      // 3.141593
    fmt.Printf("fixed:   %.2f\n", pi)    // 3.14
    fmt.Printf("sci:     %e\n", pi)      // 3.141593e+00
}
```

Output:
```
=== Weekday ===
Monday → workday
Tuesday → workday
Wednesday → workday
Thursday → workday
Friday → workday
Saturday → weekend
Sunday → weekend

=== Temperature ===
0.0°C = 32.0°F
100.0°C = 212.0°F
-40.0°C = -40.0°F
37.0°C = 98.6°F
Temperature{Celsius: 0.0}

=== Color ===
String:   rgb(255, 0, 0)
Color:    Color{R: 255, G: 0, B: 0}
Hex:      #FF0000

=== Format Verbs ===
decimal: 255
binary:  11111111
hex:     ff
octal:   377
padded:  |     255|
left:    |255     |
zeros:   |00000255|
float:   3.141593
fixed:   3.14
sci:     3.141593e+00
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Infinite recursion in String()
```go
type MyType struct{ Val int }

// ❌ Infinite recursion — %v calls String(), which calls Sprintf with %v...
func (m MyType) String() string {
    return fmt.Sprintf("%v", m)   // calls String() again → stack overflow!
}

// ✅ Use %d, %s on fields, or cast to avoid recursion
func (m MyType) String() string {
    return fmt.Sprintf("MyType(%d)", m.Val)   // use concrete verb
}

// ✅ Or cast to base type
func (m MyType) String() string {
    type plain MyType           // new type without String() method
    return fmt.Sprintf("%v", plain(m))
}
```

### 2. Pointer vs value receiver on Stringer
```go
type T struct{ val int }

func (t *T) String() string { return fmt.Sprintf("T(%d)", t.val) }

t := T{42}
fmt.Println(t)    // {42}   ← value doesn't have String()
fmt.Println(&t)   // T(42)  ← pointer has String()

// ✅ Use value receiver if you want both to work
func (t T) String() string { return fmt.Sprintf("T(%d)", t.val) }
```

### 3. Forgetting `%w` for error wrapping
```go
// ❌ %v wraps the message but NOT the error chain
err = fmt.Errorf("context: %v", originalErr)   // errors.Is won't work!

// ✅ %w wraps the error properly
err = fmt.Errorf("context: %w", originalErr)   // errors.Is works
```

### 4. Using `fmt.Println` vs `fmt.Printf` incorrectly
```go
name := "Alice"
// ❌ Literal percent sign in Println
fmt.Println("Hello %s", name)   // prints: Hello %s Alice (not formatted!)

// ✅ Use Printf for formatting
fmt.Printf("Hello %s\n", name)  // prints: Hello Alice
```

---

## 🎯 Must-Know Things (Quick Reference)

- `fmt.Stringer` = `String() string` — implement for human-readable output
- `fmt.GoStringer` = `GoString() string` — implement for `%#v` Go-syntax output
- `fmt` automatically calls `String()` with `%v`, `%s`, and `Println`
- **Avoid infinite recursion** in `String()` — don't use `%v` on the receiver itself
- Use **value receivers** for Stringer so both `T` and `*T` satisfy the interface
- `%w` in `fmt.Errorf` — wraps error for chain inspection; `%v` does not
- `fmt.Sprintf` → string, `fmt.Fprintf` → io.Writer, `fmt.Printf` → stdout
- `%+v` prints struct field names; `%#v` prints Go syntax
- `%-Nd` left-align, `%0Nd` zero-pad, `%.Nf` decimal places
- Enum types should always implement `Stringer` — printing `0` is not useful

---

## ❓ Interview Questions & Answers

---

**Q1: What is the `fmt.Stringer` interface and why is it useful?**

> `fmt.Stringer` is an interface with a single method: `String() string`. When a type implements it, the `fmt` package automatically calls `String()` whenever that value is printed — with `fmt.Println`, `fmt.Printf` using `%v` or `%s`, `fmt.Sprint`, and logging functions. It gives you full control over how your type appears in output without the caller needing to know anything special. It's especially valuable for enums (printing "North" instead of "0") and domain types (printing "USD 19.99" instead of "{1999 USD}").

---

**Q2: What is the infinite recursion pitfall with `String()`?**

> If you implement `String()` and use `fmt.Sprintf("%v", m)` where `m` is the receiver, you create infinite recursion — `%v` calls `String()`, which calls `Sprintf`, which calls `String()` again, until stack overflow. The fix is to use concrete format verbs for each field (`%d`, `%s`), or cast the receiver to a base type that doesn't have the `String()` method: `type plain MyType; fmt.Sprintf("%v", plain(m))`.

---

**Q3: What is the difference between `%v`, `%+v`, and `%#v`?**

> `%v` prints the value in its default format — for structs it prints `{field1 field2}`, and calls `String()` if available. `%+v` prints structs with field names: `{X:3 Y:4}`. `%#v` prints in Go syntax — it calls `GoString()` if available, otherwise produces something like `main.Point{X:3, Y:4}` that you could paste into Go code. `%#v` is most useful for debugging and logging where you want unambiguous type information.

---

**Q4: What is the difference between `%v` and `%w` in `fmt.Errorf`?**

> Both embed an error's message in the resulting error string. The difference is that `%w` also **wraps** the original error — the returned error implements `Unwrap() error`, making the original error accessible to `errors.Is` and `errors.As` for chain inspection. `%v` only includes the string representation — the original error is lost for programmatic inspection. Always use `%w` when wrapping errors so callers can check for specific sentinel errors in the chain.

---

**Q5: Should `String()` use a value or pointer receiver?**

> Prefer a value receiver for `String()`. With a value receiver, both `T` and `*T` satisfy `fmt.Stringer` — because `*T`'s method set includes all of `T`'s value receiver methods. With a pointer receiver, only `*T` satisfies `Stringer` — a value of type `T` prints with default formatting. Since `fmt.Println(t)` passes a value copy, using a pointer receiver means you'd have to remember to always pass `&t`. Value receivers for `Stringer` make the interface satisfaction consistent.

---

**Q6: How do `fmt.Print`, `fmt.Println`, and `fmt.Printf` differ?**

> `fmt.Print` writes arguments with default formatting, adding spaces between operands only when neither is a string — no trailing newline. `fmt.Println` writes arguments with default formatting, always adds spaces between operands, and appends a newline. `fmt.Printf` takes a format string with verbs and formats accordingly — no automatic spaces or newlines. The `S` prefix variants (`Sprintf`, `Sprint`, `Sprintln`) return strings instead of writing to stdout. The `F` prefix variants (`Fprintf`, `Fprint`, `Fprintln`) write to an `io.Writer`.

---

*Next up: **Topic 2.10 — File I/O** 🚀*
