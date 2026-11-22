# 3.9 — Databases

> **Phase:** 3 — Advanced | **Difficulty:** Advanced  
> **Prerequisites:** Topic 3.8 — HTTP Clients

---

## 📖 Overview

Go's `database/sql` package provides a generic interface for SQL databases. It handles connection pooling, prepared statements, transactions, and context propagation. You pick the driver separately.

Core concepts:
- `database/sql` — the standard interface
- Connection pooling — critical configuration
- Querying — `Query`, `QueryRow`, `Exec`
- Scanning rows into structs
- Transactions — `Begin`, `Commit`, `Rollback`
- Prepared statements
- `sqlx` — a popular extension for struct scanning
- `pgx` — the PostgreSQL-native driver

---

## 🔌 Connecting to a Database

```go
import (
    "database/sql"
    _ "github.com/lib/pq"          // PostgreSQL driver (blank import = register driver)
    _ "github.com/go-sql-driver/mysql"  // or MySQL
    _ "github.com/mattn/go-sqlite3"     // or SQLite
)

func NewDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)                  // max open connections
    db.SetMaxIdleConns(10)                  // max idle connections in pool
    db.SetConnMaxLifetime(5 * time.Minute)  // max time a connection can be reused
    db.SetConnMaxIdleTime(1 * time.Minute)  // max time a connection can be idle

    // Verify connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }
    return db, nil
}
```

### DSN formats

```go
// PostgreSQL
dsn := "postgres://user:password@localhost:5432/mydb?sslmode=disable"
// or
dsn := "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable"

// MySQL
dsn := "user:password@tcp(localhost:3306)/mydb?parseTime=true"

// SQLite
dsn := "file:./myapp.db?cache=shared&mode=rwc"
```

---

## 🔗 `sql.Open` vs `sql.DB`

```go
// sql.Open does NOT create a connection — just validates DSN format
db, err := sql.Open("postgres", dsn)

// sql.DB IS the connection pool — never close and reopen it
// Create once at startup, pass everywhere, close at shutdown
defer db.Close()
```

`sql.DB` is goroutine-safe — share it across your whole application.

---

## 📊 Connection Pool — Critical Configuration

```go
// Rule of thumb for PostgreSQL:
db.SetMaxOpenConns(25)         // ~= (num_CPU * 2) to (num_CPU * 4)
db.SetMaxIdleConns(10)         // ~= MaxOpenConns / 2
db.SetConnMaxLifetime(5 * time.Minute)   // prevent stale connections
db.SetConnMaxIdleTime(1 * time.Minute)   // release idle connections faster
```

| Setting | Too Low | Too High |
|---------|---------|----------|
| `MaxOpenConns` | Queue builds up, latency spikes | DB overwhelmed |
| `MaxIdleConns` | Too many new connections | Memory wasted |
| `ConnMaxLifetime` | N/A | Stale / closed connections |

---

## 🔍 Querying

### Single row — `QueryRowContext`

```go
type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}

func getUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`

    var u User
    err := db.QueryRowContext(ctx, query, id).Scan(
        &u.ID, &u.Name, &u.Email, &u.CreatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user %d: not found", id)
    }
    if err != nil {
        return nil, fmt.Errorf("getUser: %w", err)
    }
    return &u, nil
}
```

### Multiple rows — `QueryContext`

```go
func listUsers(ctx context.Context, db *sql.DB) ([]User, error) {
    query := `SELECT id, name, email, created_at FROM users ORDER BY id`

    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("listUsers query: %w", err)
    }
    defer rows.Close()   // ALWAYS close rows

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, fmt.Errorf("listUsers scan: %w", err)
        }
        users = append(users, u)
    }
    // ALWAYS check rows.Err() after the loop
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("listUsers rows: %w", err)
    }
    return users, nil
}
```

### Insert / Update / Delete — `ExecContext`

```go
func createUser(ctx context.Context, db *sql.DB, name, email string) (int, error) {
    query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`

    var id int
    err := db.QueryRowContext(ctx, query, name, email).Scan(&id)
    if err != nil {
        return 0, fmt.Errorf("createUser: %w", err)
    }
    return id, nil
}

func updateUser(ctx context.Context, db *sql.DB, id int, name string) error {
    query := `UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2`

    result, err := db.ExecContext(ctx, query, name, id)
    if err != nil {
        return fmt.Errorf("updateUser: %w", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("rowsAffected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("user %d: not found", id)
    }
    return nil
}

func deleteUser(ctx context.Context, db *sql.DB, id int) error {
    _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
    return err
}
```

---

## 🔒 Transactions

```go
func transferFunds(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,   // isolation level
        ReadOnly:  false,
    })
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    // Always defer rollback — if Commit is called first, Rollback is a no-op
    defer tx.Rollback()

    // Deduct from sender
    _, err = tx.ExecContext(ctx,
        `UPDATE accounts SET balance = balance - $1 WHERE id = $2`,
        amount, fromID)
    if err != nil {
        return fmt.Errorf("deduct: %w", err)
    }

    // Add to receiver
    _, err = tx.ExecContext(ctx,
        `UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
        amount, toID)
    if err != nil {
        return fmt.Errorf("credit: %w", err)
    }

    // Verify sender balance non-negative
    var balance float64
    err = tx.QueryRowContext(ctx,
        `SELECT balance FROM accounts WHERE id = $1`, fromID).Scan(&balance)
    if err != nil {
        return fmt.Errorf("check balance: %w", err)
    }
    if balance < 0 {
        return fmt.Errorf("insufficient funds")
    }

    return tx.Commit()
}
```

### The defer-rollback pattern

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()   // ✅ safe — no-op after successful Commit

// ... do work ...

return tx.Commit()   // if this succeeds, Rollback above is a no-op
```

---

## 📝 Prepared Statements

Prepared statements pre-compile SQL on the database — faster for repeated execution and safe against SQL injection:

```go
// Prepare once, use many times
stmt, err := db.PrepareContext(ctx, `SELECT id, name FROM users WHERE email = $1`)
if err != nil {
    return fmt.Errorf("prepare: %w", err)
}
defer stmt.Close()

// Use in a loop — no re-parsing overhead
for _, email := range emails {
    var id int
    var name string
    err := stmt.QueryRowContext(ctx, email).Scan(&id, &name)
    if err == sql.ErrNoRows { continue }
    if err != nil { return err }
    fmt.Printf("%d: %s (%s)\n", id, name, email)
}
```

> Note: `database/sql` also transparently prepares statements for you — explicit `Prepare` is mainly useful in loops.

---

## 🗃️ Nullable Values — `sql.NullString`, `sql.NullInt64`

When database columns can be NULL, use `sql.Null*` types:

```go
type User struct {
    ID    int
    Name  string
    Bio   sql.NullString   // nullable VARCHAR
    Age   sql.NullInt64    // nullable INTEGER
}

var u User
err := db.QueryRowContext(ctx, `SELECT id, name, bio, age FROM users WHERE id=$1`, id).
    Scan(&u.ID, &u.Name, &u.Bio, &u.Age)

if u.Bio.Valid {
    fmt.Println("bio:", u.Bio.String)
} else {
    fmt.Println("bio: (null)")
}
```

Or use pointers (simpler with `sqlx`):

```go
type User struct {
    ID   int
    Name string
    Bio  *string   // nil = NULL
    Age  *int
}
```

---

## 🚀 `sqlx` — Struct Scanning Made Easy

`sqlx` extends `database/sql` with struct scanning using field tags:

```go
import "github.com/jmoiron/sqlx"

type User struct {
    ID        int       `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}

db, _ := sqlx.Connect("postgres", dsn)

// Get a single row directly into struct
var u User
err := db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
if err == sql.ErrNoRows { /* not found */ }

// Get multiple rows into a slice
var users []User
err = db.SelectContext(ctx, &users, `SELECT * FROM users ORDER BY id`)

// Named queries — use struct fields as parameters
query := `INSERT INTO users (name, email) VALUES (:name, :email)`
_, err = db.NamedExecContext(ctx, query, User{Name: "Alice", Email: "alice@example.com"})
```

`sqlx` is a thin, battle-tested wrapper — it doesn't hide SQL, just eliminates repetitive scanning boilerplate.

---

## 🐘 `pgx` — Native PostgreSQL Driver

`pgx` is a PostgreSQL-specific driver that offers better performance, type support, and features:

```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Connection pool (preferred over single connection)
pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost/mydb")
if err != nil { log.Fatal(err) }
defer pool.Close()

// Configure pool
config, _ := pgxpool.ParseConfig(dsn)
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 5 * time.Minute
pool, err = pgxpool.NewWithConfig(ctx, config)

// Query
rows, err := pool.Query(ctx, `SELECT id, name FROM users`)
defer rows.Close()
for rows.Next() {
    var id int; var name string
    rows.Scan(&id, &name)
    fmt.Println(id, name)
}

// pgx v5: CollectRows — no manual scanning loop
rows, err = pool.Query(ctx, `SELECT id, name FROM users`)
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

### pgx advantages over `lib/pq`

- Batch queries
- Native `COPY` protocol for bulk inserts
- Better type support (arrays, JSON, custom types)
- `pgxpool` with health checks
- Generally faster

---

## 🏗️ Repository Pattern

Separate database access into a repository layer:

```go
// Interface — testable, swappable
type UserRepository interface {
    GetByID(ctx context.Context, id int) (*User, error)
    List(ctx context.Context) ([]User, error)
    Create(ctx context.Context, name, email string) (*User, error)
    Update(ctx context.Context, id int, name string) error
    Delete(ctx context.Context, id int) error
}

// PostgreSQL implementation
type postgresUserRepo struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int) (*User, error) {
    var u User
    err := r.db.QueryRowContext(ctx,
        `SELECT id, name, email, created_at FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user %d not found", id)
    }
    return &u, err
}

// Service uses repository — doesn't know about SQL
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
    return s.repo.GetByID(ctx, id)
}
```

### Testing with a fake repository

```go
type fakeUserRepo struct {
    users map[int]*User
}

func (f *fakeUserRepo) GetByID(_ context.Context, id int) (*User, error) {
    u, ok := f.users[id]
    if !ok { return nil, fmt.Errorf("user %d not found", id) }
    return u, nil
}
// ... implement other methods

func TestUserService(t *testing.T) {
    repo := &fakeUserRepo{users: map[int]*User{
        1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
    }}
    svc := &UserService{repo: repo}
    u, err := svc.GetUser(context.Background(), 1)
    if err != nil { t.Fatal(err) }
    if u.Name != "Alice" { t.Errorf("name = %q; want Alice", u.Name) }
}
```

---

## 💻 Complete Example

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/lib/pq"
)

type Product struct {
    ID        int
    Name      string
    Price     float64
    CreatedAt time.Time
}

type ProductRepo struct{ db *sql.DB }

func (r *ProductRepo) Create(ctx context.Context, name string, price float64) (*Product, error) {
    var p Product
    err := r.db.QueryRowContext(ctx,
        `INSERT INTO products (name, price) VALUES ($1, $2)
         RETURNING id, name, price, created_at`,
        name, price,
    ).Scan(&p.ID, &p.Name, &p.Price, &p.CreatedAt)
    if err != nil { return nil, fmt.Errorf("create product: %w", err) }
    return &p, nil
}

func (r *ProductRepo) List(ctx context.Context) ([]Product, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, name, price, created_at FROM products ORDER BY id`)
    if err != nil { return nil, fmt.Errorf("list products: %w", err) }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CreatedAt); err != nil {
            return nil, fmt.Errorf("scan: %w", err) 
        }
        products = append(products, p)
    }
    return products, rows.Err()
}

func (r *ProductRepo) BulkCreate(ctx context.Context, products []Product) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil { return fmt.Errorf("begin: %w", err) }
    defer tx.Rollback()

    stmt, err := tx.PrepareContext(ctx,
        `INSERT INTO products (name, price) VALUES ($1, $2)`)
    if err != nil { return fmt.Errorf("prepare: %w", err) }
    defer stmt.Close()

    for _, p := range products {
        if _, err := stmt.ExecContext(ctx, p.Name, p.Price); err != nil {
            return fmt.Errorf("insert %s: %w", p.Name, err)
        }
    }
    return tx.Commit()
}

func main() {
    dsn := "postgres://user:pass@localhost/mydb?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil { log.Fatal(err) }
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)

    ctx := context.Background()
    if err := db.PingContext(ctx); err != nil {
        log.Fatal("ping:", err)
    }

    repo := &ProductRepo{db: db}

    // Create
    p, err := repo.Create(ctx, "Go Book", 39.99)
    if err != nil { log.Fatal(err) }
    fmt.Printf("Created: #%d %s $%.2f\n", p.ID, p.Name, p.Price)

    // Bulk create in transaction
    err = repo.BulkCreate(ctx, []Product{
        {Name: "Rust Book", Price: 44.99},
        {Name: "Python Book", Price: 29.99},
    })
    if err != nil { log.Fatal(err) }

    // List
    products, err := repo.List(ctx)
    if err != nil { log.Fatal(err) }
    for _, prod := range products {
        fmt.Printf("  - %s: $%.2f\n", prod.Name, prod.Price)
    }
}
```

---

## ⚠️ Common Gotchas & Pitfalls

### 1. Not closing `rows` — connection leak
```go
rows, _ := db.QueryContext(ctx, query)
// ❌ forgot defer rows.Close() — connection held until GC

// ✅
rows, err := db.QueryContext(ctx, query)
if err != nil { return err }
defer rows.Close()
```

### 2. Ignoring `rows.Err()` — silent data truncation
```go
for rows.Next() { ... }
// ❌ if iteration stopped due to network error, you'd never know
return users, nil

// ✅
return users, rows.Err()
```

### 3. SQL injection — never interpolate user input
```go
// ❌ NEVER — SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", userInput)
db.QueryContext(ctx, query)

// ✅ Always use parameterized queries
db.QueryContext(ctx, "SELECT * FROM users WHERE name = $1", userInput)
```

### 4. Not using context — unresponsive queries run forever
```go
// ❌ No context — query runs until completion, ignores cancellation
db.Query(query)

// ✅ Always pass context
db.QueryContext(ctx, query)
```

### 5. Forgetting `defer tx.Rollback()` — held transactions
```go
tx, _ := db.BeginTx(ctx, nil)
// ❌ if Commit fails, transaction is never rolled back — connection held

// ✅ Always defer Rollback — safe even after Commit
defer tx.Rollback()
```

### 6. Opening too many connections
```go
// ❌ Default MaxOpenConns = 0 (unlimited) — can overwhelm PostgreSQL (default max 100)
db.SetMaxOpenConns(0)   // unlimited!

// ✅ Set explicit limit
db.SetMaxOpenConns(25)
```

---

## 🎯 Must-Know Things (Quick Reference)

- `sql.Open` validates DSN — **doesn't connect**; `db.PingContext` verifies connectivity
- `sql.DB` is a **connection pool** — create once, share globally, never close and reopen
- Always configure: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`
- `QueryRowContext` → single row + `Scan`; check `sql.ErrNoRows` separately
- `QueryContext` → multiple rows; **always `defer rows.Close()` + check `rows.Err()`**
- `ExecContext` → INSERT/UPDATE/DELETE; check `RowsAffected()` if needed
- **Never interpolate user input** — always use parameterized queries (`$1`, `?`)
- Transaction pattern: `BeginTx` → `defer tx.Rollback()` → work → `tx.Commit()`
- `defer tx.Rollback()` is safe after `Commit` — it's a no-op
- `sql.NullString` / `sql.NullInt64` — or pointer fields — for nullable columns
- `sqlx` — struct scanning with `db` tags; `Get` and `Select` eliminate boilerplate
- `pgx` + `pgxpool` — preferred for PostgreSQL-specific projects
- Repository pattern — hide SQL behind an interface for testability

---

## ❓ Interview Questions & Answers

---

**Q1: What is `sql.DB` and why should you create it once and share it?**

> `sql.DB` is not a single connection — it's a **connection pool** that manages a set of database connections. It handles opening new connections when needed, returning idle connections to the pool, and closing connections that have been idle too long. Creating it once and sharing it across the application means all code shares the same pool, staying within configured limits. Creating a new `sql.DB` per request creates a new pool each time — connections can't be reused across requests, destroying the performance benefits of pooling. `sql.DB` is safe for concurrent use.

---

**Q2: What happens if you don't close `sql.Rows`?**

> The underlying database connection is held open and not returned to the pool until `rows.Close()` is called (or all rows are iterated and `rows.Next()` returns false, which calls close internally). If you return early from a loop without closing, that connection is tied up until the garbage collector eventually collects the `rows` object — potentially exhausting the connection pool. Always `defer rows.Close()` immediately after a successful `QueryContext` call.

---

**Q3: Why must you always use parameterized queries?**

> Parameterized queries send the SQL and the values separately to the database — the DB driver handles escaping, so user-supplied values can never alter the SQL structure. String interpolation with `fmt.Sprintf` directly builds a SQL string, so a malicious value like `'; DROP TABLE users; --` becomes part of the SQL itself — a SQL injection attack. Parameterized queries also enable the database to cache query plans, improving performance for repeated queries.

---

**Q4: Explain the transaction pattern with `defer tx.Rollback()`.**

> After `db.BeginTx`, immediately defer `tx.Rollback()`. If any subsequent operation fails and the function returns an error, the deferred Rollback undoes all work in the transaction — leaving the DB in a consistent state. If all operations succeed and `tx.Commit()` is called, the deferred `tx.Rollback()` is still called — but it's a no-op after a successful commit. This pattern ensures the transaction is always either committed or rolled back, never left dangling, even if a panic occurs.

---

**Q5: What is the difference between `QueryContext`, `QueryRowContext`, and `ExecContext`?**

> `QueryContext` executes a query that returns multiple rows — returns `*sql.Rows` which you iterate with `rows.Next()`. `QueryRowContext` is a convenience for queries returning exactly one row — returns `*sql.Row` which you call `.Scan()` on directly; it returns `sql.ErrNoRows` if no row matched. `ExecContext` is for statements that don't return rows — INSERT, UPDATE, DELETE — it returns `sql.Result` with `RowsAffected()` and `LastInsertId()`. Using the wrong one still works but is semantically misleading and may perform worse.

---

**Q6: What does `rows.Err()` check and why is it important?**

> `rows.Next()` returns false both when iteration is complete (normal) and when an error occurs mid-iteration (network failure, context cancellation, etc.). There's no way to distinguish these from the `rows.Next()` return value alone. After the loop, `rows.Err()` returns any error that stopped iteration — if it's non-nil, your result slice is incomplete. Always check `rows.Err()` and return it if non-nil. Ignoring it can cause silent data truncation — your code thinks it fetched all users but only got the first 50 before a network hiccup.

---

**Q7: When would you use `sqlx` or `pgx` instead of plain `database/sql`?**

> Use `sqlx` when you want to eliminate repetitive scanning boilerplate — it adds `Get` and `Select` that scan directly into structs using `db` struct tags, and `NamedExec` for using struct fields as query parameters. It's a thin wrapper — still uses `database/sql` underneath, so all standard SQL patterns apply. Use `pgx` (specifically `pgxpool`) when you're PostgreSQL-only and need better performance, PostgreSQL-specific types (arrays, JSONB, custom types), batch queries, the COPY protocol for bulk inserts, or more granular connection pool configuration. Both are widely used in production Go services.

---

## 🎉 Phase 3 Complete!

All 9 topics in **Phase 3 — Advanced** are done. Full summary:

| Topic | Key Takeaway |
|-------|-------------|
| 3.1 Generics | Type params, constraints, `~T`, generic structs, no method type params |
| 3.2 Context | Cancellation tree, `defer cancel()`, values via unexported keys |
| 3.3 Concurrency Patterns | Pipeline, fan-out/in, worker pool, errgroup, semaphore |
| 3.4 Reflection | TypeOf/ValueOf, Kind switch, CanSet, 10-100x slower than direct code |
| 3.5 unsafe | Sizeof/Offsetof, uintptr danger, type punning, struct padding |
| 3.6 Testing | Table-driven, t.Helper, benchmarks, fuzz testing, httptest |
| 3.7 HTTP Servers | Handler interface, middleware chain, ServeMux 1.22, graceful shutdown |
| 3.8 HTTP Clients | Client+Transport config, context, retry on 5xx, connection pooling |
| 3.9 Databases | sql.DB pool, QueryContext/Exec/Tx, parameterized queries, repository pattern |

**Next: Phase 4 — Senior / Production Level** 🚀
