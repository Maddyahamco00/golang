package repository

import (
	"context"
	"errors"

	"github.com/agri-finance/platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrWalletExists   = errors.New("wallet already exists")
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, balance, currency, tier, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		wallet.ID,
		wallet.UserID,
		wallet.Balance,
		wallet.Currency,
		wallet.Tier,
		wallet.IsActive,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)
	return err
}

func (r *WalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, tier, is_active, created_at, updated_at
		FROM wallets WHERE id = $1
	`
	var wallet models.Wallet
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.Tier,
		&wallet.IsActive,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	return &wallet, err
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, tier, is_active, created_at, updated_at
		FROM wallets WHERE user_id = $1
	`
	var wallet models.Wallet
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.Tier,
		&wallet.IsActive,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	return &wallet, err
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, newBalance int64) error {
	query := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.Exec(ctx, query, newBalance, walletID)
	return err
}

func (r *WalletRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, tier, is_active, created_at, updated_at
		FROM wallets WHERE id = $1 FOR UPDATE
	`
	var wallet models.Wallet
	err := tx.QueryRow(ctx, query, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.Tier,
		&wallet.IsActive,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}
	return &wallet, err
}
