# 3.1 — Generics (Go 1.18+)

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 2.11 — Working with JSON

---

## 📖 What Are Generics?

Before Go 1.18, if you wanted a function that worked on multiple types, you had two options: duplicate the function for each type, or use `interface{}` and lose type safety. **Generics** solve this with **type parameters** — letting you write a single function or type that works with many types while remaining type-safe at compile time.

```go
// Before generics — duplicated for each type
func SumInts(nums []int) int {
    total := 0
    for _, n := range nums { total += n }
    return total
}
func SumFloats(nums []float64) float64 {
    total := 0.0
    for _, n := range nums { total += n }
    return total
}

// With generics — one function for all numeric types
func Sum[T int | float64](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
```

---

## 🧩 Type Parameters — Basic Syntax

A **type parameter** is declared in square brackets before the function parameters:

```go
func FuncName[T Constraint](param T) T { ... }
//            ↑ type param ↑ constraint
```

```go
// Single type parameter
func First[T any](slice []T) T {
    return slice[0]
}

// Multiple type parameters
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Type inferred from arguments — no need to specify [int] explicitly
nums := First([]int{1, 2, 3})         // T inferred as int
strs := First([]string{"a", "b"})    // T inferred as string

doubled := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
// [2, 4, 6]

lengths := Map([]string{"go", "rust"}, func(s string) int { return len(s) })
// [2, 4]
```

---

## 📐 Constraints — `any` and `comparable`

A **constraint** is an interface that restricts which types a type parameter can be:

```go
any         // any type — alias for interface{}
comparable  // any type that supports == and != (can be used as map key)
```

```go
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

Contains([]int{1, 2, 3}, 2)          // true
Contains([]string{"a", "b"}, "c")   // false
```

`comparable` is needed whenever you use `==` on a type parameter — without it, compile error.

---

## 📋 Union Constraints — `|` Operator

Constraints can be **unions** of types:

```go
// Inline union
func Min[T int | float64 | string](a, b T) T {
    if a < b { return a }
    return b
}

// Named constraint — interface with type set
type Number interface {
    int | int8 | int16 | int32 | int64 |
    float32 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
```

### The `~` prefix — underlying type

`~T` means "any type whose **underlying type** is T" — includes custom types built on T:

```go
type Celsius    float64
type Fahrenheit float64

type Temperature interface {
    ~float64   // matches float64, Celsius, Fahrenheit
}

func AbsTemp[T Temperature](t T) T {
    if t < 0 { return -t }
    return t
}

AbsTemp(float64(-3.14))     // ✅
AbsTemp(Celsius(-10.0))     // ✅ — underlying type is float64
AbsTemp(Fahrenheit(-40.0))  // ✅ — underlying type is float64
```

---

## 🏗️ Ordered Constraint

Define your own `Ordered` for types supporting `<`, `>` etc. (or use `cmp.Ordered` from Go 1.21):

```go
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64 | ~string
}

func Max[T Ordered](a, b T) T {
    if a > b { return a }
    return b
}

func Clamp[T Ordered](val, lo, hi T) T {
    if val < lo { return lo }
    if val > hi { return hi }
    return val
}

Max(3, 7)              // 7
Max("apple", "fig")    // "fig"
Clamp(15, 0, 10)       // 10
```

---

## 🏭 Generic Types — Structs and Methods

Type parameters work on structs too:

```go
// Generic Stack
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T   // zero value of T
        return zero, false
    }
    last := len(s.items) - 1
    item := s.items[last]
    s.items = s.items[:last]
    return item, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Usage — type specified at instantiation
intStack := Stack[int]{}
intStack.Push(1)
intStack.Push(2)
v, _ := intStack.Pop()   // 2

strStack := Stack[string]{}
strStack.Push("hello")
```

### Generic Pair type

```go
type Pair[A, B any] struct {
    First  A
    Second B
}

func NewPair[A, B any](a A, b B) Pair[A, B] {
    return Pair[A, B]{First: a, Second: b}
}

p := NewPair("age", 30)             // Pair[string, int]
fmt.Println(p.First, p.Second)      // age 30
```

---

## 🛠️ Common Generic Utility Functions

```go
// Filter — keep elements matching predicate
func Filter[T any](slice []T, pred func(T) bool) []T {
    var out []T
    for _, v := range slice {
        if pred(v) { out = append(out, v) }
    }
    return out
}

// Reduce — fold to a single value
func Reduce[T, U any](slice []T, init U, fn func(U, T) U) U {
    acc := init
    for _, v := range slice { acc = fn(acc, v) }
    return acc
}

// Keys — extract map keys
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m { keys = append(keys, k) }
    return keys
}

// Must — generic Must pattern (panics on error)
func Must[T any](v T, err error) T {
    if err != nil { panic(err) }
    return v
}

// Ptr — get pointer to a value (useful in tests)
func Ptr[T any](v T) *T { return &v }
```

---

## 📌 Constraints with Methods

Constraints can require methods, not just types:

```go
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Combining type set AND method requirement
type OrderedStringer interface {
    Ordered
    String() string
}
```

---

## 💻 Complete Example

```go
package main

import "fmt"

type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64 | ~string
}

func Min[T Ordered](a, b T) T { if a < b { return a }; return b }
func Max[T Ordered](a, b T) T { if a > b { return a }; return b }

func Map[T, U any](slice []T, fn func(T) U) []U {
    out := make([]U, len(slice))
    for i, v := range slice { out[i] = fn(v) }
    return out
}

func Filter[T any](slice []T, pred func(T) bool) []T {
    var out []T
    for _, v := range slice {
        if pred(v) { out = append(out, v) }
    }
    return out
}

func Reduce[T, U any](slice []T, init U, fn func(U, T) U) U {
    acc := init
    for _, v := range slice { acc = fn(acc, v) }
    return acc
}

// Generic Set
type Set[T comparable] struct{ m map[T]struct{} }

func NewSet[T comparable]() *Set[T] {
    return &Set[T]{m: make(map[T]struct{})}
}
func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

// Generic Result type
type Result[T any] struct {
    value T
    err   error
}

func Ok[T any](v T) Result[T]      { return Result[T]{value: v} }
func Err[T any](e error) Result[T] { return Result[T]{err: e} }
func (r Result[T]) Unwrap() (T, error) { return r.value, r.err }
func (r Result[T]) IsOk() bool         { return r.err == nil }

func main() {
    fmt.Println("=== Min / Max ===")
    fmt.Println(Min(3, 7))            // 3
    fmt.Println(Max("apple", "fig")) // fig

    fmt.Println("\n=== Map / Filter / Reduce ===")
    nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    evens := Filter(nums, func(n int) bool { return n%2 == 0 })
    fmt.Println("evens:", evens)   // [2 4 6 8 10]

    doubled := Map(evens, func(n int) int { return n * 2 })
    fmt.Println("doubled:", doubled)   // [4 8 12 16 20]

    sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
    fmt.Println("sum:", sum)   // 55

    words := []string{"go", "rust", "python"}
    lengths := Map(words, func(s string) int { return len(s) })
    fmt.Println("lengths:", lengths)   // [2 4 6]

    fmt.Println("\n=== Generic Set ===")
    s := NewSet[string]()
    s.Add("go"); s.Add("rust"); s.Add("go")   // duplicate ignored
    fmt.Println("has 'go':", s.Has("go"))     // true
    fmt.Println("has 'java':", s.Has("java")) // false
    fmt.Println("size:", s.Len())             // 2

    fmt.Println("\n=== Generic Result ===")
    r1 := Ok(42)
    r2 := Err[int](fmt.Errorf("something failed"))
    v, err := r1.Unwrap()
    fmt.Printf("r1: value=%d err=%v\n", v, err)        // r1: value=42 err=<nil>
    _, err = r2.Unwrap()
    fmt.Printf("r2: isOk=%v err=%v\n", r2.IsOk(), err) // r2: isOk=false err=something failed
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Cannot use operators without the right constraint
```go
// ❌ T is any — no guarantee of < operator
func Min[T any](a, b T) T {
    if a < b { return a }   // compile error
}
// ✅ Use Ordered constraint
func Min[T Ordered](a, b T) T { ... }
```

### 2. `comparable` ≠ `Ordered` — no `<` `>`
```go
// ❌ comparable only gives == and !=
func Min[T comparable](a, b T) T {
    if a < b { return a }   // compile error
}
```

### 3. No type parameters on methods
```go
type MyType struct{}
// ❌ Not allowed
func (m MyType) Do[T any](v T) {}

// ✅ Make the type itself generic
type MyType[T any] struct{}
func (m MyType[T]) Do(v T) {}
```

### 4. Type inference fails for return-only type params
```go
func Zero[T any]() T { var z T; return z }
// ❌ Can't infer T — not in any argument
_ = Zero()

// ✅ Specify explicitly
_ = Zero[int]()
```

### 5. Over-generifying code
```go
// ❌ Unnecessary — only ever called with []int anyway
func ProcessInts[T ~int](nums []T) { ... }

// ✅ Just use []int if that's the only real use
func ProcessInts(nums []int) { ... }
```

---

## 🎯 Must-Know Things (Quick Reference)

- Type parameters: `[T Constraint]` before function or type declaration
- `any` = allow any type; `comparable` = supports `==` (needed for maps/sets)
- `|` = union of types in constraint; `~T` = underlying type T
- Define `Ordered` yourself or import `cmp.Ordered` (Go 1.21+) for `<` `>`
- Generic types: `type Stack[T any] struct{}` — instantiate as `Stack[int]{}`
- Methods on generic types: `func (s Stack[T]) Push(item T)`
- **Methods cannot have their own type parameters** — only top-level funcs/types
- Type inference works from arguments — explicit `[T]` only when needed
- Generics are **compile-time** — zero runtime overhead
- Don't over-generify — concrete types are clearer when generics aren't needed

---

## ❓ Interview Questions & Answers

---

**Q1: What problem do generics solve in Go and when were they introduced?**

> Generics were introduced in Go 1.18 (March 2022). Before generics, writing type-safe code that worked across multiple types required either duplicating functions for each type or using `interface{}` — losing type safety and requiring type assertions at runtime. Generics allow type parameters so a single function or type can operate on multiple types while remaining fully type-safe at compile time. The compiler generates specialized code for each concrete type used, so there's no runtime overhead from the abstraction.

---

**Q2: What is a constraint and what are the built-in constraints?**

> A constraint is an interface that restricts which types a type parameter can be. Every type parameter must have one. The two built-in constraints are `any` (alias for `interface{}` — any type) and `comparable` (any type supporting `==` and `!=`, usable as a map key). You define custom constraints as interfaces with type sets — using `|` for unions and `~T` to include types based on underlying type. The `cmp` package (Go 1.21+) provides `cmp.Ordered` for types supporting ordering operators.

---

**Q3: What does `~T` mean in a type constraint?**

> `~T` means "any type whose underlying type is T" — it includes not just T itself but any named type defined in terms of T. For example, `~float64` matches `float64`, but also `type Celsius float64` and `type Fahrenheit float64`. Without `~`, a constraint of just `float64` wouldn't match `Celsius`. The `~` prefix makes generic code work with user-defined types built on primitives, which is critical for practical use.

---

**Q4: Can you add type parameters to methods in Go?**

> No. Go does not allow methods to have their own type parameters — only top-level functions and types can have them. If you need a method to be generic over a type, make the receiver type itself generic: `type Stack[T any] struct{}` with methods like `func (s *Stack[T]) Push(item T)`. This is a deliberate design choice — it keeps method sets predictable and interface satisfaction rules simple. The workaround for a standalone generic method is a package-level generic function.

---

**Q5: When should you use generics vs interfaces?**

> Use **interfaces** when you need runtime polymorphism — different types stored together in a collection, or passed around as abstract values where the concrete type is unknown until runtime. Use **generics** when you need compile-time type safety across multiple types with the same algorithmic structure — collections (Stack, Set), utility functions (Map, Filter, Min), or when avoiding type assertions. The key heuristic: if the type parameter is only ever resolved at the call site (not stored as a mixed-type interface), generics are the right tool.

---

**Q6: How does type inference work with generics?**

> The Go compiler infers type parameters from the types of function arguments — you usually don't need to specify them explicitly. `Max(3, 5)` infers `T = int`. Inference works when the type parameter appears in at least one function parameter. It doesn't work when the type parameter only appears in the return type with no corresponding input — like `func Zero[T any]() T` — where you must specify explicitly: `Zero[int]()`. When inference is ambiguous or unavailable, specify the type parameter in square brackets after the function name.

---

*Next up: **Topic 3.2 — Context Package** 🚀*
