# 1.12 — Methods

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.11 — Structs

---

## 📖 What is a Method?

A **method** is a function with a special **receiver** argument — it associates the function with a specific type. The receiver appears between the `func` keyword and the function name:

```go
func (receiver ReceiverType) MethodName(params) returnType {
    // body
}
```

Methods let you attach behavior to types — this is how Go does OOP without classes.

```go
type Rectangle struct {
    Width  float64
    Height float64
}

// Area is a method on Rectangle
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func main() {
    rect := Rectangle{Width: 4, Height: 5}
    fmt.Println(rect.Area())   // 20
}
```

---

## ⚖️ Value Receivers vs Pointer Receivers

This is the most important concept in this topic. The receiver can be either a **value** or a **pointer**.

### Value receiver — works on a copy

```go
func (r Rectangle) Area() float64 {
    return r.Width * r.Height   // r is a copy — original unchanged
}
```

- Gets a **copy** of the value
- Cannot modify the original
- Safe and cheap for small structs
- Calling on a pointer is fine — Go auto-dereferences: `ptr.Area()` works

### Pointer receiver — works on the original

```go
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor    // modifies the original Rectangle
    r.Height *= factor
}
```

- Gets a **pointer** — can modify the original
- Necessary when the method needs to mutate state
- More efficient for large structs (no copying)
- Calling on a value is fine — Go auto-takes address: `rect.Scale(2)` works even if `rect` is not a pointer

---

### Side-by-side example

```go
type Counter struct {
    count int
}

// Value receiver — cannot mutate
func (c Counter) Value() int {
    return c.count
}

// Pointer receiver — mutates
func (c *Counter) Increment() {
    c.count++
}

func (c *Counter) Reset() {
    c.count = 0
}

func main() {
    c := Counter{}
    c.Increment()
    c.Increment()
    c.Increment()
    fmt.Println(c.Value())   // 3
    c.Reset()
    fmt.Println(c.Value())   // 0
}
```

---

## 🔄 Go's Automatic Address/Dereference

Go automatically handles the conversion between value and pointer receivers when calling methods — as long as the variable is **addressable**:

```go
rect := Rectangle{Width: 4, Height: 5}   // value, not pointer

// Go auto-takes address for pointer receiver methods:
rect.Scale(2)   // same as (&rect).Scale(2)

ptr := &Rectangle{Width: 4, Height: 5}  // pointer

// Go auto-dereferences for value receiver methods:
ptr.Area()   // same as (*ptr).Area()
```

> This only works for **addressable** values — variables, struct fields, slice elements. It does NOT work on map values or temporary (non-addressable) values.

```go
// ❌ Map values are not addressable
m := map[string]Counter{"a": {}}
m["a"].Increment()   // compile error — cannot take address of map value

// ✅ Workaround — copy out, modify, put back
c := m["a"]
c.Increment()
m["a"] = c
```

---

## 📏 The Golden Rule — Be Consistent

**Pick one receiver type per type and stick to it.**

If ANY method on a type needs a pointer receiver (to mutate state), ALL methods should use pointer receivers — even the ones that don't mutate. This ensures consistency and prevents subtle bugs when using interfaces.

```go
// ❌ Inconsistent — some value, some pointer receivers
func (r Rectangle) Area() float64 { ... }    // value
func (r *Rectangle) Scale(f float64) { ... } // pointer

// ✅ Consistent — all pointer receivers
func (r *Rectangle) Area() float64 { ... }
func (r *Rectangle) Scale(f float64) { ... }
func (r *Rectangle) Perimeter() float64 { ... }
```

**When to use pointer receivers:**
- Method modifies the receiver
- Struct is large (avoid copying)
- Type contains a mutex or similar — must not be copied

**When value receivers are fine:**
- Small, immutable types (like `time.Duration`, `net.IP`)
- Built-in-like types (`int`, `string` wrappers)
- When you explicitly want copy semantics

---

## 🔗 Methods on Non-Struct Types

You can define methods on **any type defined in the same package** — not just structs:

```go
// Method on a named type based on int
type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
    return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
    return Celsius((f - 32) * 5 / 9)
}

func main() {
    boiling := Celsius(100)
    fmt.Printf("%.1f°C = %.1f°F\n", boiling, boiling.ToFahrenheit())   // 100.0°C = 212.0°F
}
```

> You **cannot** define methods on types from another package — e.g., you can't add methods to `int` directly. You must define your own named type first.

---

## 🔗 Method on a Slice Type

```go
type IntSlice []int

func (s IntSlice) Sum() int {
    total := 0
    for _, v := range s {
        total += v
    }
    return total
}

func (s IntSlice) Contains(n int) bool {
    for _, v := range s {
        if v == n {
            return true
        }
    }
    return false
}

func main() {
    nums := IntSlice{1, 2, 3, 4, 5}
    fmt.Println(nums.Sum())          // 15
    fmt.Println(nums.Contains(3))    // true
    fmt.Println(nums.Contains(99))   // false
}
```

---

## 🔁 Method Chaining

By returning the receiver pointer, you can chain method calls:

```go
type Builder struct {
    result string
}

func (b *Builder) Add(s string) *Builder {
    b.result += s
    return b   // return pointer to self for chaining
}

func (b *Builder) Build() string {
    return b.result
}

func main() {
    result := (&Builder{}).
        Add("Hello").
        Add(", ").
        Add("World").
        Build()

    fmt.Println(result)   // Hello, World
}
```

This pattern is used heavily in Go libraries (e.g., `strings.Builder`, query builders).

---

## 🔍 Method Values and Method Expressions

Methods can be extracted as function values:

```go
type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64 { return r.Width * r.Height }

rect := Rectangle{4, 5}

// Method value — bound to a specific receiver instance
areaFn := rect.Area   // areaFn is func() float64, bound to rect
fmt.Println(areaFn()) // 20

// Method expression — receiver passed as first argument
areaExpr := Rectangle.Area   // areaExpr is func(Rectangle) float64
fmt.Println(areaExpr(rect))  // 20
```

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "math"
)

// --- Shape types ---

type Circle struct {
    Radius float64
}

type Rectangle struct {
    Width  float64
    Height float64
}

// --- Circle methods ---

func (c *Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c *Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

func (c *Circle) Scale(factor float64) *Circle {
    c.Radius *= factor
    return c   // for chaining
}

// --- Rectangle methods ---

func (r *Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r *Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

func (r *Rectangle) Scale(factor float64) *Rectangle {
    r.Width *= factor
    r.Height *= factor
    return r
}

func (r *Rectangle) IsSquare() bool {
    return r.Width == r.Height
}

// --- Named type with method ---

type Celsius float64

func (c Celsius) ToFahrenheit() float64 {
    return float64(c)*9/5 + 32
}

func (c Celsius) String() string {
    return fmt.Sprintf("%.1f°C", float64(c))
}

func main() {
    // Circle
    c := &Circle{Radius: 5}
    fmt.Printf("Circle area: %.2f\n", c.Area())
    fmt.Printf("Circle perimeter: %.2f\n", c.Perimeter())
    c.Scale(2)
    fmt.Printf("After Scale(2), radius: %.1f\n", c.Radius)   // 10.0

    // Rectangle
    r := &Rectangle{Width: 4, Height: 5}
    fmt.Printf("\nRectangle area: %.1f\n", r.Area())
    fmt.Printf("Rectangle perimeter: %.1f\n", r.Perimeter())
    fmt.Printf("Is square? %v\n", r.IsSquare())

    r.Scale(3)
    fmt.Printf("After Scale(3): %.0fx%.0f\n", r.Width, r.Height)

    square := &Rectangle{Width: 4, Height: 4}
    fmt.Printf("Is square? %v\n", square.IsSquare())   // true

    // Named type method
    temp := Celsius(100)
    fmt.Printf("\n%s = %.1f°F\n", temp, temp.ToFahrenheit())

    // Method value
    areaFn := r.Area
    fmt.Printf("\nArea via method value: %.1f\n", areaFn())
}
```

Output:
```
Circle area: 78.54
Circle perimeter: 31.42
After Scale(2), radius: 10.0

Rectangle area: 20.0
Rectangle perimeter: 18.0
Is square? false
After Scale(3): 12x15
Is square? false
Is square? true

100.0°C = 212.0°F

Area via method value: 180.0
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Value receiver method doesn't mutate — silent bug
```go
func (c Counter) Increment() {
    c.count++   // modifies a copy — caller sees nothing!
}

c := Counter{}
c.Increment()
fmt.Println(c.count)   // still 0!

// ✅ Fix: use pointer receiver
func (c *Counter) Increment() { c.count++ }
```

### 2. Mixing receiver types breaks interface satisfaction
```go
type Stringer interface{ String() string }

type MyType struct{ val int }

func (m MyType) String() string { return fmt.Sprintf("%d", m.val) }

var s Stringer = MyType{42}    // ✅ value satisfies interface
var s2 Stringer = &MyType{42}  // ✅ pointer also satisfies (Go promotes)

func (m *MyType) String() string { ... }  // pointer receiver only

var s Stringer = MyType{42}    // ❌ value does NOT satisfy interface with pointer receiver
var s2 Stringer = &MyType{42}  // ✅ pointer satisfies
```

### 3. Calling pointer receiver method on non-addressable value
```go
// Map values are not addressable
m := map[string]Counter{"a": {}}
m["a"].Increment()   // ❌ compile error

// Function return values are not addressable
getCounter().Increment()   // ❌ compile error
```

### 4. Copying a type with a mutex
```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

// ❌ Copying copies the mutex — undefined behavior
func bad(c SafeCounter) { c.Increment() }

// ✅ Always use pointer receiver/param for types with mutexes
func good(c *SafeCounter) { c.Increment() }
```

---

## 🎯 Must-Know Things (Quick Reference)

- A method is a function with a **receiver** — `func (r ReceiverType) Name()`
- **Value receiver** — gets a copy, cannot mutate, good for small/immutable types
- **Pointer receiver** — gets the original, can mutate, efficient for large structs
- **Be consistent** — if any method needs a pointer receiver, use pointer receivers for all methods on that type
- Go **auto-dereferences** pointer receivers and **auto-takes-address** for value receivers — but only for addressable values
- Map values are **not addressable** — can't call pointer receiver methods directly on them
- Methods can be defined on **any named type** in the same package — not just structs
- **Method values** bind a method to a specific receiver instance: `rect.Area` → `func() float64`
- Types with **mutexes** must always use pointer receivers — never copy them
- Pointer receiver methods are NOT in the method set of a value type — critical for interfaces (Topic 2.1)

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between a method and a function in Go?**

> A method is a function associated with a specific type via a receiver argument. The receiver appears between `func` and the method name: `func (r Rectangle) Area() float64`. A regular function has no receiver. Methods allow you to call behavior directly on a value — `rect.Area()` — and can access the receiver's fields. Under the hood, `rect.Area()` is syntactic sugar for `Rectangle.Area(rect)`.

---

**Q2: What is the difference between value and pointer receivers?**

> A value receiver `(r Rectangle)` receives a copy of the struct — changes inside the method don't affect the original. A pointer receiver `(r *Rectangle)` receives a pointer to the original — changes are reflected in the caller. Use pointer receivers when: the method needs to mutate the receiver, the struct is large (avoid copying), or the type contains non-copyable fields like a mutex. Use value receivers for small, immutable types.

---

**Q3: How does Go handle automatic conversion between value and pointer receivers?**

> Go automatically takes the address when calling a pointer receiver method on an addressable value (`rect.Scale(2)` becomes `(&rect).Scale(2)`), and automatically dereferences when calling a value receiver method on a pointer (`ptr.Area()` becomes `(*ptr).Area()`). This only works for addressable values — variables and struct fields. Map values and function return values are not addressable, so pointer receiver methods cannot be called on them directly.

---

**Q4: Why should receiver types be consistent across all methods of a type?**

> If you mix value and pointer receivers on the same type, the method sets are different — a value type has value receiver methods in its method set, but a pointer type has both. This causes confusion with interfaces: if a type has both value and pointer receiver methods, only a pointer to that type satisfies an interface that requires all methods. The rule is: if any method needs mutation or the type shouldn't be copied (e.g., contains a mutex), all methods should use pointer receivers.

---

**Q5: Can you define methods on non-struct types in Go?**

> Yes — you can define methods on any named type defined in the same package. For example, `type Celsius float64` can have methods like `ToFahrenheit()`. You can also define methods on named slice types, named function types, etc. You cannot add methods to built-in types (`int`, `string`) directly or to types from other packages — you must define your own named type first.

---

**Q6: What is a method value vs a method expression?**

> A method value is obtained by accessing a method on a specific instance — e.g., `rect.Area`. It's a function value already bound to that receiver, so you call it with no receiver argument: `areaFn()`. A method expression is obtained from the type itself — e.g., `Rectangle.Area`. It's a function where the receiver is the first explicit argument: `areaExpr(rect)`. Method values are useful for passing as callbacks — e.g., `sort.Slice` or goroutine functions.

---

**Q7: Why can't you call pointer receiver methods on map values?**

> Map values are not addressable in Go — you can't take their address with `&`. When you access `m["key"]`, Go gives you a copy of the value, not a reference to it. Since pointer receiver methods require an addressable value to automatically take its address, this fails at compile time. The workaround is to copy the value out, call the method on the copy (or a pointer to it), and write it back: `v := m["key"]; v.Method(); m["key"] = v`.

---

## 🎉 Phase 1 Complete!

You've finished all 12 topics in **Phase 1 — Foundations**. Here's what you now know:

| Topic | Key Takeaway |
|-------|-------------|
| 1.1 Intro to Go | Philosophy, tooling, workspace |
| 1.2 First Program | package main, go run/build, exports |
| 1.3 Variables & Types | var, :=, zero values, type conversion |
| 1.4 Constants & iota | const, untyped constants, bit flags |
| 1.5 Operators | Arithmetic, logical, bitwise, no implicit conversion |
| 1.6 Control Flow | for does everything, switch auto-breaks, range |
| 1.7 Functions | Multiple returns, closures, variadic, first-class |
| 1.8 Pointers | &, *, nil safety, value vs reference |
| 1.9 Arrays & Slices | Slice internals, append, copy, shared backing array |
| 1.10 Maps | Comma-ok, nil map panic, reference type, set pattern |
| 1.11 Structs | Embedding, tags, constructors, value type |
| 1.12 Methods | Value vs pointer receivers, consistency rule |

**Next: Phase 2 — Intermediate** starts with interfaces, error handling, and concurrency! 🚀
