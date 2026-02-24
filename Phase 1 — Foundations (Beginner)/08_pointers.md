# 1.8 — Pointers

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.7 — Functions

---

## 📖 What is a Pointer?

A **pointer** is a variable that stores the **memory address** of another variable, rather than a value itself.

```
Normal variable:   x  →  [  42  ]   (holds the value directly)
Pointer variable:  p  →  [ 0xc0000b4008 ]  →  [  42  ]   (holds the address of x)
```

Go has pointers, but unlike C/C++, Go has **no pointer arithmetic** — you can't do `p++` to move to the next memory address. This makes pointers safer and simpler.

---

## 🔑 Two Core Operators

| Operator | Name | Meaning |
|----------|------|---------|
| `&` | address-of | Gets the memory address of a variable |
| `*` | dereference | Gets the value at a memory address |

```go
x := 42
p := &x          // p holds the address of x

fmt.Println(x)   // 42       — the value
fmt.Println(&x)  // 0xc000...— the address of x
fmt.Println(p)   // 0xc000...— same address (p stores it)
fmt.Println(*p)  // 42       — dereference: value at that address

// Modify x through the pointer
*p = 100
fmt.Println(x)   // 100 — x was changed via the pointer!
```

---

## 📝 Declaring Pointers

```go
// Type of a pointer to int is *int
var p *int          // p is a pointer to int — zero value is nil
fmt.Println(p)      // <nil>

x := 42
p = &x              // now p points to x
fmt.Println(*p)     // 42
```

The type `*int` means "pointer to int". `*string` means "pointer to string", etc.

---

## 🏗️ The `new` Function

`new(T)` allocates memory for a value of type `T`, initializes it to its zero value, and returns a pointer to it:

```go
p := new(int)        // allocates an int, initializes to 0, returns *int
fmt.Println(*p)      // 0
*p = 42
fmt.Println(*p)      // 42
```

`new` is less commonly used in Go — most of the time you'll use `&` with a struct literal or `:=` instead.

---

## ❓ Why Do Pointers Exist?

Two main reasons:

### 1. Sharing and mutation — pass by reference

By default, Go passes everything **by value** — a copy is made. Changes inside a function don't affect the original:

```go
func double(x int) {
    x = x * 2   // modifies the local copy — original unchanged
}

func main() {
    n := 5
    double(n)
    fmt.Println(n)   // still 5 — double had no effect!
}
```

With a pointer, you pass the address — the function can modify the original:

```go
func double(x *int) {
    *x = *x * 2   // modifies the value AT the address
}

func main() {
    n := 5
    double(&n)      // pass the address of n
    fmt.Println(n)  // 10 — double worked!
}
```

### 2. Efficiency — avoid copying large data

When you pass a large struct by value, Go copies the entire thing. With a pointer, only the address (8 bytes on 64-bit) is copied:

```go
type BigStruct struct {
    Data [1000000]int   // 8MB of data
}

// ❌ Copies 8MB every call — expensive
func processValue(s BigStruct) { ... }

// ✅ Copies only 8 bytes (a pointer) — efficient
func processPointer(s *BigStruct) { ... }
```

---

## 🏗️ Pointers to Structs

This is where you'll use pointers most often in Go. When working with structs via pointer, Go automatically dereferences — you don't need to write `(*p).field`:

```go
type Person struct {
    Name string
    Age  int
}

p := &Person{Name: "Alice", Age: 30}

// Both of these are equivalent:
fmt.Println((*p).Name)   // Alice — explicit dereference
fmt.Println(p.Name)      // Alice — Go auto-dereferences (idiomatic)

// Modify fields via pointer
p.Age = 31               // Go auto-dereferences
fmt.Println(p.Age)       // 31
```

---

## nil Pointers — The Zero Value

The zero value of any pointer type is `nil`. Dereferencing a nil pointer causes a **panic** (runtime crash):

```go
var p *int
fmt.Println(p)    // <nil>
fmt.Println(*p)   // PANIC: runtime error: invalid memory address or nil pointer dereference
```

Always check for nil before dereferencing:

```go
func printValue(p *int) {
    if p == nil {
        fmt.Println("pointer is nil")
        return
    }
    fmt.Println(*p)
}
```

---

## 🔄 Pointers vs Values — When to Use Which

This is one of the most important decisions in Go. Here are the rules:

| Situation | Use Pointer | Use Value |
|-----------|------------|-----------|
| Need to mutate the original | ✅ | ❌ |
| Large struct (avoid copying) | ✅ | ❌ |
| Method needs to modify receiver | ✅ | ❌ |
| Small struct (e.g. Point{x,y}) | ❌ | ✅ |
| Primitive types (int, string) | Rarely | ✅ |
| Need to express "optional" / absence | ✅ (can be nil) | ❌ |
| Concurrent use needs care | ✅ (shared state) | ✅ (copies are safe) |

---

## 🔁 Pointers and Functions — Practical Patterns

### Pattern 1: Mutate via pointer

```go
func increment(n *int) {
    *n++
}

x := 5
increment(&x)
fmt.Println(x)   // 6
```

### Pattern 2: Return a pointer from a function

In Go, returning a pointer to a local variable is **safe** — Go's escape analysis moves it to the heap automatically:

```go
func newPerson(name string, age int) *Person {
    p := Person{Name: name, Age: age}
    return &p   // ✅ Safe in Go — p escapes to the heap
}
```

This would be a use-after-free bug in C, but Go's garbage collector handles it.

### Pattern 3: Optional values with pointers

A pointer can be `nil`, which makes it useful for optional fields:

```go
type Config struct {
    Port    int
    Timeout *int   // nil means "use default"
}

cfg := Config{Port: 8080}
if cfg.Timeout == nil {
    fmt.Println("Using default timeout")
} else {
    fmt.Println("Timeout:", *cfg.Timeout)
}
```

---

## 💻 Complete Example

```go
package main

import "fmt"

type Rectangle struct {
    Width  float64
    Height float64
}

func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func swap(a, b *int) {
    *a, *b = *b, *a
}

func newRectangle(w, h float64) *Rectangle {
    return &Rectangle{Width: w, Height: h}
}

func main() {
    // Basic pointer usage
    x := 10
    p := &x
    fmt.Printf("x = %d, address = %v\n", x, p)
    *p = 20
    fmt.Println("After *p = 20, x =", x)   // 20

    // Swap via pointers
    a, b := 5, 10
    swap(&a, &b)
    fmt.Println("After swap:", a, b)   // 10 5

    // Pointer to struct
    r := newRectangle(4, 5)
    fmt.Printf("Area before scale: %.1f\n", r.Area())   // 20.0
    r.Scale(2)
    fmt.Printf("After Scale(2): %.1f x %.1f\n", r.Width, r.Height)   // 8.0 x 10.0
    fmt.Printf("Area after scale: %.1f\n", r.Area())    // 80.0

    // nil pointer check
    var ptr *int
    if ptr == nil {
        fmt.Println("ptr is nil — safe, no panic")
    }

    // new()
    n := new(int)
    *n = 99
    fmt.Println("new(int):", *n)   // 99
}
```

Output:
```
x = 10, address = 0xc0000b4008
After *p = 20, x = 20
After swap: 10 5
Area before scale: 20.0
After Scale(2): 8.0 x 10.0
Area after scale: 80.0
ptr is nil — safe, no panic
new(int): 99
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Dereferencing a nil pointer
```go
var p *int
*p = 42   // PANIC — nil pointer dereference
// Always check: if p != nil { *p = 42 }
```

### 2. Thinking Go passes pointers automatically for slices/maps
```go
// Slices and maps already contain a pointer internally —
// they DON'T need & to share state:
func addItem(s []int) {
    s = append(s, 99)   // ❌ won't affect the caller's slice header
}

func addItemPtr(s *[]int) {
    *s = append(*s, 99)   // ✅ modifies caller's slice
}
// (More on this in Topic 1.9 — Arrays & Slices)
```

### 3. Pointer to loop variable
```go
// ❌ Classic bug — all pointers point to the same variable!
ptrs := make([]*int, 3)
for i := 0; i < 3; i++ {
    ptrs[i] = &i   // all point to same i
}
fmt.Println(*ptrs[0], *ptrs[1], *ptrs[2])   // 3 3 3

// ✅ Fix — capture current value
for i := 0; i < 3; i++ {
    v := i
    ptrs[i] = &v   // each points to its own v
}
```

### 4. No pointer arithmetic
```go
p := &x
p++   // ❌ compile error — Go does not allow pointer arithmetic
p+1   // ❌ compile error
```

### 5. Comparing pointers
```go
a, b := 42, 42
p1 := &a
p2 := &b
p3 := &a

fmt.Println(p1 == p2)   // false — different addresses
fmt.Println(p1 == p3)   // true  — same address
fmt.Println(*p1 == *p2) // true  — same value
```

---

## 🎯 Must-Know Things (Quick Reference)

- `&x` — gets the address of `x` (creates a pointer)
- `*p` — dereferences pointer `p` (gets the value at that address)
- Zero value of any pointer is `nil` — dereferencing nil causes a **panic**
- Go passes everything by **value** — use pointers to share/mutate
- Go **auto-dereferences** struct pointer fields — `p.Name` instead of `(*p).Name`
- Returning a pointer to a local variable is **safe** — Go's GC handles it
- **No pointer arithmetic** in Go — safer than C
- `new(T)` allocates a zeroed `T` and returns `*T` — less common than `&T{}`
- Pointers make a field/param **optional** (can be `nil`)
- Loop variable pointer bug — always shadow or copy before taking address

---

## ❓ Interview Questions & Answers

---

**Q1: What is a pointer in Go and what are the two pointer operators?**

> A pointer is a variable that stores the memory address of another variable. Go has two pointer operators: `&` (address-of) gets the address of a variable, and `*` (dereference) gets the value stored at an address. For example: `p := &x` makes `p` point to `x`, and `*p` reads or writes the value of `x` through `p`.

---

**Q2: Go passes arguments by value. How do pointers change this?**

> By default, Go copies the value when passing to a function — the original is unaffected. To allow a function to modify the original, you pass a pointer (`&x`) and the function accepts `*T`. Inside the function, you dereference with `*` to read or write the original value. This is Go's way of achieving pass-by-reference semantics.

---

**Q3: What happens when you dereference a nil pointer?**

> The program panics at runtime with "invalid memory address or nil pointer dereference". The zero value of any pointer type is `nil`, meaning it points to no memory. Before dereferencing any pointer you received from outside your function, you should check `if p != nil`. This is especially important for struct pointer fields and function parameters.

---

**Q4: Is it safe to return a pointer to a local variable in Go?**

> Yes — this is safe and idiomatic in Go. Go's escape analysis detects that the variable's address is being returned and automatically allocates it on the heap instead of the stack. The garbage collector ensures it lives as long as there's a reference to it. This would be a dangerous use-after-free bug in C/C++, but Go handles it for you.

---

**Q5: Why does Go auto-dereference struct pointer fields?**

> When you have a pointer to a struct `p *Person`, Go allows you to write `p.Name` instead of the explicit `(*p).Name`. This is syntactic sugar — the compiler inserts the dereference automatically. It makes working with struct pointers as natural as working with struct values, which is important because methods often use pointer receivers.

---

**Q6: When should you use a pointer vs a value in Go?**

> Use a pointer when: (1) the function or method needs to mutate the original value; (2) the struct is large and copying would be expensive; (3) you need to express "no value" / optionality (nil pointer). Use a value when: (1) the data is small (primitives, small structs); (2) you want an independent copy; (3) the value is used concurrently and you don't want shared mutable state. For methods on a type, pick one consistently — either all pointer receivers or all value receivers (with some exceptions).

---

**Q7: What is the pointer-to-loop-variable bug?**

> When you take the address of a loop variable inside a loop, all the resulting pointers point to the same variable — the loop variable itself. After the loop ends, they all hold the variable's final value. The fix is to shadow the variable inside the loop: `v := i; ptrs[i] = &v`. This creates a new variable for each iteration. This same bug appears with goroutines in loops and is one of the most common Go mistakes.

---

*Next up: **Topic 1.9 — Arrays & Slices** 🚀*
