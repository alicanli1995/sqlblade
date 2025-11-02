package sqlblade

import (
	"context"
	"database/sql"
	"log"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// RawQuery executes a raw SQL query
type RawQuery[T any] struct {
	db      *sql.DB
	tx      *sql.Tx
	dialect dialect.Dialect
	query   string
	args    []interface{}
}

// Raw creates a new raw query builder
func Raw[T any](db *sql.DB, query string, args ...interface{}) *RawQuery[T] {
	if db == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(db.Driver())
	return &RawQuery[T]{
		db:      db,
		dialect: d,
		query:   query,
		args:    args,
	}
}

// RawTx creates a new raw query builder with transaction
func RawTx[T any](tx *sql.Tx, query string, args ...interface{}) *RawQuery[T] {
	if tx == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(nil)
	return &RawQuery[T]{
		tx:      tx,
		dialect: d,
		query:   query,
		args:    args,
	}
}

// Execute executes the raw query and returns results
func (rq *RawQuery[T]) Execute(ctx context.Context) ([]T, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	var rows *sql.Rows
	var err error

	if rq.tx != nil {
		rows, err = rq.tx.QueryContext(ctx, rq.query, rq.args...)
	} else {
		rows, err = rq.db.QueryContext(ctx, rq.query, rq.args...)
	}

	if err != nil {
		return nil, wrapQueryError(err, rq.query, rq.args)
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			log.Printf("failed to close rows: %v", closeErr)
		}
	}(rows)

	return scanRows[T](rows)
}

// First executes the raw query and returns the first result
func (rq *RawQuery[T]) First(ctx context.Context) (T, error) {
	var zero T
	results, err := rq.Execute(ctx)
	if err != nil {
		return zero, err
	}
	if len(results) == 0 {
		return zero, ErrNoRows
	}
	return results[0], nil
}

// Exec executes a raw query that doesn't return rows (INSERT, UPDATE, DELETE)
func (rq *RawQuery[T]) Exec(ctx context.Context) (sql.Result, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	var result sql.Result
	var err error

	if rq.tx != nil {
		result, err = rq.tx.ExecContext(ctx, rq.query, rq.args...)
	} else {
		result, err = rq.db.ExecContext(ctx, rq.query, rq.args...)
	}

	if err != nil {
		return nil, wrapQueryError(err, rq.query, rq.args)
	}

	return result, nil
}
