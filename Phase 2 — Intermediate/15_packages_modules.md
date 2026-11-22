# 2.3 — Packages & Modules

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.2 — Error Handling

---

## 📖 Overview

Go organizes code into **packages** (the unit of code organization) and **modules** (the unit of dependency management). Understanding both is essential for working on any real Go project.

- **Package** — a directory of `.go` files that share the same `package` declaration
- **Module** — a collection of packages with a `go.mod` file at the root, versioned together

---

## 📦 Packages

### Every file belongs to a package

```go
package main      // executable entry point
package utils     // reusable library package
package database  // another library package
```

Every `.go` file must start with a `package` declaration. All files in the same directory must have the same package name (except `_test.go` files which can use `package foo_test`).

### Package naming conventions

```go
// ✅ Short, lowercase, no underscores, no mixedCaps
package http
package json
package sync
package strconv

// ❌ Avoid
package httpUtils    // no camelCase
package http_utils   // no underscores
package HTTPUtils    // no all-caps
```

The package name is what callers use — `json.Marshal`, `http.Get`. Make it meaningful and concise.

---

## 🔑 Exported vs Unexported Identifiers

Capitalization controls visibility:

```go
package mypackage

// Exported — accessible from other packages
var PublicVar = "visible everywhere"
const PublicConst = 42
type PublicType struct{ Field string }
func PublicFunc() {}

// Unexported — only accessible within this package
var privateVar = "only in mypackage"
const privateConst = 99
type privateType struct{ field string }
func privateFunc() {}
```

### Exported fields in unexported types

```go
// The type itself is unexported, but exported fields are still accessible
type config struct {
    Host string   // exported field — accessible outside package
    port int      // unexported field — NOT accessible outside package
}
```

---

## 🗂️ Package Organization

### Single package (small projects)

```
myapp/
├── go.mod
├── main.go
├── handlers.go
├── models.go
└── utils.go
```

### Multiple packages (real projects)

```
myapp/
├── go.mod
├── main.go                    ← package main
├── internal/
│   ├── handler/
│   │   └── handler.go         ← package handler
│   ├── service/
│   │   └── service.go         ← package service
│   └── repository/
│       └── repository.go      ← package repository
├── pkg/
│   └── validator/
│       └── validator.go       ← package validator (public reusable)
└── config/
    └── config.go              ← package config
```

### The `internal` directory — Go's access control

Packages inside an `internal` directory can only be imported by code in the **parent of `internal`** and its subtrees:

```
myapp/
├── internal/
│   └── secret/        ← can only be imported by myapp/* packages
│       └── secret.go
└── pkg/
    └── public/        ← anyone can import this
        └── public.go
```

```go
// ✅ OK — within myapp
import "myapp/internal/secret"

// ❌ Compile error — external package trying to import internal
import "github.com/other/myapp/internal/secret"
```

---

## 📋 The `init` Function

Each package can have one or more `init()` functions — they run automatically before `main()`:

```go
package database

import "fmt"

var db *DB

func init() {
    // Runs automatically when package is imported
    db = connectToDatabase()
    fmt.Println("database package initialized")
}
```

Rules about `init`:
- No parameters, no return values
- Cannot be called manually
- Multiple `init` functions allowed per file (unusual but valid)
- Run order: imported packages first, then current package, in file order

```go
// Common use: registering database drivers
import _ "github.com/lib/pq"   // imports for side effect — runs pq's init()
```

---

## 🧩 Modules — `go.mod`

A **module** is a collection of Go packages with a `go.mod` file defining the module path and dependencies.

### Creating a module

```bash
mkdir myapp && cd myapp
go mod init github.com/yourname/myapp
```

### The `go.mod` file

```
module github.com/yourname/myapp   ← module path (used as import prefix)

go 1.22                             ← minimum Go version

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/stretchr/testify v1.8.4
)
```

### The `go.sum` file

Auto-generated — contains cryptographic hashes of every dependency version. **Never edit manually**. Always commit both `go.mod` and `go.sum`.

---

## 🔧 Essential `go mod` Commands

```bash
# Initialize a new module
go mod init github.com/yourname/myapp

# Add missing and remove unused dependencies
go mod tidy

# Download all dependencies to module cache
go mod download

# Verify dependencies haven't been tampered with
go mod verify

# Show module dependency graph
go mod graph

# List all dependencies
go list -m all
```

---

## 📥 Adding Dependencies

```bash
# Add latest version
go get github.com/gin-gonic/gin

# Add specific version
go get github.com/gin-gonic/gin@v1.9.1

# Upgrade all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy
```

---

## 📤 Importing Packages

### Standard library

```go
import (
    "fmt"
    "os"
    "net/http"
    "encoding/json"
)
```

### Third-party packages

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)
```

### Your own packages

```go
// Module: github.com/yourname/myapp
import (
    "github.com/yourname/myapp/internal/service"
    "github.com/yourname/myapp/config"
)
```

### Import aliases

```go
import (
    "fmt"
    myfmt "github.com/yourname/myapp/pkg/fmt"  // alias to avoid conflict
    _ "github.com/lib/pq"                       // blank — side effects only
    . "math"                                    // dot import — avoid!
)
```

---

## 📐 Versioning — Semantic Versioning

Go modules follow **semantic versioning** (`vMAJOR.MINOR.PATCH`):

| Part | When to bump | Example |
|------|-------------|---------|
| MAJOR | Breaking API changes | `v1 → v2` |
| MINOR | New features, backward compatible | `v1.2 → v1.3` |
| PATCH | Bug fixes | `v1.2.3 → v1.2.4` |

### Major version modules (v2+)

If you release a v2+ module with breaking changes, the module path must include the version:

```go
// v1 (original)
module github.com/yourname/mylib

// v2 (breaking change — path must change)
module github.com/yourname/mylib/v2
```

---

## 🏗️ Real Project Layout

```
github.com/yourname/myapp/
├── go.mod
├── go.sum
├── main.go
│
├── cmd/                        ← entry points for multiple binaries
│   ├── server/
│   │   └── main.go
│   └── worker/
│       └── main.go
│
├── internal/                   ← private — not importable externally
│   ├── handler/
│   ├── service/
│   └── repository/
│
├── pkg/                        ← public — importable by others
│   └── middleware/
│
└── config/
    └── config.go
```

---

## 💻 Complete Example — Multi-Package Project

**`go.mod`:**
```
module github.com/example/calculator

go 1.22
```

**`pkg/math/math.go`:**
```go
package math

import "errors"

var ErrDivisionByZero = errors.New("division by zero")

func Add(a, b float64) float64 {
    return a + b
}

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, ErrDivisionByZero
    }
    return a / b, nil
}
```

**`internal/logger/logger.go`:**
```go
package logger

import (
    "fmt"
    "time"
)

type Logger struct{ Prefix string }

func New(prefix string) *Logger {
    return &Logger{Prefix: prefix}
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s: %s\n", time.Now().Format("15:04:05"), l.Prefix, msg)
}
```

**`main.go`:**
```go
package main

import (
    "fmt"
    "log"

    "github.com/example/calculator/internal/logger"
    "github.com/example/calculator/pkg/math"
)

func main() {
    l := logger.New("CALC")
    l.Log("starting calculator")

    result := math.Add(10, 5)
    fmt.Println("10 + 5 =", result)

    quotient, err := math.Divide(10, 3)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("10 / 3 = %.4f\n", quotient)

    _, err = math.Divide(10, 0)
    if err != nil {
        fmt.Println("Error:", err)   // division by zero
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Circular imports — compile error
```
package A imports package B
package B imports package A   // ❌ compile error
```
Fix: Extract shared code into a third package C that both A and B import.

### 2. Package name should match directory name
```go
// In directory: internal/handler/
package handlers   // ❌ convention says match directory name
package handler    // ✅
```

### 3. Forgetting `go mod tidy`
```bash
# Always run after changing imports
go mod tidy
```

### 4. Dot imports pollute the namespace
```go
import . "math"   // ❌ Pi, Sqrt now in scope — confusing
import "math"     // ✅ math.Pi, math.Sqrt — origin is clear
```

### 5. `init()` side effects are hard to test
```go
// ❌ Hidden side effect
func init() { connectToDatabase() }

// ✅ Explicit — testable and controllable
func Initialize() error { return connectToDatabase() }
```

---

## 🎯 Must-Know Things (Quick Reference)

- **Package** = directory of `.go` files with the same `package` name
- **Module** = collection of packages with a `go.mod`, versioned together
- Uppercase = **exported** (public), lowercase = **unexported** (private)
- `internal/` packages can only be imported within the same module subtree
- `go mod init` — creates `go.mod`; `go mod tidy` — syncs dependencies
- `go get pkg@version` — adds a dependency
- `go.sum` — cryptographic hashes, always commit it
- Import paths = **module path + package directory path**
- `init()` runs automatically before `main()` — use sparingly
- Blank import `_ "pkg"` — imports only for `init()` side effects
- Avoid circular imports — restructure with a shared third package
- Major versions v2+ require `/v2` suffix in the module path

---

## ❓ Interview Questions & Answers

---

**Q1: What is the difference between a package and a module in Go?**

> A package is a single directory of `.go` files that share a `package` declaration — it's the unit of code organization and compilation. A module is a collection of packages rooted at a `go.mod` file — it's the unit of dependency management and versioning. One module contains many packages. When you `go get` a dependency, you're fetching a module; when you `import`, you're importing a package within that module.

---

**Q2: What is the `internal` package and what access restrictions does it enforce?**

> A directory named `internal` restricts which packages can import it. Only packages rooted at the parent of the `internal` directory can import packages inside it. For example, `myapp/internal/db` can only be imported by packages within `myapp/` — external modules get a compile error if they try. This is Go's way of marking implementation details as private at the module level, beyond just unexported identifiers.

---

**Q3: What does `go mod tidy` do?**

> `go mod tidy` synchronizes `go.mod` and `go.sum` with the actual imports in your code. It adds any missing dependencies your code imports but aren't in `go.mod`, and removes any dependencies listed in `go.mod` that aren't actually used. You should run it after adding/removing imports, and before committing. It also updates `go.sum` with the correct cryptographic hashes.

---

**Q4: What is the `init` function and when does it run?**

> `init` is a special function that runs automatically when a package is loaded, before `main()` executes. It takes no parameters and returns nothing. It's used for one-time package setup — registering drivers, initializing package-level state, validating configuration. Packages are initialized in dependency order — all imported packages' `init` functions run first. Multiple `init` functions per file are allowed. It can't be called manually. Use it sparingly — explicit initialization functions are usually more testable.

---

**Q5: What is a blank import and why is it used?**

> A blank import `import _ "package/path"` imports a package solely for its side effects — specifically its `init()` function — without using any of its exported identifiers. The most common use is registering database drivers: `import _ "github.com/lib/pq"` registers the PostgreSQL driver with `database/sql` via `pq`'s `init()`. Without the blank identifier, Go would give a compile error for an unused import.

---

**Q6: How does Go handle major version upgrades in modules?**

> When a module releases a v2 or higher with breaking changes, the module path must include the version suffix: `github.com/yourname/lib/v2`. This allows multiple major versions to coexist in the same build. Callers import the new version by updating their import paths to include `/v2`. This is the "import compatibility rule" and prevents silent breakage from major version upgrades.

---

**Q7: How do you resolve circular import errors in Go?**

> Circular imports — where package A imports B and B imports A — are a compile error in Go. The fix is to restructure the code: extract the shared types or functions that both packages need into a third package C that both A and B import. Often circular imports signal that two packages are too tightly coupled and should be merged, or that a common abstraction is missing. Using interfaces in one package that the other implements can also break cycles.

---

*Next up: **Topic 2.4 — Goroutines** 🚀*
