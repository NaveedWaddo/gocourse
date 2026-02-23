# 1.5 — Operators

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.4 — Constants & `iota`

---

## 📖 Overview

Operators are symbols that perform operations on values. Go has a clean, minimal set — no operator overloading, no surprises. We'll cover arithmetic, comparison, logical, bitwise, and assignment operators.

---

## ➕ Arithmetic Operators

| Operator | Name | Example | Result |
|----------|------|---------|--------|
| `+` | Addition | `5 + 3` | `8` |
| `-` | Subtraction | `5 - 3` | `2` |
| `*` | Multiplication | `5 * 3` | `15` |
| `/` | Division | `7 / 2` | `3` (integer division!) |
| `%` | Modulus (remainder) | `7 % 2` | `1` |

```go
a, b := 10, 3

fmt.Println(a + b)   // 13
fmt.Println(a - b)   // 7
fmt.Println(a * b)   // 30
fmt.Println(a / b)   // 3  — integer division, truncates
fmt.Println(a % b)   // 1  — remainder

// Float division
x, y := 7.0, 2.0
fmt.Println(x / y)   // 3.5
```

### `+` with strings — concatenation

```go
first := "Hello"
second := "World"
fmt.Println(first + ", " + second + "!")   // Hello, World!
```

---

## 🔁 Increment & Decrement

Go has `++` and `--` but they are **statements**, not expressions. You cannot use them inside an expression.

```go
x := 5
x++   // x = 6
x--   // x = 5

// ❌ These are NOT allowed in Go:
y := x++   // compile error — ++ is a statement, not an expression
if x++ > 5 { }  // compile error

// ✅ Only valid as standalone statements:
x++
x--
```

> Also, Go has **no pre-increment**. `++x` is a compile error. Only `x++` and `x--` (postfix, standalone) are valid.

---

## ⚖️ Comparison Operators

These always return a `bool` (`true` or `false`):

| Operator | Meaning | Example |
|----------|---------|---------|
| `==` | Equal | `5 == 5` → `true` |
| `!=` | Not equal | `5 != 3` → `true` |
| `<` | Less than | `3 < 5` → `true` |
| `>` | Greater than | `5 > 3` → `true` |
| `<=` | Less than or equal | `5 <= 5` → `true` |
| `>=` | Greater than or equal | `6 >= 5` → `true` |

```go
a, b := 10, 20

fmt.Println(a == b)   // false
fmt.Println(a != b)   // true
fmt.Println(a < b)    // true
fmt.Println(a > b)    // false
fmt.Println(a <= 10)  // true
fmt.Println(b >= 20)  // true
```

> ⚠️ You can only compare values of the **same type**. `10 == 10.0` is a compile error — one is `int`, the other is `float64`.

```go
var x int = 10
var y float64 = 10.0
fmt.Println(x == y)           // ❌ compile error — mismatched types
fmt.Println(float64(x) == y)  // ✅ true — explicit conversion first
```

---

## 🧠 Logical Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `&&` | AND — true if both are true | `true && false` → `false` |
| `\|\|` | OR — true if at least one is true | `true \|\| false` → `true` |
| `!` | NOT — flips the boolean | `!true` → `false` |

```go
a, b := true, false

fmt.Println(a && b)   // false
fmt.Println(a || b)   // true
fmt.Println(!a)       // false
fmt.Println(!b)       // true
```

### Short-circuit evaluation

Go uses **short-circuit evaluation** — once the result is determined, the rest is not evaluated:

```go
// If the first condition is false, the second is never evaluated
if false && expensiveCheck() {   // expensiveCheck() never called
}

// If the first condition is true, the second is never evaluated
if true || expensiveCheck() {    // expensiveCheck() never called
}
```

This matters when the right side has side effects or could panic (e.g., nil pointer dereference):

```go
// Safe nil check — if p is nil, p.value is never accessed
if p != nil && p.value > 0 {
    fmt.Println(p.value)
}
```

---

## 🔧 Assignment Operators

| Operator | Equivalent | Example |
|----------|-----------|---------|
| `=` | assign | `x = 5` |
| `+=` | add and assign | `x += 3` → `x = x + 3` |
| `-=` | subtract and assign | `x -= 3` → `x = x - 3` |
| `*=` | multiply and assign | `x *= 3` → `x = x * 3` |
| `/=` | divide and assign | `x /= 3` → `x = x / 3` |
| `%=` | modulus and assign | `x %= 3` → `x = x % 3` |

```go
x := 10

x += 5    // x = 15
x -= 3    // x = 12
x *= 2    // x = 24
x /= 4    // x = 6
x %= 4    // x = 2

fmt.Println(x)   // 2
```

---

## 🔢 Bitwise Operators

Used for low-level bit manipulation — important for flags, network protocols, performance-critical code.

| Operator | Meaning | Example |
|----------|---------|---------|
| `&` | AND — both bits must be 1 | `5 & 3` → `1` |
| `\|` | OR — at least one bit is 1 | `5 \| 3` → `7` |
| `^` | XOR — bits differ | `5 ^ 3` → `6` |
| `^x` | NOT (unary) — flips all bits | `^5` → `-6` |
| `<<` | Left shift | `1 << 3` → `8` |
| `>>` | Right shift | `8 >> 2` → `2` |
| `&^` | AND NOT (bit clear) | `5 &^ 3` → `4` |

```go
a := 5   // binary: 0101
b := 3   // binary: 0011

fmt.Println(a & b)    // 1  — 0001 (AND: both bits must be 1)
fmt.Println(a | b)    // 7  — 0111 (OR: at least one bit is 1)
fmt.Println(a ^ b)    // 6  — 0110 (XOR: bits differ)
fmt.Println(a << 1)   // 10 — 1010 (shift left by 1 = multiply by 2)
fmt.Println(a >> 1)   // 2  — 0010 (shift right by 1 = divide by 2)
fmt.Println(a &^ b)   // 4  — 0100 (AND NOT: clear bits that are set in b)
```

### Practical use — bit flags (from Topic 1.4)

```go
const (
    Read    = 1 << iota   // 001
    Write                 // 010
    Execute               // 100
)

perms := Read | Write     // 011 = 3

// Check a permission
hasRead := perms & Read != 0    // true
hasExec := perms & Execute != 0 // false

// Clear a permission (remove Write)
perms = perms &^ Write    // 001 = 1 (only Read remains)
```

---

## 📊 Operator Precedence

When multiple operators appear in an expression, higher precedence operators are evaluated first. Use parentheses to make intent clear.

| Precedence | Operators |
|-----------|-----------|
| 5 (highest) | `*`, `/`, `%`, `<<`, `>>`, `&`, `&^` |
| 4 | `+`, `-`, `\|`, `^` |
| 3 | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| 2 | `&&` |
| 1 (lowest) | `\|\|` |

```go
// Without parentheses — may be confusing
result := 2 + 3*4        // 14 (not 20) — * before +
result2 := true || false && false  // true — && before ||

// With parentheses — always clear
result3 := (2 + 3) * 4   // 20
result4 := (true || false) && false  // false
```

> **Best practice:** When in doubt, use parentheses. They make intent explicit and prevent bugs.

---

## 💻 Complete Example

```go
package main

import "fmt"

func main() {
    // Arithmetic
    a, b := 17, 5
    fmt.Printf("%d + %d = %d\n", a, b, a+b)   // 22
    fmt.Printf("%d / %d = %d\n", a, b, a/b)   // 3 (integer division)
    fmt.Printf("%d %% %d = %d\n", a, b, a%b)  // 2

    // Comparison
    fmt.Println(a > b)    // true
    fmt.Println(a == b)   // false

    // Logical with short-circuit
    x := 10
    if x > 5 && x < 20 {
        fmt.Println("x is between 5 and 20")
    }

    // Assignment operators
    count := 0
    count += 10
    count *= 2
    fmt.Println("count:", count)   // 20

    // Bit flags
    const (
        CanRead    = 1 << iota   // 1
        CanWrite                 // 2
        CanExecute               // 4
    )

    userPerms := CanRead | CanWrite   // 3

    if userPerms & CanRead != 0 {
        fmt.Println("Has read access")
    }
    if userPerms & CanExecute == 0 {
        fmt.Println("No execute access")
    }

    // Clear a permission
    userPerms = userPerms &^ CanWrite
    fmt.Printf("After removing write: %b\n", userPerms)   // 1
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Integer division truncates silently
```go
result := 7 / 2    // 3, NOT 3.5 — no warning!
// Fix: make at least one operand a float
result2 := 7.0 / 2   // 3.5
result3 := float64(7) / float64(2)  // 3.5
```

### 2. `++` is a statement, not an expression
```go
x := 5
y := x++   // ❌ compile error
if x++ > 5 { }  // ❌ compile error
x++        // ✅ standalone statement only
```

### 3. No pre-increment
```go
++x   // ❌ compile error in Go (valid in C/Java but NOT in Go)
x++   // ✅
```

### 4. Comparing different types
```go
var a int = 5
var b int32 = 5
fmt.Println(a == b)   // ❌ compile error — int vs int32
fmt.Println(a == int(b))  // ✅ explicit conversion
```

### 5. `^` is both XOR (binary) and NOT (unary)
```go
a ^ b    // XOR (binary — two operands)
^a       // bitwise NOT (unary — one operand)
```

---

## 🎯 Must-Know Things (Quick Reference)

- Integer division **truncates** — `7/2 = 3`, not `3.5`
- `++` and `--` are **statements** only — can't use them in expressions
- Go has **no pre-increment** (`++x` is invalid)
- Logical operators use **short-circuit evaluation**
- You can only compare values of the **same type**
- `&^` is Go's unique **AND NOT** (bit clear) operator — not in most other languages
- `<<` left shift = multiply by 2ⁿ, `>>` right shift = divide by 2ⁿ
- Use **parentheses** to make precedence explicit

---

## ❓ Interview Questions & Answers

---

**Q1: What is the result of `7 / 2` in Go?**

> `3`. When both operands are integers, Go performs integer division which truncates toward zero. The decimal part is lost. To get `3.5`, at least one operand must be a float: `7.0 / 2` or `float64(7) / float64(2)`.

---

**Q2: What makes Go's `++` different from C or Java?**

> In Go, `++` and `--` are **statements**, not expressions. This means they cannot appear inside an expression or assignment — `y := x++` or `if x++ > 5` are compile errors. Also, Go only has postfix `x++`, not prefix `++x`. This was a deliberate design decision to eliminate a class of confusing bugs common in C-style languages.

---

**Q3: What is short-circuit evaluation and why does it matter in Go?**

> Short-circuit evaluation means that in `a && b`, if `a` is false, `b` is never evaluated. In `a || b`, if `a` is true, `b` is never evaluated. This matters for performance (skip expensive calls) and safety — a common Go pattern is `if ptr != nil && ptr.field > 0` where the second check would panic if `ptr` were nil but is safely skipped due to short-circuiting.

---

**Q4: What does the `&^` operator do?**

> `&^` is the AND NOT (bit clear) operator, unique to Go. `x &^ y` clears the bits in `x` that are set in `y`. It's equivalent to `x & (^y)`. It's most useful for removing a flag from a bitmask: `permissions &^ Write` removes the Write bit from permissions without affecting the others.

---

**Q5: Can you compare an `int` and `int32` directly in Go?**

> No. Go requires operands of comparison operators to be the same type. Comparing `int` and `int32` is a compile error. You must convert one to match the other: `int(myInt32) == myInt` or `myInt32 == int32(myInt)`.

---

*Next up: **Topic 1.6 — Control Flow** 🚀*
