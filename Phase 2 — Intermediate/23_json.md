# 2.11 — Working with JSON

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.10 — File I/O

---

## 📖 Overview

Go's `encoding/json` package handles JSON serialization and deserialization cleanly. The two core operations are:

- **Marshal** — Go value → JSON bytes
- **Unmarshal** — JSON bytes → Go value

JSON is the lingua franca of web APIs — this topic is essential for any backend Go work.

---

## 📤 Marshal — Go to JSON

```go
import "encoding/json"

type Person struct {
    Name string
    Age  int
    Email string
}

p := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}

data, err := json.Marshal(p)
if err != nil {
    return fmt.Errorf("marshal: %w", err)
}
fmt.Println(string(data))
// {"Name":"Alice","Age":30,"Email":"alice@example.com"}
```

### Pretty printing with `json.MarshalIndent`

```go
data, err := json.MarshalIndent(p, "", "  ")
// {
//   "Name": "Alice",
//   "Age": 30,
//   "Email": "alice@example.com"
// }
```

---

## 📥 Unmarshal — JSON to Go

```go
jsonData := []byte(`{"Name":"Bob","Age":25,"Email":"bob@example.com"}`)

var p Person
if err := json.Unmarshal(jsonData, &p); err != nil {
    return fmt.Errorf("unmarshal: %w", err)
}
fmt.Println(p.Name)    // Bob
fmt.Println(p.Age)     // 25
```

> Always pass a **pointer** to `Unmarshal` — it needs to write into your variable.

---

## 🏷️ Struct Tags — Controlling JSON Output

Struct tags control the field names in JSON and other behavior:

```go
type User struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Password  string `json:"-"`              // always omit
    Age       int    `json:"age,omitempty"` // omit if zero value
    Bio       string `json:"bio,omitempty"` // omit if empty string
}

u := User{ID: 1, FirstName: "Alice", LastName: "Smith", Password: "secret", Age: 0}
data, _ := json.Marshal(u)
fmt.Println(string(data))
// {"id":1,"first_name":"Alice","last_name":"Smith"}
// Password omitted (json:"-"), Age omitted (omitempty + zero)
```

### Tag options

| Tag | Effect |
|-----|--------|
| `json:"name"` | Use "name" as the JSON key |
| `json:"-"` | Always omit this field |
| `json:"name,omitempty"` | Omit if zero value (0, "", false, nil, empty slice/map) |
| `json:",omitempty"` | Keep original field name, omit if zero |
| `json:"name,string"` | Encode as a JSON string (for numbers) |

---

## 🔄 Marshal/Unmarshal — Data Type Mapping

| Go Type | JSON Type |
|---------|-----------|
| `bool` | `true` / `false` |
| `int`, `float64` | number |
| `string` | string |
| `[]T` | array |
| `map[string]T` | object |
| `struct` | object |
| `nil` pointer/slice/map | `null` |
| `[]byte` | base64-encoded string |

---

## 🌊 Streaming — Encoder and Decoder

For large JSON or streaming (HTTP responses, files), use `json.Encoder` and `json.Decoder` — they work directly with `io.Writer` and `io.Reader`:

```go
// Encoder — write JSON to any Writer
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")   // pretty print
err := encoder.Encode(p)      // writes JSON + newline to stdout
```

```go
// Decoder — read JSON from any Reader
f, _ := os.Open("data.json")
defer f.Close()

decoder := json.NewDecoder(f)
var p Person
err := decoder.Decode(&p)
```

### Decode multiple JSON objects (NDJSON / JSON lines)

```go
decoder := json.NewDecoder(r)
for {
    var p Person
    err := decoder.Decode(&p)
    if err == io.EOF {
        break
    }
    if err != nil {
        return fmt.Errorf("decode: %w", err)
    }
    process(p)
}
```

---

## 🗺️ Working with Unknown JSON — `map[string]any`

When you don't know the JSON structure ahead of time:

```go
jsonData := []byte(`{
    "name": "Alice",
    "age": 30,
    "scores": [95, 87, 92],
    "address": {"city": "NYC", "zip": "10001"}
}`)

var result map[string]any
json.Unmarshal(jsonData, &result)

fmt.Println(result["name"])              // Alice
fmt.Println(result["age"].(float64))    // 30 — numbers are float64!
scores := result["scores"].([]any)
fmt.Println(scores[0].(float64))        // 95
```

> JSON numbers always unmarshal to `float64` when decoded into `any` — not `int`. Always type-assert carefully.

---

## 🔢 Number Handling — `json.Number`

Use `json.Number` when you need to preserve number types exactly:

```go
jsonData := []byte(`{"id": 9007199254740993}`)   // too large for float64!

var result map[string]json.Number
json.Unmarshal(jsonData, &result)

id, _ := result["id"].Int64()   // 9007199254740993 — exact!
fmt.Println(id)
```

Or configure a decoder to use `json.Number`:

```go
decoder := json.NewDecoder(r)
decoder.UseNumber()   // numbers decode as json.Number instead of float64

var result map[string]any
decoder.Decode(&result)

id := result["id"].(json.Number)
n, _ := id.Int64()
```

---

## 🔧 Custom Marshal/Unmarshal

Implement `json.Marshaler` and `json.Unmarshaler` for custom behavior:

```go
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}

type Unmarshaler interface {
    UnmarshalJSON([]byte) error
}
```

### Example — custom time format

```go
type DateTime struct {
    time.Time
}

const dtFormat = "2006-01-02 15:04:05"

func (dt DateTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(dt.Time.Format(dtFormat))
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    t, err := time.Parse(dtFormat, s)
    if err != nil {
        return fmt.Errorf("DateTime.UnmarshalJSON: %w", err)
    }
    dt.Time = t
    return nil
}

type Event struct {
    Name      string   `json:"name"`
    CreatedAt DateTime `json:"created_at"`
}
```

### Example — encode int as string

```go
type StringInt int64

func (s StringInt) MarshalJSON() ([]byte, error) {
    return json.Marshal(fmt.Sprintf("%d", int64(s)))
}

func (s *StringInt) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    n, err := strconv.ParseInt(str, 10, 64)
    if err != nil {
        return err
    }
    *s = StringInt(n)
    return nil
}
```

---

## 🔍 Partial Unmarshal with `json.RawMessage`

`json.RawMessage` defers decoding of a field — keeps it as raw JSON bytes:

```go
type Response struct {
    Status string          `json:"status"`
    Data   json.RawMessage `json:"data"`   // decode later
}

var resp Response
json.Unmarshal(jsonData, &resp)

// Now decode Data based on Status
switch resp.Status {
case "user":
    var u User
    json.Unmarshal(resp.Data, &u)
case "product":
    var p Product
    json.Unmarshal(resp.Data, &p)
}
```

---

## 🌐 JSON in HTTP — The Most Common Use

```go
// Write JSON response
func writeJSON(w http.ResponseWriter, status int, v any) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(v)
}

// Read JSON request body
func readJSON(r *http.Request, v any) error {
    defer r.Body.Close()
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()   // strict mode — error on unknown fields
    return decoder.Decode(v)
}

// Handler example
func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
        return
    }
    user, err := createUser(req)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "internal server error",
        })
        return
    }
    writeJSON(w, http.StatusCreated, user)
}
```

---

## 💻 Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
)

type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
    Zip    string `json:"zip,omitempty"`
}

type Person struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    Age     int     `json:"age,omitempty"`
    Email   string  `json:"email,omitempty"`
    Secret  string  `json:"-"`
    Address Address `json:"address"`
    Tags    []string `json:"tags,omitempty"`
}

func main() {
    // --- Marshal ---
    fmt.Println("=== Marshal ===")
    p := Person{
        ID:    1,
        Name:  "Alice",
        Age:   30,
        Email: "alice@example.com",
        Secret: "topsecret",
        Address: Address{Street: "123 Main St", City: "NYC", Zip: "10001"},
        Tags:  []string{"admin", "user"},
    }

    data, err := json.MarshalIndent(p, "", "  ")
    if err != nil {
        fmt.Println("error:", err)
        return
    }
    fmt.Println(string(data))

    // --- Unmarshal ---
    fmt.Println("\n=== Unmarshal ===")
    jsonStr := `{"id":2,"name":"Bob","address":{"street":"456 Oak Ave","city":"LA"}}`
    var p2 Person
    if err := json.Unmarshal([]byte(jsonStr), &p2); err != nil {
        fmt.Println("error:", err)
        return
    }
    fmt.Printf("Name: %s, City: %s\n", p2.Name, p2.Address.City)
    fmt.Printf("Age (omitted→zero): %d\n", p2.Age)

    // --- Streaming encoder ---
    fmt.Println("\n=== Streaming Encoder ===")
    people := []Person{
        {ID: 1, Name: "Alice", Age: 30},
        {ID: 2, Name: "Bob", Age: 25},
    }
    var sb strings.Builder
    enc := json.NewEncoder(&sb)
    for _, person := range people {
        enc.Encode(person)   // each on its own line (NDJSON)
    }
    fmt.Print(sb.String())

    // --- Streaming decoder ---
    fmt.Println("=== Streaming Decoder ===")
    ndjson := `{"id":1,"name":"Carol","address":{"city":"Chicago"}}
{"id":2,"name":"Dave","address":{"city":"Boston"}}`

    dec := json.NewDecoder(strings.NewReader(ndjson))
    for {
        var person Person
        if err := dec.Decode(&person); err != nil {
            break   // io.EOF
        }
        fmt.Printf("Decoded: %s from %s\n", person.Name, person.Address.City)
    }

    // --- map[string]any ---
    fmt.Println("\n=== Dynamic JSON ===")
    raw := `{"event":"click","x":100,"y":200,"meta":{"browser":"chrome"}}`
    var event map[string]any
    json.Unmarshal([]byte(raw), &event)
    fmt.Println("event:", event["event"])
    fmt.Printf("x: %.0f\n", event["x"].(float64))   // numbers are float64!

    // --- RawMessage ---
    fmt.Println("\n=== RawMessage ===")
    type Envelope struct {
        Type    string          `json:"type"`
        Payload json.RawMessage `json:"payload"`
    }
    envJSON := `{"type":"person","payload":{"id":99,"name":"Eve","address":{"city":"Miami"}}}`
    var env Envelope
    json.Unmarshal([]byte(envJSON), &env)
    fmt.Println("type:", env.Type)
    var pp Person
    json.Unmarshal(env.Payload, &pp)
    fmt.Printf("payload name: %s\n", pp.Name)
}
```

Output:
```json
=== Marshal ===
{
  "id": 1,
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com",
  "address": {
    "street": "123 Main St",
    "city": "NYC",
    "zip": "10001"
  },
  "tags": ["admin","user"]
}

=== Unmarshal ===
Name: Bob, City: LA
Age (omitted→zero): 0

=== Streaming Encoder ===
{"id":1,"name":"Alice","age":30,"address":{"street":"","city":""}}
{"id":2,"name":"Bob","age":25,"address":{"street":"","city":""}}

=== Streaming Decoder ===
Decoded: Carol from Chicago
Decoded: Dave from Boston

=== Dynamic JSON ===
event: click
x: 100

=== RawMessage ===
type: person
payload name: Eve
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Unexported fields are silently ignored
```go
type Person struct {
    name string   // ❌ lowercase — never marshaled or unmarshaled
    Age  int
}
p := Person{Age: 30}
json.Unmarshal(`{"name":"Alice","Age":30}`, &p)
fmt.Println(p.name)   // "" — silently ignored!
```

### 2. Numbers in `any` are `float64`, not `int`
```go
var m map[string]any
json.Unmarshal(`{"count":42}`, &m)
n := m["count"].(int)       // ❌ PANIC — it's float64!
n := m["count"].(float64)   // ✅
// Or use decoder.UseNumber()
```

### 3. nil slice vs empty slice in JSON
```go
type T struct{ Items []string }

t1 := T{Items: nil}
t2 := T{Items: []string{}}

json.Marshal(t1)   // {"Items":null}
json.Marshal(t2)   // {"Items":[]}
// If API clients care, initialize to []string{} not nil
```

### 4. Forgetting pointer in Unmarshal
```go
var p Person
json.Unmarshal(data, p)    // ❌ compiles but does nothing — not a pointer
json.Unmarshal(data, &p)   // ✅
```

### 5. DisallowUnknownFields for strict parsing
```go
dec := json.NewDecoder(r)
// Without this, unknown JSON fields are silently ignored
dec.DisallowUnknownFields()   // ✅ returns error for unexpected fields
```

### 6. Large integers lose precision as float64
```go
// JavaScript max safe int: 9007199254740991
// json.Number or string encoding needed for larger values
type BigID struct {
    ID int64 `json:"id,string"`   // encode as string to preserve precision
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- `json.Marshal(v)` → `[]byte` | `json.Unmarshal(data, &v)` → fills v
- Always pass a **pointer** to `Unmarshal`
- `json:"name"` — custom key name | `json:"-"` — always omit | `json:"name,omitempty"` — omit if zero
- Unexported fields are **silently ignored** — always export JSON fields
- `json.NewEncoder(w).Encode(v)` — stream to Writer (HTTP, files)
- `json.NewDecoder(r).Decode(&v)` — stream from Reader (HTTP, files)
- JSON numbers unmarshal to **`float64`** when target is `any` — not `int`
- nil slice → `null` in JSON | empty slice → `[]` — matters for API contracts
- `json.RawMessage` — defer decoding a field, keep as raw JSON
- `decoder.DisallowUnknownFields()` — strict mode, error on extra fields
- `decoder.UseNumber()` — numbers as `json.Number` not `float64`
- Implement `MarshalJSON()` / `UnmarshalJSON()` for custom serialization

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between `json.Marshal`/`Unmarshal` and `json.Encoder`/`Decoder`?**

> `json.Marshal` and `json.Unmarshal` work with `[]byte` in memory — marshal produces a byte slice, unmarshal reads from a byte slice. `json.Encoder` and `json.Decoder` work with `io.Writer` and `io.Reader` directly — encoding writes JSON to any writer (HTTP response, file, buffer), decoding reads from any reader (HTTP request body, file). For HTTP handlers and file I/O, use Encoder/Decoder — they avoid loading everything into memory. For small in-memory operations, Marshal/Unmarshal are simpler.

---

**Q2: What do the `json` struct tags `"-"`, `"omitempty"`, and `"string"` do?**

> `json:"-"` always excludes the field from JSON output and ignores it during input — used for sensitive fields like passwords. `json:"name,omitempty"` omits the field when it holds its zero value (0, "", false, nil, empty slice/map) — useful for optional fields in API responses. `json:"id,string"` encodes a numeric field as a JSON string — important for large int64 IDs that would lose precision as JavaScript floats. All three appear after the field name in the tag, separated by commas.

---

**Q3: Why do JSON numbers unmarshal to `float64` when decoding into `any`?**

> The JSON specification doesn't distinguish between integers and floats — all numbers are the same type. Go's `encoding/json` maps JSON numbers to `float64` when the target is `interface{}`/`any` because `float64` can represent all JSON numbers (though with precision limits for very large integers). To get exact integer values, either unmarshal into a typed struct with `int` fields, use `json.Number` as the target type, or call `decoder.UseNumber()` to decode numbers as `json.Number` strings that you convert explicitly.

---

**Q4: What is `json.RawMessage` and when would you use it?**

> `json.RawMessage` is `[]byte` with custom marshal/unmarshal methods that preserve the raw JSON. When a struct field is typed as `json.RawMessage`, it captures the raw JSON bytes without parsing them. This is useful for: (1) polymorphic payloads where the structure depends on another field's value; (2) two-phase decoding — decode the envelope first, then decode the payload based on type; (3) forwarding JSON without re-encoding. It's the idiomatic Go pattern for handling dynamic or discriminated union JSON schemas.

---

**Q5: What is the nil slice vs empty slice gotcha in JSON?**

> In Go, a nil slice and an empty slice are different values, and `encoding/json` treats them differently: a nil slice marshals to JSON `null`, while an empty slice marshals to `[]`. This matters for API contracts — many clients treat `null` and `[]` differently. If your API should always return an array (never null), initialize slices to `[]T{}` or `make([]T, 0)` instead of leaving them nil. The `omitempty` tag omits nil slices entirely, which is a third behavior.

---

**Q6: How do you implement custom JSON marshaling?**

> Implement `MarshalJSON() ([]byte, error)` and `UnmarshalJSON([]byte) error` on your type. The `MarshalJSON` method controls exactly what bytes represent the value in JSON — you can produce any valid JSON. `UnmarshalJSON` receives the raw JSON bytes for the value and populates the receiver. Common uses: custom date formats (Go's `time.Time` uses RFC3339 by default), encoding large ints as strings for JavaScript compatibility, computed fields, and type-safe enum serialization.

---

**Q7: What does `decoder.DisallowUnknownFields()` do and why use it?**

> By default, `json.Decoder` silently ignores JSON fields that have no corresponding struct field — unknown fields are discarded. `DisallowUnknownFields()` makes the decoder return an error if any JSON key doesn't match a struct field. This is useful for strict API validation — catching typos in field names, detecting when a client sends unexpected fields, or ensuring backward compatibility is maintained. For public APIs, the default (lenient) behavior is often better — it allows the API to add fields without breaking old clients.

---

## 🎉 Phase 2 Complete!

You've finished all 11 topics in **Phase 2 — Intermediate**. Here's what you now know:

| Topic | Key Takeaway |
|-------|-------------|
| 2.1 Interfaces | Implicit satisfaction, type assertions, empty interface |
| 2.2 Error Handling | error interface, wrapping with %w, errors.Is/As |
| 2.3 Packages & Modules | go.mod, internal/, init(), blank imports |
| 2.4 Goroutines | go keyword, M:N scheduler, leaks, race detector |
| 2.5 Channels | Unbuffered/buffered, directional, close rules, pipelines |
| 2.6 Select | Multiplexing, timeout, non-blocking, nil disables case |
| 2.7 sync Package | WaitGroup, Mutex, RWMutex, Once, Pool, atomic |
| 2.8 Defer/Panic/Recover | Cleanup, LIFO, recover in defer, Must pattern |
| 2.9 Stringer & fmt | String(), format verbs, infinite recursion pitfall |
| 2.10 File I/O | os/io/bufio, Reader/Writer interfaces, streaming |
| 2.11 JSON | Marshal/Unmarshal, struct tags, streaming, RawMessage |

**Next: Phase 3 — Advanced** starts with Generics, Context, and advanced concurrency patterns! 🚀
