# 1.6 — Control Flow

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.5 — Operators

---

## 📖 Overview

Control flow determines the order in which code executes. Go keeps this minimal and clean — there's no `while`, no `do-while`, no `foreach`. Just `if/else`, `switch`, and `for` — and `for` handles everything loop-related.

---

## 🔀 `if` / `else`

### Basic syntax

```go
x := 10

if x > 5 {
    fmt.Println("x is greater than 5")
}
```

> ⚠️ Unlike many languages, Go does **not** require parentheses around the condition. But the curly braces `{}` are **mandatory** — even for single-line bodies.

```go
// ❌ No curly braces — compile error
if x > 5
    fmt.Println("hi")

// ✅ Always use curly braces
if x > 5 {
    fmt.Println("hi")
}
```

### `if` / `else if` / `else`

```go
score := 75

if score >= 90 {
    fmt.Println("A")
} else if score >= 80 {
    fmt.Println("B")
} else if score >= 70 {
    fmt.Println("C")
} else {
    fmt.Println("F")
}
```

---

### The `if` initializer statement — very Go-idiomatic

Go lets you run a short statement **before** the condition, scoped to the `if` block:

```go
if x := computeValue(); x > 10 {
    fmt.Println("big:", x)
} else {
    fmt.Println("small:", x)
}
// x is NOT accessible here — scoped to the if/else block
```

This is used constantly with error handling:

```go
if err := doSomething(); err != nil {
    fmt.Println("error:", err)
    return
}
```

It keeps the variable scoped tightly and avoids polluting the outer scope.

---

## 🔃 `switch`

Go's `switch` is much more powerful and flexible than in C or Java.

### Basic `switch`

```go
day := "Monday"

switch day {
case "Monday":
    fmt.Println("Start of the work week")
case "Friday":
    fmt.Println("End of the work week")
case "Saturday", "Sunday":   // multiple values in one case
    fmt.Println("Weekend!")
default:
    fmt.Println("Midweek")
}
```

> **Key difference from C/Java:** Go does **NOT** fall through by default. Each case breaks automatically. No need to write `break` everywhere.

---

### Switch with no condition — acts like `if/else if`

```go
x := 42

switch {
case x < 0:
    fmt.Println("negative")
case x == 0:
    fmt.Println("zero")
case x < 100:
    fmt.Println("small positive")   // this runs
default:
    fmt.Println("large positive")
}
```

This is clean and readable — preferred over long `if/else if` chains.

---

### Switch with an initializer

Just like `if`, `switch` can have an initializer statement:

```go
switch os := runtime.GOOS; os {
case "darwin":
    fmt.Println("macOS")
case "linux":
    fmt.Println("Linux")
default:
    fmt.Println("Other:", os)
}
```

---

### `fallthrough` — explicit opt-in

If you DO want C-style fall-through, use the `fallthrough` keyword explicitly:

```go
x := 1

switch x {
case 1:
    fmt.Println("one")
    fallthrough   // explicitly fall to next case
case 2:
    fmt.Println("two")   // this also runs when x == 1
case 3:
    fmt.Println("three")
}
// Output:
// one
// two
```

> `fallthrough` is rare and generally discouraged — it makes code harder to reason about. Use it only when you have a clear reason.

---

### Type switch — powerful pattern

A type switch checks the **dynamic type** of an interface value. You'll use this a lot with `interface{}` and `any`:

```go
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("int: %d\n", v)
    case string:
        fmt.Printf("string: %s\n", v)
    case bool:
        fmt.Printf("bool: %t\n", v)
    default:
        fmt.Printf("unknown type: %T\n", v)
    }
}

describe(42)      // int: 42
describe("hi")    // string: hi
describe(true)    // bool: true
describe(3.14)    // unknown type: float64
```

---

## 🔁 `for` — Go's Only Loop

Go has **one loop keyword: `for`**. It covers all looping patterns:

---

### Classic C-style for loop

```go
for i := 0; i < 5; i++ {
    fmt.Println(i)   // 0 1 2 3 4
}
```

Three parts: `init; condition; post` — all optional.

---

### `for` as a `while` loop

Omit init and post — just keep the condition:

```go
n := 1
for n < 100 {
    n *= 2
}
fmt.Println(n)   // 128
```

---

### Infinite loop

Omit everything:

```go
for {
    // runs forever
    // use break to exit
}
```

This is the idiomatic Go way to write an infinite loop (like a server loop):

```go
for {
    conn, err := listener.Accept()
    if err != nil {
        break
    }
    go handleConn(conn)
}
```

---

### `for range` — iterate over collections

`range` is Go's powerful iteration construct. It works on strings, arrays, slices, maps, and channels.

**Over a slice:**
```go
nums := []int{10, 20, 30, 40}

for i, v := range nums {
    fmt.Printf("index %d: value %d\n", i, v)
}
// index 0: value 10
// index 1: value 20
// ...
```

**Ignore the index with `_`:**
```go
for _, v := range nums {
    fmt.Println(v)
}
```

**Ignore the value (index only):**
```go
for i := range nums {
    fmt.Println(i)
}
```

**Over a string (iterates runes, not bytes):**
```go
for i, ch := range "Go🐹" {
    fmt.Printf("index %d: %c\n", i, ch)
}
// index 0: G
// index 1: o
// index 2: 🐹  (index is byte position — 🐹 starts at byte 2)
```

**Over a map:**
```go
ages := map[string]int{"Alice": 30, "Bob": 25}

for name, age := range ages {
    fmt.Printf("%s is %d years old\n", name, age)
}
// Order is NOT guaranteed with maps
```

**Over a channel:**
```go
ch := make(chan int)
// ... (used in goroutines, covered in Topic 2.5)
for v := range ch {
    fmt.Println(v)
}
```

---

### `break` and `continue`

```go
// break — exits the loop entirely
for i := 0; i < 10; i++ {
    if i == 5 {
        break
    }
    fmt.Println(i)   // 0 1 2 3 4
}

// continue — skips the rest of the current iteration
for i := 0; i < 10; i++ {
    if i%2 == 0 {
        continue   // skip even numbers
    }
    fmt.Println(i)   // 1 3 5 7 9
}
```

---

### Labeled `break` and `continue` — for nested loops

```go
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if j == 1 {
            break outer   // breaks the OUTER loop, not just inner
        }
        fmt.Printf("i=%d j=%d\n", i, j)
    }
}
// Output:
// i=0 j=0
```

Without the label, `break` would only exit the inner loop. Labels let you target a specific loop.

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
)

func classify(n int) string {
    switch {
    case n < 0:
        return "negative"
    case n == 0:
        return "zero"
    case n < 10:
        return "small"
    case n < 100:
        return "medium"
    default:
        return "large"
    }
}

func main() {
    // if with initializer
    if score := 85; score >= 80 {
        fmt.Println("Pass with score:", score)
    }

    // switch — multiple values per case
    for _, day := range []string{"Mon", "Sat", "Sun", "Fri"} {
        switch day {
        case "Sat", "Sun":
            fmt.Println(day, "→ Weekend")
        case "Mon", "Fri":
            fmt.Println(day, "→ Bookend of work week")
        default:
            fmt.Println(day, "→ Midweek")
        }
    }

    // classify with conditionless switch
    for _, n := range []int{-5, 0, 7, 42, 999} {
        fmt.Printf("%d is %s\n", n, classify(n))
    }

    // for range with index and value
    fruits := []string{"apple", "banana", "cherry"}
    for i, fruit := range fruits {
        fmt.Printf("%d: %s\n", i, fruit)
    }

    // while-style loop
    power := 1
    for power < 1000 {
        power *= 2
    }
    fmt.Println("First power of 2 >= 1000:", power)   // 1024

    // loop with continue
    fmt.Print("Odd numbers: ")
    for i := 1; i <= 10; i++ {
        if i%2 == 0 {
            continue
        }
        fmt.Print(i, " ")
    }
    fmt.Println()
}
```

Output:
```
Pass with score: 85
Mon → Bookend of work week
Sat → Weekend
Sun → Weekend
Fri → Bookend of work week
-5 is negative
0 is zero
7 is small
42 is medium
999 is large
0: apple
1: banana
2: cherry
First power of 2 >= 1000: 1024
Odd numbers: 1 3 5 7 9
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Curly braces are mandatory
```go
// ❌ compile error
if x > 5
    fmt.Println("hi")

// ✅
if x > 5 {
    fmt.Println("hi")
}
```

### 2. No parentheses around condition
```go
if (x > 5) {   // ⚠️ compiles, but not idiomatic — gofmt won't remove them though
}
if x > 5 {     // ✅ idiomatic Go
}
```

### 3. Switch falls through by default? NO
```go
// In Go, each case auto-breaks — no fallthrough by default
switch x {
case 1:
    fmt.Println("one")
    // implicit break here — "two" will NOT print
case 2:
    fmt.Println("two")
}
```

### 4. Map iteration order is not guaranteed
```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v)   // order changes every run
}
```

### 5. Modifying a slice while ranging over it
```go
nums := []int{1, 2, 3, 4, 5}
for i, v := range nums {
    nums[i] = v * 2   // ✅ modifying via index is fine
}
// But appending during range can cause issues — range captures original length
```

### 6. Variable shadowing in `if` initializer
```go
err := doFirst()
if err := doSecond(); err != nil {   // this err is a NEW variable, shadows outer err
    fmt.Println(err)
}
// outer err is unchanged here — be careful!
```

---

## 🎯 Must-Know Things (Quick Reference)

- Go has **no `while`**, **no `do-while`**, **no `foreach`** — just `for` for everything
- Curly braces `{}` are **always required** — even for single-line bodies
- No parentheses around `if`/`switch`/`for` conditions
- `switch` cases **do NOT fall through** by default — `break` is implicit
- Use `fallthrough` explicitly when you want C-style fall-through
- `if` and `switch` support an **initializer statement** before the condition
- `for range` iterates strings by **rune** (Unicode), not bytes
- Map iteration order is **random** — never rely on it
- Use **labeled `break`/`continue`** to control nested loops
- `switch` with no expression = cleaner `if/else if` chain

---

## ❓ Interview Questions & Answers

---

**Q1: How do you write a `while` loop in Go?**

> Go has no `while` keyword. You use `for` with only a condition: `for condition { }`. For example: `for n < 100 { n *= 2 }`. An infinite loop is `for { }`, and you exit with `break`. This design keeps the language simple — one loop construct that covers all cases.

---

**Q2: Does Go's `switch` fall through by default?**

> No. Unlike C and Java, Go's `switch` cases automatically break at the end of each case. There is no need to write `break` explicitly. If you want fall-through behavior, you must use the `fallthrough` keyword explicitly. This default is safer and eliminates an entire class of bugs where developers forget to write `break` in C-style languages.

---

**Q3: What is the `if` initializer statement in Go?**

> Go allows a short statement before the condition in an `if`, separated by a semicolon: `if x := compute(); x > 10 { }`. The variable `x` is scoped to the `if/else` block and not accessible outside. This is the standard pattern for error handling: `if err := doSomething(); err != nil { }`. It keeps the code concise and avoids polluting the outer scope with temporary variables.

---

**Q4: What does `for range` return for different types?**

> For slices and arrays: index (int) and value. For maps: key and value. For strings: byte index (int) and rune (Unicode character). For channels: just the value. You can discard either return value with `_`. For strings specifically, `range` iterates over Unicode runes, not bytes — so a multi-byte character like an emoji appears as a single iteration with its code point.

---

**Q5: What is a type switch in Go?**

> A type switch checks the dynamic type of an interface value using the syntax `switch v := i.(type) { case int: ... case string: ... }`. It's used when you have a value of type `interface{}` (or `any`) and need to handle different concrete types differently. It's a clean alternative to multiple type assertions with `if` statements.

---

**Q6: How do labeled `break` and `continue` work?**

> You can attach a label to a `for` loop and use it with `break` or `continue` to target that specific loop instead of the innermost one. For example, `break outer` will break out of the loop labeled `outer:`, even from inside a nested inner loop. This is useful in cases like searching a 2D grid where you want to exit all loops when a value is found.

---

**Q7: Is map iteration order guaranteed in Go?**

> No. Iterating over a map with `for range` produces keys in a randomized order — Go deliberately randomizes this to prevent programs from accidentally depending on insertion order (which was consistent in early Go versions and caused bugs). If you need ordered output, collect the keys into a slice, sort the slice, then iterate.

---

*Next up: **Topic 1.7 — Functions** 🚀*
