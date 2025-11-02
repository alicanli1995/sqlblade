package sqlblade

import (
	"context"
	"database/sql"
	"strings"
)

// Aggregate functions
type AggregateFunc string

const (
	Count AggregateFunc = "COUNT"
	Sum   AggregateFunc = "SUM"
	Avg   AggregateFunc = "AVG"
	Min   AggregateFunc = "MIN"
	Max   AggregateFunc = "MAX"
)

// AggregateResult represents the result of an aggregate query
type AggregateResult struct {
	Value interface{}
}

// Count executes a COUNT query
func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	val, err := qb.aggregate(ctx, Count, "*")
	if err != nil {
		return 0, err
	}
	if i, ok := val.(int64); ok {
		return i, nil
	}
	if f, ok := val.(float64); ok {
		return int64(f), nil
	}
	return 0, nil
}

// Sum executes a SUM query
func (qb *QueryBuilder[T]) Sum(ctx context.Context, column string) (float64, error) {
	val, err := qb.aggregate(ctx, Sum, column)
	if err != nil {
		return 0, err
	}
	if f, ok := val.(float64); ok {
		return f, nil
	}
	if i, ok := val.(int64); ok {
		return float64(i), nil
	}
	return 0, nil
}

// Avg executes an AVG query
func (qb *QueryBuilder[T]) Avg(ctx context.Context, column string) (float64, error) {
	val, err := qb.aggregate(ctx, Avg, column)
	if err != nil {
		return 0, err
	}
	if f, ok := val.(float64); ok {
		return f, nil
	}
	return 0, nil
}

// Min executes a MIN query
func (qb *QueryBuilder[T]) Min(ctx context.Context, column string) (interface{}, error) {
	return qb.aggregate(ctx, Min, column)
}

// Max executes a MAX query
func (qb *QueryBuilder[T]) Max(ctx context.Context, column string) (interface{}, error) {
	return qb.aggregate(ctx, Max, column)
}

// aggregate executes an aggregate function
func (qb *QueryBuilder[T]) aggregate(ctx context.Context, fn AggregateFunc, column string) (interface{}, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	var buf strings.Builder
	paramIndex := 0
	var args []interface{}

	// SELECT aggregate function
	buf.WriteString("SELECT ")
	buf.WriteString(string(fn))
	buf.WriteString("(")
	if column == "*" {
		buf.WriteString("*")
	} else {
		buf.WriteString(qb.dialect.QuoteIdentifier(column))
	}
	buf.WriteString(")")

	// FROM
	buf.WriteString(" FROM ")
	buf.WriteString(qb.dialect.QuoteIdentifier(qb.tableName))

	// JOINs
	for _, join := range qb.joins {
		buf.WriteString(" ")
		buf.WriteString(qb.dialect.BuildJoin(join))
	}

	// WHERE
	whereSQL, whereArgs := buildWhereClause(qb.dialect, qb.whereClauses, &paramIndex)
	if whereSQL != "" {
		buf.WriteString(" ")
		buf.WriteString(whereSQL)
		args = append(args, whereArgs...)
	}

	// GROUP BY
	if len(qb.groupBy) > 0 {
		buf.WriteString(" GROUP BY ")
		quotedCols := make([]string, len(qb.groupBy))
		for i, col := range qb.groupBy {
			quotedCols[i] = qb.dialect.QuoteIdentifier(col)
		}
		buf.WriteString(strings.Join(quotedCols, ", "))
	}

	// HAVING
	if len(qb.having) > 0 {
		havingSQL, havingArgs := buildWhereClause(qb.dialect, qb.having, &paramIndex)
		if havingSQL != "" {
			buf.WriteString(" ")
			buf.WriteString(strings.Replace(havingSQL, "WHERE", "HAVING", 1))
			args = append(args, havingArgs...)
		}
	}

	sqlStr := buf.String()

	var row *sql.Row
	if qb.tx != nil {
		row = qb.tx.QueryRowContext(ctx, sqlStr, args...)
	} else {
		row = qb.db.QueryRowContext(ctx, sqlStr, args...)
	}

	var result interface{}
	err := row.Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRows
		}
		return nil, err
	}

	return result, nil
}
