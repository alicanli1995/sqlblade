package sqlblade

import (
	"context"
	"fmt"
	"strings"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// QueryPreview provides a way to preview SQL without executing
type QueryPreview[T any] struct {
	builder *QueryBuilder[T]
}

// Preview returns a query preview for inspecting SQL without execution
func (qb *QueryBuilder[T]) Preview() *QueryPreview[T] {
	return &QueryPreview[T]{builder: qb}
}

// SQL returns the generated SQL query string
func (qp *QueryPreview[T]) SQL() string {
	sql, _ := qp.builder.buildSQL()
	return sql
}

// SQLWithArgs returns the SQL query with arguments substituted for readability
func (qp *QueryPreview[T]) SQLWithArgs() string {
	sql, args := qp.builder.buildSQL()
	return SubstituteArgs(sql, args)
}

// Args returns the query arguments
func (qp *QueryPreview[T]) Args() []interface{} {
	_, args := qp.builder.buildSQL()
	return args
}

// String returns a formatted string representation of the query
func (qp *QueryPreview[T]) String() string {
	sql, args := qp.builder.buildSQL()
	var sb strings.Builder
	sb.WriteString("SQL: ")
	sb.WriteString(sql)
	if len(args) > 0 {
		sb.WriteString("\nArgs: [")
		for i, arg := range args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v", arg))
		}
		sb.WriteString("]")
	}
	return sb.String()
}

// PrettyPrint prints a formatted version of the query
func (qp *QueryPreview[T]) PrettyPrint() {
	sql, args := qp.builder.buildSQL()

	debugQuery := &DebugQuery{
		SQL:       sql,
		Args:      args,
		Table:     qp.builder.tableName,
		Operation: "SELECT",
	}

	fmt.Print(formatQuery(debugQuery))
}

// Execute still allows execution after preview
func (qp *QueryPreview[T]) Execute(ctx context.Context) ([]T, error) {
	return qp.builder.Execute(ctx)
}

// QueryFragment represents a reusable query fragment
type QueryFragment struct {
	whereClauses []WhereClause
	joins        []dialect.Join
	orderBy      []dialect.OrderBy
	selectCols   []string
	groupBy      []string
	having       []WhereClause
	distinct     bool
	limit        *int
	offset       *int
}

// NewQueryFragment creates a new query fragment
func NewQueryFragment() *QueryFragment {
	return &QueryFragment{
		whereClauses: make([]WhereClause, 0),
		joins:        make([]dialect.Join, 0),
		orderBy:      make([]dialect.OrderBy, 0),
		selectCols:   make([]string, 0),
		groupBy:      make([]string, 0),
		having:       make([]WhereClause, 0),
	}
}

// Where adds a WHERE condition to the fragment
func (qf *QueryFragment) Where(column string, operator string, value interface{}) *QueryFragment {
	qf.whereClauses = append(qf.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return qf
}

// OrWhere adds an OR WHERE condition to the fragment
func (qf *QueryFragment) OrWhere(column string, operator string, value interface{}) *QueryFragment {
	qf.whereClauses = append(qf.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      false,
	})
	return qf
}

// Join adds a JOIN to the fragment
func (qf *QueryFragment) Join(table string, condition string) *QueryFragment {
	qf.joins = append(qf.joins, dialect.Join{
		Type:      dialect.InnerJoin,
		Table:     table,
		Condition: condition,
	})
	return qf
}

// LeftJoin adds a LEFT JOIN to the fragment
func (qf *QueryFragment) LeftJoin(table string, condition string) *QueryFragment {
	qf.joins = append(qf.joins, dialect.Join{
		Type:      dialect.LeftJoin,
		Table:     table,
		Condition: condition,
	})
	return qf
}

// OrderBy adds an ORDER BY clause to the fragment
func (qf *QueryFragment) OrderBy(column string, order dialect.OrderDirection) *QueryFragment {
	qf.orderBy = append(qf.orderBy, dialect.OrderBy{
		Column: column,
		Order:  order,
	})
	return qf
}

// Select adds columns to select
func (qf *QueryFragment) Select(columns ...string) *QueryFragment {
	qf.selectCols = append(qf.selectCols, columns...)
	return qf
}

// GroupBy adds a GROUP BY clause
func (qf *QueryFragment) GroupBy(columns ...string) *QueryFragment {
	qf.groupBy = append(qf.groupBy, columns...)
	return qf
}

// Having adds a HAVING clause
func (qf *QueryFragment) Having(column string, operator string, value interface{}) *QueryFragment {
	qf.having = append(qf.having, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		And:      true,
	})
	return qf
}

// Distinct sets distinct flag
func (qf *QueryFragment) Distinct() *QueryFragment {
	qf.distinct = true
	return qf
}

// Limit sets the limit
func (qf *QueryFragment) Limit(limit int) *QueryFragment {
	qf.limit = &limit
	return qf
}

// Offset sets the offset
func (qf *QueryFragment) Offset(offset int) *QueryFragment {
	qf.offset = &offset
	return qf
}

// Apply applies the fragment to a query builder (method on QueryBuilder)
func (qb *QueryBuilder[T]) Apply(qf *QueryFragment) *QueryBuilder[T] {
	// Apply where clauses
	qb.whereClauses = append(qb.whereClauses, qf.whereClauses...)

	// Apply joins
	qb.joins = append(qb.joins, qf.joins...)

	// Apply order by
	qb.orderBy = append(qb.orderBy, qf.orderBy...)

	// Apply select columns
	if len(qf.selectCols) > 0 {
		if len(qb.selectCols) == 0 {
			qb.selectCols = qf.selectCols
		} else {
			qb.selectCols = append(qb.selectCols, qf.selectCols...)
		}
	}

	// Apply group by
	qb.groupBy = append(qb.groupBy, qf.groupBy...)

	// Apply having
	qb.having = append(qb.having, qf.having...)

	// Apply distinct
	if qf.distinct {
		qb.distinct = true
	}

	// Apply limit (only if not already set)
	if qf.limit != nil && qb.limit == nil {
		qb.limit = qf.limit
	}

	// Apply offset (only if not already set)
	if qf.offset != nil && qb.offset == nil {
		qb.offset = qf.offset
	}

	return qb
}

// Subquery represents a subquery that can be used in WHERE clauses
type Subquery struct {
	sql  string
	args []interface{}
}

// NewSubquery creates a new subquery from a QueryBuilder
func NewSubquery[T any](qb *QueryBuilder[T]) *Subquery {
	sql, args := qb.buildSQL()
	return &Subquery{
		sql:  sql,
		args: args,
	}
}

// SQL returns the SQL of the subquery
func (sq *Subquery) SQL() string {
	return "(" + sq.sql + ")"
}

// Args returns the arguments of the subquery
func (sq *Subquery) Args() []interface{} {
	return sq.args
}

// WhereSubquery adds a WHERE condition using a subquery
func (qb *QueryBuilder[T]) WhereSubquery(column string, operator string, subquery *Subquery) *QueryBuilder[T] {
	// We need to handle subqueries specially in buildWhereClause
	// For now, we'll store it as a special WhereClause
	qb.whereClauses = append(qb.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    subquery, // Store subquery as value
		And:      true,
	})
	return qb
}

// OrWhereSubquery adds an OR WHERE condition using a subquery
func (qb *QueryBuilder[T]) OrWhereSubquery(column string, operator string, subquery *Subquery) *QueryBuilder[T] {
	qb.whereClauses = append(qb.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    subquery,
		And:      false,
	})
	return qb
}
