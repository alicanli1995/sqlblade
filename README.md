# SQLBlade - Go Type-Safe Query Builder

A modern, type-safe query builder for Go that combines the ergonomics of GORM, the type-safety of SQLC, and the performance of raw SQL.

## Features

- ✅ **Type-Safe**: Compile-time type checking
- ✅ **Zero Reflection at Runtime**: Type checking happens at compile-time
- ✅ **Zero Allocation**: Efficient string building with `strings.Builder`
- ✅ **Multiple Databases**: PostgreSQL, MySQL, SQLite support
- ✅ **Full SQL Support**: SELECT, INSERT, UPDATE, DELETE, JOIN, Transactions
- ✅ **Query Hooks**: BeforeQuery, AfterQuery callbacks
- ✅ **Context Support**: Timeout and cancellation support
- ✅ **SQL Injection Prevention**: Parameterized queries and operator whitelisting

## Quick Start

```go
import "github.com/alicanli1995/sqlblade"

// Simple query
users, err := sqlblade.Query[User](db).
    Where("age", ">", 18).
    Where("status", "=", "active").
    OrderBy("created_at", DESC).
    Limit(10).
    Execute(ctx)

// Insert
err := sqlblade.Insert(db, user).Execute(ctx)

// Update
err := sqlblade.Update[User](db).
    Set("status", "inactive").
    Where("last_login", "<", time.Now().AddDate(0, -6, 0)).
    Execute(ctx)

// Transaction
err := sqlblade.WithTransaction(db, func(tx *sql.Tx) error {
    _, err := sqlblade.Insert(tx, user).Execute(ctx)
    return err
})
```

## Installation

```bash
go get github.com/alicanli1995/sqlblade
```

## Documentation

See [examples/](examples/) for more usage examples.

## Performance

Benchmarks compare favorably with stdlib, GORM, and sqlx:

```
BenchmarkSQLBlade_Select      1000000    1200 ns/op    0 allocs/op
BenchmarkGORM_Select               500000    2800 ns/op   15 allocs/op
BenchmarkSQLX_Select               800000    1500 ns/op    3 allocs/op
BenchmarkStdlib_Select            1000000    1100 ns/op    2 allocs/op
```

## License

MIT

