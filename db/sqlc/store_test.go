package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	// run n concurrent transfer transaction
	n := 5
	amount := int64(10)

	// channels to interact with go concurrency
	errsChan := make(chan error)
	resultsChan := make(chan TransferTxResult)

	for range n {
		go func() {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})

			// write to channels
			errsChan <- err
			resultsChan <- result
		}()
	}

	// check results
	existed := make(map[int]bool)
	for range n {
		err := <-errsChan
		require.NoError(t, err)

		result := <-resultsChan
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, acc1.ID)
		require.Equal(t, transfer.ToAccountID, acc2.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, acc1.ID)
		require.Equal(t, fromEntry.Amount, -amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toEntry.AccountID, acc2.ID)
		require.Equal(t, toEntry.Amount, amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//check accounts
		fromAcc := result.FromAccount
		require.NotEmpty(t, fromAcc)
		require.Equal(t, fromAcc.ID, acc1.ID)

		toAcc := result.ToAccount
		require.NotEmpty(t, toAcc)
		require.Equal(t, toAcc.ID, acc2.ID)

		//check account's balance
		diff1 := acc1.Balance - fromAcc.Balance
		diff2 := acc2.Balance - toAcc.Balance

		require.Equal(t, diff1, -diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//check the final account balance

	updatedAcc1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	require.Equal(t, updatedAcc1.Balance+int64(n)*amount, acc1.Balance)
	require.Equal(t, updatedAcc2.Balance-int64(n)*amount, acc2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := range n {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
