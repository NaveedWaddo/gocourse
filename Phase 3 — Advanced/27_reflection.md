# 3.4 — Reflection

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.3 — Advanced Concurrency Patterns

---

## 📖 What is Reflection?

Reflection lets a program **inspect and manipulate its own types and values at runtime** — without knowing them at compile time. Go's `reflect` package provides this capability.

Use reflection when you genuinely cannot know the type at compile time — encoding libraries, ORMs, dependency injection frameworks, test helpers. Most application code should never use it.

```go
import "reflect"
```

The two core types: `reflect.Type` and `reflect.Value`.

---

## 🔍 `reflect.TypeOf` and `reflect.ValueOf`

```go
x := 42
t := reflect.TypeOf(x)   // reflect.Type
v := reflect.ValueOf(x)  // reflect.Value

fmt.Println(t)            // int
fmt.Println(t.Kind())     // int
fmt.Println(v)            // 42
fmt.Println(v.Int())      // 42
```

### `Type` vs `Kind`

```go
type Celsius float64
var c Celsius = 37.0

t := reflect.TypeOf(c)
fmt.Println(t.Name())    // Celsius   — the named type
fmt.Println(t.Kind())    // float64   — the underlying kind
```

`Type` is the specific named type (`Celsius`, `[]int`, `MyStruct`). `Kind` is the underlying category (`float64`, `slice`, `struct`). You almost always switch on `Kind`.

### All Kinds

```go
reflect.Bool
reflect.Int, Int8, Int16, Int32, Int64
reflect.Uint, Uint8, Uint16, Uint32, Uint64
reflect.Float32, Float64
reflect.Complex64, Complex128
reflect.Array, Slice
reflect.Chan
reflect.Func
reflect.Interface
reflect.Map
reflect.Pointer  // (reflect.Ptr in older Go)
reflect.String
reflect.Struct
reflect.UnsafePointer
reflect.Invalid  // zero Value
```

---

## 🏗️ Inspecting Structs

```go
type Person struct {
    Name string `json:"name" validate:"required"`
    Age  int    `json:"age"  validate:"min=0,max=150"`
    email string   // unexported — still visible via reflect
}

p := Person{Name: "Alice", Age: 30}
t := reflect.TypeOf(p)
v := reflect.ValueOf(p)

fmt.Println("Type:", t.Name())           // Person
fmt.Println("Fields:", t.NumField())     // 3

for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)                  // reflect.StructField
    value := v.Field(i)                  // reflect.Value

    fmt.Printf("Field: %-8s Type: %-8s Value: %v Exported: %v\n",
        field.Name,
        field.Type,
        value,
        field.IsExported(),
    )
    // Tags
    jsonTag := field.Tag.Get("json")
    validateTag := field.Tag.Get("validate")
    if jsonTag != "" {
        fmt.Printf("  json=%q validate=%q\n", jsonTag, validateTag)
    }
}
```

Output:
```
Type: Person
Fields: 3
Field: Name     Type: string   Value: Alice  Exported: true
  json="name" validate="required"
Field: Age      Type: int      Value: 30     Exported: true
  json="age" validate="min=0,max=150"
Field: email    Type: string   Value:        Exported: false
```

---

## ✏️ Setting Values via Reflection

To set a value via reflection, you need a **pointer** to it, then dereference with `.Elem()`:

```go
x := 42
v := reflect.ValueOf(&x).Elem()   // pointer → dereference to get settable value

fmt.Println(v.CanSet())   // true
v.SetInt(100)
fmt.Println(x)            // 100
```

### Setting struct fields

```go
p := Person{Name: "Alice", Age: 30}
v := reflect.ValueOf(&p).Elem()

nameField := v.FieldByName("Name")
if nameField.CanSet() {
    nameField.SetString("Bob")
}

ageField := v.FieldByName("Age")
if ageField.CanSet() {
    ageField.SetInt(25)
}

fmt.Println(p)   // {Bob 25 }
```

> `CanSet()` returns `false` for unexported fields — you cannot set them via reflection.

---

## 🔧 Calling Functions via Reflection

```go
func Add(a, b int) int { return a + b }

fn := reflect.ValueOf(Add)

// Build argument list
args := []reflect.Value{
    reflect.ValueOf(10),
    reflect.ValueOf(20),
}

results := fn.Call(args)
fmt.Println(results[0].Int())   // 30
```

### Calling methods on a value

```go
type Calculator struct{}

func (c Calculator) Multiply(a, b int) int { return a * b }

calc := Calculator{}
method := reflect.ValueOf(calc).MethodByName("Multiply")

results := method.Call([]reflect.Value{
    reflect.ValueOf(6),
    reflect.ValueOf(7),
})
fmt.Println(results[0].Int())   // 42
```

---

## 🗺️ Inspecting Maps and Slices

### Maps

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
v := reflect.ValueOf(m)

fmt.Println("Kind:", v.Kind())              // map
fmt.Println("Key type:", v.Type().Key())    // string
fmt.Println("Elem type:", v.Type().Elem())  // int
fmt.Println("Len:", v.Len())               // 3

for _, key := range v.MapKeys() {
    fmt.Printf("%v → %v\n", key, v.MapIndex(key))
}

// Set a map value
v.SetMapIndex(reflect.ValueOf("d"), reflect.ValueOf(4))
fmt.Println(m["d"])   // 4
```

### Slices

```go
s := []int{1, 2, 3}
v := reflect.ValueOf(s)

fmt.Println("Len:", v.Len())    // 3
fmt.Println("Cap:", v.Cap())    // 3

for i := 0; i < v.Len(); i++ {
    fmt.Println(v.Index(i).Int())
}

// Append via reflection
v = reflect.Append(v, reflect.ValueOf(4))
fmt.Println(v.Interface())   // [1 2 3 4]
```

---

## 🏭 Creating Values with `reflect.New` and `reflect.MakeSlice`

```go
// Create a new zero value of a type
t := reflect.TypeOf(Person{})
newVal := reflect.New(t)           // *Person (pointer)
p := newVal.Interface().(*Person)  // type assert back
p.Name = "created via reflect"

// Create a slice of a type
sliceType := reflect.SliceOf(t)                    // reflect.Type for []Person
slice := reflect.MakeSlice(sliceType, 0, 10)       // empty []Person cap 10
slice = reflect.Append(slice, reflect.ValueOf(Person{Name: "Alice"}))
people := slice.Interface().([]Person)
fmt.Println(people[0].Name)   // Alice

// Create a map
mapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0))
m := reflect.MakeMap(mapType)
m.SetMapIndex(reflect.ValueOf("key"), reflect.ValueOf(42))
result := m.Interface().(map[string]int)
fmt.Println(result["key"])   // 42
```

---

## 🌐 Real Use Cases

### 1. Generic deep copy

```go
func DeepCopy[T any](src T) T {
    srcVal := reflect.ValueOf(src)
    dstVal := reflect.New(srcVal.Type()).Elem()
    copyValue(dstVal, srcVal)
    return dstVal.Interface().(T)
}
```

### 2. Struct validator (like `validator` library)

```go
func Validate(v any) []string {
    var errors []string
    t := reflect.TypeOf(v)
    val := reflect.ValueOf(v)

    if t.Kind() == reflect.Pointer {
        t = t.Elem()
        val = val.Elem()
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        value := val.Field(i)
        tag := field.Tag.Get("validate")

        if tag == "required" && value.IsZero() {
            errors = append(errors, field.Name+" is required")
        }
    }
    return errors
}

type User struct {
    Name  string `validate:"required"`
    Email string `validate:"required"`
    Age   int
}

u := User{Name: "Alice"}
errs := Validate(u)
fmt.Println(errs)   // [Email is required]
```

### 3. Struct to map conversion

```go
func StructToMap(v any) map[string]any {
    result := make(map[string]any)
    t := reflect.TypeOf(v)
    val := reflect.ValueOf(v)

    if t.Kind() == reflect.Pointer {
        t = t.Elem()
        val = val.Elem()
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if !field.IsExported() { continue }

        key := field.Tag.Get("json")
        if key == "" || key == "-" { key = field.Name }
        // strip omitempty suffix
        if idx := len(key); idx > 0 {
            for j, c := range key {
                if c == ',' { key = key[:j]; break }
            }
        }
        result[key] = val.Field(i).Interface()
    }
    return result
}
```

---

## 💻 Complete Example

```go
package main

import (
    "fmt"
    "reflect"
    "strings"
)

// --- Struct tag inspector ---
type Config struct {
    Host     string `env:"HOST"     default:"localhost"`
    Port     int    `env:"PORT"     default:"8080"`
    Debug    bool   `env:"DEBUG"    default:"false"`
    Secret   string `env:"SECRET"   default:""`
}

func loadFromEnv(cfg any, lookup func(string) (string, bool)) {
    t := reflect.TypeOf(cfg).Elem()
    v := reflect.ValueOf(cfg).Elem()

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fval := v.Field(i)

        envKey := field.Tag.Get("env")
        def := field.Tag.Get("default")

        raw, ok := lookup(envKey)
        if !ok { raw = def }
        if raw == "" { continue }

        switch fval.Kind() {
        case reflect.String:
            fval.SetString(raw)
        case reflect.Int, reflect.Int64:
            var n int64
            fmt.Sscan(raw, &n)
            fval.SetInt(n)
        case reflect.Bool:
            fval.SetBool(raw == "true" || raw == "1")
        }
    }
}

// --- Type inspector ---
func Inspect(v any) {
    t := reflect.TypeOf(v)
    val := reflect.ValueOf(v)

    if t.Kind() == reflect.Pointer {
        t = t.Elem()
        val = val.Elem()
    }

    fmt.Printf("Type: %s (Kind: %s)\n", t.Name(), t.Kind())
    if t.Kind() != reflect.Struct { return }

    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        fv := val.Field(i)
        fmt.Printf("  .%-10s %-10s = %v\n", f.Name, f.Type, fv)
    }
}

func main() {
    // --- Config loader ---
    fmt.Println("=== Config Loader ===")
    fakeEnv := map[string]string{
        "HOST":  "prod-server",
        "PORT":  "9090",
        "DEBUG": "true",
    }
    cfg := &Config{}
    loadFromEnv(cfg, func(key string) (string, bool) {
        v, ok := fakeEnv[key]
        return v, ok
    })
    Inspect(cfg)

    // --- Function calling ---
    fmt.Println("\n=== Dynamic Function Call ===")
    ops := map[string]any{
        "upper": strings.ToUpper,
        "lower": strings.ToLower,
        "title": strings.Title,
    }
    for name, fn := range ops {
        result := reflect.ValueOf(fn).Call([]reflect.Value{
            reflect.ValueOf("Hello World"),
        })
        fmt.Printf("%s: %s\n", name, result[0].String())
    }

    // --- Kind switching ---
    fmt.Println("\n=== Kind Switch ===")
    values := []any{42, "hello", 3.14, true, []int{1, 2, 3}}
    for _, val := range values {
        v := reflect.ValueOf(val)
        switch v.Kind() {
        case reflect.Int:
            fmt.Printf("int:    %d\n", v.Int())
        case reflect.String:
            fmt.Printf("string: %q\n", v.String())
        case reflect.Float64:
            fmt.Printf("float:  %.2f\n", v.Float())
        case reflect.Bool:
            fmt.Printf("bool:   %v\n", v.Bool())
        case reflect.Slice:
            fmt.Printf("slice:  len=%d\n", v.Len())
        }
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Panics from wrong Kind
```go
v := reflect.ValueOf("hello")
v.Int()   // ❌ PANIC — String value, called Int()
// Always check Kind before calling type-specific methods
if v.Kind() == reflect.Int { v.Int() }
```

### 2. Setting unexported fields panics
```go
type T struct{ secret string }
v := reflect.ValueOf(&T{}).Elem().FieldByName("secret")
v.SetString("x")   // ❌ PANIC — can't set unexported field
// Always check CanSet() first
if v.CanSet() { v.SetString("x") }
```

### 3. Not dereferencing pointers before operating
```go
p := &Person{Name: "Alice"}
v := reflect.ValueOf(p)
v.FieldByName("Name")   // ❌ PANIC — p is a pointer, not a struct

// ✅ Dereference first
v = reflect.ValueOf(p).Elem()
v.FieldByName("Name")   // works
```

### 4. Forgetting to get `Interface()` to extract value
```go
v := reflect.ValueOf(42)
result := v + 1   // ❌ compile error — v is reflect.Value, not int
result := v.Int() + 1   // ✅ extract as int64 first
original := v.Interface().(int)   // ✅ get original type back
```

### 5. Using reflection where generics or interfaces suffice
```go
// ❌ Reflection for something generics handle better
func Sum(nums any) float64 {
    v := reflect.ValueOf(nums)
    // ... painful reflection code
}

// ✅ Use generics
func Sum[T Number](nums []T) T { ... }
```

---

## 🎯 Must-Know Things (Quick Reference)

- `reflect.TypeOf(v)` → `Type`; `reflect.ValueOf(v)` → `Value`
- `Type.Name()` = named type (`Person`); `Type.Kind()` = underlying kind (`struct`)
- Always **switch on Kind**, not Type, for generic code
- To set a value: need a pointer → `reflect.ValueOf(&x).Elem()` → `.CanSet()`
- Unexported fields: visible via reflect, but **`CanSet()` = false** — cannot set
- `reflect.New(t)` — creates `*T`; `.Elem()` to get the `T`
- `v.Interface()` — converts `reflect.Value` back to `interface{}`
- Panics from: wrong Kind method call, setting non-settable value, operating on pointer without `.Elem()`
- Use reflection for: encoding libs, ORMs, DI frameworks, struct validators
- **Never use reflection in hot paths** — it's 10-100x slower than direct code
- When generics or interfaces can solve the problem, use those instead

---

## ❓ Interview Questions & Answers

---

**Q1: What is reflection in Go and when should you use it?**

> Reflection allows a program to inspect and manipulate types and values at runtime without knowing them at compile time. Go's `reflect` package provides this through `reflect.Type` (type metadata) and `reflect.Value` (value manipulation). Use reflection when you genuinely can't know types at compile time — encoding/decoding libraries like `encoding/json`, ORMs mapping structs to DB rows, dependency injection containers, struct validators, and test assertion libraries. Most application code should never use reflection — generics, interfaces, and type switches handle the vast majority of cases more safely and efficiently.

---

**Q2: What is the difference between `Type` and `Kind` in reflect?**

> `Type` is the specific named type — `Person`, `[]int`, `map[string]int`. `Kind` is the underlying category — `struct`, `slice`, `map`. A `type Celsius float64` has Type `Celsius` but Kind `float64`. When writing generic reflection code, you almost always switch on `Kind` — because it tells you how to access the value (`.Float()`, `.Int()`, `.Field(i)`, etc). You use `Type` when you need the actual named type, method set, or struct field metadata.

---

**Q3: How do you set a value using reflection and what can go wrong?**

> To set a value via reflection, you must pass a pointer: `reflect.ValueOf(&x).Elem()`. The `.Elem()` dereferences the pointer to get a settable `Value`. Before setting, always call `CanSet()` — it returns false for unexported struct fields (panics if you try) and for values not obtained through a pointer. Common panics: calling a wrong-Kind method (`v.Int()` on a string), setting an unexported field, or forgetting to call `.Elem()` on a pointer value before accessing fields.

---

**Q4: How does `encoding/json` use reflection internally?**

> `encoding/json` uses reflection to inspect struct fields at runtime: it calls `reflect.TypeOf` to get the struct type, iterates `NumField()` / `Field(i)` to examine each field's name, type, and `json` struct tag. For marshaling, it reads field values with `reflect.ValueOf(v).Field(i).Interface()`. For unmarshaling, it creates new values with `reflect.New` and sets fields with `SetString`, `SetInt`, etc. This is why JSON works with any struct without code generation — but also why it's slower than hand-written or generated serializers.

---

**Q5: What are the performance implications of reflection?**

> Reflection is significantly slower than direct code — typically 10-100x for type inspection and value access, because all type information that the compiler knows statically must be looked up dynamically at runtime through interface dispatch and memory indirection. Allocation rates are also higher. This matters for hot paths — serializing millions of small requests per second. For non-hot paths (startup configuration, one-time setup, test utilities), reflection overhead is usually acceptable. Libraries that need performance often use reflection during initialization and cache the results, or use `go generate` / code generation to produce direct code.

---

**Q6: Why is `CanSet()` false for unexported struct fields?**

> Go's encapsulation rules apply at the package level — unexported fields are private to their package. Reflection respects this: even though you can *see* an unexported field's value via reflection (useful for debugging), you cannot *set* it. This prevents external packages from bypassing Go's access control via reflection. `CanSet()` returns false when a value is not addressable or when it's an unexported field. The `unsafe` package can bypass this (Topic 3.5), but doing so is considered unsafe and non-portable.

---

*Next up: **Topic 3.5 — `unsafe` Package** 🚀*
