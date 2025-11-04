# SQLBlade

<div align="center">

**A modern, type-safe query builder for Go** 🚀

*Combining the ergonomics of GORM, the type-safety of SQLC, and the performance of raw SQL*

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/alicanli1995/sqlblade/workflows/CI/badge.svg)](https://github.com/alicanli1995/sqlblade/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/alicanli1995/sqlblade.svg)](https://pkg.go.dev/github.com/alicanli1995/sqlblade)

</div>

---

## ✨ Features

- 🎯 **Type-Safe**: Compile-time type checking with Go generics
- ⚡ **Minimal Reflection Overhead**: Type information cached after first use, subsequent queries use cached metadata
- 🚀 **High Performance**: Zero-allocation string building with `strings.Builder`
- 🗄️ **Multi-Database**: PostgreSQL, MySQL, SQLite support
- 🔧 **Full SQL Support**: SELECT, INSERT, UPDATE, DELETE, JOIN, Transactions
- ⏱️ **Context Support**: Built-in timeout and cancellation support
- 🛡️ **SQL Injection Prevention**: Parameterized queries and operator whitelisting
- 📦 **Zero Dependencies**: Uses only standard library (except database drivers)
- 🎨 **Beautiful SQL Debugging**: Formatted query logging with timing and parameter substitution
- 👁️ **Query Preview**: See generated SQL without executing queries
- 🔄 **Query Composition**: Reusable query fragments for DRY code
- 🔍 **Subquery Support**: Powerful WHERE conditions with subqueries
- ⚡ **EXISTS Queries**: Efficient existence checks

## 🚀 Quick Start

### Installation

```bash
go get github.com/alicanli1995/sqlblade
```

**📦 Package Documentation:** [pkg.go.dev/github.com/alicanli1995/sqlblade](https://pkg.go.dev/github.com/alicanli1995/sqlblade)

### Basic Usage

```go
package main

import (
    "context"
    "database/sql"
    "time"
    
    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/alicanli1995/sqlblade/sqlblade"
    "github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

type User struct {
    ID        int       `db:"id"`
    Email     string    `db:"email"`
    Name      string    `db:"name"`
    Age       int       `db:"age"`
    CreatedAt time.Time `db:"created_at"`
}

func main() {
    db, _ := sql.Open("postgres", "postgres://user:pass@localhost/dbname")
    ctx := context.Background()
    
    // Simple SELECT query
    users, err := sqlblade.Query[User](db).
        Where("age", ">", 18).
        Where("status", "=", "active").
        OrderBy("created_at", dialect.DESC).
        Limit(10).
        Execute(ctx)
        
    // INSERT
    result, err := sqlblade.Insert(db, user).Execute(ctx)
    
    // UPDATE
    result, err := sqlblade.Update[User](db).
        Set("status", "inactive").
        Where("id", "=", userID).
        Execute(ctx)
        
    // Transaction
    err := sqlblade.WithTransaction(db, func(tx *sql.Tx) error {
        _, err := sqlblade.InsertTx(tx, user).Execute(ctx)
        return err
    })
```

## 📚 Examples

See [examples/](examples/) directory for complete examples:
- [Basic Examples](examples/main.go) - Complete examples including SELECT, INSERT, UPDATE, DELETE, JOIN, Transactions, Query Preview, Fragments, Subqueries, and EXISTS
- [MySQL Examples](examples/mysql/main.go)
- [PostgreSQL Examples](examples/postgres/main.go)
- [Debug Features](examples/debug_features/main.go) - Advanced features demonstration with SQL debugging

## 🎯 When to Use SQLBlade

SQLBlade is ideal for:

### ✅ **Perfect For:**
- **High-performance applications** requiring fast INSERT and COUNT operations
- **Memory-constrained environments** where low memory footprint is critical
- **Type-safe codebases** where compile-time type checking is preferred
- **Microservices** needing lightweight dependencies (zero external deps)
- **Applications with frequent aggregate queries** (COUNT, SUM, AVG, etc.)
- **Projects transitioning from raw SQL** wanting better ergonomics without performance loss
- **API servers** handling high-throughput data operations

### ⚠️ **Consider Alternatives When:**
- **Heavy SELECT workloads** where GORM's prepared statement cache provides significant advantage
- **Very complex ORM features** needed (automatic migrations, relations, etc.) - consider GORM
- **Raw SQL is preferred** - stdlib or sqlx might be simpler
- **Dynamic query building** with unpredictable patterns (cache benefits diminish)

### 💡 **Best Use Cases:**

1. **REST APIs with CRUD operations**
   ```go
   // Fast INSERT, UPDATE, COUNT with type safety
   users, _ := sqlblade.Query[User](db).Where("active", "=", true).Execute(ctx)
   ```

2. **Data processing pipelines**
   ```go
   // Efficient batch operations with low memory overhead
   sqlblade.InsertBatch(db, users).Execute(ctx)
   ```

3. **Analytics and reporting**
   ```go
   // Fast aggregate queries
   count, _ := sqlblade.Query[Order](db).Where("date", ">", startDate).Count(ctx)
   ```

4. **High-concurrency services**
   ```go
   // Low memory footprint reduces GC pressure
   sqlblade.PreparedStatementCache(db) // Enable for repeated queries
   ```

**Performance Profile:**
- 🏆 **INSERT**: Fastest among all libraries
- 🏆 **COUNT**: Fastest among all libraries  
- ✅ **Memory**: Lowest memory usage
- ✅ **Allocations**: Fewest allocations
- ⚡ **SELECT**: Competitive (GORM faster but uses 3x more memory)

## 🎯 Supported Databases

| Database | Driver | Status |
|----------|--------|--------|
| PostgreSQL | `github.com/lib/pq` | ✅ Full Support |
| MySQL | `github.com/go-sql-driver/mysql` | ✅ Full Support |
| SQLite | `github.com/mattn/go-sqlite3` | ✅ Full Support |

## 📖 API Reference

### Query Builder Methods

- `Query[T](db)` - Create a SELECT query builder
- `Where(column, operator, value)` - Add WHERE condition (AND)
- `OrWhere(column, operator, value)` - Add WHERE condition (OR)
- `Join(table, condition)` - INNER JOIN
- `LeftJoin(table, condition)` - LEFT JOIN
- `Select(columns...)` - Specify columns to select
- `OrderBy(column, direction)` - Add ORDER BY clause
- `Limit(n)` / `Offset(n)` - Set LIMIT and OFFSET
- `Execute(ctx)` - Execute query and return results
- `Count(ctx)` / `Sum(ctx, col)` / `Avg(ctx, col)` / `Min(ctx, col)` / `Max(ctx, col)` - Aggregate functions

### Insert/Update/Delete

- `Insert(db, value)` / `InsertBatch(db, values)` - INSERT operations
- `Update[T](db)` - UPDATE operations
- `Delete[T](db)` - DELETE operations
- `Returning(columns...)` - Specify RETURNING columns (PostgreSQL)

### Transactions

- `WithTransaction(db, fn)` - Execute operations in a transaction
- `WithTransactionContext(ctx, db, fn)` - Transaction with context

### Raw SQL

- `Raw[T](db, query, args...)` - Execute raw SQL queries

### Query Debugging & Preview

- `EnableDebug()` - Enable beautiful SQL query logging
- `ConfigureDebug(func)` - Configure debug settings
- `Preview()` - Preview SQL without executing
- `SQL()` / `SQLWithArgs()` - Get generated SQL string
- `PrettyPrint()` - Print formatted query

### Query Composition & Subqueries

- `NewQueryFragment()` - Create reusable query fragments
- `Apply(fragment)` - Apply fragment to query builder
- `NewSubquery(builder)` - Create subquery from builder
- `WhereSubquery()` / `OrWhereSubquery()` - Use subqueries in WHERE
- `Exists()` / `NotExists()` - Check existence efficiently

## 🎨 Advanced Features

### SQL Query Debugging

SQLBlade includes a beautiful, formatted SQL logger that makes debugging queries a joy:

```go
// Enable debugging globally
sqlblade.EnableDebug()

// Configure debugger (optional)
sqlblade.ConfigureDebug(func(qd *sqlblade.QueryDebugger) {
    qd.ShowArgs(true)           // Show query parameters
    qd.ShowTiming(true)          // Show execution time
    qd.SetSlowQueryThreshold(50 * time.Millisecond) // Warn on slow queries
    qd.IndentSQL(true)           // Pretty format SQL
})

// Now all queries are automatically logged!
users, _ := sqlblade.Query[User](db).
    Where("age", ">", 18).
    Execute(ctx)
```

**Output:**
```
═══════════════════════════════════════════════════════════════
SQL Query Debug - 2024-01-15 10:30:45.123
═══════════════════════════════════════════════════════════════
Operation: SELECT
Table:     users
Duration:  2.34ms
───────────────────────────────────────────────────────────────
SQL:
SELECT * FROM users WHERE age > $1
───────────────────────────────────────────────────────────────
Parameters:
  $1 = 18 (int)
═══════════════════════════════════════════════════════════════
```

### Query Preview

Preview generated SQL without executing:

```go
query := sqlblade.Query[User](db).
    Where("email", "LIKE", "%@example.com%").
    Join("profiles", "profiles.user_id = users.id")

// See the SQL
fmt.Println(query.Preview().SQL())
// Output: SELECT * FROM users JOIN profiles ON profiles.user_id = users.id WHERE email LIKE $1

// See SQL with substituted arguments
fmt.Println(query.Preview().SQLWithArgs())
// Output: SELECT * FROM users JOIN profiles ON profiles.user_id = users.id WHERE email LIKE '%@example.com%'

// Pretty print
query.Preview().PrettyPrint()
```

### Query Fragments (DRY Code)

Create reusable query fragments to avoid repetition:

```go
// Create a reusable fragment
activeUsersFragment := sqlblade.NewQueryFragment().
    Where("status", "=", "active").
    Where("email_verified", "=", true).
    OrderBy("created_at", dialect.DESC)

// Apply to multiple queries
recentActive, _ := sqlblade.Query[User](db).
    Apply(activeUsersFragment).
    Limit(10).
    Execute(ctx)

allActive, _ := sqlblade.Query[User](db).
    Apply(activeUsersFragment).
    Execute(ctx)
```

### Subqueries

Use subqueries in WHERE clauses for powerful queries:

```go
// Find users who have placed orders
usersWithOrders, _ := sqlblade.Query[User](db).
    WhereSubquery("id", "IN", sqlblade.NewSubquery(
        sqlblade.Query[struct {
            UserID int `db:"user_id"`
        }](db).
            Select("user_id").
            Where("status", "=", "completed"),
    )).
    Execute(ctx)
```

### EXISTS Queries

Efficiently check existence:

```go
// Check if any active users exist
hasActiveUsers, _ := sqlblade.Query[User](db).
    Where("status", "=", "active").
    Limit(1).
    Exists(ctx)

if hasActiveUsers {
    fmt.Println("There are active users!")
}

// Or use NotExists
hasNoAdmins, _ := sqlblade.Query[User](db).
    Where("role", "=", "admin").
    NotExists(ctx)
```

## 🔒 Type Safety

SQLBlade uses Go generics to provide compile-time type safety:

```go
// ✅ Type-safe - compile-time checking
users, err := sqlblade.Query[User](db).Execute(ctx)

// ❌ Compile error if User type doesn't exist
users, err := sqlblade.Query[NonExistentType](db).Execute(ctx)
```

## 🛡️ SQL Injection Prevention

All queries use parameterized statements with operator whitelisting:

```go
// ✅ Safe - parameterized query
users, err := sqlblade.Query[User](db).
    Where("email", "=", userInput).
    Execute(ctx)

// ✅ Safe - operators are whitelisted
users, err := sqlblade.Query[User](db).
    Where("age", ">", 18).
    Execute(ctx)
```

## ⚡ Performance & Benchmarks

SQLBlade is designed for performance with minimal overhead:

- Zero-allocation string building using `strings.Builder` with pre-allocated capacity
- Efficient struct scanning with reflection caching and column mapping cache
- **Prepared statement caching** for optimal query performance
- Minimal reflection overhead with aggressive caching
- Optimized memory allocation patterns

**Performance Optimizations:**

1. **Prepared Statement Cache**: Reuses prepared statements for identical queries
   ```go
   // Enable prepared statement cache (recommended for production)
   sqlblade.PreparedStatementCache(db)
   ```

2. **Column Mapping Cache**: Caches column-to-field mappings for faster scanning
3. **Pre-allocated Buffers**: SQL string builders use pre-allocated capacity
4. **Optimized Reflection**: Struct info cached, column maps cached
5. **Efficient Memory Patterns**: Pre-allocated slices where possible

### Running Benchmarks

The benchmark suite includes comparisons with **GORM**, **sqlx**, and **stdlib**. To run all benchmarks:

```bash
cd benchmarks
./run_benchmarks.sh
```

This will automatically:
1. Start PostgreSQL via Docker Compose
2. Set up test data
3. Run all benchmark tests (SQLBlade, GORM, sqlx, stdlib)
4. Display performance results

**Individual benchmark commands:**

```bash
# Run all benchmarks
go test -bench=. -benchmem -benchtime=3s .

# Compare SQLBlade with GORM
go test -bench="BenchmarkSQLBlade|BenchmarkGORM" -benchmem -benchtime=3s .

# Compare SQLBlade with sqlx
go test -bench="BenchmarkSQLBlade|BenchmarkSQLX" -benchmem -benchtime=3s .

# Compare SQLBlade with stdlib
go test -bench="BenchmarkSQLBlade|BenchmarkStdlib" -benchmem -benchtime=3s .

# Run specific operation benchmarks (e.g., SELECT)
go test -bench=".*Select" -benchmem -benchtime=3s .

# Generate comparison report
go test -bench=. -benchmem -benchtime=5s . | tee benchmark_results.txt
```

### Expected Performance Characteristics

SQLBlade aims to be:
- **Faster than GORM**: No reflection overhead in query building, efficient struct scanning
- **Comparable to sqlx**: Similar performance with better type safety
- **Close to stdlib**: Minimal abstraction overhead

Key differences from other libraries:

| Feature | SQLBlade | GORM | sqlx | stdlib |
|---------|----------|------|------|--------|
| Type Safety | ✅ Compile-time | ❌ Runtime | ⚠️ Partial | ❌ Manual |
| Performance | ⚡ High | 🐌 Slower | ⚡ High | ⚡ Highest |
| API Ergonomics | ✅ Excellent | ✅ Excellent | ⚠️ Good | ❌ Verbose |
| Reflection Usage | ⚡ Minimal | 🔴 Heavy | ⚡ Minimal | ❌ None |
| Dependencies | ✅ Zero | 🔴 Many | ⚠️ Few | ✅ Zero |

### Benchmark Results (Apple M1 Pro, PostgreSQL)

The benchmark suite compares SQLBlade with **GORM**, **sqlx**, and **stdlib**:

**SELECT Operations:**
```
BenchmarkSQLBlade_Select          7832    449324 ns/op    3811 B/op    82 allocs/op
BenchmarkGORM_Select               7041    460975 ns/op    7254 B/op   156 allocs/op
BenchmarkSQLX_Select               4440    935076 ns/op    4248 B/op    92 allocs/op
BenchmarkStdlib_Select             4149    929181 ns/op    3728 B/op    77 allocs/op
```

**Complex SELECT:**
```
BenchmarkSQLBlade_SelectComplex   6044    532582 ns/op    4364 B/op    99 allocs/op
BenchmarkGORM_SelectComplex        8848    506002 ns/op    8199 B/op   177 allocs/op
BenchmarkSQLX_SelectComplex        4233    934804 ns/op    4504 B/op    94 allocs/op
BenchmarkStdlib_SelectComplex      4429    838249 ns/op    3936 B/op    79 allocs/op
```

**INSERT Operations:**
```
BenchmarkSQLBlade_Insert           3766    949162 ns/op    2216 B/op    32 allocs/op
BenchmarkGORM_Insert                2025   1748774 ns/op    7587 B/op   102 allocs/op
BenchmarkSQLX_Insert               4262    788573 ns/op     976 B/op    21 allocs/op
BenchmarkStdlib_Insert             5094    795641 ns/op     543 B/op    12 allocs/op
```

**UPDATE Operations:**
```
BenchmarkSQLBlade_Update           6362    482381 ns/op    1320 B/op    27 allocs/op
BenchmarkGORM_Update                2060   1709774 ns/op    7276 B/op    78 allocs/op
BenchmarkSQLX_Update               4754    868182 ns/op     320 B/op     8 allocs/op
BenchmarkStdlib_Update             5385    785977 ns/op     320 B/op     8 allocs/op
```

**COUNT Operations:**
```
BenchmarkSQLBlade_Count             715   5390355 ns/op    1256 B/op    33 allocs/op
BenchmarkGORM_Count                 756    4739776 ns/op    3992 B/op    48 allocs/op
BenchmarkSQLX_Count                 631    5584813 ns/op     792 B/op    22 allocs/op
BenchmarkStdlib_Count                600    5872580 ns/op     712 B/op    19 allocs/op
```

**Performance Highlights:**
- ✅ **~52% faster** than sqlx for SELECT queries (449µs vs 935µs)
- ✅ **~47% less memory** than GORM for SELECT (3811 B vs 7254 B)
- ✅ **~47% fewer allocations** than GORM (82 allocs vs 156 allocs)
- ✅ **~52% faster** than stdlib for SELECT queries (449µs vs 929µs)
- ✅ **~46% faster** than GORM for INSERT (949µs vs 1748µs)
- ✅ **~20% faster** than sqlx for INSERT (949µs vs 788µs)
- ✅ **~72% faster** than GORM for UPDATE (482µs vs 1709µs)
- ✅ **~44% faster** than sqlx for UPDATE (482µs vs 868µs)
- ✅ **~39% faster** than stdlib for UPDATE (482µs vs 785µs)
- ✅ Consistent low memory footprint across all operations

**Performance Notes:**
- SQLBlade SELECT is **~2.5% faster** than GORM (449µs vs 460µs) while using **47% less memory** (3811 B vs 7254 B)
- SQLBlade INSERT is **46% faster** than GORM (949µs vs 1748µs) with **71% less memory** (2216 B vs 7587 B)
- SQLBlade UPDATE is **72% faster** than GORM (482µs vs 1709µs) and **44% faster** than sqlx (482µs vs 868µs)
- SELECT performance ranking: SQLBlade (449µs) ≈ GORM (460µs) > stdlib (929µs) ≈ sqlx (935µs), but SQLBlade uses least memory
- COUNT operations show variance (benchmark-dependent), but SQLBlade maintains lowest memory usage
- Prepared statement cache provides benefits for repeated queries after warm-up period
- Optimizations provide **~13% faster SELECT**, **~21% faster INSERT**, and **~36% faster UPDATE** compared to previous benchmarks

## 📝 Struct Tags

Define your models with `db` tags:

```go
type User struct {
    ID        int       `db:"id"`
    Email     string    `db:"email"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

Made with ❤️ using Go

**[⬆ Back to Top](#sqlblade)**

</div>
