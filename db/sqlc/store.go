package db

import (
	"context"
	"database/sql"
	"fmt"
)

// *sql.DB
// Connection Pooling: Represents a pool of connections to the database. It manages multiple connections and can be used concurrently by multiple goroutines.
// Execution of Queries: You use *sql.DB to execute queries that do not require transaction management. This includes operations like Query(), Exec(), and Prepare().
// Lifecycle: *sql.DB remains open for the lifetime of your application (or as long as you need it). You can use it to create multiple transactions or run multiple queries.
// *sql.Tx
// Transaction Management: Represents a single transaction. You create a *sql.Tx when you begin a transaction using *sql.DB.Begin().
// Isolation: Operations executed within a *sql.Tx are isolated from other transactions until the transaction is committed or rolled back. This ensures data integrity.
// Lifecycle: The *sql.Tx is short-lived, typically used to execute a series of statements that must be atomic. You must call Commit() or Rollback() on it to finalize or discard the transaction.

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// store define all the functions to execute sql queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB // it required to create a new db transaction
}

// NewStore create a new store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx execute the function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{AccountID: arg.FromAccountID, Amount: -arg.Amount})
		if err != nil {
			return err
		}
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{AccountID: arg.ToAccountID, Amount: arg.Amount})
		if err != nil {
			return err
		}

		//update accounts balance
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1 int64, amount1 int64, accountID2 int64, amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{ID: accountID1, Amount: amount1})
	if err != nil {
		return account1, account2, err
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{ID: accountID2, Amount: amount2})
	if err != nil {
		return account1, account2, err
	}
	return account1, account2, err
}
