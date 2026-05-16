package db

import (
	"context"
	"testing"
	"time"

	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, acc1, acc2 Account) Transfer {
	args := CreateTransferParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc2.ID,
		Amount:        utils.RandomMoney(),
	}

	transfer, err := testStore.CreateTransfer(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, args.FromAccountID, transfer.FromAccountID)
	require.Equal(t, args.ToAccountID, transfer.ToAccountID)
	require.Equal(t, args.Amount, transfer.Amount)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

// <-- ----------------------------------***-------------------------------------- -->

func TestCreateTransfer(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	createRandomTransfer(t, acc1, acc2)
}

func TestGetTransfer(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	transfer := createRandomTransfer(t, acc1, acc2)

	found, err := testStore.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, found)

	require.Equal(t, found, transfer)
	require.WithinDuration(t, transfer.CreatedAt, found.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	for i := 1; i <= 5; i++ {
		createRandomTransfer(t, acc1, acc2)
		createRandomTransfer(t, acc2, acc1)
	}

	args := ListTransfersParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc1.ID,
		Limit:         5,
		Offset:        5,
	}

	listTransfers, err := testStore.ListTransfers(context.Background(), args)

	require.NoError(t, err)
	require.Len(t, listTransfers, 5)

	for _, transfer := range listTransfers {
		require.NotEmpty(t, transfer)
		require.True(t, transfer.FromAccountID == acc1.ID || transfer.ToAccountID == acc1.ID)
	}
}
