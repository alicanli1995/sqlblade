// Package sqlblade provides a modern, type-safe query builder for Go.
//
// SQLBlade combines the ergonomics of GORM, the type-safety of SQLC, and the performance of raw SQL.
// It uses Go generics to provide compile-time type safety and supports PostgreSQL, MySQL, and SQLite.
//
// Features:
//
//   - Type-safe queries with compile-time type checking
//   - Zero reflection overhead at runtime
//   - High performance with zero-allocation string building
//   - Multi-database support (PostgreSQL, MySQL, SQLite)
//   - Full SQL support (SELECT, INSERT, UPDATE, DELETE, JOIN, Transactions)
//   - Context support for timeout and cancellation
//   - SQL injection prevention with parameterized queries
//   - Zero dependencies (except database drivers)
//
// Example usage:
//
//	import (
//	    "context"
//	    "database/sql"
//	    _ "github.com/lib/pq"
//	    "github.com/alicanli1995/sqlblade/sqlblade"
//	    "github.com/alicanli1995/sqlblade/sqlblade/dialect"
//	)
//
//	type User struct {
//	    ID    int    `db:"id"`
//	    Email string `db:"email"`
//	    Name  string `db:"name"`
//	}
//
//	// Query
//	users, err := sqlblade.Query[User](db).
//	    Where("age", ">", 18).
//	    OrderBy("created_at", dialect.DESC).
//	    Limit(10).
//	    Execute(ctx)
//
//	// Insert
//	result, err := sqlblade.Insert(db, user).Execute(ctx)
//
//	// Update
//	result, err := sqlblade.Update[User](db).
//	    Set("status", "inactive").
//	    Where("id", "=", userID).
//	    Execute(ctx)
//
//	// Transaction
//	err := sqlblade.WithTransaction(db, func(tx *sql.Tx) error {
//	    _, err := sqlblade.InsertTx(tx, user).Execute(ctx)
//	    return err
//	})
//
// For more information, visit: https://github.com/alicanli1995/sqlblade
package sqlblade
