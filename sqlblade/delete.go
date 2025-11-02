package sqlblade

import (
	"context"
	"database/sql"
	"reflect"
	"strings"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// DeleteBuilder handles DELETE operations
type DeleteBuilder[T any] struct {
	db           *sql.DB
	tx           *sql.Tx
	dialect      dialect.Dialect
	tableName    string
	whereClauses []WhereClause
	returning    []string
}

// Delete creates a new DELETE builder
func Delete[T any](db *sql.DB) *DeleteBuilder[T] {
	if db == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(db.Driver())
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		info = &structInfo{
			tableName: toSnakeCase(typ.Name()),
		}
	}

	return &DeleteBuilder[T]{
		db:           db,
		dialect:      d,
		tableName:    info.tableName,
		whereClauses: make([]WhereClause, 0),
		returning:    make([]string, 0),
	}
}

// DeleteTx creates a new DELETE builder with transaction
func DeleteTx[T any](tx *sql.Tx) *DeleteBuilder[T] {
	if tx == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(nil)
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		info = &structInfo{
			tableName: toSnakeCase(typ.Name()),
		}
	}

	return &DeleteBuilder[T]{
		tx:           tx,
		dialect:      d,
		tableName:    info.tableName,
		whereClauses: make([]WhereClause, 0),
		returning:    make([]string, 0),
	}
}

// Where adds a WHERE condition
func (db *DeleteBuilder[T]) Where(column string, operator string, value interface{}) *DeleteBuilder[T] {
	db.whereClauses = append(db.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return db
}

// Returning specifies columns to return (PostgreSQL)
func (db *DeleteBuilder[T]) Returning(columns ...string) *DeleteBuilder[T] {
	db.returning = columns
	return db
}

// Execute executes the DELETE statement
func (db *DeleteBuilder[T]) Execute(ctx context.Context) (sql.Result, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	var buf strings.Builder
	paramIndex := 0
	var args []interface{}

	buf.WriteString("DELETE FROM ")
	buf.WriteString(db.dialect.QuoteIdentifier(db.tableName))

	whereSQL, whereArgs := buildWhereClause(db.dialect, db.whereClauses, &paramIndex)
	if whereSQL != "" {
		buf.WriteString(" ")
		buf.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	if len(db.returning) > 0 && db.dialect.Name() == "postgres" {
		buf.WriteString(" RETURNING ")
		returningCols := make([]string, len(db.returning))
		for i, col := range db.returning {
			returningCols[i] = db.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(returningCols, ", "))
	}

	sqlStr := buf.String()

	var result sql.Result
	var err error

	if db.tx != nil {
		result, err = db.tx.ExecContext(ctx, sqlStr, args...)
	} else {
		result, err = db.db.ExecContext(ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, wrapQueryError(err, sqlStr, args)
	}

	return result, nil
}
