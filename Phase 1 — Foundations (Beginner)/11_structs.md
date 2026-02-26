# 1.11 — Structs

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.10 — Maps

---

## 📖 What is a Struct?

A **struct** is a composite type that groups together fields of different types under one name. It's Go's primary way to model real-world entities and data — the equivalent of a class in other languages, but without inheritance.

```go
type Person struct {
    Name string
    Age  int
    Email string
}
```

---

## 🏗️ Defining and Creating Structs

### Definition

```go
type Point struct {
    X float64
    Y float64
}

type Employee struct {
    Name       string
    Age        int
    Department string
    Salary     float64
    Active     bool
}
```

### Creating struct instances

```go
// Method 1: Named fields (idiomatic — order doesn't matter)
p := Person{
    Name:  "Alice",
    Age:   30,
    Email: "alice@example.com",
}

// Method 2: Positional (not recommended — fragile if fields change)
p2 := Person{"Bob", 25, "bob@example.com"}

// Method 3: Zero value then set fields
var p3 Person
p3.Name = "Carol"
p3.Age = 28

// Method 4: new() — returns a pointer
p4 := new(Person)
p4.Name = "Dave"

// Method 5: Address of literal — most common for pointer structs
p5 := &Person{
    Name: "Eve",
    Age:  35,
}
```

### Zero value of a struct

When you declare a struct without initializing, all fields get their zero values:

```go
var p Person
fmt.Println(p.Name)   // "" (empty string)
fmt.Println(p.Age)    // 0
fmt.Println(p.Active) // false
```

---

## 🔑 Accessing and Modifying Fields

```go
p := Person{Name: "Alice", Age: 30}

// Access
fmt.Println(p.Name)   // Alice

// Modify
p.Age = 31
fmt.Println(p.Age)    // 31

// Via pointer — Go auto-dereferences
ptr := &p
fmt.Println(ptr.Name)   // Alice (same as (*ptr).Name)
ptr.Age = 32
fmt.Println(p.Age)      // 32 — original modified!
```

---

## 📦 Structs are Value Types

Assigning a struct copies all its fields — structs are value types just like `int`:

```go
a := Person{Name: "Alice", Age: 30}
b := a           // b is a complete copy

b.Name = "Bob"
fmt.Println(a.Name)   // Alice — unchanged
fmt.Println(b.Name)   // Bob
```

> Use pointers when you need shared mutation or want to avoid copying large structs.

---

## 🏗️ Constructor Functions

Go has no built-in constructor keyword. The convention is a `New...` function that returns a pointer to an initialized struct:

```go
type Rectangle struct {
    Width  float64
    Height float64
}

// Constructor — validates and initializes
func NewRectangle(w, h float64) (*Rectangle, error) {
    if w <= 0 || h <= 0 {
        return nil, fmt.Errorf("width and height must be positive, got %v and %v", w, h)
    }
    return &Rectangle{Width: w, Height: h}, nil
}

func main() {
    r, err := NewRectangle(4, 5)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(r.Width, r.Height)   // 4 5
}
```

---

## 🔗 Struct Embedding — Composition over Inheritance

Go has no inheritance. Instead it uses **embedding** — including one struct inside another to reuse its fields and methods:

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal        // embedded — no field name, just the type
    Breed string
}

func main() {
    d := Dog{
        Animal: Animal{Name: "Rex"},
        Breed:  "Labrador",
    }

    // Access Animal's fields directly — promoted to Dog
    fmt.Println(d.Name)     // Rex (same as d.Animal.Name)
    fmt.Println(d.Speak())  // Rex makes a sound (Animal's method promoted)
    fmt.Println(d.Breed)    // Labrador
}
```

### Overriding promoted methods

```go
func (d Dog) Speak() string {
    return d.Name + " barks!"
}

d.Speak()          // "Rex barks!" — Dog's own method takes precedence
d.Animal.Speak()   // "Rex makes a sound" — access embedded explicitly
```

### Embedding multiple structs

```go
type Logger struct{ Prefix string }
func (l Logger) Log(msg string) { fmt.Println(l.Prefix+":", msg) }

type Server struct {
    Logger              // embedded
    Host   string
    Port   int
}

s := Server{
    Logger: Logger{Prefix: "SERVER"},
    Host:   "localhost",
    Port:   8080,
}

s.Log("started")   // SERVER: started — promoted from Logger
```

---

## 🏷️ Struct Tags

Struct tags are metadata strings attached to fields — used by packages like `encoding/json`, `gorm`, `validate`, etc.:

```go
type User struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Password  string `json:"-"`           // "-" means skip this field
    Age       int    `json:"age,omitempty"` // omitempty: skip if zero value
}
```

```go
user := User{ID: 1, FirstName: "Alice", LastName: "Smith", Password: "secret", Age: 0}

data, _ := json.Marshal(user)
fmt.Println(string(data))
// {"id":1,"first_name":"Alice","last_name":"Smith"}
// Password is omitted (json:"-"), Age is omitted (omitempty + zero value)
```

Tags are just strings — you can define your own for custom use:

```go
type Config struct {
    Host string `env:"HOST" default:"localhost"`
    Port int    `env:"PORT" default:"8080"`
}
```

Access tags via reflection (Topic 3.4):
```go
t := reflect.TypeOf(Config{})
field, _ := t.FieldByName("Host")
fmt.Println(field.Tag.Get("env"))     // HOST
fmt.Println(field.Tag.Get("default")) // localhost
```

---

## 🔵 Anonymous Structs

A struct without a type name — useful for one-off data shapes, test cases, and JSON decoding:

```go
// Inline anonymous struct
point := struct {
    X, Y int
}{X: 10, Y: 20}
fmt.Println(point.X, point.Y)   // 10 20

// Common use: table-driven tests
tests := []struct {
    input    string
    expected int
}{
    {"hello", 5},
    {"hi", 2},
    {"", 0},
}

for _, tt := range tests {
    result := len(tt.input)
    if result != tt.expected {
        fmt.Printf("FAIL: len(%q) = %d, want %d\n", tt.input, result, tt.expected)
    }
}
```

---

## 🔄 Comparing Structs

Structs are comparable with `==` if **all their fields are comparable**:

```go
type Point struct{ X, Y int }

p1 := Point{1, 2}
p2 := Point{1, 2}
p3 := Point{3, 4}

fmt.Println(p1 == p2)   // true  — all fields equal
fmt.Println(p1 == p3)   // false
```

If a struct has non-comparable fields (slices, maps, functions), `==` is a compile error:

```go
type Bad struct {
    Data []int   // slice is not comparable
}

b1 := Bad{Data: []int{1, 2}}
b2 := Bad{Data: []int{1, 2}}
fmt.Println(b1 == b2)   // ❌ compile error — use reflect.DeepEqual instead
```

---

## 💻 Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
)

// Base struct
type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
    Zip    string `json:"zip"`
}

// Struct with embedding and tags
type Person struct {
    Address                    // embedded — fields promoted
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email,omitempty"`
}

// Constructor
func NewPerson(name string, age int, city string) (*Person, error) {
    if name == "" {
        return nil, fmt.Errorf("name cannot be empty")
    }
    if age < 0 || age > 150 {
        return nil, fmt.Errorf("invalid age: %d", age)
    }
    return &Person{
        Name: name,
        Age:  age,
        Address: Address{City: city},
    }, nil
}

func main() {
    // Constructor usage
    p, err := NewPerson("Alice", 30, "New York")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Accessing embedded fields directly
    p.Street = "123 Main St"
    p.Zip = "10001"

    fmt.Println("Name:", p.Name)
    fmt.Println("City:", p.City)   // promoted from Address
    fmt.Println("Street:", p.Address.Street)   // also accessible explicitly

    // Struct copy (value type)
    p2 := *p   // dereference pointer to copy the struct value
    p2.Name = "Bob"
    fmt.Println("Original:", p.Name)   // Alice — unchanged
    fmt.Println("Copy:", p2.Name)      // Bob

    // JSON marshaling with tags
    data, _ := json.MarshalIndent(p, "", "  ")
    fmt.Println(string(data))

    // Anonymous struct
    config := struct {
        Debug bool
        Port  int
    }{Debug: true, Port: 9090}
    fmt.Printf("Config: %+v\n", config)
}
```

Output:
```
Name: Alice
City: New York
Street: 123 Main St
Original: Alice
Copy: Bob
{
  "street": "123 Main St",
  "city": "New York",
  "zip": "10001",
  "name": "Alice",
  "age": 30
}
Config: {Debug:true Port:9090}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Positional struct literal — fragile
```go
type Point struct{ X, Y, Z float64 }

// ❌ Breaks silently if fields are reordered or added
p := Point{1.0, 2.0, 3.0}

// ✅ Named fields — robust
p := Point{X: 1.0, Y: 2.0, Z: 3.0}
```

### 2. Modifying a struct copy, not the original
```go
p := Person{Name: "Alice"}
modify(p)          // ❌ modifies a copy
fmt.Println(p.Name) // still "Alice"

func modify(p *Person) { p.Name = "Bob" }
modify(&p)         // ✅ pass pointer
```

### 3. Embedding name conflicts
```go
type A struct{ Name string }
type B struct{ Name string }
type C struct {
    A
    B
}

c := C{}
// c.Name   // ❌ ambiguous — compiler error
c.A.Name = "from A"   // ✅ explicit
c.B.Name = "from B"   // ✅ explicit
```

### 4. Struct tags must be formatted exactly right
```go
// ❌ Wrong — spaces around = or missing quotes
Name string `json: "name"` // space after colon — tag ignored!

// ✅ Correct
Name string `json:"name"`
```

### 5. Comparing structs with non-comparable fields
```go
type Bad struct{ Data []int }
b1, b2 := Bad{[]int{1}}, Bad{[]int{1}}
b1 == b2   // ❌ compile error

// ✅ Use reflect.DeepEqual
reflect.DeepEqual(b1, b2)   // true
```

---

## 🎯 Must-Know Things (Quick Reference)

- Structs group fields of different types — Go's primary data modeling tool
- Always use **named fields** in struct literals — positional is fragile
- Structs are **value types** — assignment copies all fields
- Use **pointer to struct** when you need to mutate or avoid large copies
- **Constructor function** convention: `NewTypeName(...)` returning `(*Type, error)`
- **Embedding** promotes fields and methods — Go's composition over inheritance
- **Struct tags** are metadata for packages like `encoding/json`, `gorm`, `validate`
- `json:"-"` skips a field, `json:"name,omitempty"` skips zero values
- **Anonymous structs** are great for one-off shapes and table-driven tests
- Structs with non-comparable fields (slice, map) cannot use `==` — use `reflect.DeepEqual`

---

## ❓ Interview Questions & Answers

---

**Q1: How does Go achieve object-oriented behavior without classes?**

> Go uses structs for data, methods on structs for behavior, and interfaces for polymorphism. Instead of inheritance, Go uses embedding — you include one struct inside another to promote its fields and methods. This is "composition over inheritance". The result is more flexible and avoids the fragile base class problem of traditional OOP. Go's approach is often described as achieving OOP goals without OOP mechanisms.

---

**Q2: What is struct embedding and how does promotion work?**

> Struct embedding means including one struct type inside another without giving it a field name — just the type. The embedded struct's fields and methods are "promoted" to the outer struct, meaning you can access them directly as if they were defined on the outer struct. For example, if `Dog` embeds `Animal`, you can call `d.Speak()` instead of `d.Animal.Speak()`. If there's a name conflict, the outer struct's own field/method takes precedence, and you can always access the embedded one explicitly.

---

**Q3: What are struct tags and how are they used?**

> Struct tags are string literals attached to struct fields using backticks. They're metadata that other packages can read via reflection. The most common use is `encoding/json` — `json:"name"` controls the JSON key name, `json:"-"` skips the field, and `json:"name,omitempty"` skips it when it holds a zero value. Other common users are `gorm` for database column mapping and `validate` for input validation. Tags must follow an exact format — `key:"value"` with no spaces around the colon.

---

**Q4: Are structs value types or reference types in Go?**

> Structs are value types. Assigning a struct to another variable or passing it to a function creates a complete copy of all fields. Changes to the copy don't affect the original. This is different from maps and slices, which are reference types. To share a struct across functions or to mutate it, pass a pointer. For large structs, pointers also avoid expensive copying.

---

**Q5: What is the idiomatic way to create a constructor in Go?**

> Go has no constructor keyword. The convention is a function named `NewTypeName` that takes initialization parameters, validates them, and returns `(*TypeName, error)`. Returning a pointer allows the caller to modify the instance, and returning an error allows validation failures to be communicated cleanly. The function is just a regular function — no special syntax. Some simple types use just `&TypeName{field: value}` directly without a constructor if no validation is needed.

---

**Q6: When would you use an anonymous struct?**

> Anonymous structs — defined inline without a type name — are useful for: (1) table-driven tests, where each test case is a struct with input and expected output fields; (2) one-off JSON decoding where you don't want to define a full named type; (3) grouping related local variables that don't warrant a package-level type. They're a lightweight alternative when you need structure but the type won't be reused.

---

*Next up: **Topic 1.12 — Methods** 🚀*
