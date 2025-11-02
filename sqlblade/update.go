package sqlblade

import (
	"context"
	"database/sql"
	"reflect"
	"strings"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// UpdateBuilder handles UPDATE operations
type UpdateBuilder[T any] struct {
	db           *sql.DB
	tx           *sql.Tx
	dialect      dialect.Dialect
	tableName    string
	sets         map[string]interface{}
	whereClauses []WhereClause
	returning    []string
}

// Update creates a new UPDATE builder
func Update[T any](db *sql.DB) *UpdateBuilder[T] {
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

	return &UpdateBuilder[T]{
		db:           db,
		dialect:      d,
		tableName:    info.tableName,
		sets:         make(map[string]interface{}),
		whereClauses: make([]WhereClause, 0),
		returning:    make([]string, 0),
	}
}

// UpdateTx creates a new UPDATE builder with transaction
func UpdateTx[T any](tx *sql.Tx) *UpdateBuilder[T] {
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

	return &UpdateBuilder[T]{
		tx:           tx,
		dialect:      d,
		tableName:    info.tableName,
		sets:         make(map[string]interface{}),
		whereClauses: make([]WhereClause, 0),
		returning:    make([]string, 0),
	}
}

// Set sets a column value
func (ub *UpdateBuilder[T]) Set(column string, value interface{}) *UpdateBuilder[T] {
	ub.sets[column] = value
	return ub
}

// Where adds a WHERE condition
func (ub *UpdateBuilder[T]) Where(column string, operator string, value interface{}) *UpdateBuilder[T] {
	ub.whereClauses = append(ub.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return ub
}

// Returning specifies columns to return (PostgreSQL)
func (ub *UpdateBuilder[T]) Returning(columns ...string) *UpdateBuilder[T] {
	ub.returning = columns
	return ub
}

// Execute executes the UPDATE statement
func (ub *UpdateBuilder[T]) Execute(ctx context.Context) (sql.Result, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if len(ub.sets) == 0 {
		return nil, ErrEmptySet
	}

	var buf strings.Builder
	buf.Grow(256)
	paramIndex := 0
	args := make([]interface{}, 0, len(ub.sets)+len(ub.whereClauses))

	buf.WriteString("UPDATE ")
	buf.WriteString(ub.dialect.QuoteIdentifier(ub.tableName))
	buf.WriteString(" SET ")

	setParts := make([]string, 0, len(ub.sets))
	for col, val := range ub.sets {
		paramIndex++
		setParts = append(setParts, ub.dialect.QuoteIdentifier(col)+" = "+ub.dialect.Placeholder(paramIndex))
		args = append(args, val)
	}
	buf.WriteString(strings.Join(setParts, ", "))

	whereSQL, whereArgs := buildWhereClause(ub.dialect, ub.whereClauses, &paramIndex)
	if whereSQL != "" {
		buf.WriteString(" ")
		buf.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	if len(ub.returning) > 0 && ub.dialect.Name() == "postgres" {
		buf.WriteString(" RETURNING ")
		returningCols := make([]string, len(ub.returning))
		for i, col := range ub.returning {
			returningCols[i] = ub.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(returningCols, ", "))
	}

	sqlStr := buf.String()

	var result sql.Result
	var err error

	if ub.tx == nil && globalStmtCache != nil && globalStmtCache.db == ub.db {
		stmt, stmtErr := globalStmtCache.getStmt(ctx, sqlStr)
		if stmtErr == nil {
			result, err = stmt.ExecContext(ctx, args...)
			if err == nil {
				return result, nil
			}
			return nil, wrapQueryError(err, sqlStr, args)
		}
		return nil, wrapQueryError(stmtErr, sqlStr, args)
	}

	if ub.tx != nil {
		result, err = ub.tx.ExecContext(ctx, sqlStr, args...)
	} else {
		result, err = ub.db.ExecContext(ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, wrapQueryError(err, sqlStr, args)
	}

	return result, nil
}
