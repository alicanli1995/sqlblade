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

// QueryBuilder is the main query builder struct
type QueryBuilder[T any] struct {
	db           *sql.DB
	tx           *sql.Tx
	dialect      dialect.Dialect
	tableName    string
	whereClauses []WhereClause
	joins        []dialect.Join
	orderBy      []dialect.OrderBy
	limit        *int
	offset       *int
	selectCols   []string
	groupBy      []string
	having       []WhereClause
	distinct     bool
}

// Query creates a new SELECT query builder
func Query[T any](db *sql.DB) *QueryBuilder[T] {
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

	return &QueryBuilder[T]{
		db:           db,
		dialect:      d,
		tableName:    info.tableName,
		whereClauses: make([]WhereClause, 0),
		joins:        make([]dialect.Join, 0),
		orderBy:      make([]dialect.OrderBy, 0),
		selectCols:   make([]string, 0),
		groupBy:      make([]string, 0),
		having:       make([]WhereClause, 0),
	}
}

// QueryTx creates a new SELECT query builder with transaction
func QueryTx[T any](tx *sql.Tx) *QueryBuilder[T] {
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

	return &QueryBuilder[T]{
		tx:           tx,
		dialect:      d,
		tableName:    info.tableName,
		whereClauses: make([]WhereClause, 0),
		joins:        make([]dialect.Join, 0),
		orderBy:      make([]dialect.OrderBy, 0),
		selectCols:   make([]string, 0),
		groupBy:      make([]string, 0),
		having:       make([]WhereClause, 0),
	}
}

// detectDialect detects database dialect from driver
func detectDialect(driver interface{}) dialect.Dialect {
	if driver == nil {
		return dialect.NewPostgreSQL()
	}

	driverType := reflect.TypeOf(driver).String()
	switch {
	case strings.Contains(driverType, "pq") || strings.Contains(driverType, "postgres"):
		return dialect.NewPostgreSQL()
	case strings.Contains(driverType, "mysql"):
		return dialect.NewMySQL()
	case strings.Contains(driverType, "sqlite"):
		return dialect.NewSQLite()
	default:
		return dialect.NewPostgreSQL()
	}
}

// Where adds a WHERE condition (AND)
func (qb *QueryBuilder[T]) Where(column string, operator string, value interface{}) *QueryBuilder[T] {
	qb.whereClauses = append(qb.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return qb
}

// OrWhere adds a WHERE condition (OR)
func (qb *QueryBuilder[T]) OrWhere(column string, operator string, value interface{}) *QueryBuilder[T] {
	qb.whereClauses = append(qb.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      false,
	})
	return qb
}

// Select specifies columns to select
func (qb *QueryBuilder[T]) Select(columns ...string) *QueryBuilder[T] {
	qb.selectCols = columns
	return qb
}

// Distinct adds DISTINCT keyword
func (qb *QueryBuilder[T]) Distinct() *QueryBuilder[T] {
	qb.distinct = true
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder[T]) Join(table string, condition string) *QueryBuilder[T] {
	return qb.joinWithType(dialect.InnerJoin, table, condition)
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder[T]) LeftJoin(table string, condition string) *QueryBuilder[T] {
	return qb.joinWithType(dialect.LeftJoin, table, condition)
}

// RightJoin adds a RIGHT JOIN clause
func (qb *QueryBuilder[T]) RightJoin(table string, condition string) *QueryBuilder[T] {
	return qb.joinWithType(dialect.RightJoin, table, condition)
}

// FullJoin adds a FULL JOIN clause
func (qb *QueryBuilder[T]) FullJoin(table string, condition string) *QueryBuilder[T] {
	return qb.joinWithType(dialect.FullJoin, table, condition)
}

// joinWithType adds a JOIN with specific type
func (qb *QueryBuilder[T]) joinWithType(joinType dialect.JoinType, table string, condition string) *QueryBuilder[T] {
	qb.joins = append(qb.joins, dialect.Join{
		Type:      joinType,
		Table:     table,
		Condition: condition,
	})
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder[T]) OrderBy(column string, order dialect.OrderDirection) *QueryBuilder[T] {
	qb.orderBy = append(qb.orderBy, dialect.OrderBy{
		Column: column,
		Order:  order,
	})
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder[T]) GroupBy(columns ...string) *QueryBuilder[T] {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Having adds a HAVING clause
func (qb *QueryBuilder[T]) Having(column string, operator string, value interface{}) *QueryBuilder[T] {
	qb.having = append(qb.having, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	qb.limit = &limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	qb.offset = &offset
	return qb
}

func (qb *QueryBuilder[T]) buildSQL() (string, []interface{}) {
	var buf strings.Builder
	buf.Grow(sqlBuilderBufferSize)
	paramIndex := 0
	args := make([]interface{}, 0, argsInitialCapacity)

	buf.WriteString("SELECT ")
	if qb.distinct {
		buf.WriteString("DISTINCT ")
	}

	if len(qb.selectCols) > 0 {
		quotedCols := make([]string, len(qb.selectCols))
		for i, col := range qb.selectCols {
			quotedCols[i] = qb.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(quotedCols, ", "))
	} else {
		buf.WriteString("*")
	}

	buf.WriteString(" FROM ")
	buf.WriteString(qb.dialect.QuoteIdentifier(qb.tableName))

	for _, join := range qb.joins {
		buf.WriteString(" ")
		buf.WriteString(qb.dialect.BuildJoin(join))
	}

	whereSQL, whereArgs := buildWhereClause(qb.dialect, qb.whereClauses, &paramIndex)
	if whereSQL != "" {
		buf.WriteString(" ")
		buf.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	if len(qb.groupBy) > 0 {
		buf.WriteString(" GROUP BY ")
		quotedCols := make([]string, len(qb.groupBy))
		for i, col := range qb.groupBy {
			quotedCols[i] = qb.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(quotedCols, ", "))
	}

	if len(qb.having) > 0 {
		havingSQL, havingArgs := buildWhereClause(qb.dialect, qb.having, &paramIndex)
		if havingSQL != "" {
			buf.WriteString(" ")
			buf.WriteString(strings.Replace(havingSQL, "WHERE", "HAVING", 1))
			args = append(args, havingArgs...)
		}
	}

	if len(qb.orderBy) > 0 {
		buf.WriteString(" ")
		buf.WriteString(qb.dialect.BuildOrderBy(qb.orderBy))
	}

	if qb.limit != nil || qb.offset != nil {
		buf.WriteString(" ")
		buf.WriteString(qb.dialect.BuildLimitOffset(qb.limit, qb.offset))
	}

	return buf.String(), args
}

// Execute executes the query and returns results
func (qb *QueryBuilder[T]) Execute(ctx context.Context) ([]T, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	sqlStr, args := qb.buildSQL()
	startTime := time.Now()

	// Execute before hooks
	if err := DefaultHooks.ExecuteBeforeHooks(ctx, sqlStr, args); err != nil {
		return nil, err
	}

	// Debug logging
	if globalDebugger.enabled {
		debugQuery := &DebugQuery{
			SQL:       sqlStr,
			Args:      args,
			Table:     qb.tableName,
			Operation: "SELECT",
			Timestamp: startTime,
		}
		defer func() {
			debugQuery.Duration = time.Since(startTime)
			globalDebugger.Log(debugQuery)
		}()
	}

	var rows *sql.Rows
	var err error

	if qb.tx == nil && globalStmtCache != nil && globalStmtCache.db == qb.db {
		stmt, stmtErr := globalStmtCache.getStmt(ctx, sqlStr)
		if stmtErr == nil {
			rows, err = stmt.QueryContext(ctx, args...)
			if err == nil {
				defer func(rows *sql.Rows) {
					closeErr := rows.Close()
					if closeErr != nil {
						log.Printf("failed to close rows: %v", closeErr)
					}
				}(rows)
				result, err := scanRowsOptimized[T](rows)
				if err == nil {
					DefaultHooks.ExecuteAfterHooks(ctx, sqlStr, args)
				}
				return result, err
			}
			return nil, wrapQueryError(err, sqlStr, args)
		}
		return nil, wrapQueryError(stmtErr, sqlStr, args)
	}

	if qb.tx != nil {
		rows, err = qb.tx.QueryContext(ctx, sqlStr, args...)
	} else {
		rows, err = qb.db.QueryContext(ctx, sqlStr, args...)
	}

	if err != nil {
		return nil, wrapQueryError(err, sqlStr, args)
	}
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			log.Printf("failed to close rows: %v", closeErr)
		}
	}(rows)

	result, err := scanRowsOptimized[T](rows)
	if err == nil {
		DefaultHooks.ExecuteAfterHooks(ctx, sqlStr, args)
	}
	return result, err
}

// NotExists creates a NOT EXISTS subquery
func (qb *QueryBuilder[T]) NotExists(ctx context.Context) (bool, error) {
	exists, err := qb.Exists(ctx)
	return !exists, err
}

// Exists creates an EXISTS subquery
func (qb *QueryBuilder[T]) Exists(ctx context.Context) (bool, error) {
	sql, args := qb.buildSQL()
	existsSQL := "SELECT EXISTS(" + sql + ")"

	var result bool
	if qb.tx != nil {
		row := qb.tx.QueryRowContext(ctx, existsSQL, args...)
		err := row.Scan(&result)
		if err != nil {
			return false, wrapQueryError(err, existsSQL, args)
		}
		return result, nil
	}

	row := qb.db.QueryRowContext(ctx, existsSQL, args...)
	err := row.Scan(&result)
	if err != nil {
		return false, wrapQueryError(err, existsSQL, args)
	}

	return result, nil
}
