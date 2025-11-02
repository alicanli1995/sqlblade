package sqlblade

import (
	"context"
	"database/sql"
	"fmt"
)

// WithTransaction executes a function within a database transaction
func WithTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("transaction rollback failed: %v\n", err)
				return
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %v)", rbErr, err)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("%w: %v", ErrTransactionCommit, commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}

// WithTransactionContext executes a function within a database transaction with context
func WithTransactionContext(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %v)", rbErr, err)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("%w: %v", ErrTransactionCommit, commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}
