package db

import (
	"context"
	"testing"
	"time"

	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, account Account) Entry {
	args := CreateEntryParams{
		AccountID: account.ID,
		Amount:    utils.RandomMoney(),
	}

	entry, err := testStore.CreateEntry(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, args.AccountID, entry.AccountID)
	require.Equal(t, args.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

// <-- ------------------------------***------------------------------- -->

func TestCreateEntry(t *testing.T) {
	acc := createRandomAccount(t)
	createRandomEntry(t, acc)
}

func TestGetEntry(t *testing.T) {
	acc := createRandomAccount(t)
	entry := createRandomEntry(t, acc)

	found, err := testStore.GetEntry(context.Background(), entry.ID)

	require.NoError(t, err)
	require.NotEmpty(t, found)

	require.Equal(t, entry, found)
	require.WithinDuration(t, entry.CreatedAt, found.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	acc := createRandomAccount(t)

	for i := 0; i <= 10; i++ {
		createRandomEntry(t, acc)
	}

	args := ListEntriesParams{
		AccountID: acc.ID,
		Limit:     5,
		Offset:    5,
	}

	listEntries, err := testStore.ListEntries(context.Background(), args)

	require.NoError(t, err)
	require.Len(t, listEntries, 5)

	for _, entry := range listEntries {
		require.NotEmpty(t, entry)
		require.Equal(t, args.AccountID, entry.AccountID)
	}
}
