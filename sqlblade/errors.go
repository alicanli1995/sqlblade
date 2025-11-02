package sqlblade

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNoRows is returned when a query returns no rows
	ErrNoRows = errors.New("sqlblade: no rows in result set")

	// ErrInvalidOperator is returned when an invalid operator is used in WHERE clause
	ErrInvalidOperator = errors.New("sqlblade: invalid operator")

	// ErrNilDB is returned when a nil database connection is provided
	ErrNilDB = errors.New("sqlblade: nil database connection")

	// ErrNilContext is returned when a nil context is provided
	ErrNilContext = errors.New("sqlblade: nil context")

	// ErrInvalidModel is returned when a model doesn't implement required interface
	ErrInvalidModel = errors.New("sqlblade: invalid model")

	// ErrNoTableName is returned when table name cannot be determined
	ErrNoTableName = errors.New("sqlblade: cannot determine table name")

	// ErrInvalidColumn is returned when a column doesn't exist
	ErrInvalidColumn = errors.New("sqlblade: invalid column")

	// ErrEmptySet is returned when trying to insert/update with empty data
	ErrEmptySet = errors.New("sqlblade: empty data set")

	// ErrTransactionCommit is returned when transaction commit fails
	ErrTransactionCommit = errors.New("sqlblade: transaction commit failed")
)

// QueryError wraps a database error with query context
type QueryError struct {
	Query string
	Args  []interface{}
	Err   error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("sqlblade: query failed: %v (query: %s, args: %v)", e.Err, e.Query, e.Args)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// IsNoRows checks if the error is ErrNoRows or sql.ErrNoRows
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}

// IsDuplicateKey checks if the error is a duplicate key constraint violation
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "duplicate key") ||
		contains(errStr, "unique constraint") ||
		contains(errStr, "duplicate entry") ||
		contains(errStr, "UNIQUE constraint failed")
}

// IsForeignKeyViolation checks if the error is a foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "foreign key constraint") ||
		contains(errStr, "foreign key") ||
		contains(errStr, "violates foreign key constraint")
}

// IsConnectionError checks if the error is a database connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "connection") ||
		contains(errStr, "network") ||
		contains(errStr, "timeout") ||
		contains(errStr, "refused")
}

func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// wrapQueryError wraps a database error with query context
func wrapQueryError(err error, query string, args []interface{}) error {
	if err == nil {
		return nil
	}
	return &QueryError{
		Query: query,
		Args:  args,
		Err:   err,
	}
}
