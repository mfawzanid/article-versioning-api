package transactionutil

import (
	"database/sql"
	"fmt"
)

func SettleTransaction(tx *sql.Tx, err error) error {
	if tx == nil {
		return nil
	}

	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			return fmt.Errorf("error db: rollback error")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error db: commit error")
	}

	return nil
}
