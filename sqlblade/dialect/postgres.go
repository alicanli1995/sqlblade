package dialect

import (
	"fmt"
	"strings"
)

// PostgreSQL implements the Dialect interface for PostgreSQL
type PostgreSQL struct{}

// NewPostgreSQL creates a new PostgreSQL dialect
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

// Name returns the name of the dialect
func (p *PostgreSQL) Name() string {
	return "postgres"
}

// Placeholder returns the placeholder format for PostgreSQL ($1, $2, etc.)
func (p *PostgreSQL) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

// QuoteIdentifier quotes an identifier using double quotes
func (p *PostgreSQL) QuoteIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
	}
	return strings.Join(quoted, ".")
}

// EscapeString escapes a string literal
func (p *PostgreSQL) EscapeString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// BuildLimitOffset builds LIMIT and OFFSET clauses for PostgreSQL
func (p *PostgreSQL) BuildLimitOffset(limit, offset *int) string {
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
func (p *PostgreSQL) BuildOrderBy(orderBy []OrderBy) string {
	if len(orderBy) == 0 {
		return ""
	}
	var parts []string
	for _, ob := range orderBy {
		order := "ASC"
		if ob.Order == DESC {
			order = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s", p.QuoteIdentifier(ob.Column), order))
	}
	return "ORDER BY " + strings.Join(parts, ", ")
}

// BuildJoin builds JOIN clause
func (p *PostgreSQL) BuildJoin(join Join) string {
	return fmt.Sprintf("%s %s ON %s", join.Type.String(), p.QuoteIdentifier(join.Table), join.Condition)
}

// SupportLastInsertID returns false for PostgreSQL (uses RETURNING instead)
func (p *PostgreSQL) SupportLastInsertID() bool {
	return false
}

// LastInsertIDReturning returns the SQL for returning last insert ID
func (p *PostgreSQL) LastInsertIDReturning(tableName string, idColumn string) string {
	return fmt.Sprintf("RETURNING %s", p.QuoteIdentifier(idColumn))
}
