package dialect

// Dialect defines the interface for database-specific SQL generation
type Dialect interface {
	// Name returns the name of the dialect
	Name() string

	// Placeholder returns the placeholder format for parameterized queries
	// Returns the format string and whether to use positional placeholders
	Placeholder(index int) string

	// QuoteIdentifier quotes an identifier (table/column name)
	QuoteIdentifier(identifier string) string

	// EscapeString escapes a string literal
	EscapeString(s string) string

	// BuildLimitOffset builds LIMIT and OFFSET clauses
	BuildLimitOffset(limit, offset *int) string

	// BuildOrderBy builds ORDER BY clause
	BuildOrderBy(orderBy []OrderBy) string

	// BuildJoin builds JOIN clause
	BuildJoin(join Join) string

	// SupportLastInsertID returns whether the dialect supports LastInsertId()
	SupportLastInsertID() bool

	// LastInsertIDReturning returns the SQL for returning last insert ID (PostgreSQL)
	LastInsertIDReturning(tableName string, idColumn string) string
}

// OrderBy represents an ORDER BY clause
type OrderBy struct {
	Column string
	Order  OrderDirection
}

// OrderDirection represents the order direction
type OrderDirection int

const (
	ASC OrderDirection = iota
	DESC
)

// Join represents a JOIN clause
type Join struct {
	Type      JoinType
	Table     string
	Condition string
}

// JoinType represents the type of JOIN
type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullJoin
)

const (
	joinInner = "INNER JOIN"
	orderASC  = "ASC"
	orderDESC = "DESC"
)

// String returns the SQL string representation of JoinType
func (jt JoinType) String() string {
	switch jt {
	case InnerJoin:
		return joinInner
	case LeftJoin:
		return "LEFT JOIN"
	case RightJoin:
		return "RIGHT JOIN"
	case FullJoin:
		return "FULL JOIN"
	default:
		return joinInner
	}
}
