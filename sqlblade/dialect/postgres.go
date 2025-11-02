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

func (p *PostgreSQL) Placeholder(index int) string {
	if index >= 0 && index < len(placeholderCache) {
		return placeholderCache[index]
	}
	return "$" + fastIntToString(index)
}

var placeholderCache = [100]string{
	"$0", "$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9",
	"$10", "$11", "$12", "$13", "$14", "$15", "$16", "$17", "$18", "$19",
	"$20", "$21", "$22", "$23", "$24", "$25", "$26", "$27", "$28", "$29",
	"$30", "$31", "$32", "$33", "$34", "$35", "$36", "$37", "$38", "$39",
	"$40", "$41", "$42", "$43", "$44", "$45", "$46", "$47", "$48", "$49",
	"$50", "$51", "$52", "$53", "$54", "$55", "$56", "$57", "$58", "$59",
	"$60", "$61", "$62", "$63", "$64", "$65", "$66", "$67", "$68", "$69",
	"$70", "$71", "$72", "$73", "$74", "$75", "$76", "$77", "$78", "$79",
	"$80", "$81", "$82", "$83", "$84", "$85", "$86", "$87", "$88", "$89",
	"$90", "$91", "$92", "$93", "$94", "$95", "$96", "$97", "$98", "$99",
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
		order := orderASC
		if ob.Order == DESC {
			order = orderDESC
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
