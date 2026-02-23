# 1.2 — Your First Go Program

> **Phase:** 1 — Foundations | **Difficulty:** Beginner  
> **Prerequisites:** Topic 1.1 — Introduction to Go

---

## 📖 Anatomy of a Go Program

Let's start with the classic Hello World and dissect every single part:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

Run it:
```bash
go run main.go
# Output: Hello, World!
```

Simple — but every word here has a specific meaning. Let's break it down.

---

## 🧩 Part 1: `package main`

```go
package main
```

- Every Go file **must** start with a `package` declaration.
- A **package** is a collection of Go files in the same directory that share the same package name.
- `package main` is **special** — it tells the Go compiler: *"this is an executable program, not a library."*
- The `main` package is the **entry point** of any Go application.

> If you're writing a reusable library, you'd use something like `package utils` or `package math` instead.

---

## 🧩 Part 2: `import "fmt"`

```go
import "fmt"
```

- `import` brings in packages so you can use their functionality.
- `"fmt"` is a package from Go's **standard library** — it handles formatted I/O (print, scan, format strings).
- The string `"fmt"` is the **import path** — for standard library packages it's just the name; for third-party packages it's a full URL like `"github.com/gin-gonic/gin"`.

### Importing multiple packages

```go
// Style 1: multiple import statements (uncommon)
import "fmt"
import "math"

// Style 2: grouped import — the idiomatic Go way ✅
import (
    "fmt"
    "math"
    "strings"
)
```

### The blank identifier `_` for imports

Sometimes you import a package **only for its side effects** (e.g., registering a database driver) and don't call any of its functions directly. Use `_`:

```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // registers the postgres driver, never called directly
)
```

### Aliasing imports

```go
import (
    "fmt"
    f "fmt"          // alias — now use f.Println()
    mth "math"       // alias for math
)
```

> ⚠️ Remember: **unused imports are a compile error in Go.**

---

## 🧩 Part 3: `func main()`

```go
func main() {
    // code here
}
```

- `func` is the keyword to declare a function.
- `main` is the **entry point** — when you run a Go program, execution starts here.
- There can only be **one** `main()` function in `package main`.
- It takes **no arguments** and returns **nothing**.
- Command-line arguments are accessed via `os.Args`, not via `main()` parameters (unlike C).

---

## 🧩 Part 4: `fmt.Println`

```go
fmt.Println("Hello, World!")
```

- `fmt` — the package name
- `.` — the dot accesses an exported name from that package
- `Println` — a function in the `fmt` package (capital `P` = exported/public)
- `"Hello, World!"` — a string argument

### Common `fmt` functions

```go
fmt.Println("Hello")           // prints with newline
fmt.Print("Hello")             // prints without newline
fmt.Printf("Name: %s\n", name) // formatted print (like C's printf)
fmt.Sprintf("Age: %d", 25)     // returns formatted string (doesn't print)
fmt.Errorf("error: %w", err)   // creates formatted error
```

### Format verbs (used with Printf/Sprintf)

| Verb | Meaning | Example |
|------|---------|---------|
| `%v` | Default format (any type) | `fmt.Printf("%v", 42)` → `42` |
| `%+v` | Struct with field names | `fmt.Printf("%+v", p)` → `{Name:Alice Age:30}` |
| `%T` | Type of value | `fmt.Printf("%T", 3.14)` → `float64` |
| `%d` | Integer (decimal) | `fmt.Printf("%d", 42)` → `42` |
| `%f` | Float | `fmt.Printf("%.2f", 3.14159)` → `3.14` |
| `%s` | String | `fmt.Printf("%s", "hi")` → `hi` |
| `%q` | Quoted string | `fmt.Printf("%q", "hi")` → `"hi"` |
| `%b` | Binary | `fmt.Printf("%b", 5)` → `101` |
| `%x` | Hex | `fmt.Printf("%x", 255)` → `ff` |
| `%p` | Pointer address | `fmt.Printf("%p", &x)` → `0xc0000b4008` |

---

## 🔨 `go run` vs `go build` vs `go install`

These are three different ways to execute Go code. Understanding the difference is essential.

### `go run`

```bash
go run main.go
go run .         # run the package in the current directory
```

- Compiles and runs the program **in one step**.
- The binary is **temporary** — stored in a temp directory, deleted after execution.
- Great for **development and quick testing**.
- Slightly slower because it compiles every time.

### `go build`

```bash
go build main.go      # produces binary named "main" (or "main.exe" on Windows)
go build -o app .     # produces binary named "app"
go build ./...        # build all packages
```

- Compiles and produces a **permanent binary file** in the current directory.
- The binary is **self-contained** — you can copy it anywhere and run it.
- No Go installation needed on the target machine.
- Use this for **production builds**.

```bash
go build -o myapp .
./myapp               # run the binary
```

### `go install`

```bash
go install github.com/yourname/myapp@latest
```

- Compiles and places the binary in `$GOPATH/bin` (or `$GOBIN`).
- Used for **installing Go tools** that you want available system-wide.
- Example: `go install golang.org/x/tools/gopls@latest` installs the Go language server.

### Cross Compilation (Bonus — very powerful!)

Go can compile for **any OS/architecture** from any machine:

```bash
# Build a Linux binary from macOS
GOOS=linux GOARCH=amd64 go build -o app-linux .

# Build a Windows binary from macOS
GOOS=windows GOARCH=amd64 go build -o app.exe .

# Build for ARM (e.g., Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o app-arm .
```

> This is one of Go's superpowers — no special toolchain needed for cross-compilation.

---

## 📁 Go Project Structure

### Minimal single-file project

```
hello-go/
├── go.mod
└── main.go
```

### Typical small project

```
my-app/
├── go.mod
├── go.sum
├── main.go
├── README.md
└── internal/
    └── utils/
        └── utils.go
```

### `go.mod` file explained

```
module github.com/yourname/my-app   ← module name (import path prefix)

go 1.22                              ← minimum Go version required

require (
    github.com/gin-gonic/gin v1.9.1  ← dependencies
)
```

---

## 📝 Writing a Slightly More Complete Program

Let's write a program that uses multiple packages and shows some real structure:

```go
package main

import (
    "fmt"
    "math"
    "os"
    "strings"
)

func greet(name string) string {
    return fmt.Sprintf("Hello, %s! Welcome to Go.", name)
}

func circleArea(radius float64) float64 {
    return math.Pi * radius * radius
}

func main() {
    // Basic output
    fmt.Println("=== My First Go Program ===")

    // Using a custom function
    message := greet("Gopher")
    fmt.Println(message)

    // String manipulation
    upper := strings.ToUpper("go is awesome")
    fmt.Println(upper)

    // Math
    area := circleArea(5.0)
    fmt.Printf("Circle area (r=5): %.2f\n", area)

    // Getting command-line arguments
    fmt.Println("Program name:", os.Args[0])
    if len(os.Args) > 1 {
        fmt.Println("Arguments:", os.Args[1:])
    }
}
```

```bash
go run main.go hello world
# Output:
# === My First Go Program ===
# Hello, Gopher! Welcome to Go.
# GO IS AWESOME
# Circle area (r=5): 78.54
# Program name: /tmp/go-build.../main
# Arguments: [hello world]
```

---

## 🔑 Exported vs Unexported Names

This is a fundamental Go concept — **capitalization determines visibility**:

```go
// Exported (public) — starts with uppercase
fmt.Println    // ✅ accessible from outside the package
math.Pi        // ✅ accessible from outside the package

// Unexported (private) — starts with lowercase
// Only accessible within the same package
myPrivateFunc()   // only callable within this package
```

> Go has **no** `public`, `private`, or `protected` keywords. Capitalization IS the access modifier.

---

## 🔧 Essential Go Tools You Should Know

```bash
# Format code (always run this before committing)
gofmt -w main.go        # format and write to file
go fmt ./...            # format entire project

# Static analysis (catches bugs gofmt won't)
go vet ./...

# Download dependencies
go mod tidy

# View documentation
go doc fmt
go doc fmt.Println

# Run tests
go test ./...

# Check for race conditions
go run -race main.go
go test -race ./...
```

---

## ⚠️ Common Mistakes & Gotchas

### 1. Brace on wrong line
```go
// ❌ Compile error — Go's parser inserts a semicolon after ")"
func main()
{
    fmt.Println("Hi")
}

// ✅ Correct
func main() {
    fmt.Println("Hi")
}
```

### 2. Unused import
```go
// ❌ Compile error
import "fmt"
import "math"  // if math is never used — compile error!

// ✅ Only import what you use
import "fmt"
```

### 3. Using := outside a function
```go
// ❌ Compile error — := is only valid inside functions
package main
x := 5  // syntax error!

// ✅ Use var at package level
var x = 5
```

### 4. Forgetting `go.mod`
```bash
# ❌ Running without initializing the module
go run main.go  # might fail if you use third-party packages

# ✅ Always initialize first
go mod init myapp
go run main.go
```

---

## 🎯 Must-Know Things Summary

- Every Go file starts with `package <name>`
- `package main` + `func main()` = executable program
- Imports use the full path string — `"fmt"`, `"github.com/..."` etc.
- Unused imports = **compile error**
- `go run` → compile + run temporarily
- `go build` → compile to permanent binary
- `go install` → compile + put in `$GOPATH/bin`
- **Uppercase = exported (public), lowercase = unexported (private)**
- Go produces **single self-contained binaries** — no runtime needed
- Cross-compilation is built-in via `GOOS` and `GOARCH`

---

## ❓ Interview Questions & Answers

---

**Q1: What is `package main` and why is it special?**

> `package main` is the special package that defines an executable program in Go. Every Go application that should be run directly must have a `package main` declaration and a `func main()` function. Other packages (libraries) use different package names and cannot be executed directly — they're meant to be imported.

---

**Q2: What is the difference between `go run`, `go build`, and `go install`?**

> - `go run` compiles and executes a program in one step, producing a temporary binary that's deleted after execution. Best for development.
> - `go build` compiles the program and creates a permanent binary in the current directory. The binary is self-contained and doesn't need Go installed to run. Best for production.
> - `go install` compiles and places the binary in `$GOPATH/bin`, making it available as a system-wide command. Best for tools you want to use globally.

---

**Q3: How does Go handle visibility (public vs private)?**

> Go uses **capitalization** to determine visibility. Any identifier (function, type, variable, constant) that starts with an uppercase letter is **exported** (public) and accessible from other packages. Anything starting with a lowercase letter is **unexported** (private) and only accessible within the same package. There are no `public`, `private`, or `protected` keywords in Go.

---

**Q4: What is `fmt.Println` vs `fmt.Printf` vs `fmt.Sprintf`?**

> - `fmt.Println` prints values with default formatting, adds spaces between arguments, and appends a newline. Simple and convenient.
> - `fmt.Printf` prints with a format string using verbs like `%s`, `%d`, `%f`. Like C's `printf`. Does NOT add a newline automatically.
> - `fmt.Sprintf` works like `Printf` but returns a formatted string instead of printing it. Use it to build strings.

---

**Q5: What happens if you have an unused import in Go?**

> The program will **fail to compile**. Go treats unused imports as a hard error, not a warning. This is a deliberate design choice to keep code clean, reduce compilation overhead, and avoid confusion from dead code. If you need an import only for side effects, use the blank identifier: `import _ "package/path"`.

---

**Q6: What is cross-compilation in Go and how do you do it?**

> Cross-compilation means compiling a binary for a different operating system or architecture than the one you're currently on. Go supports this out of the box using environment variables. For example, to build a Linux binary on macOS:
> ```bash
> GOOS=linux GOARCH=amd64 go build -o myapp .
> ```
> No special toolchain is required. This is possible because Go compiles to machine code directly and bundles its own runtime.

---

**Q7: What does `go mod init` do?**

> `go mod init <module-name>` initializes a new Go module by creating a `go.mod` file. This file defines the module's name (import path) and tracks all dependencies with their versions. It's the equivalent of `npm init` in Node.js or `pip`'s `requirements.txt`. The module name is typically a URL like `github.com/username/project`.

---

**Q8: Can a Go program have multiple `main()` functions?**

> No. A Go program can only have one `main()` function within `package main`. Having two `main()` functions in the same package causes a compile error. However, different packages can each have their own `main` package and `main()` function — they'd just be separate executable programs.

---

**Q9: What is `go vet` and how is it different from `gofmt`?**

> `gofmt` is a code **formatter** — it fixes whitespace, indentation, and style automatically. It doesn't care about logic.
> `go vet` is a **static analyzer** — it looks for code that is syntactically valid but likely incorrect, such as wrong format verbs in `Printf` calls, unreachable code, or suspicious struct tags. Think of `gofmt` as a stylist and `go vet` as a code reviewer.

---

**Q10: How does Go's self-contained binary work?**

> When you run `go build`, Go compiles everything — your code, all dependencies, and the Go runtime — into a **single statically-linked binary**. This binary has no external dependencies and can run on any machine with a compatible OS and architecture, even if Go is not installed. This makes deployment extremely simple compared to languages like Python (which need the interpreter) or Java (which needs the JVM).

---

*Next up: **Topic 1.3 — Variables & Data Types** 🚀*
