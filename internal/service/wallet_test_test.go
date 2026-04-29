package service

import (
	"context"
	"testing"

	"github.com/agri-finance/platform/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletService_CreateWallet(t *testing.T) {
	// Skip integration test if no DB
	ctx := context.Background()
	connStr := "postgres://postgres:postgres@localhost:5432/agri_finance_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Skip("Database not available:", err)
	}
	defer pool.Close()

	walletRepo := repository.NewWalletRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)

	svc := NewWalletService(walletRepo, txRepo, nil, pool)

	userID := uuid.New()
	wallet, err := svc.CreateWallet(ctx, CreateWalletRequest{
		UserID:   userID,
		Currency: "NGN",
		Tier:     1,
	})

	require.NoError(t, err)
	assert.Equal(t, userID, wallet.UserID)
	assert.Equal(t, int64(0), wallet.Balance)
	assert.True(t, wallet.IsActive)
	assert.Equal(t, "NGN", wallet.Currency)
}

func TestWalletService_GetWallet(t *testing.T) {
	// Create wallet first
	ctx := context.Background()
	connStr := "postgres://postgres:postgres@localhost:5432/agri_finance_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Skip("Database not available:", err)
	}
	defer pool.Close()

	walletRepo := repository.NewWalletRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	svc := NewWalletService(walletRepo, txRepo, nil, pool)

	userID := uuid.New()
	createdWallet, err := svc.CreateWallet(ctx, CreateWalletRequest{
		UserID:   userID,
		Currency: "NGN",
		Tier:     1,
	})
	require.NoError(t, err)

	// Get wallet
	fetchedWallet, err := svc.GetWallet(ctx, createdWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, createdWallet.ID, fetchedWallet.ID)
	assert.Equal(t, createdWallet.Balance, fetchedWallet.Balance)
}

