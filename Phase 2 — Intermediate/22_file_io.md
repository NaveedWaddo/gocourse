# 2.10 — File I/O

> **Phase:** 2 — Intermediate | **Difficulty:** Intermediate  
> **Prerequisites:** Topic 2.9 — Stringer & fmt Interface

---

## 📖 Overview

Go provides three main packages for file and I/O operations:

- **`os`** — file system operations (open, create, delete, stat)
- **`io`** — core I/O interfaces and utilities (`io.Reader`, `io.Writer`, `io.Copy`)
- **`bufio`** — buffered I/O for efficient reading/writing line by line

Understanding `io.Reader` and `io.Writer` is essential — they're the interfaces that tie everything together.

---

## 🔑 The Core Interfaces — `io.Reader` and `io.Writer`

Everything in Go I/O is built on these two interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Any type that implements `Read` is a `Reader` — files, network connections, byte buffers, strings, HTTP bodies. Any type that implements `Write` is a `Writer` — files, stdout, network connections, byte buffers.

This means functions written against `io.Reader` work with **any** readable source:

```go
func countBytes(r io.Reader) (int64, error) {
    return io.Copy(io.Discard, r)   // works with file, HTTP body, buffer, anything
}
```

---

## 📂 Reading Files

### Simple read — entire file at once

```go
import "os"

// Read entire file into memory — simplest approach
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("readFile: %w", err)
}
fmt.Println(string(data))
```

### Open and read manually

```go
f, err := os.Open("data.txt")   // read-only
if err != nil {
    return fmt.Errorf("open: %w", err)
}
defer f.Close()   // always close!

data, err := io.ReadAll(f)
if err != nil {
    return fmt.Errorf("readAll: %w", err)
}
fmt.Println(string(data))
```

### Read in chunks

```go
f, err := os.Open("largefile.bin")
if err != nil { return err }
defer f.Close()

buf := make([]byte, 4096)   // 4KB buffer
for {
    n, err := f.Read(buf)
    if n > 0 {
        process(buf[:n])   // process n bytes read
    }
    if err == io.EOF {
        break   // end of file — normal termination
    }
    if err != nil {
        return fmt.Errorf("read: %w", err)
    }
}
```

> `io.EOF` is a **sentinel error** — it means "no more data", not a failure. Always check `n > 0` before checking `err` because the last read may return data AND `io.EOF` simultaneously.

---

## ✏️ Writing Files

### Simple write — entire content at once

```go
data := []byte("Hello, Go!\n")
err := os.WriteFile("output.txt", data, 0644)
if err != nil {
    return fmt.Errorf("writeFile: %w", err)
}
```

File permission `0644`:
- Owner: read + write
- Group: read only
- Others: read only

### Open and write manually

```go
// os.Create creates or truncates the file
f, err := os.Create("output.txt")
if err != nil {
    return fmt.Errorf("create: %w", err)
}
defer f.Close()

_, err = f.WriteString("Hello, Go!\n")
if err != nil {
    return fmt.Errorf("write: %w", err)
}

_, err = fmt.Fprintf(f, "Value: %d\n", 42)   // fmt.Fprintf works on any Writer
```

### Append to existing file

```go
f, err := os.OpenFile("log.txt",
    os.O_APPEND|os.O_CREATE|os.O_WRONLY,   // flags
    0644,                                   // permissions
)
if err != nil {
    return fmt.Errorf("openFile: %w", err)
}
defer f.Close()

fmt.Fprintln(f, "new log entry")
```

### `os.OpenFile` flags

| Flag | Meaning |
|------|---------|
| `os.O_RDONLY` | Read only |
| `os.O_WRONLY` | Write only |
| `os.O_RDWR` | Read and write |
| `os.O_CREATE` | Create if not exists |
| `os.O_TRUNC` | Truncate to zero on open |
| `os.O_APPEND` | Append on each write |
| `os.O_EXCL` | Error if file exists (with O_CREATE) |

---

## 📖 `bufio` — Buffered Reading

Buffered I/O groups small reads/writes into larger chunks, dramatically improving performance for line-by-line operations.

### Reading line by line — `bufio.Scanner`

```go
f, err := os.Open("data.txt")
if err != nil { return err }
defer f.Close()

scanner := bufio.NewScanner(f)
lineNum := 0
for scanner.Scan() {
    lineNum++
    line := scanner.Text()   // current line (without newline)
    fmt.Printf("%d: %s\n", lineNum, line)
}

if err := scanner.Err(); err != nil {
    return fmt.Errorf("scanner: %w", err)
}
```

> Always check `scanner.Err()` after the loop — `Scan()` returns `false` both on EOF (no error) and on errors.

### Scanner with custom split function

```go
// Scan word by word instead of line by line
scanner := bufio.NewScanner(f)
scanner.Split(bufio.ScanWords)

for scanner.Scan() {
    word := scanner.Text()
    fmt.Println(word)
}
```

Built-in split functions: `bufio.ScanLines` (default), `bufio.ScanWords`, `bufio.ScanBytes`, `bufio.ScanRunes`.

### Buffered reader — `bufio.Reader`

```go
reader := bufio.NewReader(f)

// Read a line including the delimiter
line, err := reader.ReadString('\n')

// Peek without consuming
peeked, err := reader.Peek(10)   // look at next 10 bytes

// Read single byte
b, err := reader.ReadByte()

// Read single rune
r, size, err := reader.ReadRune()
```

### Buffered writer — `bufio.Writer`

```go
f, err := os.Create("output.txt")
if err != nil { return err }
defer f.Close()

w := bufio.NewWriter(f)
fmt.Fprintln(w, "buffered line 1")
fmt.Fprintln(w, "buffered line 2")

// MUST flush — buffered writes don't hit disk until Flush
if err := w.Flush(); err != nil {
    return fmt.Errorf("flush: %w", err)
}
```

> Always call `Flush()` on a `bufio.Writer` — unwritten data stays in the buffer until flushed.

---

## 🔧 File System Operations

```go
// Check if file/dir exists
_, err := os.Stat("myfile.txt")
if os.IsNotExist(err) {
    fmt.Println("file does not exist")
}

// Get file info
info, err := os.Stat("myfile.txt")
fmt.Println("size:", info.Size())
fmt.Println("modified:", info.ModTime())
fmt.Println("is dir:", info.IsDir())
fmt.Println("permissions:", info.Mode())

// Rename / move
err = os.Rename("old.txt", "new.txt")

// Delete
err = os.Remove("file.txt")       // delete file
err = os.RemoveAll("mydir/")      // delete directory tree

// Create directory
err = os.Mkdir("newdir", 0755)            // one level
err = os.MkdirAll("a/b/c", 0755)         // all levels

// List directory contents
entries, err := os.ReadDir(".")
for _, entry := range entries {
    fmt.Printf("%s (dir: %v)\n", entry.Name(), entry.IsDir())
}

// Temp file
f, err := os.CreateTemp("", "prefix-*.txt")
defer os.Remove(f.Name())   // clean up

// Temp directory
dir, err := os.MkdirTemp("", "myapp-*")
defer os.RemoveAll(dir)
```

---

## 🔄 `io` Utilities

```go
// Copy from reader to writer — most important utility
n, err := io.Copy(dst, src)   // returns bytes copied

// Copy limited bytes
n, err := io.CopyN(dst, src, 1024)   // copy at most 1024 bytes

// Read exactly n bytes
buf := make([]byte, 100)
n, err := io.ReadFull(r, buf)   // error if fewer than 100 bytes available

// Discard — useful for draining a reader
io.Copy(io.Discard, resp.Body)   // drain HTTP response body

// Multi-reader — chain readers sequentially
r := io.MultiReader(r1, r2, r3)   // reads from r1, then r2, then r3

// Multi-writer — write to multiple destinations at once
w := io.MultiWriter(file, os.Stdout)   // writes go to both
io.Copy(w, src)

// TeeReader — read and copy simultaneously
tee := io.TeeReader(src, logWriter)   // reads from src, writes copy to logWriter
io.Copy(dst, tee)

// String reader
r := strings.NewReader("hello world")   // io.Reader from string

// Byte buffer — in-memory reader/writer
var buf bytes.Buffer
buf.WriteString("hello")
buf.WriteByte(' ')
fmt.Fprintf(&buf, "world %d", 42)
fmt.Println(buf.String())   // "hello world 42"
```

---

## 💻 Complete Example

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"
)

// wordCount counts words in any Reader
func wordCount(r io.Reader) (map[string]int, error) {
    counts := make(map[string]int)
    scanner := bufio.NewScanner(r)
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        word := strings.ToLower(scanner.Text())
        counts[word]++
    }
    return counts, scanner.Err()
}

// copyFile copies src to dst
func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("copyFile open src: %w", err)
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return fmt.Errorf("copyFile create dst: %w", err)
    }
    defer out.Close()

    if _, err := io.Copy(out, in); err != nil {
        return fmt.Errorf("copyFile copy: %w", err)
    }
    return out.Sync()   // flush to disk
}

func main() {
    // --- Write a file ---
    content := "the quick brown fox\njumps over the lazy dog\nthe fox ran away\n"
    if err := os.WriteFile("sample.txt", []byte(content), 0644); err != nil {
        fmt.Println("write error:", err)
        return
    }
    defer os.Remove("sample.txt")

    // --- Read line by line ---
    fmt.Println("=== Line by Line ===")
    f, _ := os.Open("sample.txt")
    scanner := bufio.NewScanner(f)
    for i := 1; scanner.Scan(); i++ {
        fmt.Printf("%d: %s\n", i, scanner.Text())
    }
    f.Close()

    // --- Word count via io.Reader ---
    fmt.Println("\n=== Word Count ===")
    r := strings.NewReader(content)
    counts, _ := wordCount(r)
    fmt.Println("'the' appears:", counts["the"])   // 3
    fmt.Println("'fox' appears:", counts["fox"])   // 2

    // --- Buffered write ---
    fmt.Println("\n=== Buffered Write ===")
    out, _ := os.Create("buffered.txt")
    defer os.Remove("buffered.txt")
    w := bufio.NewWriter(out)
    for i := 1; i <= 5; i++ {
        fmt.Fprintf(w, "line %d\n", i)
    }
    w.Flush()   // must flush!
    out.Close()

    data, _ := os.ReadFile("buffered.txt")
    fmt.Print(string(data))

    // --- io.MultiWriter — write to file AND stdout simultaneously ---
    fmt.Println("\n=== MultiWriter ===")
    logFile, _ := os.Create("multi.txt")
    defer os.Remove("multi.txt")
    defer logFile.Close()
    mw := io.MultiWriter(logFile, os.Stdout)
    fmt.Fprintln(mw, "this goes to file AND stdout")

    // --- File stat ---
    fmt.Println("\n=== File Info ===")
    info, _ := os.Stat("sample.txt")
    fmt.Printf("name: %s, size: %d bytes\n", info.Name(), info.Size())
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Not closing files — resource leak
```go
// ❌ File handle leaked if function returns early
f, _ := os.Open("file.txt")
// ... code that might return early ...
f.Close()

// ✅ Always defer Close immediately after Open
f, err := os.Open("file.txt")
if err != nil { return err }
defer f.Close()
```

### 2. Forgetting to flush buffered writer
```go
w := bufio.NewWriter(f)
fmt.Fprintln(w, "data")
// ❌ forgot w.Flush() — data stays in buffer, never written to file!

// ✅
defer w.Flush()   // or call explicitly before closing
```

### 3. Ignoring `scanner.Err()` after the loop
```go
for scanner.Scan() { ... }
// ❌ if scan stopped due to error, you'd never know
scanner.Err()   // ✅ always check this
```

### 4. Treating io.EOF as an error
```go
n, err := r.Read(buf)
if err == io.EOF {
    // ✅ this is NORMAL — end of data, not a failure
    break
}
if err != nil {
    return err   // actual error
}
```

### 5. Reading entire large files into memory
```go
// ❌ Will OOM if file is several GB
data, _ := os.ReadFile("huge.log")

// ✅ Stream with bufio.Scanner
scanner := bufio.NewScanner(f)
for scanner.Scan() { processLine(scanner.Text()) }
```

### 6. Not handling Close errors on write
```go
// ❌ Close errors on writable files are silently lost
defer f.Close()

// ✅ For files you write to, check Close error
defer func() {
    if cerr := f.Close(); cerr != nil && err == nil {
        err = cerr
    }
}()
```

---

## 🎯 Must-Know Things (Quick Reference)

- `os.ReadFile` / `os.WriteFile` — simplest one-shot read/write
- `os.Open` (read) / `os.Create` (write) / `os.OpenFile` (full control)
- **Always `defer f.Close()`** immediately after successful open
- `io.Reader` and `io.Writer` — the universal interfaces tying all I/O together
- `bufio.Scanner` — idiomatic line-by-line reading; always check `scanner.Err()`
- `bufio.Writer` — **must call `Flush()`** or data is lost
- `io.EOF` is not an error — it means "end of data", handle it separately
- `io.Copy(dst, src)` — copy any reader to any writer efficiently
- `io.MultiWriter` — fan out writes to multiple destinations
- `strings.NewReader` / `bytes.Buffer` — in-memory readers/writers for testing
- Never read entire large files into memory — stream with `bufio.Scanner`

---

## ❓ Interview Questions & Answers

---

**Q1: What is the relationship between `os`, `io`, and `bufio` packages?**

> `os` provides OS-level file operations — opening, creating, deleting files, returning `*os.File` which implements both `io.Reader` and `io.Writer`. The `io` package defines the core interfaces (`Reader`, `Writer`) and utilities (`Copy`, `ReadAll`, `MultiWriter`) that work with any implementation. `bufio` wraps any `io.Reader` or `io.Writer` with buffering — grouping many small reads/writes into fewer, larger system calls for better performance. Together they form a layered I/O system: `os` for file handles, `io` for interfaces and utilities, `bufio` for efficiency.

---

**Q2: Why is `io.Reader` such an important interface in Go?**

> `io.Reader`'s single method `Read(p []byte) (n int, err error)` is implemented by files, network connections, HTTP request/response bodies, byte buffers, string readers, zip archives, and more. Functions written against `io.Reader` work with any of these sources without modification — a function that counts bytes works on a file today and an HTTP body tomorrow. This is Go's composition-based approach to I/O: small interfaces that compose into powerful abstractions.

---

**Q3: What is the difference between `bufio.Scanner` and `bufio.Reader`?**

> `bufio.Scanner` is higher-level — it splits input into tokens (lines by default) and provides a simple `Scan()` / `Text()` API. It's the idiomatic way to read line-by-line. `bufio.Reader` is lower-level — it buffers reads and provides methods like `ReadString`, `ReadByte`, `ReadRune`, and `Peek`. Use `Scanner` for simple line/word/token iteration; use `Reader` when you need more control — like reading until a specific delimiter or peeking ahead.

---

**Q4: What is `io.EOF` and how should you handle it?**

> `io.EOF` is a sentinel error from the `io` package signaling end of input — no more data to read. It is NOT a failure — it's normal termination. When `Read` returns `(n, io.EOF)`, `n` bytes were read AND the end was reached — you should process those bytes before stopping. Never return `io.EOF` to your callers as an error; translate it to a nil error or a domain-specific signal. `bufio.Scanner` handles this automatically — you only call `scanner.Err()` to check for real errors.

---

**Q5: Why must you call `Flush()` on a `bufio.Writer`?**

> `bufio.Writer` accumulates writes in an internal buffer and only flushes to the underlying writer when the buffer is full or `Flush()` is called explicitly. If you don't call `Flush()` before closing the file, buffered data that hasn't been written yet is simply lost — the file ends up incomplete with no error. Always call `Flush()`, ideally with `defer w.Flush()` right after creating the writer, or check its error explicitly.

---

**Q6: How does `io.Copy` work and why is it preferred over manual read/write loops?**

> `io.Copy(dst, src)` reads from `src` in chunks and writes to `dst`, returning the total bytes copied and any error. It handles the read-write loop, buffer management, and `io.EOF` detection internally. It's preferred because: it's correct (handles partial reads, EOF detection), efficient (uses an internal 32KB buffer or calls `WriteTo`/`ReadFrom` for zero-copy when available), and concise. Writing a manual read/write loop correctly — handling partial reads, distinguishing EOF from errors — is surprisingly tricky.

---

**Q7: How do you test code that reads from or writes to files?**

> Use `io.Reader` and `io.Writer` interfaces in function signatures instead of `*os.File`. In tests, substitute `strings.NewReader("test data")` for file reading and `bytes.Buffer` for file writing — no real files needed. This makes tests fast, hermetic, and parallelizable. For functions that need actual file paths, use `os.CreateTemp` to create a real temp file in the test and `defer os.Remove(f.Name())` to clean up. The key insight is that if your code depends on `io.Reader` not `*os.File`, it's automatically testable with in-memory data.

---

*Next up: **Topic 2.11 — Working with JSON** 🚀*
