# 3.5 — `unsafe` Package

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.4 — Reflection

---

## 📖 What is `unsafe`?

The `unsafe` package lets you **bypass Go's type system and memory safety guarantees**. It gives direct access to memory layout, pointer arithmetic, and size/alignment information. It is genuinely unsafe — bugs can cause silent data corruption, crashes, and non-portable programs.

```go
import "unsafe"
```

The package is built into the compiler — it has no Go source. Its "functions" are handled specially by the compiler.

**Rule of thumb:** If you're not writing a systems library, serializer, or working at the OS/hardware boundary, you almost certainly don't need `unsafe`. When in doubt, don't.

---

## 📐 `unsafe.Sizeof`, `unsafe.Alignof`, `unsafe.Offsetof`

These are the three most useful and safe parts of `unsafe` — they return memory layout information at compile time:

```go
type Point struct {
    X int32   // 4 bytes
    Y int32   // 4 bytes
}

type Mixed struct {
    A bool    // 1 byte
    B int64   // 8 bytes — but 7 bytes padding after A!
    C int32   // 4 bytes
    D bool    // 1 byte + 3 bytes padding at end
}

fmt.Println(unsafe.Sizeof(Point{}))    // 8   (4+4)
fmt.Println(unsafe.Sizeof(Mixed{}))    // 24  (1+7pad+8+4+1+3pad)
fmt.Println(unsafe.Sizeof(int64(0)))   // 8
fmt.Println(unsafe.Sizeof(""))         // 16  (ptr + len on 64-bit)
fmt.Println(unsafe.Sizeof([]int{}))    // 24  (ptr + len + cap)

fmt.Println(unsafe.Alignof(Mixed{}.B))   // 8
fmt.Println(unsafe.Offsetof(Mixed{}.B))  // 8 (1 byte A + 7 pad)
fmt.Println(unsafe.Offsetof(Mixed{}.C))  // 16
```

### Struct field padding — minimize with field ordering

```go
// ❌ Wasteful — 24 bytes due to padding
type Bad struct {
    A bool    // 1 + 7 padding
    B int64   // 8
    C bool    // 1 + 7 padding
}

// ✅ Compact — 10 bytes (with 6 padding = 16 total on 64-bit)
type Good struct {
    B int64   // 8  — largest fields first
    A bool    // 1
    C bool    // 1  + 6 padding
}

fmt.Println(unsafe.Sizeof(Bad{}))    // 24
fmt.Println(unsafe.Sizeof(Good{}))   // 16
```

---

## 🔑 `unsafe.Pointer` — The Core Type

`unsafe.Pointer` is a special pointer type that:
- Can be converted to/from **any pointer type**
- Can be converted to/from `uintptr` (for arithmetic)
- Bypasses Go's type safety

```go
// The four legal conversions involving unsafe.Pointer:
// 1. *T → unsafe.Pointer
// 2. unsafe.Pointer → *T
// 3. unsafe.Pointer → uintptr  (for arithmetic ONLY, immediately back to Pointer)
// 4. uintptr → unsafe.Pointer  (only valid if uintptr came from Pointer in the same expression)
```

### Reinterpreting memory — type punning

Convert bytes to float64 without copying (used in low-level serializers):

```go
// Interpret the bits of a float64 as uint64 (IEEE 754 tricks)
func float64bits(f float64) uint64 {
    return *(*uint64)(unsafe.Pointer(&f))
}

fmt.Printf("%x\n", float64bits(1.0))   // 3ff0000000000000
```

### Convert between slice types (zero-copy)

```go
// Convert []byte to string without copying
func bytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

// Convert string to []byte without copying
// ⚠️ The resulting slice must NOT be written to — string is immutable!
func stringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
```

> These zero-copy conversions are used in hot paths in libraries like `fasthttp`. In normal code, just use `string(b)` and `[]byte(s)`.

---

## 📍 Pointer Arithmetic with `uintptr`

`uintptr` is an integer type large enough to hold any pointer value. Combined with `unsafe.Pointer`, you can do pointer arithmetic — but it's **extremely dangerous**:

```go
type Point struct { X, Y int32 }
p := Point{X: 10, Y: 20}

// Access field Y by offset arithmetic
yPtr := (*int32)(unsafe.Pointer(
    uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.Y),
))
fmt.Println(*yPtr)   // 20
*yPtr = 99
fmt.Println(p.Y)     // 99
```

> **Critical rule:** Never store a `uintptr` in a variable between the conversion from `unsafe.Pointer` and back. The GC can move objects — a stored `uintptr` can become a dangling pointer. The conversion must happen in a **single expression**.

```go
// ❌ DANGEROUS — GC may move p between these two lines
addr := uintptr(unsafe.Pointer(&p))
// ... GC runs, p is moved ...
ptr := (*int32)(unsafe.Pointer(addr))   // addr is now invalid!

// ✅ SAFE — single expression
ptr := (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + offset))
```

---

## 🔗 `unsafe.Slice` and `unsafe.String` (Go 1.17+)

Create slices and strings from raw pointers:

```go
// unsafe.Slice(ptr *T, len) []T — create a slice from pointer + length
data := [4]byte{1, 2, 3, 4}
slice := unsafe.Slice(&data[0], 4)   // []byte{1, 2, 3, 4}

// unsafe.SliceData(s []T) *T — get pointer to slice's backing array
ptr := unsafe.SliceData(slice)

// unsafe.String(ptr *byte, len) string — create string from pointer + length
str := unsafe.String(&data[0], 4)

// unsafe.StringData(s string) *byte — get pointer to string's data
strPtr := unsafe.StringData(str)
_ = strPtr
```

---

## 🏗️ Memory Layout — Why It Matters

Understanding struct layout helps with:
- **Performance** — cache-line alignment, reducing padding
- **Interop** — matching C struct layouts for cgo
- **Serialization** — reading binary protocols directly into structs

### Cache line alignment (64 bytes on x86)

```go
// ❌ False sharing — X and Y on same cache line, contested by two goroutines
type Counters struct {
    X int64
    Y int64
}

// ✅ Pad to separate cache lines — each goroutine owns its cache line
type PaddedCounter struct {
    Value int64
    _     [56]byte   // pad to 64 bytes
}

type Counters struct {
    X PaddedCounter
    Y PaddedCounter
}
```

---

## 🔄 Accessing Unexported Fields (Carefully)

One use: accessing unexported fields of structs from external packages — for testing or interop. This is fragile and breaks with Go version changes:

```go
package main

import (
    "reflect"
    "unsafe"
)

type externalStruct struct {
    public  int
    private int   // unexported
}

func readPrivate(s *externalStruct) int {
    t := reflect.TypeOf(*s)
    field, _ := t.FieldByName("private")
    ptr := unsafe.Pointer(uintptr(unsafe.Pointer(s)) + field.Offset)
    return *(*int)(ptr)
}
```

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "unsafe"
)

// --- Struct layout analysis ---
type Inefficient struct {
    A bool    // 1 + 7 pad
    B int64   // 8
    C bool    // 1 + 7 pad
    D int32   // 4 + 4 pad
}

type Efficient struct {
    B int64   // 8
    D int32   // 4
    A bool    // 1
    C bool    // 1 + 2 pad
}

// --- Zero-copy string/bytes ---
func unsafeString(b []byte) string {
    if len(b) == 0 { return "" }
    return unsafe.String(&b[0], len(b))
}

// --- Type punning: float64 bits ---
func f64bits(f float64) uint64 {
    return *(*uint64)(unsafe.Pointer(&f))
}

func f64fromBits(u uint64) float64 {
    return *(*float64)(unsafe.Pointer(&u))
}

// --- Field access via offset ---
type Config struct {
    Host string
    Port int
    debug bool   // unexported
}

func setDebug(c *Config, v bool) {
    t := struct{ Host string; Port int; debug bool }{}
    _ = t
    // Use offsetof pattern
    off := unsafe.Offsetof(struct{ a string; b int; c bool }{}.c)
    // Simpler approach using known offsets:
    ptr := (*bool)(unsafe.Pointer(
        uintptr(unsafe.Pointer(c)) +
        unsafe.Sizeof(c.Host) +
        unsafe.Sizeof(c.Port),
    ))
    *ptr = v
    _ = off
}

func main() {
    // Struct sizes
    fmt.Println("=== Struct Layout ===")
    fmt.Printf("Inefficient: %d bytes\n", unsafe.Sizeof(Inefficient{}))   // 32
    fmt.Printf("Efficient:   %d bytes\n", unsafe.Sizeof(Efficient{}))     // 16

    fmt.Println("\nInefficient offsets:")
    var bad Inefficient
    fmt.Printf("  A: offset=%d\n", unsafe.Offsetof(bad.A))   // 0
    fmt.Printf("  B: offset=%d\n", unsafe.Offsetof(bad.B))   // 8
    fmt.Printf("  C: offset=%d\n", unsafe.Offsetof(bad.C))   // 16
    fmt.Printf("  D: offset=%d\n", unsafe.Offsetof(bad.D))   // 24

    fmt.Println("\nEfficient offsets:")
    var good Efficient
    fmt.Printf("  B: offset=%d\n", unsafe.Offsetof(good.B))  // 0
    fmt.Printf("  D: offset=%d\n", unsafe.Offsetof(good.D))  // 8
    fmt.Printf("  A: offset=%d\n", unsafe.Offsetof(good.A))  // 12
    fmt.Printf("  C: offset=%d\n", unsafe.Offsetof(good.C))  // 13

    // Type sizes reference
    fmt.Println("\n=== Type Sizes ===")
    fmt.Printf("bool:      %d\n", unsafe.Sizeof(false))
    fmt.Printf("int:       %d\n", unsafe.Sizeof(int(0)))
    fmt.Printf("int32:     %d\n", unsafe.Sizeof(int32(0)))
    fmt.Printf("int64:     %d\n", unsafe.Sizeof(int64(0)))
    fmt.Printf("float64:   %d\n", unsafe.Sizeof(float64(0)))
    fmt.Printf("string:    %d\n", unsafe.Sizeof(""))
    fmt.Printf("[]int:     %d\n", unsafe.Sizeof([]int{}))
    fmt.Printf("map:       %d\n", unsafe.Sizeof(map[int]int{}))
    fmt.Printf("interface: %d\n", unsafe.Sizeof((*interface{})(nil)))

    // Float bits
    fmt.Println("\n=== Float Bit Tricks ===")
    fmt.Printf("1.0  bits: %016x\n", f64bits(1.0))
    fmt.Printf("-1.0 bits: %016x\n", f64bits(-1.0))
    fmt.Printf("NaN  bits: %016x\n", f64bits(f64fromBits(0x7FF8000000000001)))

    // Zero-copy string
    fmt.Println("\n=== Zero-Copy String ===")
    b := []byte("hello, unsafe world")
    s := unsafeString(b)
    fmt.Println(s)   // hello, unsafe world
    // ⚠️ s shares memory with b — modifying b changes s!
    b[0] = 'H'
    fmt.Println(s)   // Hello, unsafe world (shares backing array)
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Storing uintptr in a variable — dangling pointer
```go
// ❌ CRITICAL BUG — GC may move p between lines
addr := uintptr(unsafe.Pointer(&p))
// GC may run here and move p
ptr := (*int)(unsafe.Pointer(addr))   // addr is now stale!

// ✅ Single expression — GC can't interrupt an expression
ptr := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + offset))
```

### 2. Writing through a zero-copy string-to-bytes slice
```go
s := "immutable"
b := *(*[]byte)(unsafe.Pointer(&s))
b[0] = 'X'   // ❌ UNDEFINED BEHAVIOR — string data is read-only!
```

### 3. `unsafe.Sizeof` returns the header size, not deep size
```go
s := "hello"
fmt.Println(unsafe.Sizeof(s))   // 16 (ptr+len) — NOT 5!
// For actual content size, use len(s)
```

### 4. Struct layout is not guaranteed across Go versions
```go
// ❌ Hardcoded offset — breaks if struct fields are reordered
ptr := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + 8))

// ✅ Use unsafe.Offsetof — reorder-safe
ptr := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + unsafe.Offsetof(s.field)))
```

### 5. `unsafe` code is not portable across architectures
```go
// Pointer sizes differ: 4 bytes on 32-bit, 8 bytes on 64-bit
fmt.Println(unsafe.Sizeof(uintptr(0)))   // 8 on amd64, 4 on arm32
```

---

## 🎯 Must-Know Things (Quick Reference)

- `unsafe.Sizeof(x)` — size in bytes of x's type (compile-time constant)
- `unsafe.Alignof(x)` — alignment requirement of x's type
- `unsafe.Offsetof(s.f)` — byte offset of field f within struct s
- `unsafe.Pointer` — can convert to/from any pointer type and `uintptr`
- **Never store `uintptr` from `unsafe.Pointer` in a variable** — GC can move objects
- Pointer arithmetic must be a **single expression** — not split across lines
- `unsafe.Slice(ptr, n)` — create `[]T` from pointer + length (Go 1.17+)
- `unsafe.String(ptr, n)` — create `string` from `*byte` + length (Go 1.20+)
- Zero-copy `[]byte`↔`string` — shares memory, never write through the slice
- Order struct fields **largest to smallest** to minimize padding
- `unsafe` code is **not portable** across architectures or guaranteed across Go versions
- Most code never needs `unsafe` — if in doubt, don't use it

---

## ❓ Interview Questions & Answers

---

**Q1: What is the `unsafe` package and when is it appropriate to use it?**

> The `unsafe` package bypasses Go's type system and memory safety guarantees — it allows raw pointer arithmetic, type punning, and direct memory layout access. It's appropriate in a narrow set of cases: systems-level libraries that need zero-copy conversions between types (e.g., `[]byte` to `string` without allocation), interoperating with C via cgo where struct layouts must match exactly, lock-free data structures using atomic pointer operations, and performance-critical hot paths where the overhead of safe conversions is measurable. Most application code should never use it — the risks (silent data corruption, non-portable behavior, GC interference) far outweigh the benefits.

---

**Q2: Why is storing a `uintptr` from `unsafe.Pointer` dangerous?**

> Go's garbage collector can move objects in memory (it's a moving GC in some implementations, and may become one more broadly). A `uintptr` is just an integer — the GC doesn't know it represents a pointer, so it won't update it if the object moves. If you store a `uintptr` in a variable and the GC runs between the conversion from `unsafe.Pointer` and the conversion back, the `uintptr` is now a dangling pointer to freed or reallocated memory. The safe pattern is to perform the pointer arithmetic in a single expression so the compiler ensures the original pointer stays live throughout.

---

**Q3: What does `unsafe.Sizeof` return for a string or slice?**

> `unsafe.Sizeof` returns the size of the **header** — the static data structure — not the dynamic content. For a `string`, it returns 16 bytes on 64-bit systems (a pointer + length). For a `[]int`, it returns 24 bytes (pointer + length + capacity). The actual content — the characters or elements — lives on the heap and is not counted. To get the content size, use `len(s)` for the element count or `len(s) * unsafe.Sizeof(s[0])` for the byte size of slice contents.

---

**Q4: How do you minimize struct padding in Go?**

> Order fields from largest to smallest alignment. The Go compiler inserts padding between fields to satisfy each field's alignment requirement (e.g., `int64` must be 8-byte aligned). If you place a `bool` (1 byte) before an `int64` (8 bytes), 7 bytes of padding are inserted. By ordering `int64` fields first, then `int32`, then smaller types, you minimize wasted space. Use `unsafe.Sizeof` and `unsafe.Offsetof` to verify the layout. This matters for structs with millions of instances in memory or structs mapped to binary protocols.

---

**Q5: What is type punning and when is it used?**

> Type punning is treating the raw bits of one type as if they were another type — bypassing normal conversion rules. In Go, it's done with `*(*TargetType)(unsafe.Pointer(&sourceValue))`. Common uses: interpreting the IEEE 754 bit pattern of a `float64` as `uint64` for mathematical tricks (like fast inverse square root), zero-copy conversion between `[]byte` and `string` (they have compatible memory layouts), and reading C binary protocol data directly into Go structs. It's inherently non-portable — it assumes specific type sizes and memory layouts that vary by platform.

---

*Next up: **Topic 3.6 — Testing in Go** 🚀*
