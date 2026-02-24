# 1.7 — Functions

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.6 — Control Flow

---

## 📖 Overview

Functions are the core building block of Go programs. Go functions are first-class values — they can be assigned to variables, passed as arguments, and returned from other functions. Go also has some unique features: multiple return values, named returns, variadic functions, and defer (covered in Topic 2.8).

---

## 🧩 Basic Function Syntax

```go
func functionName(param1 type1, param2 type2) returnType {
    // body
    return value
}
```

### Simple example

```go
func add(x int, y int) int {
    return x + y
}

func main() {
    result := add(3, 4)
    fmt.Println(result)   // 7
}
```

---

## 🔤 Parameter Shorthand

When consecutive parameters share the same type, you can group them:

```go
// Verbose form
func add(x int, y int) int { ... }

// Shorthand — same type, grouped
func add(x, y int) int { ... }

// Multiple groups
func mix(a, b int, c, d string, e float64) { ... }
```

---

## 🔁 Multiple Return Values

This is one of Go's most distinctive features. Functions can return **more than one value**.

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide by zero")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(result)   // 5
}
```

This is Go's idiomatic error handling pattern — return the result AND an error. The caller always checks the error.

### Ignoring return values with `_`

```go
result, _ := divide(10, 2)   // ignore the error (only do this if you're sure it's safe)
_, err := divide(10, 0)      // ignore the result, only care about the error
```

---

## 🏷️ Named Return Values

You can name the return values in the function signature. They act as pre-declared variables inside the function:

```go
func minMax(nums []int) (min, max int) {
    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min {
            min = n
        }
        if n > max {
            max = n
        }
    }
    return   // "naked return" — returns named values min and max
}

func main() {
    min, max := minMax([]int{3, 1, 4, 1, 5, 9, 2, 6})
    fmt.Println(min, max)   // 1 9
}
```

### When to use named returns

Named returns are good for:
- **Documentation** — makes the signature self-explanatory
- **Short functions** where the naked return is clear
- **Defer** scenarios where you need to modify return values

Avoid naked returns in **long functions** — they hurt readability.

```go
// ✅ Good — short function, names add clarity
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return
}

// ❌ Bad — long function with naked return is confusing
func complexCalc(data []int) (result int, err error) {
    // ... 50 lines of code ...
    return   // what are we returning? Hard to tell
}
```

---

## 📦 Functions as Values (First-Class Functions)

In Go, functions are first-class citizens — you can assign them to variables and pass them around:

```go
// Assign a function to a variable
add := func(x, y int) int {
    return x + y
}

fmt.Println(add(3, 4))   // 7

// Pass a function as an argument
func apply(x, y int, op func(int, int) int) int {
    return op(x, y)
}

result := apply(3, 4, add)
fmt.Println(result)   // 7
```

---

## 🔒 Anonymous Functions (Closures)

An anonymous function is a function without a name. When it captures variables from the surrounding scope, it becomes a **closure**.

```go
// Anonymous function — immediately invoked
func main() {
    func() {
        fmt.Println("I'm anonymous!")
    }()   // the () at the end calls it immediately
}
```

### Closure — captures outer variable

```go
func makeCounter() func() int {
    count := 0
    return func() int {
        count++   // captures and modifies count from outer scope
        return count
    }
}

func main() {
    counter := makeCounter()
    fmt.Println(counter())   // 1
    fmt.Println(counter())   // 2
    fmt.Println(counter())   // 3

    counter2 := makeCounter()   // independent counter
    fmt.Println(counter2())     // 1
}
```

The inner function "closes over" the `count` variable — it keeps a reference to it even after `makeCounter` returns. This is a **closure**.

---

## 🔄 Variadic Functions

A variadic function accepts a **variable number of arguments** of the same type, using `...`:

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func main() {
    fmt.Println(sum(1, 2))            // 3
    fmt.Println(sum(1, 2, 3, 4, 5))  // 15

    // Spread a slice into variadic function
    nums := []int{1, 2, 3, 4}
    fmt.Println(sum(nums...))         // 10
}
```

Inside the function, `nums` is a `[]int` slice. `fmt.Println` itself is variadic — that's how it accepts any number of arguments.

> The variadic parameter must be **last** in the parameter list.

---

## 🔁 Recursive Functions

Go supports recursion. The classic example:

```go
func factorial(n int) int {
    if n == 0 {
        return 1
    }
    return n * factorial(n-1)
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    fmt.Println(factorial(5))    // 120
    fmt.Println(fibonacci(10))   // 55
}
```

> Go doesn't optimize tail recursion, so deep recursion can stack overflow. For very deep recursion, prefer an iterative approach.

---

## 🏗️ Function Types

You can declare a named function type — useful for callbacks and interfaces:

```go
// Declare a function type
type MathOp func(int, int) int

// Use it as a parameter type
func compute(x, y int, op MathOp) int {
    return op(x, y)
}

func main() {
    multiply := func(a, b int) int { return a * b }
    fmt.Println(compute(3, 4, multiply))   // 12
}
```

---

## 💻 Complete Example

```go
package main

import (
    "errors"
    "fmt"
)

// Basic function with multiple returns
func safeDivide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Named return values
func stats(nums []float64) (mean, variance float64) {
    n := float64(len(nums))
    for _, v := range nums {
        mean += v
    }
    mean /= n
    for _, v := range nums {
        diff := v - mean
        variance += diff * diff
    }
    variance /= n
    return   // naked return
}

// Variadic
func joinStrings(sep string, parts ...string) string {
    result := ""
    for i, p := range parts {
        if i > 0 {
            result += sep
        }
        result += p
    }
    return result
}

// Closure — adder factory
func makeAdder(x int) func(int) int {
    return func(y int) int {
        return x + y   // closes over x
    }
}

func main() {
    // Multiple returns
    if result, err := safeDivide(10, 3); err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Printf("10 / 3 = %.4f\n", result)   // 3.3333
    }

    // Named returns
    data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
    mean, variance := stats(data)
    fmt.Printf("Mean: %.2f, Variance: %.2f\n", mean, variance)

    // Variadic
    fmt.Println(joinStrings(", ", "Go", "is", "awesome"))   // Go, is, awesome

    // Closures
    add5 := makeAdder(5)
    add10 := makeAdder(10)
    fmt.Println(add5(3))    // 8
    fmt.Println(add10(3))   // 13

    // Function as value
    ops := map[string]func(int, int) int{
        "add": func(a, b int) int { return a + b },
        "mul": func(a, b int) int { return a * b },
    }
    fmt.Println(ops["add"](3, 4))   // 7
    fmt.Println(ops["mul"](3, 4))   // 12
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Ignoring error return values
```go
// ❌ Dangerous — silently ignores errors
result, _ := os.Open("file.txt")

// ✅ Always handle errors
f, err := os.Open("file.txt")
if err != nil {
    log.Fatal(err)
}
```

### 2. Closure captures variable by reference, not value
```go
// ❌ Classic bug — all goroutines see the same i (its final value)
funcs := make([]func(), 5)
for i := 0; i < 5; i++ {
    funcs[i] = func() { fmt.Println(i) }
}
for _, f := range funcs {
    f()   // prints 5 5 5 5 5 — NOT 0 1 2 3 4!
}

// ✅ Fix — capture current value by passing as argument
for i := 0; i < 5; i++ {
    i := i   // shadow i with a new variable scoped to the loop body
    funcs[i] = func() { fmt.Println(i) }
}
```

### 3. Naked returns in long functions
```go
// ❌ Confusing in a long function
func process(data []int) (result int, err error) {
    // ... 60 lines ...
    return   // what's being returned?
}

// ✅ Explicit returns are clearer
func process(data []int) (result int, err error) {
    // ...
    return result, err
}
```

### 4. Variadic vs slice parameter
```go
func sumSlice(nums []int) int { ... }   // must pass a slice
func sumVariadic(nums ...int) int { ... }  // can pass individual values OR spread a slice

sumSlice(1, 2, 3)          // ❌ compile error
sumVariadic(1, 2, 3)       // ✅
sumVariadic([]int{1,2,3}...) // ✅ spread slice
```

---

## 🎯 Must-Know Things (Quick Reference)

- Functions can return **multiple values** — the foundation of Go's error handling
- Use `_` to discard unwanted return values
- **Named returns** pre-declare return variables — useful for documentation and short functions
- **Naked returns** (just `return`) should only be used in short functions
- Functions are **first-class values** — assign to variables, pass around, return from functions
- **Closures** capture variables from the enclosing scope **by reference** — classic gotcha in loops
- **Variadic** functions use `...type` — spread a slice with `slice...`
- Go has **no function overloading** — each function name must be unique in its package
- Go has **no default parameter values** — use variadic or option structs instead

---

## ❓ Interview Questions & Answers

---

**Q1: How does Go handle multiple return values?**

> Go natively supports returning multiple values from a function by listing them in the return signature: `func f() (int, error)`. The caller unpacks them with multiple assignment: `result, err := f()`. This is the foundation of Go's error handling idiom — functions return both a result and an error, and the caller always checks the error. You can discard any return value using `_`.

---

**Q2: What are named return values and when should you use them?**

> Named return values declare variables in the function signature that are automatically returned when you use a bare `return` (called a naked return). For example: `func split(n int) (x, y int)`. They're good for short functions where the names document what's returned, and in `defer` scenarios where you need to modify the return value. They should be avoided in long functions because naked returns make it hard to track what's actually being returned.

---

**Q3: What is a closure in Go?**

> A closure is a function value that references variables from outside its own body — it "closes over" those variables. The function can read and modify them even after the outer function has returned. For example, `makeCounter()` returns an inner function that keeps incrementing a `count` variable defined in the outer function. Each call to `makeCounter()` creates an independent `count` — closures are a powerful way to encapsulate state without structs.

---

**Q4: What is the closure-in-loop gotcha?**

> When you create closures inside a loop, all closures capture the same loop variable by reference — not by value. So when they execute later, they all see the variable's final value. The fix is to shadow the variable inside the loop body with `:=` to create a new variable for each iteration: `i := i`. This is one of the most common Go bugs, especially with goroutines.

---

**Q5: What is a variadic function and how do you pass a slice to one?**

> A variadic function accepts zero or more arguments of the same type using `...type` syntax, e.g. `func sum(nums ...int) int`. Inside the function, `nums` is a `[]int` slice. To pass an existing slice to a variadic function, you spread it with `...`: `sum(mySlice...)`. You cannot pass a slice directly without spreading — it would be a type mismatch.

---

**Q6: Does Go support function overloading?**

> No. Go does not support function overloading — you cannot have two functions with the same name but different parameter types in the same package. This is intentional — it keeps code simpler and more readable. Common alternatives include variadic functions, option structs, or simply naming functions differently (e.g., `ParseInt`, `ParseFloat`, `ParseBool`).

---

**Q7: How do you implement default parameter values in Go?**

> Go has no built-in default parameter values. The common patterns are: (1) variadic parameters where zero args means "use default"; (2) an options struct where callers only set what they need; (3) separate functions with descriptive names. The options struct pattern is widely used in libraries — you pass a config struct and the function uses defaults for zero values.

---

*Next up: **Topic 1.8 — Pointers** 🚀*
