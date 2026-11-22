# 2.1 — Interfaces

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 1.12 — Methods

---

## 📖 What is an Interface?

An **interface** is a type that defines a set of method signatures. Any type that implements all those methods **automatically satisfies** the interface — no explicit declaration needed. This is called **implicit implementation** and it's one of Go's most powerful design choices.

```go
type Animal interface {
    Speak() string
    Move() string
}
```

Any type with both `Speak() string` and `Move() string` methods is an `Animal` — whether it knows it or not.

---

## 🔑 Implicit Implementation — No `implements` Keyword

In Java/C# you write `class Dog implements Animal`. In Go, you just define the methods:

```go
type Dog struct{ Name string }

func (d Dog) Speak() string { return "Woof!" }
func (d Dog) Move() string  { return "runs" }

type Bird struct{ Name string }

func (b Bird) Speak() string { return "Tweet!" }
func (b Bird) Move() string  { return "flies" }

// Both satisfy Animal without any declaration
var a Animal = Dog{Name: "Rex"}    // ✅
var b Animal = Bird{Name: "Tweety"} // ✅
```

This decouples the interface definition from the implementation — a type can satisfy interfaces defined in packages it has never heard of.

---

## 🏗️ Defining and Using Interfaces

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct{ Radius float64 }
type Rectangle struct{ Width, Height float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

// Function that accepts any Shape
func printShapeInfo(s Shape) {
    fmt.Printf("Area: %.2f, Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

func main() {
    shapes := []Shape{
        Circle{Radius: 5},
        Rectangle{Width: 4, Height: 3},
    }
    for _, s := range shapes {
        printShapeInfo(s)
    }
}
```

---

## 🔬 Interface Internals — Two-Word Structure

An interface value is internally a two-word pair:

```
Interface value:
┌──────────────┐
│  type ptr    │ → points to the concrete type's metadata
│  data ptr    │ → points to the concrete value
└──────────────┘
```

This is why an interface can hold any type — it stores both what the value is and what type it is.

```go
var s Shape = Circle{Radius: 5}
// type ptr → Circle
// data ptr → Circle{Radius: 5}

s = Rectangle{Width: 4, Height: 3}
// type ptr → Rectangle
// data ptr → Rectangle{Width:4, Height:3}
```

---

## 🎭 The Empty Interface — `interface{}` and `any`

The empty interface has no methods — every type satisfies it:

```go
var x interface{} = 42
x = "hello"
x = true
x = []int{1, 2, 3}
```

Since Go 1.18, `any` is an alias for `interface{}` — use `any` in modern code:

```go
var x any = 42
x = "hello"
```

Common uses:
- `fmt.Println(a ...any)` — accepts anything
- Generic containers before Go 1.18 generics
- JSON decoding into `map[string]any`

> Avoid `any` when you know the type — you lose compile-time type safety.

---

## 🔍 Type Assertions

A type assertion extracts the concrete value from an interface:

```go
var i interface{} = "hello"

// Assertion — panics if wrong type
s := i.(string)
fmt.Println(s)   // hello

// Comma-ok — safe, no panic
s, ok := i.(string)
fmt.Println(s, ok)   // hello true

n, ok := i.(int)
fmt.Println(n, ok)   // 0 false — safe
```

> Always use the comma-ok form unless you are 100% certain of the type. A failed assertion without comma-ok causes a **panic**.

---

## 🔀 Type Switch

The cleanest way to handle multiple possible types:

```go
func describe(i any) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("int: %d", v)
    case string:
        return fmt.Sprintf("string: %q (len=%d)", v, len(v))
    case bool:
        return fmt.Sprintf("bool: %t", v)
    case []int:
        return fmt.Sprintf("[]int with %d elements", len(v))
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown: %T", v)
    }
}
```

---

## 📦 Interface Composition

Interfaces can embed other interfaces — building larger contracts from smaller ones:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ReadWriter is composed from Reader and Writer
type ReadWriter interface {
    Reader
    Writer
}

// Closer adds Close
type ReadWriteCloser interface {
    ReadWriter
    Close() error
}
```

This is exactly how Go's `io` package works. `io.ReadWriter`, `io.ReadCloser`, `io.ReadWriteCloser` are all composed interfaces.

---

## 🔑 Method Sets and Pointer vs Value Receivers

This is where the receiver consistency rule from Topic 1.12 becomes critical.

**The rule:**
- A **value type** `T` only has value receiver methods in its method set
- A **pointer type** `*T` has BOTH value and pointer receiver methods

```go
type Stringer interface {
    String() string
}

type MyType struct{ val int }

// Pointer receiver
func (m *MyType) String() string { return fmt.Sprintf("%d", m.val) }

var s Stringer = &MyType{42}   // ✅ *MyType satisfies Stringer
var s2 Stringer = MyType{42}   // ❌ MyType does NOT — String() is on *MyType
```

```go
// Value receiver
func (m MyType) String() string { return fmt.Sprintf("%d", m.val) }

var s Stringer = MyType{42}    // ✅ MyType satisfies Stringer
var s2 Stringer = &MyType{42}  // ✅ *MyType also satisfies (pointer includes value methods)
```

> **Rule of thumb:** If you use pointer receivers, assign pointers to interfaces. If you use value receivers, both values and pointers work.

---

## 🧱 Small, Focused Interfaces — Go's Philosophy

Go's standard library prefers tiny interfaces over large ones:

```go
// io.Reader — just one method
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer — just one method
type Writer interface {
    Write(p []byte) (n int, err error)
}

// io.Closer — just one method
type Closer interface {
    Close() error
}
```

> **"The bigger the interface, the weaker the abstraction."** — Rob Pike

Design interfaces with the minimum methods needed. This makes them easy to satisfy and highly reusable.

---

## 🧪 Nil Interface vs Interface Holding Nil

One of Go's most surprising gotchas:

```go
// A nil interface — both type and value are nil
var s Shape
fmt.Println(s == nil)   // true

// An interface holding a nil pointer — type is set, value is nil
var c *Circle = nil
var s2 Shape = c
fmt.Println(s2 == nil)   // FALSE! — s2 has type info (*Circle), just nil data
fmt.Println(c == nil)    // true
```

This causes a famous bug with error handling:

```go
// ❌ Classic bug
func getError() error {
    var err *MyError = nil
    // ... some logic ...
    return err   // returns interface holding (*MyError, nil) — NOT a nil interface!
}

if err := getError(); err != nil {
    // This runs even though we "returned nil"!
}

// ✅ Fix — return nil directly
func getError() error {
    return nil   // returns a true nil interface
}
```

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "math"
    "sort"
)

// --- Interface definition ---

type Shape interface {
    Area() float64
    Perimeter() float64
    String() string
}

// --- Implementations ---

type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string     { return fmt.Sprintf("Circle(r=%.1f)", c.Radius) }

type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) String() string {
    return fmt.Sprintf("Rect(%.1fx%.1f)", r.Width, r.Height)
}

type Triangle struct{ A, B, C float64 } // side lengths

func (t Triangle) Area() float64 {
    s := (t.A + t.B + t.C) / 2
    return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}
func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) String() string {
    return fmt.Sprintf("Triangle(%.1f,%.1f,%.1f)", t.A, t.B, t.C)
}

// --- Functions using the interface ---

func totalArea(shapes []Shape) float64 {
    total := 0.0
    for _, s := range shapes {
        total += s.Area()
    }
    return total
}

func largest(shapes []Shape) Shape {
    if len(shapes) == 0 {
        return nil
    }
    max := shapes[0]
    for _, s := range shapes[1:] {
        if s.Area() > max.Area() {
            max = s
        }
    }
    return max
}

func main() {
    shapes := []Shape{
        Circle{Radius: 5},
        Rectangle{Width: 4, Height: 6},
        Triangle{A: 3, B: 4, C: 5},
        Circle{Radius: 2},
    }

    // Polymorphism
    for _, s := range shapes {
        fmt.Printf("%s → area=%.2f perimeter=%.2f\n",
            s, s.Area(), s.Perimeter())
    }

    fmt.Printf("\nTotal area: %.2f\n", totalArea(shapes))
    fmt.Printf("Largest: %s\n", largest(shapes))

    // Sort by area using interface
    sort.Slice(shapes, func(i, j int) bool {
        return shapes[i].Area() < shapes[j].Area()
    })
    fmt.Println("\nSorted by area:")
    for _, s := range shapes {
        fmt.Printf("  %s: %.2f\n", s, s.Area())
    }

    // Type assertion
    var s Shape = Circle{Radius: 3}
    if c, ok := s.(Circle); ok {
        fmt.Printf("\nExtracted Circle radius: %.1f\n", c.Radius)
    }

    // Type switch
    for _, shape := range shapes {
        switch v := shape.(type) {
        case Circle:
            fmt.Printf("Circle with radius %.1f\n", v.Radius)
        case Rectangle:
            fmt.Printf("Rectangle %.1f x %.1f\n", v.Width, v.Height)
        case Triangle:
            fmt.Printf("Triangle with sides %.1f, %.1f, %.1f\n", v.A, v.B, v.C)
        }
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Pointer receiver + value assigned to interface
```go
type Greeter interface{ Greet() string }
type Person struct{ Name string }

func (p *Person) Greet() string { return "Hi, " + p.Name }

var g Greeter = Person{"Alice"}    // ❌ compile error
var g2 Greeter = &Person{"Alice"}  // ✅
```

### 2. Nil interface vs interface holding nil pointer
```go
var err *MyError = nil
return err   // ❌ returns non-nil interface!
return nil   // ✅ returns true nil interface
```

### 3. Type asserting without comma-ok
```go
var i any = "hello"
n := i.(int)   // ❌ PANIC — use comma-ok form
n, ok := i.(int)  // ✅ n=0, ok=false
```

### 4. Over-engineering with large interfaces
```go
// ❌ Too large — hard to satisfy, hard to mock in tests
type Storage interface {
    Get(key string) ([]byte, error)
    Set(key string, val []byte) error
    Delete(key string) error
    List() ([]string, error)
    Flush() error
    Stats() StorageStats
    // ...10 more methods
}

// ✅ Smaller, focused interfaces
type Getter interface { Get(key string) ([]byte, error) }
type Setter interface { Set(key string, val []byte) error }
```

---

## 🎯 Must-Know Things (Quick Reference)

- Interfaces are satisfied **implicitly** — no `implements` keyword
- An interface value has two parts: **type** and **data**
- `any` = `interface{}` — every type satisfies the empty interface
- **Type assertion**: `v, ok := i.(T)` — always use comma-ok form
- **Type switch**: `switch v := i.(type)` — cleanest multi-type handling
- **Interface composition** — embed interfaces to build larger ones
- Pointer receiver methods are NOT in the method set of value types — assign `*T` to interface
- A **nil interface** (`var s Shape`) is truly nil — an **interface holding nil pointer** is NOT
- Prefer **small, focused interfaces** — the Go standard library models this
- Returning a typed nil pointer as an error is a classic bug — always `return nil`

---

## ❓ Interview Questions & Answers

---

**Q1: How does Go's interface implementation differ from Java or C#?**

> Go uses implicit (structural) interface satisfaction — a type satisfies an interface simply by having the required methods, with no `implements` declaration needed. Java and C# require explicit declaration (`implements`, `:`). This decouples interface definition from implementation — a type can satisfy interfaces from packages it has never imported. It also means you can define an interface after the fact that existing types already satisfy.

---

**Q2: What is the internal structure of an interface value?**

> An interface value is a two-word pair: a pointer to the type descriptor (containing type metadata and method table) and a pointer to the concrete data. This allows an interface to hold any type dynamically. A nil interface has both words as nil. An interface holding a nil pointer has the type word set (non-nil) and the data word nil — which is why it doesn't equal nil even though the underlying value is nil.

---

**Q3: What is the nil interface vs interface holding nil pointer problem?**

> When you store a typed nil pointer in an interface, the interface is NOT nil — its type field is set. This causes the famous error-returning bug: `var err *MyError = nil; return err` returns a non-nil `error` interface. The caller's `if err != nil` is true even though you intended to return "no error". The fix is to always `return nil` directly for nil errors, never return a typed nil through an interface.

---

**Q4: What are the method set rules for interfaces?**

> A value type `T` has only value receiver methods in its method set. A pointer type `*T` has both value and pointer receiver methods. This means: if you use pointer receivers, you must store `*T` (not `T`) in the interface — only `*T` satisfies an interface requiring pointer receiver methods. If you use value receivers, both `T` and `*T` satisfy the interface.

---

**Q5: What is interface composition and why is it useful?**

> Interface composition is embedding one or more interfaces inside another to create a larger contract. For example, `io.ReadWriter` embeds both `io.Reader` and `io.Writer`. It allows building complex interfaces from small primitives without repeating method signatures. It promotes the Go philosophy of small, focused interfaces — you compose them only when you need a function to require multiple behaviors simultaneously.

---

**Q6: What is a type assertion and when should you use the comma-ok form?**

> A type assertion `v := i.(T)` extracts the concrete value of type `T` from interface `i`. If `i` doesn't hold a `T`, it panics. The comma-ok form `v, ok := i.(T)` is safe — it returns the zero value and `false` if the type doesn't match, without panicking. Always use comma-ok unless you are absolutely certain of the type — which is rarely the case in production code.

---

**Q7: What does "the bigger the interface, the weaker the abstraction" mean?**

> This Rob Pike quote captures a key Go design principle. A large interface (many methods) is harder to satisfy — fewer types will implement it, and mocking it in tests requires implementing many methods. A small interface (one or two methods) is easy to satisfy, highly reusable, and can be composed. `io.Reader` with its single `Read` method is satisfied by files, network connections, byte buffers, strings, and more — it's powerful precisely because it's minimal.

---

*Next up: **Topic 2.2 — Error Handling** 🚀*
