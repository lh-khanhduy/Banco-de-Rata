package db

import (
	"context"
	"testing"
	"time"

	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	args := CreateUserParams{
		Username:       utils.RandomOwner(),
		HashedPassword: "secret",
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

// <-- --------------------------***----------------------- -->

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	found, err := testQueries.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, found)

	require.Equal(t, found, user)
	require.WithinDuration(t, user.PasswordChangedAt, found.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user.CreatedAt, found.CreatedAt, time.Second)

}
