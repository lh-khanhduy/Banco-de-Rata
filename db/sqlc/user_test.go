package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword("secret")
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	args := CreateUserParams{
		Username:       utils.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), args)

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

	found, err := testStore.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, found)

	require.Equal(t, found, user)
	require.WithinDuration(t, user.PasswordChangedAt, found.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user.CreatedAt, found.CreatedAt, time.Second)

}

func TestUpdateOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := utils.RandomOwner()

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newFullName, updatedUser.FullName)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Email, updatedUser.Email)

	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
}

func TestUpdateOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := utils.RandomEmail()

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newEmail, updatedUser.Email)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)

	require.NotEqual(t, oldUser.Email, updatedUser.Email)

}

func TestUpdateOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newPassword := utils.RandomString(10)
	newHashedPassword, err := utils.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)

	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateAllField(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := utils.RandomOwner()
	newEmail := utils.RandomEmail()

	newPassword := utils.RandomString(10)
	newHashedPassword, err := utils.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		HashedPassword: pgtype.Text{
			String: newHashedPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)

	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)

	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}
