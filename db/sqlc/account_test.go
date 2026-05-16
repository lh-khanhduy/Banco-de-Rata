package db

import (
	"context"
	"testing"
	"time"

	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)

	args := CreateAccountParams{
		Owner:    user.Username,
		Balance:  utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}

	account, err := testStore.CreateAccount(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

// <-- --------------------------***----------------------- -->

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)

	found, err := testStore.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.NotEmpty(t, found)

	require.Equal(t, found, account)
	require.WithinDuration(t, account.CreatedAt, found.CreatedAt, time.Second)

}

func TestUpdateAccount(t *testing.T) {
	originalAcc := createRandomAccount(t)

	updateArgs := UpdateAccountParams{
		ID:      originalAcc.ID,
		Balance: utils.RandomMoney(),
	}

	updatedAcc, err := testStore.UpdateAccount(context.Background(), updateArgs)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc)

	require.Equal(t, originalAcc.ID, updatedAcc.ID)
	require.Equal(t, originalAcc.Owner, updatedAcc.Owner)
	require.NotEqual(t, originalAcc.Balance, updatedAcc.Balance)
	require.Equal(t, updateArgs.Balance, updatedAcc.Balance)
	require.Equal(t, originalAcc.Currency, updatedAcc.Currency)
	require.WithinDuration(t, originalAcc.CreatedAt, updatedAcc.CreatedAt, time.Second)

}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testStore.DeleteAccount(context.Background(), account.ID)

	require.NoError(t, err)

	found, err := testStore.GetAccount(context.Background(), account.ID)

	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, found)
}

func TestListAccounts(t *testing.T) {
	var lastAccount Account

	for range 10 {
		lastAccount = createRandomAccount(t)
	}

	args := ListAccountParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	listAccounts, err := testStore.ListAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, listAccounts)

	for _, account := range listAccounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}
}
