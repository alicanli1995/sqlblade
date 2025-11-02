package sqlblade

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// WithTransaction executes a function within a database transaction
func WithTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Printf("transaction rollback failed: %v", rollbackErr)
				return
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %w)", rbErr, err)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("%w: %w", ErrTransactionCommit, commitErr)
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
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Printf("transaction rollback failed: %v", rollbackErr)
				return
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction rollback failed: %w (original error: %w)", rbErr, err)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("%w: %w", ErrTransactionCommit, commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}
