# 1.1 — Introduction to Go

> **Phase:** 1 — Foundations | **Difficulty:** Beginner

---

## 📖 What is Go?

Go (also called **Golang**) is an open-source, statically typed, compiled programming language created at **Google** in 2007 by **Robert Griesemer, Rob Pike, and Ken Thompson**. It was publicly released in **2009** and reached version 1.0 in **2012**.

Go was designed to solve real frustrations engineers at Google faced with existing languages:
- C/C++ were fast but slow to compile and hard to manage.
- Python/Ruby were easy to write but slow at runtime.
- Java had too much boilerplate and verbosity.

Go's philosophy: **simple, readable, fast, and built for concurrency.**

---

## 🧠 Go's Core Philosophy

| Principle | What it means |
|-----------|---------------|
| **Simplicity** | A small language spec — few keywords, no inheritance, no operator overloading |
| **Readability** | Code should read like plain English; formatting is enforced by `gofmt` |
| **Explicitness** | No magic. Errors are returned, not thrown. No hidden control flow |
| **Concurrency first** | Goroutines and channels are built into the language, not bolted on |
| **Fast compilation** | Even large codebases compile in seconds |
| **One way to do things** | Reduces bikeshedding and keeps codebases consistent |

> **Go's unofficial motto:** *"Clear is better than clever."*

---

## 🏗️ Where is Go Used?

Go shines in backend and infrastructure engineering:

- **Web servers & APIs** — Go's `net/http` package is production-ready out of the box
- **Microservices** — fast startup, low memory footprint
- **CLI tools** — `kubectl`, `terraform`, `docker`, `hugo` are all written in Go
- **Cloud infrastructure** — Kubernetes, Docker, Prometheus, etcd, Consul
- **Networking tools** — built-in TCP/UDP support
- **DevOps & SRE tooling**

---

## ⚙️ Installing Go

### Step 1: Download

Visit [https://go.dev/dl/](https://go.dev/dl/) and download the installer for your OS.

### Step 2: Verify installation

```bash
go version
# Output: go version go1.22.0 linux/amd64
```

### Step 3: Check your environment

```bash
go env
```

Key environment variables:

| Variable | Description |
|----------|-------------|
| `GOROOT` | Where Go is installed (the toolchain) |
| `GOPATH` | Your workspace — where downloaded modules are cached (`~/go` by default) |
| `GOMODCACHE` | Where module cache lives (`$GOPATH/pkg/mod`) |
| `GOOS` | Target OS (e.g., `linux`, `darwin`, `windows`) |
| `GOARCH` | Target architecture (e.g., `amd64`, `arm64`) |

---

## 🗂️ Go Workspace & Module System

### Before Go 1.11 (old way — GOPATH mode)
All Go code had to live inside `$GOPATH/src/`. This was rigid and annoying.

### After Go 1.11 (modern way — Module mode) ✅
You can put your project **anywhere** on your filesystem and use `go mod` to manage dependencies.

```bash
# Create a new project
mkdir hello-go && cd hello-go

# Initialize a Go module
go mod init github.com/yourname/hello-go
```

This creates a `go.mod` file — Go's equivalent of `package.json` or `requirements.txt`.

```
module github.com/yourname/hello-go

go 1.22
```

---

## 🔑 Key Go Tooling

These are the tools you'll use daily:

| Command | Purpose |
|---------|---------|
| `go run file.go` | Compile and run a file directly |
| `go build` | Compile and produce a binary |
| `go fmt ./...` | Format all code (enforced style) |
| `go vet ./...` | Static analysis — catches common bugs |
| `go test ./...` | Run all tests |
| `go mod tidy` | Clean up `go.mod` and `go.sum` |
| `go get package` | Add a dependency |
| `go doc fmt.Println` | View documentation |
| `go install` | Compile and install binary to `$GOPATH/bin` |

---

## 🧩 Go's Key Characteristics (vs other languages)

| Feature | Go | Python | Java |
|---------|-----|--------|------|
| Typing | Static | Dynamic | Static |
| Compilation | Compiled | Interpreted | Compiled (JVM) |
| Concurrency | Goroutines (built-in) | Threads/async | Threads |
| Error handling | Return values | Exceptions | Exceptions |
| Garbage collection | Yes (low-latency GC) | Yes | Yes |
| Generics | Yes (since 1.18) | Yes | Yes |
| Inheritance | No (uses composition) | Yes | Yes |
| Runtime | Standalone binary | Needs interpreter | Needs JVM |

---

## 💻 Your First Peek at Go Code

```go
package main  // Every executable Go program must be in package "main"

import "fmt"  // Import the format package from standard library

func main() {  // main() is the entry point — execution starts here
    fmt.Println("Hello, Go!")
}
```

**Run it:**
```bash
go run main.go
# Output: Hello, Go!
```

We'll break this down fully in **Topic 1.2**, but notice:
- No semicolons
- Curly brace `{` must be on the same line (enforced!)
- `fmt.Println` — capitalized = exported (public)

---

## ⚠️ Common Gotchas for Beginners

1. **Unused imports are a compile error** — if you import a package and don't use it, Go won't compile.
2. **Unused variables are a compile error** — declare a variable and never use it? Compile error.
3. **Opening brace `{` must be on the same line** — Go's parser requires this.
4. **`gofmt` is law** — Go enforces code formatting. Run it always (most editors do this automatically).

```go
// ❌ This will NOT compile
import "fmt"   // imported and not used

// ❌ This will NOT compile
x := 5   // x declared and not used

// ✅ This compiles fine
x := 5
fmt.Println(x)
```

---

## 🎯 Must-Know Things (Quick Reference)

- Go was created at Google in 2007, open sourced in 2009
- Designed by Robert Griesemer, Rob Pike, Ken Thompson
- Statically typed, compiled, garbage collected
- No classes, no inheritance — uses **composition and interfaces**
- **Concurrency is a first-class citizen** — goroutines are cheap (~2KB stack)
- A Go binary is **self-contained** — no runtime needed on the target machine
- Uses **modules** (`go.mod`) for dependency management since Go 1.11
- `gofmt` enforces a single code style — no debates about formatting
- Unused imports and variables are **compile errors**

---

## ❓ Interview Questions & Answers

---

**Q1: What is Go and who created it?**

> Go is a statically typed, compiled, open-source programming language developed at Google. It was created by Robert Griesemer, Rob Pike, and Ken Thompson, released publicly in 2009. It was designed to combine the performance of C with the readability of Python while having built-in support for concurrency.

---

**Q2: Why was Go created? What problems does it solve?**

> Go was created to address pain points at Google: C/C++ was fast but had slow compile times and complex memory management; Python/Ruby were easy but slow; Java had too much boilerplate. Go aimed to be fast to compile, easy to write, and safe to run — especially for large-scale distributed systems.

---

**Q3: What are the key features of Go?**

> - Statically typed with type inference (`:=`)
> - Fast compilation (seconds even for large projects)
> - Built-in concurrency with goroutines and channels
> - Garbage collection with low-latency GC
> - No classes — uses structs and interfaces
> - Single binary deployment (no runtime dependencies)
> - Strong standard library
> - Built-in tooling: `gofmt`, `go vet`, `go test`

---

**Q4: How is Go different from Java or Python?**

> **vs Java:** Go compiles to a native binary (no JVM), has simpler syntax, no classes/inheritance, and goroutines instead of threads. Go programs start faster and use less memory.

> **vs Python:** Go is statically typed and compiled — much faster at runtime. Go has explicit error handling (return values) vs Python's exception model. However, Python has a richer ecosystem for data science.

---

**Q5: What is `GOPATH` and `GOROOT`?**

> `GOROOT` is where Go itself is installed (the compiler, standard library). `GOPATH` is the workspace directory where Go downloads and caches packages. In modern Go (1.11+) with modules, you don't need to organize your code inside `GOPATH` anymore — you can work from any directory.

---

**Q6: What is a Go module?**

> A Go module is the unit of dependency management in Go. It's initialized with `go mod init <module-name>` and creates a `go.mod` file that tracks the module name and its dependencies (with their versions). The `go.sum` file contains cryptographic hashes for security. Modules replaced the older `GOPATH`-based approach.

---

**Q7: Why are unused imports and variables compile errors in Go?**

> This is a deliberate design decision to keep code clean. Unused imports slow compilation and indicate dead code. Unused variables are usually a sign of a bug (e.g., forgetting to use a result). By making these compile errors, Go forces developers to write cleaner code.

---

**Q8: What is `gofmt` and why does it matter?**

> `gofmt` is Go's official code formatter. It enforces a single, canonical style for all Go code — indentation, spacing, brace placement, etc. This eliminates style debates in teams and makes all Go code look consistent. In Go, there's only one way to format your code. Most editors run `gofmt` automatically on save.

---

**Q9: Is Go object-oriented?**

> Go is not object-oriented in the traditional sense — it has no classes, no inheritance, and no constructors. However, it supports **object-oriented patterns** through structs (data), methods (behavior on structs), interfaces (polymorphism), and composition (embedding structs). Go favors **composition over inheritance**.

---

**Q10: What types of applications is Go best suited for?**

> Go excels at: backend web services and REST APIs, microservices, CLI tools, networking tools, DevOps/infrastructure tooling, and distributed systems. It's used extensively at Google, Uber, Cloudflare, Dropbox, and in projects like Kubernetes, Docker, Terraform, and Prometheus.

---

## 📚 Resources

- Official site: [https://go.dev](https://go.dev)
- Go Playground (run code in browser): [https://go.dev/play/](https://go.dev/play/)
- Standard library docs: [https://pkg.go.dev/std](https://pkg.go.dev/std)
- Effective Go (must-read): [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
- Go specification: [https://go.dev/ref/spec](https://go.dev/ref/spec)

---

*Next up: **Topic 1.2 — Your First Go Program** 🚀*
