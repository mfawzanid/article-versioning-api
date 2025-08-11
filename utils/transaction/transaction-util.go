package transactionutil

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
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


var (
	ErrRollback = errors.New("error db: rollback error")
	ErrCommit   = errors.New("error db: commit error")
)

type Transaction interface {
	InitTransaction() *gorm.DB
	SettleTransaction(tx *gorm.DB, err error) error
}

type transaction struct {
	db *gorm.DB
}

func NewConnection(db *gorm.DB) Transaction {
	return &transaction{
		db: db,
	}
}

func (t *transaction) InitTransaction() *gorm.DB {
	return t.db.Begin()
}

func (t *transaction) SettleTransaction(tx *gorm.DB, err error) error {
	if tx == nil {
		return nil
	}
	if err != nil {
		errRollback := tx.Rollback().Error
		if errRollback != nil {
			return ErrRollback
		}
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return ErrCommit
	}
	return nil
}

func GetTransaction(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 {
		return tx[0]
	}
	return nil
}
