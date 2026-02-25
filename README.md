# 🐹 Go (Golang) Learning Roadmap — Beginner to Senior

> A structured, topic-by-topic guide to mastering Go from the ground up to senior engineer level.
> Each topic has its own dedicated notes file with explanations, examples, and interview Q&A.

---

## 🟢 Phase 1 — Foundations (Beginner)

| # | Topic | Notes File | Status |
|---|-------|------------|--------|
| 1.1 | Introduction to Go — history, philosophy, use cases, installation, workspace | `01_intro_to_go.md` | ⬜ |
| 1.2 | Your First Go Program — `Hello World`, `go run`, `go build`, package `main` | `02_first_program.md` | ⬜ |
| 1.3 | Variables & Data Types — `var`, `:=`, basic types | `03_variables_and_types.md` | ⬜ |
| 1.4 | Constants & `iota` | `04_constants_iota.md` | ⬜ |
| 1.5 | Operators — arithmetic, comparison, logical, bitwise | `05_operators.md` | ⬜ |
| 1.6 | Control Flow — `if/else`, `switch`, `for` | `06_control_flow.md` | ⬜ |
| 1.7 | Functions — params, return values, multiple returns, named returns | `07_functions.md` | ⬜ |
| 1.8 | Pointers — `*`, `&`, when and why | `08_pointers.md` | ⬜ |
| 1.9 | Arrays & Slices — internals, `append`, `copy`, slicing tricks | `09_arrays_slices.md` | ⬜ |
| 1.10 | Maps — CRUD, checking existence, iteration | `10_maps.md` | ⬜ |
| 1.11 | Structs — defining, embedding, anonymous structs | `11_structs.md` | ⬜ |
| 1.12 | Methods — value vs pointer receivers | `12_methods.md` | ⬜ |

---

## 🔵 Phase 2 — Intermediate

| # | Topic | Notes File | Status |
|---|-------|------------|--------|
| 2.1 | Interfaces — defining, implementing, empty interface, type assertions | `13_interfaces.md` | ⬜ |
| 2.2 | Error Handling — `error` type, `errors.New`, `fmt.Errorf`, `%w` wrapping | `14_error_handling.md` | ⬜ |
| 2.3 | Packages & Modules — `go mod`, visibility, organizing code | `15_packages_modules.md` | ⬜ |
| 2.4 | Goroutines — launching, the Go scheduler, lightweight threads | `16_goroutines.md` | ⬜ |
| 2.5 | Channels — unbuffered, buffered, directional, `close` | `17_channels.md` | ⬜ |
| 2.6 | Select Statement — multiplexing channels | `18_select.md` | ⬜ |
| 2.7 | `sync` Package — `WaitGroup`, `Mutex`, `RWMutex`, `Once` | `19_sync_package.md` | ⬜ |
| 2.8 | Defer, Panic & Recover | `20_defer_panic_recover.md` | ⬜ |
| 2.9 | Stringer & `fmt` Interface | `21_stringer_fmt.md` | ⬜ |
| 2.10 | File I/O — `os`, `bufio`, `io` packages | `22_file_io.md` | ⬜ |
| 2.11 | Working with JSON — `encoding/json`, marshal/unmarshal, struct tags | `23_json.md` | ⬜ |

---

## 🟠 Phase 3 — Advanced

| # | Topic | Notes File | Status |
|---|-------|------------|--------|
| 3.1 | Generics (Go 1.18+) — type parameters, constraints, generic functions & types | `24_generics.md` | ⬜ |
| 3.2 | Context Package — cancellation, deadlines, timeouts, passing values | `25_context.md` | ⬜ |
| 3.3 | Advanced Concurrency Patterns — pipelines, fan-in/out, worker pools | `26_concurrency_patterns.md` | ⬜ |
| 3.4 | Reflection — `reflect` package, use cases and pitfalls | `27_reflection.md` | ⬜ |
| 3.5 | `unsafe` Package — memory layout, when to use | `28_unsafe.md` | ⬜ |
| 3.6 | Testing in Go — `testing` pkg, table-driven tests, benchmarks, fuzz testing | `29_testing.md` | ⬜ |
| 3.7 | HTTP Servers — `net/http`, handlers, middleware, routing | `30_http_servers.md` | ⬜ |
| 3.8 | HTTP Clients — making requests, timeouts, retries | `31_http_clients.md` | ⬜ |
| 3.9 | Databases — `database/sql`, connection pooling, `sqlx` / `pgx` | `32_databases.md` | ⬜ |

---

## 🔴 Phase 4 — Senior / Production Level

| # | Topic | Notes File | Status |
|---|-------|------------|--------|
| 4.1 | Project Structure & Design Patterns — Clean Architecture, repository pattern, DI | `33_project_structure.md` | ⬜ |
| 4.2 | gRPC & Protobuf — defining services, streaming, interceptors | `34_grpc_protobuf.md` | ⬜ |
| 4.3 | Performance & Profiling — `pprof`, CPU/memory profiling, escape analysis | `35_performance_profiling.md` | ⬜ |
| 4.4 | Memory Management — GC internals, stack vs heap, reducing allocations | `36_memory_management.md` | ⬜ |
| 4.5 | CLI Tools — building with `cobra`, `flag` | `37_cli_tools.md` | ⬜ |
| 4.6 | Configuration Management — `viper`, env vars, 12-factor app | `38_config_management.md` | ⬜ |
| 4.7 | Logging & Observability — structured logging (`slog`, `zerolog`), tracing, metrics | `39_logging_observability.md` | ⬜ |
| 4.8 | Containerization & Deployment — multi-stage Docker builds, distroless images | `40_containerization.md` | ⬜ |
| 4.9 | Security Best Practices — input validation, secrets management, secure HTTP | `41_security.md` | ⬜ |
| 4.10 | Reading Large Go Codebases & Idiomatic Go | `42_idiomatic_go.md` | ⬜ |

---

## 📌 How to Use This Roadmap

- Work through topics **sequentially** — later topics build on earlier ones.
- Each topic's `.md` notes file contains:
  - 📖 Concept explanation with real-world context
  - 💻 Code examples (runnable)
  - ⚠️ Common pitfalls & gotchas
  - 🎯 Must-know things summary
  - ❓ Interview questions & answers
- Mark topics **✅** as you complete them.
- Revisit earlier topics when you encounter them in later phases.

---

## 🛠️ Setup Reminder

```bash
# Install Go: https://go.dev/dl/
go version          # verify installation
go env GOPATH       # check your workspace

# Create a new module
mkdir my-go-project && cd my-go-project
go mod init github.com/yourname/my-go-project
```

---

*Happy coding! Go is simple by design — embrace the simplicity.* 🐹
