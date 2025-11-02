package dialect

import (
	"fmt"
	"strings"
)

// MySQL implements the Dialect interface for MySQL
type MySQL struct{}

// NewMySQL creates a new MySQL dialect
func NewMySQL() *MySQL {
	return &MySQL{}
}

// Name returns the name of the dialect
func (m *MySQL) Name() string {
	return "mysql"
}

// Placeholder returns the placeholder format for MySQL (?)
func (m *MySQL) Placeholder(index int) string {
	return "?"
}

// QuoteIdentifier quotes an identifier using backticks
func (m *MySQL) QuoteIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = "`" + strings.ReplaceAll(part, "`", "``") + "`"
	}
	return strings.Join(quoted, ".")
}

// EscapeString escapes a string literal
func (m *MySQL) EscapeString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// BuildLimitOffset builds LIMIT and OFFSET clauses for MySQL
func (m *MySQL) BuildLimitOffset(limit, offset *int) string {
	if limit == nil && offset == nil {
		return ""
	}
	if limit != nil && offset != nil {
		return fmt.Sprintf("LIMIT %d OFFSET %d", *limit, *offset)
	}
	if limit != nil {
		return fmt.Sprintf("LIMIT %d", *limit)
	}
	return fmt.Sprintf("LIMIT 18446744073709551615 OFFSET %d", *offset) // MySQL requires LIMIT when using OFFSET
}

// BuildOrderBy builds ORDER BY clause
func (m *MySQL) BuildOrderBy(orderBy []OrderBy) string {
	if len(orderBy) == 0 {
		return ""
	}
	var parts []string
	for _, ob := range orderBy {
		order := orderASC
		if ob.Order == DESC {
			order = orderDESC
		}
		parts = append(parts, fmt.Sprintf("%s %s", m.QuoteIdentifier(ob.Column), order))
	}
	return "ORDER BY " + strings.Join(parts, ", ")
}

// BuildJoin builds JOIN clause
func (m *MySQL) BuildJoin(join Join) string {
	return fmt.Sprintf("%s %s ON %s", join.Type.String(), m.QuoteIdentifier(join.Table), join.Condition)
}

// SupportLastInsertID returns true for MySQL
func (m *MySQL) SupportLastInsertID() bool {
	return true
}

// LastInsertIDReturning returns empty string for MySQL (uses LastInsertId())
func (m *MySQL) LastInsertIDReturning(tableName string, idColumn string) string {
	return ""
}
