package dialect

import (
	"fmt"
	"strings"
)

// SQLite implements the Dialect interface for SQLite
type SQLite struct{}

// NewSQLite creates a new SQLite dialect
func NewSQLite() *SQLite {
	return &SQLite{}
}

// Name returns the name of the dialect
func (s *SQLite) Name() string {
	return "sqlite"
}

// Placeholder returns the placeholder format for SQLite (?)
func (s *SQLite) Placeholder(index int) string {
	return "?"
}

// QuoteIdentifier quotes an identifier using double quotes
func (s *SQLite) QuoteIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(quoted, ".")
}

// EscapeString escapes a string literal
func (s *SQLite) EscapeString(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}

// BuildLimitOffset builds LIMIT and OFFSET clauses for SQLite
func (s *SQLite) BuildLimitOffset(limit, offset *int) string {
	var parts []string
	if limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *limit))
	}
	if offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %d", *offset))
	}
	return strings.Join(parts, " ")
}

// BuildOrderBy builds ORDER BY clause
func (s *SQLite) BuildOrderBy(orderBy []OrderBy) string {
	if len(orderBy) == 0 {
		return ""
	}
	var parts []string
	for _, ob := range orderBy {
		order := orderASC
		if ob.Order == DESC {
			order = orderDESC
		}
		parts = append(parts, fmt.Sprintf("%s %s", s.QuoteIdentifier(ob.Column), order))
	}
	return "ORDER BY " + strings.Join(parts, ", ")
}

// BuildJoin builds JOIN clause
func (s *SQLite) BuildJoin(join Join) string {
	return fmt.Sprintf("%s %s ON %s", join.Type.String(), s.QuoteIdentifier(join.Table), join.Condition)
}

// SupportLastInsertID returns true for SQLite
func (s *SQLite) SupportLastInsertID() bool {
	return true
}

// LastInsertIDReturning returns empty string for SQLite (uses LastInsertId())
func (s *SQLite) LastInsertIDReturning(tableName string, idColumn string) string {
	return ""
}
