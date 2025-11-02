package sqlblade

import (
	"strings"

	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
)

// WhereClause represents a WHERE condition
type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
	And      bool // true = AND, false = OR
}

// Valid operators for WHERE clauses
var validOperators = map[string]bool{
	"=":           true,
	"!=":          true,
	"<>":          true,
	">":           true,
	">=":          true,
	"<":           true,
	"<=":          true,
	"IN":          true,
	"NOT IN":      true,
	"LIKE":        true,
	"NOT LIKE":    true,
	"IS NULL":     true,
	"IS NOT NULL": true,
	"BETWEEN":     true,
	"NOT BETWEEN": true,
}

// isValidOperator checks if an operator is valid
func isValidOperator(op string) bool {
	return validOperators[strings.ToUpper(strings.TrimSpace(op))]
}

// buildWhereClause builds WHERE clause SQL
func buildWhereClause(d dialect.Dialect, clauses []WhereClause, paramIndex *int) (string, []interface{}) {
	if len(clauses) == 0 {
		return "", nil
	}

	var parts []string
	var args []interface{}

	for i, clause := range clauses {
		var condition string
		op := strings.ToUpper(strings.TrimSpace(clause.Operator))

		if !isValidOperator(op) {
			continue // Skip invalid operators
		}

		// Build condition based on operator
		switch op {
		case "IS NULL", "IS NOT NULL":
			condition = d.QuoteIdentifier(clause.Column) + " " + op
		case "IN", "NOT IN":
			if values, ok := clause.Value.([]interface{}); ok && len(values) > 0 {
				placeholders := make([]string, len(values))
				for j := range values {
					*paramIndex++
					placeholders[j] = d.Placeholder(*paramIndex)
					args = append(args, values[j])
				}
				condition = d.QuoteIdentifier(clause.Column) + " " + op + " (" + strings.Join(placeholders, ", ") + ")"
			}
		case "BETWEEN", "NOT BETWEEN":
			if values, ok := clause.Value.([]interface{}); ok && len(values) == 2 {
				*paramIndex++
				ph1 := d.Placeholder(*paramIndex)
				*paramIndex++
				ph2 := d.Placeholder(*paramIndex)
				condition = d.QuoteIdentifier(clause.Column) + " " + op + " " + ph1 + " AND " + ph2
				args = append(args, values[0], values[1])
			}
		default:
			*paramIndex++
			condition = d.QuoteIdentifier(clause.Column) + " " + op + " " + d.Placeholder(*paramIndex)
			args = append(args, clause.Value)
		}

		if condition != "" {
			if i > 0 {
				if clause.And {
					parts = append(parts, "AND")
				} else {
					parts = append(parts, "OR")
				}
			}
			parts = append(parts, condition)
		}
	}

	if len(parts) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(parts, " "), args
}
