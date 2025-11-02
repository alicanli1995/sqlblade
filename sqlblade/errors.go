package sqlblade

import "errors"

var (
	// ErrNoRows is returned when a query returns no rows
	ErrNoRows = errors.New("qb: no rows in result set")

	// ErrInvalidOperator is returned when an invalid operator is used in WHERE clause
	ErrInvalidOperator = errors.New("qb: invalid operator")

	// ErrNilDB is returned when a nil database connection is provided
	ErrNilDB = errors.New("qb: nil database connection")

	// ErrNilContext is returned when a nil context is provided
	ErrNilContext = errors.New("qb: nil context")

	// ErrInvalidModel is returned when a model doesn't implement required interface
	ErrInvalidModel = errors.New("qb: invalid model")

	// ErrNoTableName is returned when table name cannot be determined
	ErrNoTableName = errors.New("qb: cannot determine table name")

	// ErrInvalidColumn is returned when a column doesn't exist
	ErrInvalidColumn = errors.New("qb: invalid column")

	// ErrEmptySet is returned when trying to insert/update with empty data
	ErrEmptySet = errors.New("qb: empty data set")
)

