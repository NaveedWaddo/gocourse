# 3.8 — HTTP Clients

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.7 — HTTP Servers

---

## 📖 Overview

Go's `net/http` package includes a powerful HTTP client. The key insight is that the default `http.DefaultClient` is **not safe for production** — you must configure your own client with timeouts, connection pooling, and retry logic.

Core concepts:
- `http.Client` — the configurable client
- `http.Transport` — connection pooling and TLS
- Building and sending requests
- Reading and handling responses
- Timeouts and context
- Retries and resilience
- Testing with `httptest.Server`

---

## ⚠️ Never Use `http.DefaultClient` in Production

```go
// ❌ DefaultClient has NO timeout — hangs forever on slow servers
resp, err := http.Get("https://api.example.com/users")

// ✅ Always create a client with timeout
client := &http.Client{
    Timeout: 10 * time.Second,
}
resp, err := client.Get("https://api.example.com/users")
```

`http.DefaultClient` is a zero-value `http.Client` with no timeout. One slow server can hang your goroutine forever.

---

## 🔧 Configuring `http.Client`

```go
client := &http.Client{
    // Total request timeout (including redirects, reading body)
    Timeout: 30 * time.Second,

    // Custom transport for connection pooling, TLS, proxy
    Transport: &http.Transport{
        MaxIdleConns:        100,              // max idle connections total
        MaxIdleConnsPerHost: 10,               // max idle per host
        MaxConnsPerHost:     100,              // max total per host
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        DisableCompression:  false,
        // Dial timeout
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
    },

    // Control redirect behaviour
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        if len(via) >= 5 {
            return fmt.Errorf("too many redirects")
        }
        return nil
    },
}
```

### Timeout hierarchy

```
client.Timeout             ← total: dial + connect + TLS + send + read headers + read body
Transport.DialContext      ← TCP dial timeout
Transport.TLSHandshakeTimeout ← TLS timeout
context.WithTimeout        ← per-request timeout (overrides client.Timeout if shorter)
```

---

## 🛠️ Making Requests

### GET

```go
func getUser(ctx context.Context, client *http.Client, id int) (*User, error) {
    url := fmt.Sprintf("https://api.example.com/users/%d", id)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    return &user, nil
}
```

### POST with JSON body

```go
func createUser(ctx context.Context, client *http.Client, u CreateUserRequest) (*User, error) {
    body, err := json.Marshal(u)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost,
        "https://api.example.com/users",
        bytes.NewReader(body),
    )
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
    }

    var created User
    if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    return &created, nil
}
```

### Always use `http.NewRequestWithContext` — not `http.NewRequest`

```go
// ❌ No context — can't cancel, can't propagate deadline
req, _ := http.NewRequest("GET", url, nil)

// ✅ With context — cancellable, deadline-aware
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

---

## 📥 Reading Responses Correctly

```go
resp, err := client.Do(req)
if err != nil { return nil, err }
defer resp.Body.Close()   // ALWAYS close body

// Check status BEFORE reading body
if resp.StatusCode >= 400 {
    // Limit body read to prevent OOM from huge error responses
    body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
    return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
}

// Read body safely
body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))   // 10 MB limit
if err != nil {
    return nil, fmt.Errorf("read body: %w", err)
}

// Or decode JSON directly (no need to ReadAll first)
var result MyStruct
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, fmt.Errorf("decode: %w", err)
}
```

---

## 🔁 Retry Logic

Retries with exponential backoff for transient failures:

```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func withRetry(ctx context.Context, cfg RetryConfig, fn func() (*http.Response, error)) (*http.Response, error) {
    var lastErr error
    delay := cfg.BaseDelay

    for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
        resp, err := fn()

        // Success or non-retryable error
        if err == nil && !isRetryable(resp.StatusCode) {
            return resp, nil
        }

        if err == nil {
            resp.Body.Close()   // drain and close before retry
            lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
        } else {
            lastErr = err
        }

        if attempt == cfg.MaxAttempts { break }

        // Wait before retry — respect context cancellation
        select {
        case <-time.After(delay):
        case <-ctx.Done():
            return nil, ctx.Err()
        }

        // Exponential backoff with cap
        delay = time.Duration(float64(delay) * cfg.Multiplier)
        if delay > cfg.MaxDelay { delay = cfg.MaxDelay }
    }
    return nil, fmt.Errorf("after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

func isRetryable(status int) bool {
    // Retry on 429, 500, 502, 503, 504
    switch status {
    case http.StatusTooManyRequests,
        http.StatusInternalServerError,
        http.StatusBadGateway,
        http.StatusServiceUnavailable,
        http.StatusGatewayTimeout:
        return true
    }
    return false
}
```

---

## 🏗️ Reusable API Client Pattern

Structure clients as a struct with the `http.Client` embedded and per-method functions:

```go
type APIClient struct {
    baseURL    string
    httpClient *http.Client
    apiKey     string
}

func NewAPIClient(baseURL, apiKey string) *APIClient {
    return &APIClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}

func (c *APIClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
    var bodyReader io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil { return nil, fmt.Errorf("marshal: %w", err) }
        bodyReader = bytes.NewReader(data)
    }

    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
    if err != nil { return nil, fmt.Errorf("new request: %w", err) }

    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("Accept", "application/json")
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := c.httpClient.Do(req)
    if err != nil { return nil, fmt.Errorf("do: %w", err) }

    if resp.StatusCode >= 400 {
        defer resp.Body.Close()
        errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errBody)
    }
    return resp, nil
}

func (c *APIClient) GetUser(ctx context.Context, id int) (*User, error) {
    resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/users/%d", id), nil)
    if err != nil { return nil, err }
    defer resp.Body.Close()

    var user User
    return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func (c *APIClient) CreateUser(ctx context.Context, u CreateUserRequest) (*User, error) {
    resp, err := c.do(ctx, http.MethodPost, "/users", u)
    if err != nil { return nil, err }
    defer resp.Body.Close()

    var created User
    return &created, json.NewDecoder(resp.Body).Decode(&created)
}
```

---

## 🧪 Testing HTTP Clients — `httptest.Server`

```go
func TestGetUser(t *testing.T) {
    // Start a local test server
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify the request
        if r.URL.Path != "/users/1" {
            t.Errorf("unexpected path: %s", r.URL.Path)
        }
        if r.Header.Get("Authorization") == "" {
            t.Error("missing Authorization header")
        }

        // Return fake response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(User{ID: 1, Name: "Alice", Age: 30})
    }))
    defer ts.Close()

    // Point client at test server
    client := NewAPIClient(ts.URL, "test-key")
    user, err := client.GetUser(context.Background(), 1)

    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if user.Name != "Alice" { t.Errorf("name = %q; want Alice", user.Name) }
}

func TestGetUser_ServerError(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "service unavailable", http.StatusServiceUnavailable)
    }))
    defer ts.Close()

    client := NewAPIClient(ts.URL, "test-key")
    _, err := client.GetUser(context.Background(), 1)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

---

## 💻 Complete Example

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Post struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
    UserID int    `json:"userId"`
}

type JSONPlaceholderClient struct {
    base   string
    client *http.Client
}

func NewJSONPlaceholderClient() *JSONPlaceholderClient {
    return &JSONPlaceholderClient{
        base: "https://jsonplaceholder.typicode.com",
        client: &http.Client{
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                MaxIdleConnsPerHost: 5,
                IdleConnTimeout:     60 * time.Second,
            },
        },
    }
}

func (c *JSONPlaceholderClient) GetPost(ctx context.Context, id int) (*Post, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet,
        fmt.Sprintf("%s/posts/%d", c.base, id), nil)
    if err != nil { return nil, err }
    req.Header.Set("Accept", "application/json")

    resp, err := c.client.Do(req)
    if err != nil { return nil, fmt.Errorf("get post: %w", err) }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
        return nil, fmt.Errorf("status %d: %s", resp.StatusCode, b)
    }

    var post Post
    return &post, json.NewDecoder(resp.Body).Decode(&post)
}

func (c *JSONPlaceholderClient) CreatePost(ctx context.Context, p Post) (*Post, error) {
    data, err := json.Marshal(p)
    if err != nil { return nil, err }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost,
        c.base+"/posts", bytes.NewReader(data))
    if err != nil { return nil, err }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil { return nil, fmt.Errorf("create post: %w", err) }
    defer resp.Body.Close()

    var created Post
    return &created, json.NewDecoder(resp.Body).Decode(&created)
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    client := NewJSONPlaceholderClient()

    // GET
    fmt.Println("=== GET Post ===")
    post, err := client.GetPost(ctx, 1)
    if err != nil {
        fmt.Println("error:", err)
    } else {
        fmt.Printf("Post #%d: %s\n", post.ID, post.Title)
    }

    // POST
    fmt.Println("\n=== POST Create ===")
    newPost, err := client.CreatePost(ctx, Post{
        Title:  "My Go Post",
        Body:   "Learning HTTP clients in Go",
        UserID: 1,
    })
    if err != nil {
        fmt.Println("error:", err)
    } else {
        fmt.Printf("Created post #%d: %s\n", newPost.ID, newPost.Title)
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Not closing response body — connection leak
```go
resp, err := client.Do(req)
if err != nil { return err }
// ❌ Forgot defer resp.Body.Close() — connection can't be reused!

// ✅ Always close immediately after checking error
defer resp.Body.Close()
```

### 2. Not draining body before closing
```go
// ❌ If body not drained, connection cannot be reused (TCP overhead)
resp.Body.Close()

// ✅ Drain then close — connection returns to pool
io.Copy(io.Discard, resp.Body)
resp.Body.Close()

// Or just io.ReadAll — reads and drains in one step
body, _ := io.ReadAll(resp.Body)
```

### 3. Not checking the status code
```go
resp, _ := client.Do(req)
defer resp.Body.Close()
json.NewDecoder(resp.Body).Decode(&result)   // ❌ decodes error body as success!

// ✅ Always check status first
if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}
```

### 4. Sharing http.Client across goroutines — safe!
```go
// ✅ http.Client is goroutine-safe — share it freely
var client = &http.Client{Timeout: 10 * time.Second}

func handler1() { client.Get(url1) }
func handler2() { client.Get(url2) }
```

### 5. Creating a new http.Client per request — destroys connection pooling
```go
// ❌ New client = new connection pool = no reuse
func fetchUser(id int) {
    client := &http.Client{Timeout: 10 * time.Second}   // new every call!
    client.Get(...)
}

// ✅ Create once, reuse everywhere
var httpClient = &http.Client{Timeout: 10 * time.Second}
func fetchUser(id int) { httpClient.Get(...) }
```

### 6. Reading body after error — meaningless
```go
resp, err := client.Do(req)
if err != nil {
    body, _ := io.ReadAll(resp.Body)   // ❌ resp is nil when err != nil!
}
```

---

## 🎯 Must-Know Things (Quick Reference)

- **Never use `http.DefaultClient`** in production — no timeout
- `client.Timeout` = total time including reading body; set to 10–30s typically
- `http.Transport` controls connection pooling — **create once and reuse the client**
- Always use `http.NewRequestWithContext(ctx, ...)` — not `http.NewRequest`
- Always `defer resp.Body.Close()` — immediately after the error check
- Drain body before closing for connection reuse: `io.Copy(io.Discard, resp.Body)`
- Check `resp.StatusCode` before decoding the body
- Limit body reads: `io.LimitReader(resp.Body, maxBytes)` — prevent OOM
- `http.Client` is goroutine-safe — share one instance across the whole app
- Retry on 429, 500, 502, 503, 504 — not on 400, 401, 403, 404 (client errors)
- Use exponential backoff with jitter for retries
- Test with `httptest.NewServer` — point your client at the test server URL

---

## ❓ Interview Questions & Answers

---

**Q1: Why shouldn't you use `http.DefaultClient` in production?**

> `http.DefaultClient` is a zero-value `http.Client` with no timeout set. If a server is slow or unresponsive, requests hang forever, eventually exhausting goroutines or file descriptors. It also uses the default transport with default connection pool settings, which may not match your load. In production, always create an explicit `http.Client` with `Timeout` set and a configured `Transport` for appropriate connection pool limits.

---

**Q2: Why should you create one `http.Client` and reuse it?**

> `http.Client` contains an `http.Transport` which maintains a pool of persistent TCP connections per host. Creating a new client per request means creating a new transport with an empty pool — every request pays the cost of a new TCP handshake (and TLS handshake for HTTPS). Reusing one client allows connections to be reused across requests, dramatically reducing latency and server load. `http.Client` is safe for concurrent use from multiple goroutines.

---

**Q3: Why must you always close `resp.Body`?**

> The response body is a network connection wrapped in a reader. If you don't close it, the underlying TCP connection is never returned to the pool and is eventually abandoned — leaking file descriptors. For connection reuse, the body must also be fully read before closing. If you don't need the body, drain it with `io.Copy(io.Discard, resp.Body)` before closing. The pattern is: check error → `defer resp.Body.Close()` → check status → read body.

---

**Q4: What is the difference between `client.Timeout` and a context timeout?**

> `client.Timeout` is a deadline for the entire request lifecycle — from dialing through reading the last byte of the body — and applies to every request made with that client. A context timeout (`context.WithTimeout`) is per-request and can be shorter or longer than `client.Timeout`. The effective timeout is the minimum of the two — whichever expires first wins. Use `client.Timeout` as a safety net maximum; use context timeouts to set per-operation deadlines that propagate through the call stack.

---

**Q5: Which HTTP status codes should you retry and which should you not?**

> Retry on server-side and transient errors: 429 (Too Many Requests — throttled), 500 (Internal Server Error), 502 (Bad Gateway), 503 (Service Unavailable), 504 (Gateway Timeout). These indicate the server had a transient problem. Do not retry on client errors: 400 (Bad Request — fix your request), 401 (Unauthorized — fix auth), 403 (Forbidden), 404 (Not Found), 422 (Unprocessable Entity). Retrying client errors wastes resources and won't help. Also retry on network errors (connection refused, timeout) but not on context cancellation.

---

**Q6: How do you test code that makes HTTP requests?**

> Use `httptest.NewServer(handler)` to start a real local HTTP server on a random port. Point your client's base URL at `ts.URL`. The handler is a normal `http.Handler` where you verify the incoming request (path, headers, body) and return a controlled response. This tests the full request/response cycle without hitting real servers. Make sure to `defer ts.Close()`. For unit-testing at the function level, inject the `http.Client` as an interface or parameter so tests can substitute a fake transport.

---

## 🎉 Phase 3 Complete!

All 9 topics in **Phase 3 — Advanced** are done (after the final topic 3.9 Databases). You now know:

| Topic | Key Takeaway |
|-------|-------------|
| 3.1 Generics | Type params, constraints, `~T`, generic types, no method type params |
| 3.2 Context | Cancellation tree, `defer cancel()`, values, `r.Context()` |
| 3.3 Concurrency Patterns | Pipeline, fan-out/in, worker pool, errgroup, semaphore |
| 3.4 Reflection | TypeOf/ValueOf, Kind switch, CanSet, real-world uses |
| 3.5 unsafe | Sizeof/Alignof/Offsetof, pointer rules, uintptr danger |
| 3.6 Testing | Table-driven, t.Helper, benchmarks, fuzz, httptest |
| 3.7 HTTP Servers | Handler interface, middleware, ServeMux 1.22, graceful shutdown |
| 3.8 HTTP Clients | Client config, transport pooling, retries, context, testing |

*Next: **Topic 3.9 — Databases** then Phase 4 — Senior/Production Level!* 🚀
