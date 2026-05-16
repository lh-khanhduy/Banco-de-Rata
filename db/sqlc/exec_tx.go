package db

import (
	"context"
	"fmt"
)

// execTx executes a function within a database transaction
func (s *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	query := New(tx)
	err = fn(query)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rollbackErr)
		}

		return err
	}

	return tx.Commit(ctx)
}
