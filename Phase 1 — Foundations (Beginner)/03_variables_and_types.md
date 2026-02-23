# 1.3 — Variables & Data Types

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.2 — Your First Go Program

---

## 📖 Overview

Variables are named storage for data. Go is **statically typed** — every variable has a fixed type that's known at compile time. But Go also has **type inference** — you don't always have to write the type explicitly, the compiler can figure it out.

In this topic we'll cover:
- How to declare variables (`var` and `:=`)
- All of Go's basic built-in types
- Zero values
- Type conversion
- Constants (brief intro — full coverage in Topic 1.4)

---

## 🧩 Declaring Variables

Go has **two ways** to declare variables. Both are valid — knowing when to use each is important.

---

### Method 1: `var` keyword (explicit)

```go
var name string = "Alice"
var age  int    = 30
var pi   float64 = 3.14
```

You can also omit the type if the value makes it obvious — Go infers it:

```go
var name = "Alice"    // Go infers string
var age  = 30         // Go infers int
var pi   = 3.14       // Go infers float64
```

Or declare the variable with just a type (no value) — it gets the **zero value**:

```go
var name string   // ""    (empty string)
var age  int      // 0
var active bool   // false
```

You can also declare multiple variables at once:

```go
var (
    name   string = "Alice"
    age    int    = 30
    active bool   = true
)
```

---

### Method 2: `:=` short variable declaration (idiomatic)

```go
name := "Alice"   // string
age  := 30        // int
pi   := 3.14      // float64
```

This is the **most common way** to declare variables in Go. It's shorter and Go infers the type automatically.

> ⚠️ **Critical rule:** `:=` can **only be used inside functions**. At the package level (outside any function), you must use `var`.

```go
package main

var globalName = "Bob"  // ✅ package-level — must use var

func main() {
    localAge := 25      // ✅ inside function — can use :=
}
```

---

### Side-by-side comparison

```go
// These are all equivalent:

var x int = 10      // full explicit form
var x = 10          // type inferred
x := 10             // short form (only inside functions)
```

---

### Multiple assignment with `:=`

```go
x, y := 10, 20
name, age := "Alice", 30

// Swap variables elegantly
x, y = y, x
```

A key rule: `:=` requires **at least one new variable** on the left:

```go
x := 10
x, y := 20, 30  // ✅ y is new — allowed
x, y := 20, 30  // ❌ if both exist already — compile error
x, y = 20, 30   // ✅ use = (not :=) to reassign both
```

---

## 🔢 Basic Data Types

Go has a rich set of built-in types. These are the ones you'll use constantly:

---

### Integer Types

Go has both **signed** (positive and negative) and **unsigned** (non-negative only) integers:

| Type | Size | Range |
|------|------|-------|
| `int8` | 8-bit | -128 to 127 |
| `int16` | 16-bit | -32,768 to 32,767 |
| `int32` | 32-bit | ~-2.1B to ~2.1B |
| `int64` | 64-bit | very large |
| `int` | platform-dependent (32 or 64-bit) | depends on OS |
| `uint8` | 8-bit unsigned | 0 to 255 |
| `uint16` | 16-bit unsigned | 0 to 65,535 |
| `uint32` | 32-bit unsigned | 0 to ~4.3B |
| `uint64` | 64-bit unsigned | 0 to very large |
| `uint` | platform-dependent | depends on OS |
| `byte` | alias for `uint8` | 0 to 255 |
| `rune` | alias for `int32` | Unicode code point |

**Which integer type to use?**
- Use **`int`** for most cases — it's the default and most idiomatic
- Use `int64` when you need guaranteed 64-bit (e.g. timestamps, large IDs)
- Use `byte` for raw binary data
- Use `rune` for individual characters/Unicode

```go
age := 30          // int (inferred)
var bigNum int64 = 9_000_000_000  // underscores for readability (Go 1.13+)
var b byte = 'A'   // byte = 72 (ASCII value of A)
var r rune = '🐹'  // rune = Unicode code point
```

---

### Floating-Point Types

| Type | Size | Precision |
|------|------|-----------|
| `float32` | 32-bit | ~7 decimal digits |
| `float64` | 64-bit | ~15 decimal digits |

**Always use `float64`** — it's the default and has better precision. Only use `float32` for specific memory-critical applications.

```go
pi := 3.14159265358979  // float64 (inferred)
var ratio float32 = 1.5

fmt.Printf("%.2f\n", pi)   // prints: 3.14
```

---

### Boolean Type

```go
var active bool = true
isReady := false

fmt.Println(active && isReady)  // false — AND
fmt.Println(active || isReady)  // true  — OR
fmt.Println(!active)            // false — NOT
```

> Go has **no truthy/falsy** concept. Only `true` and `false` are booleans. `0`, `""`, `nil` are NOT automatically false like in Python/JavaScript.

---

### String Type

Strings in Go are **immutable sequences of bytes** (UTF-8 encoded by default).

```go
name := "Alice"
greeting := "Hello, " + name   // concatenation

// Multi-line string with backtick (raw string literal)
message := `This is
a multi-line
string`

// String length — in bytes (not characters!)
fmt.Println(len("hello"))   // 5
fmt.Println(len("héllo"))   // 6 (é is 2 bytes in UTF-8)

// Accessing individual bytes
s := "Go"
fmt.Println(s[0])           // 71 (byte value of 'G')
fmt.Printf("%c\n", s[0])   // G (print as character)

// Iterating over characters (runes) — handles Unicode correctly
for i, ch := range "Go🐹" {
    fmt.Printf("index %d: %c\n", i, ch)
}
```

**Double quotes `"` vs backtick `` ` ``:**

```go
"Hello\nWorld"   // interpreted string — \n is a newline
`Hello\nWorld`   // raw string literal — \n is literally backslash-n (no escape processing)
```

---

### Complex Types (Rare)

Go has `complex64` and `complex128` for complex numbers. You'll almost never need these in backend/systems work.

```go
c := complex(3, 4)          // 3 + 4i
fmt.Println(real(c))        // 3
fmt.Println(imag(c))        // 4
```

---

## 🧊 Zero Values

A key Go feature: **every type has a zero value**. If you declare a variable without assigning it, it gets this default. No null pointer surprises!

| Type | Zero Value |
|------|-----------|
| `int`, `int64`, etc. | `0` |
| `float32`, `float64` | `0.0` |
| `bool` | `false` |
| `string` | `""` (empty string) |
| Pointer | `nil` |
| Slice | `nil` |
| Map | `nil` |
| Interface | `nil` |
| Struct | all fields set to their zero values |

```go
var count int      // 0
var name string    // ""
var active bool    // false
var ptr *int       // nil

fmt.Println(count, name, active, ptr)  // 0  false <nil>
```

> This is different from many languages where uninitialized variables are undefined/garbage. In Go, everything is always initialized.

---

## 🔄 Type Conversion

Go is **strictly typed** — there are no implicit type conversions. You must convert explicitly.

```go
var x int = 42
var y float64 = float64(x)    // int → float64
var z int = int(y)            // float64 → int (truncates, doesn't round)

// String conversions
s := string(65)               // "A" — converts int to Unicode character!
// To convert number to string representation, use fmt or strconv:
import "strconv"
s2 := strconv.Itoa(42)        // "42" — int to string
n, err := strconv.Atoi("42")  // 42  — string to int (returns error too)
```

> ⚠️ `string(65)` gives you `"A"` (the character), NOT `"65"`. Use `strconv.Itoa(65)` for `"65"`.

---

## 🔤 Type Aliases: `byte` and `rune`

These are just aliases — they exist to make code more readable:

```go
var b byte = 'A'    // same as uint8 — for binary data, ASCII
var r rune = '界'    // same as int32 — for Unicode characters

// byte and uint8 are interchangeable
var u uint8 = b  // ✅ no conversion needed — same type

// Iterating strings with rune (handles Unicode)
for _, r := range "Hello, 世界" {
    fmt.Printf("%c ", r)  // H e l l o ,   世 界
}
```

---

## 🔍 Checking Types at Runtime

Use `fmt.Printf("%T", value)` to see a variable's type:

```go
x := 42
y := 3.14
z := "hello"
w := true

fmt.Printf("%T\n", x)   // int
fmt.Printf("%T\n", y)   // float64
fmt.Printf("%T\n", z)   // string
fmt.Printf("%T\n", w)   // bool
```

---

## 💻 Complete Example — Putting It All Together

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // --- Variable declarations ---
    var firstName string = "Alice"
    lastName := "Smith"
    fullName := firstName + " " + lastName

    var age int = 30
    height := 5.9   // float64

    var isActive bool = true

    // --- Zero values ---
    var score int    // 0
    var grade string // ""
    fmt.Println("Score:", score, "| Grade:", grade)

    // --- Multiple assignment ---
    x, y := 10, 20
    fmt.Println("Before swap:", x, y)
    x, y = y, x
    fmt.Println("After swap:", x, y)

    // --- Type conversion ---
    ageAsFloat := float64(age)
    heightAsInt := int(height)  // truncates to 5

    // --- String / int conversion ---
    ageStr := strconv.Itoa(age)
    fmt.Println("Age as string: " + ageStr)

    // --- Types ---
    fmt.Printf("fullName: %v (%T)\n", fullName, fullName)
    fmt.Printf("age:      %v (%T)\n", age, age)
    fmt.Printf("height:   %v (%T)\n", height, height)
    fmt.Printf("isActive: %v (%T)\n", isActive, isActive)

    // Avoid "declared and not used" errors
    _ = ageAsFloat
    _ = heightAsInt
}
```

Output:
```
Score: 0 | Grade:
Before swap: 10 20
After swap: 20 10
Age as string: 42
fullName: Alice Smith (string)
age:      30 (int)
height:   5.9 (float64)
isActive: true (bool)
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. `:=` only works inside functions
```go
// ❌ Compile error — := at package level
package main
x := 10   // syntax error

// ✅
var x = 10
```

### 2. Declaring a variable and not using it
```go
// ❌ Compile error
func main() {
    x := 5  // x declared and not used
}

// ✅ Use it or discard it with _
func main() {
    x := 5
    fmt.Println(x)   // used
    _ = someValue    // intentionally discarded
}
```

### 3. No implicit type conversion
```go
var x int = 10
var y float64 = x   // ❌ cannot use x (int) as float64

var y float64 = float64(x)  // ✅ explicit conversion required
```

### 4. `string(number)` is NOT string representation
```go
s := string(65)           // "A"  — NOT "65"!
s2 := strconv.Itoa(65)    // "65" — this is what you want
s3 := fmt.Sprintf("%d", 65) // "65" — also works
```

### 5. Integer division truncates
```go
result := 7 / 2    // 3, NOT 3.5 — both operands are int
result2 := 7.0 / 2 // 3.5 — one is float, so float division
```

### 6. Overflow is NOT an error at runtime
```go
var x int8 = 127
x++           // x becomes -128 (wraps around!) — no panic
fmt.Println(x) // -128
```

---

## 🎯 Must-Know Things (Quick Reference)

- Go has **two ways** to declare variables: `var` (package or function level) and `:=` (function only)
- `:=` is the **idiomatic** choice inside functions
- Every type has a **zero value** — no uninitialized variables
- Go is **strictly typed** — no implicit conversions, must convert explicitly
- Use **`int`** for integers and **`float64`** for decimals unless you have a specific reason not to
- `string(65)` gives `"A"`, not `"65"` — use `strconv.Itoa()` for number-to-string
- `byte` = `uint8` (raw bytes), `rune` = `int32` (Unicode characters)
- Use `%T` format verb to print a variable's type
- Integer division **truncates** (7/2 = 3, not 3.5)
- Unused variables are a **compile error**

---

## ❓ Interview Questions & Answers

---

**Q1: What are the two ways to declare a variable in Go? When do you use each?**

> Go has `var` and `:=`. The `var` keyword works everywhere — inside and outside functions — and is explicit about the type. The short declaration `:=` is only valid inside functions, is more concise, and infers the type automatically. In practice, `:=` is preferred inside functions for local variables, and `var` is required at the package level.

---

**Q2: What is a zero value in Go? Why does it matter?**

> Every type in Go has a default zero value assigned when a variable is declared without an explicit value. For example: `int` → `0`, `string` → `""`, `bool` → `false`, pointer/slice/map → `nil`. This matters because it eliminates an entire class of bugs — you can never have an "uninitialized" variable with garbage data. Go guarantees memory is always initialized to a predictable value.

---

**Q3: Does Go support implicit type conversion?**

> No. Go does not support implicit type conversion. You must always convert explicitly using the type as a function: `float64(myInt)`, `int(myFloat)`. This prevents subtle bugs like precision loss or unexpected behavior that can happen in languages like C where integers silently become floats.

---

**Q4: What is the difference between `int`, `int32`, and `int64` in Go?**

> `int32` is always exactly 32 bits, and `int64` is always exactly 64 bits, regardless of the platform. `int` is platform-dependent — it's 32 bits on 32-bit systems and 64 bits on 64-bit systems. For most use cases, `int` is idiomatic. Use `int64` when you need a guaranteed 64-bit range (e.g., timestamps, large counts), and use `int32`/`int64` when working with binary protocols that specify exact sizes.

---

**Q5: What is the difference between `byte` and `rune`?**

> `byte` is an alias for `uint8` and represents a single byte — it's used for raw binary data and ASCII characters. `rune` is an alias for `int32` and represents a Unicode code point. Since Go strings are UTF-8, a single character might take 1-4 bytes. When you need to iterate over characters correctly (including non-ASCII like Chinese, emoji), use `rune`. When working with raw bytes or ASCII, use `byte`.

---

**Q6: What does `string(65)` produce in Go?**

> It produces `"A"`, not `"65"`. In Go, `string(n)` where `n` is an integer creates a string containing the Unicode character with that code point. `65` is the code point for `'A'`. To convert a number to its string representation (e.g., `65` → `"65"`), use `strconv.Itoa(65)` or `fmt.Sprintf("%d", 65)`.

---

**Q7: Why does `7 / 2` equal `3` in Go?**

> When both operands of a division are integers, Go performs integer division, which truncates the result toward zero. So `7 / 2` is `3`, not `3.5`. To get a float result, at least one operand must be a float: `7.0 / 2` or `float64(7) / 2` both give `3.5`.

---

**Q8: Can you use `:=` to redeclare an existing variable?**

> Not exactly, but there's a nuance. If you use `:=` with multiple variables and at least one is new, it's allowed — the existing variable is simply reassigned, and the new variable is created. For example, `x, err := someFunc()` is valid even if `err` already exists, as long as `x` is new. But if ALL variables on the left already exist, `:=` is a compile error — you must use `=` instead.

---

**Q9: What is the difference between `var name string` and `name := ""`?**

> Both declare a string variable named `name` initialized to the empty string `""`. Functionally they're identical in this case. However, `var name string` is explicit — it shows the type clearly and works at the package level. `name := ""` is shorthand that infers the type, only works inside functions. In style, `var name string` is preferred when the zero value is intentional and you want to signal that — it communicates "I want a string starting empty" more clearly.

---

**Q10: How do you print the type of a variable in Go?**

> Use the `%T` format verb with `fmt.Printf`:
> ```go
> x := 42
> fmt.Printf("%T\n", x)  // int
> ```
> At runtime, you can also use `reflect.TypeOf(x)` from the `reflect` package, which returns a `reflect.Type` value you can inspect programmatically.

---

*Next up: **Topic 1.4 — Constants & `iota`** 🚀*
