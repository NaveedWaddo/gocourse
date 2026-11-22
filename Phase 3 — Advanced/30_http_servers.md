# 3.7 — HTTP Servers

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.6 — Testing in Go

---

## 📖 Overview

Go's `net/http` package provides a production-ready HTTP server out of the box — no framework required. Understanding the standard library deeply is essential before reaching for third-party routers like `chi`, `gin`, or `echo`.

Core concepts:
- `http.Handler` interface — the heart of all HTTP in Go
- `http.ServeMux` — the built-in router (upgraded in Go 1.22)
- Middleware — composing handlers
- `httptest` — testing handlers without a real server
- Graceful shutdown

---

## 🔑 The `http.Handler` Interface

Everything in Go's HTTP is built on one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type implementing `ServeHTTP` is a handler. `http.HandlerFunc` lets you use a plain function as a handler:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

So `func(w http.ResponseWriter, r *http.Request)` **is** a handler via `http.HandlerFunc`.

---

## 🚦 Basic Server

```go
package main

import (
    "fmt"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, Go!")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", helloHandler)

    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    fmt.Println("server listening on :8080")
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}
```

> Always set timeouts on `http.Server` — the zero-value leaves all timeouts infinite, which is a security and reliability risk.

---

## 🗺️ ServeMux — Routing (Go 1.22+)

Go 1.22 dramatically upgraded `http.ServeMux` with method matching and path parameters:

```go
mux := http.NewServeMux()

// Method + path matching (Go 1.22+)
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("PUT /users/{id}", updateUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)

// Wildcard — matches everything under /static/
mux.HandleFunc("GET /static/{path...}", serveStatic)
```

### Extracting path parameters (Go 1.22+)

```go
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")   // extract {id} from path
    fmt.Fprintf(w, "getting user: %s\n", id)
}
```

### Before Go 1.22 — manual routing

```go
func usersRouter(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        listUsers(w, r)
    case http.MethodPost:
        createUser(w, r)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}
```

---

## 📤 Writing Responses

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Set headers BEFORE writing status or body
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", "abc-123")

    // Write status code — call once, before writing body
    w.WriteHeader(http.StatusCreated)   // 201

    // Write body
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
```

### Common response helpers

```go
// Plain text error
http.Error(w, "not found", http.StatusNotFound)

// Redirect
http.Redirect(w, r, "/new-location", http.StatusMovedPermanently)

// Serve a file
http.ServeFile(w, r, "index.html")

// JSON helper (write your own or use a framework's)
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        log.Printf("writeJSON: %v", err)
    }
}

// Error JSON helper
func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}
```

### Header ordering rules

```go
// ✅ Correct order
w.Header().Set("Content-Type", "application/json")  // 1. set headers
w.WriteHeader(http.StatusOK)                         // 2. write status
w.Write([]byte(`{"ok":true}`))                       // 3. write body

// ❌ Wrong — headers ignored after WriteHeader or Write
w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "application/json")  // too late!
```

---

## 📥 Reading Requests

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Method and URL
    fmt.Println(r.Method)           // GET, POST, etc.
    fmt.Println(r.URL.Path)         // /users/42
    fmt.Println(r.URL.RawQuery)     // page=1&limit=10

    // Query parameters
    page := r.URL.Query().Get("page")
    limit := r.URL.Query().Get("limit")

    // Path value (Go 1.22+)
    id := r.PathValue("id")

    // Headers
    authHeader := r.Header.Get("Authorization")
    contentType := r.Header.Get("Content-Type")

    // Body — always close it
    defer r.Body.Close()
    body, err := io.ReadAll(r.Body)

    // JSON body
    var payload CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    // Form data (after r.ParseForm())
    r.ParseForm()
    name := r.FormValue("name")

    // Context — carries cancellation, deadline, request-scoped values
    ctx := r.Context()

    _ = page; _ = limit; _ = id; _ = authHeader
    _ = contentType; _ = body; _ = name; _ = ctx
}
```

---

## 🔗 Middleware — Composing Handlers

Middleware is a function that wraps a handler, adding behaviour before and/or after:

```go
// Middleware signature: takes a Handler, returns a Handler
type Middleware func(http.Handler) http.Handler

// Logging middleware
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// Auth middleware
func requireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return   // don't call next
        }
        next.ServeHTTP(w, r)
    })
}

// Recovery middleware — catch panics
func recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// CORS middleware
func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Chaining middleware

```go
// Manual chaining — innermost handler first, outermost last
handler := recovery(logging(requireAuth(mux)))

// Helper to chain cleanly
func chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Usage
handler := chain(mux, recovery, logging, cors)
```

---

## 🔑 Passing Values via Context

Use `r.Context()` and middleware to attach request-scoped values:

```go
type contextKey string
const keyUserID contextKey = "userID"

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        userID, err := validateToken(token)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        // Attach userID to request context
        ctx := context.WithValue(r.Context(), keyUserID, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getUser(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(keyUserID).(int)
    fmt.Fprintf(w, "hello user %d\n", userID)
}
```

---

## 🏗️ Response Writer Wrapper — Capturing Status Code

The standard `ResponseWriter` doesn't expose the status code after writing. Wrap it:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (r *statusRecorder) WriteHeader(status int) {
    r.status = status
    r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Status() int {
    if r.status == 0 { return http.StatusOK }
    return r.status
}

// Use in logging middleware
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{ResponseWriter: w}
        start := time.Now()
        next.ServeHTTP(rec, r)
        log.Printf("%s %s → %d (%v)", r.Method, r.URL.Path, rec.Status(), time.Since(start))
    })
}
```

---

## 🧪 Testing HTTP Handlers — `httptest`

Test handlers without starting a real server:

```go
import "net/http/httptest"

func TestHelloHandler(t *testing.T) {
    // Create a request
    req := httptest.NewRequest(http.MethodGet, "/hello", nil)
    req.Header.Set("Authorization", "Bearer token123")

    // Create a response recorder
    rec := httptest.NewRecorder()

    // Call the handler directly
    helloHandler(rec, req)

    // Inspect the response
    res := rec.Result()
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        t.Errorf("status = %d; want %d", res.StatusCode, http.StatusOK)
    }

    body, _ := io.ReadAll(res.Body)
    if !strings.Contains(string(body), "Hello") {
        t.Errorf("body = %q; want it to contain 'Hello'", body)
    }
}

// Test with a real server (for integration tests)
func TestServer(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /hello", helloHandler)

    ts := httptest.NewServer(mux)   // starts a local test server
    defer ts.Close()

    resp, err := http.Get(ts.URL + "/hello")
    if err != nil { t.Fatal(err) }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d; want 200", resp.StatusCode)
    }
}
```

---

## 🛑 Graceful Shutdown

A production server must drain in-flight requests on shutdown:

```go
func main() {
    mux := http.NewServeMux()
    // ... register routes

    server := &http.Server{
        Addr:         ":8080",
        Handler:      chain(mux, recovery, logging),
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()
    log.Println("server started on :8080")

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("shutting down server...")

    // Give in-flight requests 30 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("forced shutdown: %v", err)
    }
    log.Println("server stopped")
}
```

---

## 💻 Complete Example — REST API

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "sync"
    "time"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// In-memory store
type Store struct {
    mu    sync.RWMutex
    users map[int]User
    next  int
}

func NewStore() *Store {
    return &Store{users: make(map[int]User), next: 1}
}

func (s *Store) List() []User {
    s.mu.RLock(); defer s.mu.RUnlock()
    out := make([]User, 0, len(s.users))
    for _, u := range s.users { out = append(out, u) }
    return out
}

func (s *Store) Create(u User) User {
    s.mu.Lock(); defer s.mu.Unlock()
    u.ID = s.next; s.next++
    s.users[u.ID] = u
    return u
}

func (s *Store) Get(id int) (User, bool) {
    s.mu.RLock(); defer s.mu.RUnlock()
    u, ok := s.users[id]
    return u, ok
}

// Handler with dependency
type UserHandler struct{ store *Store }

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, h.store.List())
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
    var u User
    if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
        return
    }
    if u.Name == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
        return
    }
    created := h.store.Create(u)
    writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
        return
    }
    u, ok := h.store.Get(id)
    if !ok {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
        return
    }
    writeJSON(w, http.StatusOK, u)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func main() {
    store := NewStore()
    h := &UserHandler{store: store}

    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", h.list)
    mux.HandleFunc("POST /users", h.create)
    mux.HandleFunc("GET /users/{id}", h.get)

    server := &http.Server{
        Addr:         ":8080",
        Handler:      logging(mux),
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    fmt.Println("listening on :8080")
    log.Fatal(server.ListenAndServe())
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Writing headers after the body
```go
// ❌ Headers must be set before WriteHeader or Write
w.Write([]byte("body"))
w.Header().Set("Content-Type", "application/json")  // ignored!

// ✅ Headers first, then status, then body
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte("body"))
```

### 2. Not setting timeouts on http.Server
```go
// ❌ Zero-value server — no timeouts — vulnerable to slow-loris attacks
http.ListenAndServe(":8080", mux)

// ✅ Always set timeouts
server := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### 3. Returning after http.Error without `return`
```go
// ❌ Continues executing after sending error response
func handler(w http.ResponseWriter, r *http.Request) {
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        // falls through — may write again!
    }
    // more code runs even after error response sent
}

// ✅
if err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return   // stop here
}
```

### 4. Not closing request body
```go
// ❌ Leaks connection
body, _ := io.ReadAll(r.Body)

// ✅
defer r.Body.Close()
body, _ := io.ReadAll(r.Body)
```

### 5. Using default ServeMux with `http.HandleFunc`
```go
// ❌ Registers on the global default mux — hard to test, no isolation
http.HandleFunc("/users", handler)
http.ListenAndServe(":8080", nil)   // nil uses default mux

// ✅ Create an explicit mux
mux := http.NewServeMux()
mux.HandleFunc("/users", handler)
```

---

## 🎯 Must-Know Things (Quick Reference)

- `http.Handler` interface: `ServeHTTP(ResponseWriter, *Request)` — everything implements this
- `http.HandlerFunc` — converts `func(w, r)` to a `Handler`
- Always set `ReadTimeout`, `WriteTimeout`, `IdleTimeout` on `http.Server`
- Headers → `WriteHeader` → body — order is critical; headers after write are ignored
- Go 1.22 `ServeMux`: method matching (`GET /users`), path params (`{id}`), `r.PathValue("id")`
- Middleware: `func(http.Handler) http.Handler` — chain with a helper for clarity
- `r.Context()` carries cancellation and values; `r.WithContext(ctx)` updates it
- `httptest.NewRecorder()` + `httptest.NewRequest()` — test handlers without a server
- `httptest.NewServer(mux)` — real local server for integration tests
- Graceful shutdown: `server.Shutdown(ctx)` — drains in-flight requests
- Always `return` after writing an error response
- Inject dependencies via struct fields — not global variables

---

## ❓ Interview Questions & Answers

---

**Q1: What is `http.Handler` and why is it central to Go's HTTP model?**

> `http.Handler` is an interface with a single method `ServeHTTP(ResponseWriter, *Request)`. It's central because the entire `net/http` package is built around it — the server calls `ServeHTTP` on whatever handler you provide, and that handler can delegate to others, inspect the request, write the response, or compose with middleware. `http.HandlerFunc` makes any matching function satisfy the interface. This design means middleware, routers, and handlers are all the same type, composable without any framework magic.

---

**Q2: How does middleware work in Go's HTTP model?**

> Middleware is a function that takes an `http.Handler` and returns an `http.Handler`. The returned handler wraps the original — it can execute code before calling the inner handler (pre-processing: auth, logging), after (post-processing: response headers, timing), or conditionally skip it (abort: reject unauthenticated requests). Multiple middleware are chained by wrapping: `logging(auth(mux))`. The outermost middleware runs first. A chain helper reverses the slice for readable left-to-right ordering.

---

**Q3: What are the timeout fields on `http.Server` and why do they matter?**

> `ReadTimeout` is the max time to read the entire request including body. `WriteTimeout` is the max time to write the response. `IdleTimeout` is the max time a keep-alive connection can be idle. Without timeouts, a slow or malicious client can hold connections open indefinitely — a slow-loris attack. The convenience function `http.ListenAndServe` uses a zero-value server with no timeouts, which is why production code should always create an explicit `http.Server` with timeouts set.

---

**Q4: How do you test an HTTP handler in Go without starting a real server?**

> Use `net/http/httptest`. Create a request with `httptest.NewRequest(method, path, body)` and a recorder with `httptest.NewRecorder()`. Call the handler function directly: `handler(recorder, request)`. Then inspect `recorder.Code` for the status and `recorder.Body` for the response body. This is fast, parallel-safe, and doesn't require network ports. For integration tests that need real HTTP semantics (connection handling, middleware chains), use `httptest.NewServer(handler)` which starts a real local server on a random port.

---

**Q5: What changed in `http.ServeMux` in Go 1.22?**

> Go 1.22 added method-based routing (`GET /users`, `POST /users`) and named path parameters (`/users/{id}`). Before 1.22, `ServeMux` only matched on path prefix — no method filtering and no path parameters, requiring manual `switch r.Method` and string parsing. With 1.22, you access path values via `r.PathValue("id")`. The trailing `{path...}` wildcard matches the rest of the path. These additions cover most routing needs without a third-party router, though libraries like `chi` still offer features like middleware grouping and regex constraints.

---

**Q6: How do you implement graceful shutdown?**

> Start `server.ListenAndServe()` in a goroutine. In the main goroutine, listen for OS signals (`SIGINT`, `SIGTERM`) via `signal.Notify`. When a signal arrives, call `server.Shutdown(ctx)` with a timeout context — it stops accepting new connections and waits for in-flight requests to complete, up to the timeout. `ListenAndServe` returns `http.ErrServerClosed` when shutdown completes — this is expected and should not be logged as an error. Without graceful shutdown, active requests are abruptly terminated when the process exits.

---

*Next up: **Topic 3.8 — HTTP Clients** 🚀*
