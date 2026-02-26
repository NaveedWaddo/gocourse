# 1.10 — Maps

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.9 — Arrays & Slices

---

## 📖 What is a Map?

A **map** is Go's built-in hash table — an unordered collection of **key-value pairs** where every key is unique. Maps are the go-to data structure for lookups, grouping, counting, and caching.

```
map[string]int
     │        └── value type
     └── key type
```

---

## 🏗️ Creating Maps

### Using `make` (idiomatic for empty maps)

```go
// make(map[KeyType]ValueType)
ages := make(map[string]int)
ages["Alice"] = 30
ages["Bob"] = 25
```

### Map literal (idiomatic when you have initial data)

```go
ages := map[string]int{
    "Alice": 30,
    "Bob":   25,
    "Carol": 35,
}
```

### Nil map — zero value

```go
var m map[string]int   // nil map
fmt.Println(m == nil)  // true
fmt.Println(len(m))    // 0 — safe to read

m["key"] = 1           // ❌ PANIC — cannot write to a nil map!
```

> Always initialize with `make` or a literal before writing. Reading from a nil map is safe (returns zero value), but writing panics.

---

## 🔧 CRUD Operations

### Create / Update

```go
m := make(map[string]int)

m["apple"] = 5    // create
m["apple"] = 10   // update — same syntax, just overwrites
```

### Read

```go
count := m["apple"]   // 10
missing := m["grape"] // 0 — zero value, no error!
```

### Check existence — the comma-ok idiom

```go
value, ok := m["apple"]
if ok {
    fmt.Println("Found:", value)
} else {
    fmt.Println("Not found")
}

// One-liner
if count, ok := m["apple"]; ok {
    fmt.Println("apple count:", count)
}
```

> This is critical — reading a missing key returns the zero value silently. Always use comma-ok when you need to distinguish "key exists with zero value" from "key doesn't exist".

### Delete

```go
delete(m, "apple")   // removes the key — no error if key doesn't exist
```

### Length

```go
fmt.Println(len(m))   // number of key-value pairs
```

---

## 🔁 Iterating Over Maps

```go
ages := map[string]int{"Alice": 30, "Bob": 25, "Carol": 35}

for name, age := range ages {
    fmt.Printf("%s is %d years old\n", name, age)
}
```

> ⚠️ **Map iteration order is NOT guaranteed** — it's randomized on purpose. Never rely on a specific order.

### Iterate keys only

```go
for name := range ages {
    fmt.Println(name)
}
```

### Iterate in sorted order

```go
import "sort"

keys := make([]string, 0, len(ages))
for k := range ages {
    keys = append(keys, k)
}
sort.Strings(keys)

for _, k := range keys {
    fmt.Printf("%s: %d\n", k, ages[k])
}
```

---

## 🔑 Valid Key Types

Map keys must be **comparable** — they must support `==`. This means:

✅ Valid key types: `string`, `int`, `float64`, `bool`, `pointer`, `array`, `struct` (if all fields are comparable)

❌ Invalid key types: `slice`, `map`, `function` — these are not comparable

```go
// ✅ All valid
map[string]int{}
map[int]string{}
map[[3]int]string{}   // array as key
map[bool]int{}

// ❌ Compile errors
map[[]int]string{}    // slice is not comparable
map[map[string]int]string{}  // map is not comparable
```

---

## 🧩 Common Patterns

### Pattern 1: Counting occurrences

```go
words := []string{"go", "is", "great", "go", "is", "go"}

counts := make(map[string]int)
for _, w := range words {
    counts[w]++   // zero value of int is 0 — this works even on first access!
}

fmt.Println(counts)   // map[go:3 great:1 is:2]
```

### Pattern 2: Grouping (map of slices)

```go
people := []struct{ Name, City string }{
    {"Alice", "NYC"}, {"Bob", "LA"}, {"Carol", "NYC"}, {"Dave", "LA"},
}

byCity := make(map[string][]string)
for _, p := range people {
    byCity[p.City] = append(byCity[p.City], p.Name)
}

fmt.Println(byCity["NYC"])   // [Alice Carol]
fmt.Println(byCity["LA"])    // [Bob Dave]
```

### Pattern 3: Set (map with empty struct values)

```go
// map[T]struct{} — the idiomatic Go set
// struct{} takes 0 bytes — most memory-efficient value

seen := make(map[string]struct{})

words := []string{"go", "is", "go", "great"}
for _, w := range words {
    seen[w] = struct{}{}
}

// Check membership
if _, ok := seen["go"]; ok {
    fmt.Println("go is in the set")
}

fmt.Println(len(seen))   // 3 — duplicates removed
```

### Pattern 4: Cache / memoization

```go
cache := make(map[int]int)

func fibMemo(n int) int {
    if n <= 1 {
        return n
    }
    if v, ok := cache[n]; ok {
        return v   // return cached result
    }
    result := fibMemo(n-1) + fibMemo(n-2)
    cache[n] = result
    return result
}
```

---

## 🔄 Maps are Reference Types

Like slices, maps are reference types — assigning a map to another variable gives you another reference to the **same underlying hash table**:

```go
a := map[string]int{"x": 1}
b := a          // b points to the same map as a

b["x"] = 99
fmt.Println(a["x"])   // 99 — a was modified through b!
```

To copy a map, you must manually copy it:

```go
func copyMap(m map[string]int) map[string]int {
    result := make(map[string]int, len(m))
    for k, v := range m {
        result[k] = v
    }
    return result
}
```

---

## 🔒 Maps are NOT Safe for Concurrent Use

If multiple goroutines read and write the same map concurrently, it's a **data race** and will crash:

```go
// ❌ Dangerous — concurrent map writes cause panic
go func() { m["key"] = 1 }()
go func() { m["key"] = 2 }()
```

Solutions:
- Use `sync.Mutex` to protect the map (Topic 2.7)
- Use `sync.Map` from the `sync` package (designed for concurrent use)

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    // --- Basic CRUD ---
    inventory := map[string]int{
        "apples":  10,
        "bananas": 5,
        "oranges": 8,
    }

    // Read
    fmt.Println("Apples:", inventory["apples"])   // 10

    // Comma-ok: check existence
    if count, ok := inventory["grapes"]; ok {
        fmt.Println("Grapes:", count)
    } else {
        fmt.Println("Grapes not found")   // this runs
    }

    // Update
    inventory["apples"] += 5
    fmt.Println("Apples after restock:", inventory["apples"])   // 15

    // Delete
    delete(inventory, "bananas")
    fmt.Println("After delete, len:", len(inventory))   // 2

    // --- Sorted iteration ---
    fmt.Println("\nInventory (sorted):")
    keys := make([]string, 0, len(inventory))
    for k := range inventory {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, k := range keys {
        fmt.Printf("  %s: %d\n", k, inventory[k])
    }

    // --- Count words ---
    text := []string{"the", "quick", "brown", "fox", "the", "fox"}
    wordCount := make(map[string]int)
    for _, w := range text {
        wordCount[w]++
    }
    fmt.Println("\nWord counts:", wordCount)

    // --- Set pattern ---
    unique := make(map[string]struct{})
    for _, w := range text {
        unique[w] = struct{}{}
    }
    fmt.Println("Unique words:", len(unique))   // 4

    // --- Map of slices ---
    scores := map[string][]int{
        "Alice": {95, 87, 92},
        "Bob":   {78, 85, 90},
    }
    scores["Alice"] = append(scores["Alice"], 99)
    fmt.Println("Alice's scores:", scores["Alice"])   // [95 87 92 99]
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Writing to a nil map panics
```go
var m map[string]int
m["key"] = 1   // ❌ PANIC: assignment to entry in nil map
m = make(map[string]int)
m["key"] = 1   // ✅
```

### 2. Missing key returns zero value — not an error
```go
m := map[string]int{"a": 1}
v := m["missing"]   // v = 0 — no error, no panic!
// Use comma-ok to distinguish:
v, ok := m["missing"]   // v=0, ok=false
```

### 3. Map iteration order is random
```go
// ❌ Don't rely on order
for k, v := range m { ... }   // different order each run

// ✅ Sort keys first if order matters
```

### 4. Maps are reference types — assignment doesn't copy
```go
a := map[string]int{"x": 1}
b := a        // b is NOT a copy — same map!
b["x"] = 99   // modifies a too!
```

### 5. Concurrent map access causes panic
```go
// ❌ Race condition — use sync.Mutex or sync.Map
go func() { m["a"] = 1 }()
go func() { _ = m["a"] }()
```

---

## 🎯 Must-Know Things (Quick Reference)

- Maps are **unordered** key-value stores — iteration order is random
- Zero value is `nil` — reading is safe, **writing to nil map panics**
- Always use **comma-ok** to check if a key exists: `v, ok := m[key]`
- `delete(m, key)` — no error if key doesn't exist
- Maps are **reference types** — assignment shares the same underlying data
- Key types must be **comparable** — no slices, maps, or functions as keys
- `map[T]struct{}` is the idiomatic **set** pattern — zero memory overhead
- Maps are **NOT safe for concurrent use** — use `sync.Mutex` or `sync.Map`
- Pre-size with `make(map[K]V, hint)` when you know approximate size

---

## ❓ Interview Questions & Answers

---

**Q1: What is the zero value of a map in Go and what are the implications?**

> The zero value of a map is `nil`. A nil map behaves like an empty map for reads — accessing a missing key returns the zero value silently. However, writing to a nil map causes a runtime panic. Always initialize with `make(map[K]V)` or a map literal before writing. This is a common bug for beginners who declare `var m map[string]int` and then try to assign to it.

---

**Q2: How do you check if a key exists in a map?**

> Use the comma-ok idiom: `value, ok := m[key]`. If the key exists, `ok` is `true` and `value` holds the associated value. If the key doesn't exist, `ok` is `false` and `value` is the zero value for the map's value type. This distinction matters because a key can legitimately hold the zero value — you can't tell from the value alone whether it was set or not.

---

**Q3: Is map iteration order guaranteed in Go?**

> No. Go deliberately randomizes map iteration order to prevent programs from accidentally depending on it (early versions had consistent order which caused bugs). If you need to iterate in a specific order, collect the keys into a slice, sort it with `sort.Strings` or `sort.Ints`, and then iterate the sorted keys.

---

**Q4: How do you implement a set in Go?**

> Go has no built-in set type. The idiomatic pattern is `map[T]struct{}` — using an empty struct as the value type. `struct{}` takes zero bytes of memory, so it's the most efficient option. To add: `s[key] = struct{}{}`. To check membership: `_, ok := s[key]`. To remove: `delete(s, key)`. Alternatively, `map[T]bool` works but wastes a byte per entry.

---

**Q5: Are maps reference types in Go?**

> Yes. When you assign a map to a new variable or pass it to a function, both variables point to the same underlying hash table. Modifications through either variable affect the same data. To get an independent copy, you must manually iterate and copy every key-value pair into a new map. This behavior is similar to slices but different from structs, which are value types.

---

**Q6: Why are maps not safe for concurrent use in Go?**

> Go's built-in map is not designed for concurrent access — simultaneous reads and writes from multiple goroutines cause a data race, which will crash the program with a "concurrent map read and map write" error (detected by Go's race detector). The solutions are: (1) protect with `sync.Mutex` — lock before any read or write; (2) use `sync.Map` from the sync package, which is optimized for concurrent read-heavy workloads; (3) use channels to serialize access.

---

**Q7: What types can be used as map keys?**

> Map keys must be comparable — they must support the `==` operator. Valid key types include all basic types (`string`, `int`, `bool`, `float64`), pointers, arrays (if element type is comparable), and structs (if all fields are comparable). Invalid key types are slices, maps, and functions — these are not comparable. Attempting to use them as keys is a compile error.

---

*Next up: **Topic 1.11 — Structs** 🚀*
