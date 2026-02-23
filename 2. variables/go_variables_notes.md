# Go Variables -- Complete Notes (Beginner to Senior Level)

## 1. What is a Variable?

A variable is a named memory location used to store data.

Example:

``` go
var age int = 25
```

-   `age` = variable name
-   `int` = data type
-   `25` = value

Variables store data in RAM during program execution.

------------------------------------------------------------------------

## 2. Ways to Declare Variables

### Method 1: Standard declaration

``` go
var name string = "Naveed"
```

### Method 2: Type inference

``` go
var name = "Naveed"
```

Compiler automatically detects type.

### Method 3: Short declaration (MOST COMMON)

``` go
name := "Naveed"
```

Only allowed inside functions.

------------------------------------------------------------------------

## 3. Variable Declaration Without Initialization

``` go
var age int
```

Default value (zero value):

    0

------------------------------------------------------------------------

## 4. Zero Values (VERY IMPORTANT)

  Type      Zero Value
  --------- ------------
  int       0
  float     0.0
  string    ""
  bool      false
  pointer   nil
  slice     nil
  map       nil

Example:

``` go
var name string
fmt.Println(name) // ""
```

------------------------------------------------------------------------

## 5. Multiple Variable Declaration

``` go
var a, b, c int = 1, 2, 3
```

Block declaration:

``` go
var (
 name = "Naveed"
 age = 25
 isActive = true
)
```

------------------------------------------------------------------------

## 6. Short Declaration Rules

Valid:

``` go
x := 10
```

Invalid (outside function):

``` go
var x := 10 ❌
```

------------------------------------------------------------------------

## 7. Scope of Variables

### Local Variable

Declared inside function.

``` go
func main() {
 x := 10
}
```

### Global Variable

Declared outside function.

``` go
var x = 10
```

------------------------------------------------------------------------

## 8. Variable Memory Allocation (Senior Level)

Variables stored in:

-   Stack (fast)
-   Heap (slower)

Stack example:

``` go
x := 10
```

Heap example:

``` go
func test() *int {
 x := 10
 return &x
}
```

Compiler moves x to heap (escape analysis).

------------------------------------------------------------------------

## 9. Stack vs Heap Comparison

  Stack               Heap
  ------------------- -------------------
  Faster              Slower
  Automatic cleanup   Garbage collected
  Small memory        Large memory

Senior engineers minimize heap allocation.

------------------------------------------------------------------------

## 10. Variable Shadowing (Important)

``` go
x := 10

if true {
 x := 20
}
```

Inner x is different variable.

------------------------------------------------------------------------

## 11. Naming Conventions

Good:

``` go
userName
totalCount
isActive
```

Bad:

``` go
x
a
temp
```

------------------------------------------------------------------------

## 12. Exported vs Unexported Variables

Exported:

``` go
var UserName string
```

Unexported:

``` go
var userName string
```

Capital letter = exported.

------------------------------------------------------------------------

## 13. Type Safety

Go is statically typed.

Invalid:

``` go
var x int = "hello" ❌
```

------------------------------------------------------------------------

## 14. Explicit Type Conversion

``` go
var x int = 10
var y float64 = float64(x)
```

------------------------------------------------------------------------

## 15. Constants vs Variables

Variable:

``` go
var x = 10
x = 20
```

Constant:

``` go
const x = 10
```

Cannot change.

------------------------------------------------------------------------

## 16. Senior-Level Concepts

### Escape Analysis

Compiler decides stack vs heap.

### Memory Alignment

Optimizes struct layout.

### Variable Lifetime

Defined by scope.

### Garbage Collection

Heap variables cleaned automatically.

------------------------------------------------------------------------

## 17. Best Practices (Senior Engineers)

-   Use short declaration inside functions
-   Use int64 for backend systems
-   Avoid unnecessary heap allocation
-   Use meaningful variable names
-   Minimize global variables

------------------------------------------------------------------------

## 18. Backend Example

``` go
type User struct {
 ID int64
 Name string
 Email string
}

func getUser() *User {
 return &User{
  ID: 1,
  Name: "Naveed",
 }
}
```

Pointer avoids copying large struct.

------------------------------------------------------------------------

## 19. Interview Questions

-   What is variable?
-   What is zero value?
-   Difference between := and var?
-   Stack vs heap?
-   What is escape analysis?
-   What is variable scope?

------------------------------------------------------------------------

## Summary

Variables store data in memory. Go supports multiple declaration
methods, strict typing, zero values, and efficient memory allocation.
Senior engineers must understand stack vs heap, escape analysis, and
memory optimization.
