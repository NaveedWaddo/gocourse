# 1.9 — Arrays & Slices

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.8 — Pointers

---

## 📖 Overview

Go has two sequence types: **arrays** (fixed size, rarely used directly) and **slices** (dynamic, used everywhere). Understanding how slices work internally — the backing array, length, and capacity — is essential for writing correct and efficient Go.

---

## 📦 Arrays

### Declaration and initialization

```go
// Declare an array of 5 ints — all zero-initialized
var a [5]int
fmt.Println(a)   // [0 0 0 0 0]

// Array literal
b := [5]int{10, 20, 30, 40, 50}
fmt.Println(b)   // [10 20 30 40 50]

// Let compiler count the length
c := [...]int{10, 20, 30}
fmt.Println(len(c))   // 3

// Partial initialization — rest are zero
d := [5]int{1, 2}
fmt.Println(d)   // [1 2 0 0 0]
```

### Accessing elements

```go
a := [3]string{"Go", "is", "great"}
fmt.Println(a[0])   // Go
fmt.Println(a[2])   // great

a[1] = "really"
fmt.Println(a)   // [Go really great]
```

### Key properties of arrays

- The **size is part of the type** — `[3]int` and `[5]int` are different types, not interchangeable
- Arrays are **value types** — assigning or passing copies the entire array
- Mostly used as the foundation for slices — you'll rarely use arrays directly in Go

```go
a := [3]int{1, 2, 3}
b := a           // b is a full copy
b[0] = 99
fmt.Println(a)   // [1 2 3] — unchanged
fmt.Println(b)   // [99 2 3]
```

---

## 🍕 Slices — The Real Workhorse

A slice is a **dynamically-sized, flexible view into an array**. It's the go-to sequence type in Go.

### Slice internals — critical to understand

Under the hood, a slice is a small struct with three fields:

```
Slice header:
┌──────────────┐
│   ptr        │ → points to the first element in the backing array
│   len = 3    │ → number of elements currently in the slice
│   cap = 5    │ → total capacity of the backing array from ptr onward
└──────────────┘
```

This is why slices are cheap to pass — only the 24-byte header is copied, not the data.

---

### Creating slices

```go
// Slice literal — creates a backing array automatically
s := []int{1, 2, 3, 4, 5}

// make([]T, len, cap) — creates a slice with given len and cap
s2 := make([]int, 3)        // len=3, cap=3, all zeros
s3 := make([]int, 3, 10)    // len=3, cap=10, all zeros

// From an array
a := [5]int{1, 2, 3, 4, 5}
s4 := a[1:4]   // slice of a — elements at index 1,2,3

// nil slice — zero value
var s5 []int
fmt.Println(s5 == nil)   // true
fmt.Println(len(s5))     // 0 — safe to use
```

---

### Slicing syntax — `s[low:high]`

```go
s := []int{0, 1, 2, 3, 4, 5}

fmt.Println(s[1:4])   // [1 2 3]   — indices 1, 2, 3 (high is exclusive)
fmt.Println(s[:3])    // [0 1 2]   — from start to index 2
fmt.Println(s[3:])    // [3 4 5]   — from index 3 to end
fmt.Println(s[:])     // [0 1 2 3 4 5] — entire slice
```

> `s[low:high]` gives elements from index `low` up to but **not including** `high`. Length = `high - low`.

### Slices share the backing array — important!

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:4]   // b shares a's backing array

b[0] = 99
fmt.Println(a)   // [1 99 3 4 5] — a was modified through b!
fmt.Println(b)   // [99 3 4]
```

This is a common source of bugs. Use `copy` to get an independent slice.

---

### `len` and `cap`

```go
s := make([]int, 3, 10)
fmt.Println(len(s))   // 3 — current number of elements
fmt.Println(cap(s))   // 10 — total capacity

s2 := []int{1, 2, 3, 4, 5}[1:3]
fmt.Println(len(s2))  // 2
fmt.Println(cap(s2))  // 4 — capacity from index 1 to end of backing array
```

---

### `append` — adding elements

```go
s := []int{1, 2, 3}
s = append(s, 4)           // append one element
s = append(s, 5, 6, 7)    // append multiple
s = append(s, []int{8, 9}...)  // append another slice (spread with ...)

fmt.Println(s)   // [1 2 3 4 5 6 7 8 9]
```

> **Always assign the result back**: `s = append(s, x)`. If `append` needs more capacity it creates a new backing array — if you don't capture the new slice header you'll lose it.

### How `append` grows capacity

When `append` exceeds the current capacity, Go allocates a new, larger backing array and copies data over. The growth strategy roughly doubles capacity for small slices:

```go
s := make([]int, 0)
for i := 0; i < 8; i++ {
    s = append(s, i)
    fmt.Printf("len=%d cap=%d\n", len(s), cap(s))
}
// len=1 cap=1
// len=2 cap=2
// len=3 cap=4
// len=4 cap=4
// len=5 cap=8
// len=6 cap=8
// len=7 cap=8
// len=8 cap=8
```

If you know the final size ahead of time, pre-allocate with `make` to avoid repeated reallocations:

```go
// ✅ Efficient — pre-allocate
s := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    s = append(s, i)
}
```

---

### `copy` — independent copy of a slice

```go
src := []int{1, 2, 3, 4, 5}
dst := make([]int, len(src))
n := copy(dst, src)

dst[0] = 99
fmt.Println(src)   // [1 2 3 4 5] — unchanged
fmt.Println(dst)   // [99 2 3 4 5]
fmt.Println(n)     // 5 — number of elements copied
```

`copy(dst, src)` copies `min(len(dst), len(src))` elements.

---

### Deleting an element from a slice

Go has no built-in delete for slices. Common patterns:

```go
s := []int{1, 2, 3, 4, 5}
i := 2   // delete index 2 (value 3)

// Method 1: preserve order
s = append(s[:i], s[i+1:]...)
fmt.Println(s)   // [1 2 4 5]

// Method 2: fast (doesn't preserve order — swap with last)
s[i] = s[len(s)-1]
s = s[:len(s)-1]
fmt.Println(s)   // [1 2 5 4]
```

---

### 2D Slices

```go
// Create a 3x3 grid
grid := make([][]int, 3)
for i := range grid {
    grid[i] = make([]int, 3)
}

grid[0][0] = 1
grid[1][1] = 5
grid[2][2] = 9

for _, row := range grid {
    fmt.Println(row)
}
// [1 0 0]
// [0 5 0]
// [0 0 9]
```

---

## 💻 Complete Example

```go
package main

import "fmt"

func main() {
    // --- Arrays ---
    arr := [3]int{10, 20, 30}
    fmt.Println("Array:", arr)

    // Arrays are values — copy on assign
    arr2 := arr
    arr2[0] = 99
    fmt.Println("Original array:", arr)    // [10 20 30] — unchanged
    fmt.Println("Copy array:", arr2)       // [99 20 30]

    // --- Slice basics ---
    s := []string{"Go", "Python", "Rust"}
    fmt.Println("Slice:", s)
    fmt.Printf("len=%d cap=%d\n", len(s), cap(s))

    // --- Slicing ---
    fmt.Println("s[1:]:", s[1:])    // [Python Rust]
    fmt.Println("s[:2]:", s[:2])    // [Go Python]

    // --- append ---
    s = append(s, "Java", "C++")
    fmt.Println("After append:", s)

    // --- Shared backing array ---
    a := []int{1, 2, 3, 4, 5}
    b := a[1:3]
    b[0] = 99
    fmt.Println("a after modifying b:", a)   // [1 99 3 4 5]

    // --- Independent copy ---
    c := make([]int, len(a))
    copy(c, a)
    c[0] = 777
    fmt.Println("a after modifying c:", a)   // [1 99 3 4 5] — unchanged
    fmt.Println("c:", c)                     // [777 99 3 4 5]

    // --- Delete element ---
    nums := []int{10, 20, 30, 40, 50}
    nums = append(nums[:2], nums[3:]...)  // delete index 2 (value 30)
    fmt.Println("After delete:", nums)    // [10 20 40 50]

    // --- Pre-allocated slice ---
    squares := make([]int, 0, 10)
    for i := 1; i <= 10; i++ {
        squares = append(squares, i*i)
    }
    fmt.Println("Squares:", squares)
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Forgetting to reassign after append
```go
s := []int{1, 2, 3}
append(s, 4)        // ❌ result is discarded — compile warning
s = append(s, 4)    // ✅
```

### 2. Slice shares backing array — unexpected mutations
```go
a := []int{1, 2, 3, 4, 5}
b := a[:3]
b[0] = 99       // ❌ also modifies a!
// Use copy if you need independence
```

### 3. Appending to a sub-slice can overwrite the original
```go
a := []int{1, 2, 3, 4, 5}
b := a[:3]                // b shares a's backing array, cap=5
b = append(b, 99)         // fits in cap — overwrites a[3]!
fmt.Println(a)            // [1 2 3 99 5] — a[3] is now 99!
```

### 4. nil slice vs empty slice
```go
var s []int          // nil slice — s == nil is true
s2 := []int{}        // empty slice — s2 == nil is false
s3 := make([]int, 0) // empty slice — s3 == nil is false

// All have len=0 and are safe to append to
// But only nil slice == nil
```

### 5. Array size is part of the type
```go
var a [3]int
var b [5]int
a = b   // ❌ compile error — [3]int and [5]int are different types
```

---

## 🎯 Must-Know Things (Quick Reference)

- Arrays are **fixed size**, value types — size is part of the type
- Slices are **dynamic views** over a backing array — described by pointer, len, cap
- `s[low:high]` — high is **exclusive**
- Slices share the backing array — modifying one affects the other
- **Always assign `append` result back**: `s = append(s, x)`
- `copy(dst, src)` makes an **independent** copy — copies `min(len(dst), len(src))` elements
- Pre-allocate with `make([]T, 0, cap)` when size is known — avoids reallocation
- `nil` slice and empty slice both have `len=0` and are safe to use — only `nil` slice `== nil`
- Use `append(s[:i], s[i+1:]...)` to delete an element (order-preserving)

---

## ❓ Interview Questions & Answers

---

**Q1: What is the internal structure of a slice in Go?**

> A slice is a three-field struct: a pointer to the first element of a backing array, a length (current number of elements), and a capacity (total elements available from the pointer to the end of the backing array). This is why slices are cheap to pass — only the 24-byte header is copied, not the data. Multiple slices can share the same backing array, which is efficient but requires care to avoid unintended mutations.

---

**Q2: What is the difference between `len` and `cap` on a slice?**

> `len` is the number of elements currently accessible in the slice. `cap` is the total number of elements in the backing array from the slice's start pointer to the end. You can grow a slice up to its capacity without allocation using `s = s[:newLen]` (as long as `newLen <= cap`). Exceeding the capacity requires `append` to allocate a new backing array.

---

**Q3: What happens when `append` exceeds capacity?**

> Go allocates a new, larger backing array (roughly doubling for small slices), copies all existing elements to it, appends the new element, and returns a new slice header pointing to the new array. The original backing array is unaffected. This is why you must always assign the result: `s = append(s, x)` — if a reallocation happened, the old slice header is now stale.

---

**Q4: What is the difference between a nil slice and an empty slice?**

> A nil slice (`var s []int`) has a nil pointer, len=0, cap=0, and `s == nil` is true. An empty slice (`s := []int{}` or `make([]int, 0)`) has a non-nil pointer, len=0, cap=0, and `s == nil` is false. Both are safe to use — you can `append` to either, `len` returns 0 for both, and `range` iterates zero times. The distinction matters when you need to signal "no value" vs "empty collection" — e.g., JSON marshaling: nil slice becomes `null`, empty slice becomes `[]`.

---

**Q5: How do you safely copy a slice in Go?**

> Use the built-in `copy(dst, src)`. First create a destination slice with `make([]T, len(src))`, then call `copy`. It copies `min(len(dst), len(src))` elements and returns the count. The result is an independent slice — modifying it won't affect the source. Alternatively, `append([]T{}, src...)` creates a copy in one line but is slightly less efficient.

---

**Q6: What is the append-to-subslice gotcha?**

> When you create a sub-slice `b := a[:3]`, `b` shares `a`'s backing array and has the remaining capacity. If you `append` to `b` and it fits within the capacity, it overwrites elements of `a` beyond the sub-slice — `a[3]` gets silently overwritten. The fix is to use a three-index slice `b := a[:3:3]` to set the capacity equal to the length, forcing any append to allocate a new backing array.

---

**Q7: How do you delete an element from a slice in Go?**

> Go has no built-in delete for slices. The order-preserving way is `s = append(s[:i], s[i+1:]...)` which re-slices around the element and appends the tail. The faster but order-changing way is to swap the element with the last: `s[i] = s[len(s)-1]; s = s[:len(s)-1]`. Both modify the underlying array in-place, so make a copy first if you need to preserve the original.

---

*Next up: **Topic 1.10 — Maps** 🚀*
