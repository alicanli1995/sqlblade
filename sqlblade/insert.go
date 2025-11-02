package sqlblade

import (
	"context"
	"database/sql"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// InsertBuilder handles INSERT operations
type InsertBuilder[T any] struct {
	db        *sql.DB
	tx        *sql.Tx
	dialect   dialect.Dialect
	tableName string
	values    []T
	columns   []string
	returning []string
}

// Insert creates a new INSERT builder
func Insert[T any](db *sql.DB, value T) *InsertBuilder[T] {
	if db == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(db.Driver())
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		info = &structInfo{
			tableName: toSnakeCase(typ.Name()),
		}
	}

	return &InsertBuilder[T]{
		db:        db,
		dialect:   d,
		tableName: info.tableName,
		values:    []T{value},
		columns:   make([]string, 0),
		returning: make([]string, 0),
	}
}

// InsertTx creates a new INSERT builder with transaction
func InsertTx[T any](tx *sql.Tx, value T) *InsertBuilder[T] {
	if tx == nil {
		panic(ErrNilDB)
	}

	d := detectDialect(nil)
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		info = &structInfo{
			tableName: toSnakeCase(typ.Name()),
		}
	}

	return &InsertBuilder[T]{
		tx:        tx,
		dialect:   d,
		tableName: info.tableName,
		values:    []T{value},
		columns:   make([]string, 0),
		returning: make([]string, 0),
	}
}

// InsertBatch creates a new batch INSERT builder
func InsertBatch[T any](db *sql.DB, values []T) *InsertBuilder[T] {
	if db == nil {
		panic(ErrNilDB)
	}
	if len(values) == 0 {
		panic(ErrEmptySet)
	}

	d := detectDialect(db.Driver())
	typ := reflect.TypeOf(values[0])
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		info = &structInfo{
			tableName: toSnakeCase(typ.Name()),
		}
	}

	return &InsertBuilder[T]{
		db:        db,
		dialect:   d,
		tableName: info.tableName,
		values:    values,
		columns:   make([]string, 0),
		returning: make([]string, 0),
	}
}

// Columns specifies which columns to insert
func (ib *InsertBuilder[T]) Columns(columns ...string) *InsertBuilder[T] {
	ib.columns = columns
	return ib
}

// Returning specifies columns to return (PostgreSQL)
func (ib *InsertBuilder[T]) Returning(columns ...string) *InsertBuilder[T] {
	ib.returning = columns
	return ib
}

// Execute executes the INSERT statement
func (ib *InsertBuilder[T]) Execute(ctx context.Context) (sql.Result, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if len(ib.values) == 0 {
		return nil, ErrEmptySet
	}

	typ := reflect.TypeOf(ib.values[0])
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	info, err := getStructInfo(typ)
	if err != nil {
		return nil, err
	}

	columns := ib.resolveColumns(info)
	sqlStr, args := ib.buildInsertSQL(info, columns)

	startTime := time.Now()
	if err := DefaultHooks.ExecuteBeforeHooks(ctx, sqlStr, args); err != nil {
		return nil, err
	}

	var result sql.Result
	var execErr error

	if globalDebugger.enabled {
		debugQuery := &DebugQuery{
			SQL:       sqlStr,
			Args:      args,
			Table:     ib.tableName,
			Operation: "INSERT",
			Timestamp: startTime,
		}
		defer func() {
			debugQuery.Duration = time.Since(startTime)
			if result != nil {
				rowsAffected, err := result.RowsAffected()
				if err == nil {
					debugQuery.RowsAffected = rowsAffected
				}
			}
			globalDebugger.Log(debugQuery)
		}()
	}

	if ib.tx != nil {
		result, execErr = ib.tx.ExecContext(ctx, sqlStr, args...)
	} else {
		result, execErr = ib.db.ExecContext(ctx, sqlStr, args...)
	}

	if execErr != nil {
		return nil, wrapQueryError(execErr, sqlStr, args)
	}

	if hookErr := DefaultHooks.ExecuteAfterHooks(ctx, sqlStr, args); hookErr != nil {
		log.Printf("after query hook error: %v", hookErr)
	}

	return result, nil
}

func (ib *InsertBuilder[T]) resolveColumns(info *structInfo) []string {
	if len(ib.columns) > 0 {
		return ib.columns
	}

	if len(ib.values) == 0 {
		columns := make([]string, 0, len(info.fields))
		for _, field := range info.fields {
			columns = append(columns, field.dbColumn)
		}
		return columns
	}

	valRef := reflect.ValueOf(ib.values[0])
	if valRef.Kind() == reflect.Ptr {
		valRef = valRef.Elem()
	}

	columns := make([]string, 0, len(info.fields))
	for _, field := range info.fields {
		if strings.Contains(field.dbColumn, "id") {
			fieldVal := valRef.Field(field.index)
			if fieldVal.IsValid() && fieldVal.IsZero() {
				continue
			}
		}
		columns = append(columns, field.dbColumn)
	}
	return columns
}

func (ib *InsertBuilder[T]) buildInsertSQL(info *structInfo, columns []string) (string, []interface{}) {
	var buf strings.Builder
	estimatedSize := insertBufferSize
	if len(ib.values) > 1 {
		estimatedSize = batchInsertBufferSize
	}
	buf.Grow(estimatedSize)
	paramIndex := 0
	var args []interface{}

	buf.WriteString("INSERT INTO ")
	buf.WriteString(ib.dialect.QuoteIdentifier(ib.tableName))
	buf.WriteString(" (")

	quotedCols := make([]string, len(columns))
	for i, col := range columns {
		quotedCols[i] = ib.dialect.QuoteIdentifier(col)
	}
	buf.WriteString(strings.Join(quotedCols, ", "))
	buf.WriteString(") VALUES ")

	fieldMap := make(map[string]int, len(info.fields))
	for idx, field := range info.fields {
		fieldMap[field.dbColumn] = idx
	}

	valueParts := ib.buildValueParts(columns, fieldMap, &paramIndex, &args)
	buf.WriteString(strings.Join(valueParts, ", "))

	if len(ib.returning) > 0 && ib.dialect.Name() == dialectPostgres {
		buf.WriteString(" RETURNING ")
		returningCols := make([]string, len(ib.returning))
		for i, col := range ib.returning {
			returningCols[i] = ib.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(returningCols, ", "))
	}

	return buf.String(), args
}

func (ib *InsertBuilder[T]) buildValueParts(columns []string, fieldMap map[string]int, paramIndex *int, args *[]interface{}) []string {
	valueParts := make([]string, len(ib.values))
	for i, val := range ib.values {
		valRef := reflect.ValueOf(val)
		if valRef.Kind() == reflect.Ptr {
			valRef = valRef.Elem()
		}

		placeholders := make([]string, len(columns))
		for j, col := range columns {
			*paramIndex++
			placeholders[j] = ib.dialect.Placeholder(*paramIndex)

			var fieldValue interface{}
			colLower := strings.ToLower(col)
			if fieldIdx, ok := fieldMap[colLower]; ok {
				fieldVal := valRef.Field(fieldIdx)
				if fieldVal.IsValid() {
					fieldValue = fieldVal.Interface()
				}
			}
			*args = append(*args, fieldValue)
		}
		valueParts[i] = "(" + strings.Join(placeholders, ", ") + ")"
	}
	return valueParts
}
