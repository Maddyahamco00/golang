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
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrDuplicateReference  = errors.New("duplicate reference")
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *pgxpool.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, wallet_id, type, amount, balance_before, balance_after, status, description, reference, related_wallet_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := (*tx).Exec(ctx, query,
		transaction.ID,
		transaction.WalletID,
		transaction.Type,
		transaction.Amount,
		transaction.BalanceBefore,
		transaction.BalanceAfter,
		transaction.Status,
		transaction.Description,
		transaction.Reference,
		transaction.RelatedWalletID,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)
	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_before, balance_after, status, description, reference, related_wallet_id, created_at, updated_at
		FROM transactions WHERE id = $1
	`
	var tx models.Transaction
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.Type,
		&tx.Amount,
		&tx.BalanceBefore,
		&tx.BalanceAfter,
		&tx.Status,
		&tx.Description,
		&tx.Reference,
		&tx.RelatedWalletID,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTransactionNotFound
	}
	return &tx, err
}

func (r *TransactionRepository) GetByReference(ctx context.Context, reference string) (*models.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_before, balance_after, status, description, reference, related_wallet_id, created_at, updated_at
		FROM transactions WHERE reference = $1
	`
	var tx models.Transaction
	err := r.db.QueryRow(ctx, query, reference).Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.Type,
		&tx.Amount,
		&tx.BalanceBefore,
		&tx.BalanceAfter,
		&tx.Status,
		&tx.Description,
		&tx.Reference,
		&tx.RelatedWalletID,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTransactionNotFound
	}
	return &tx, err
}

func (r *TransactionRepository) GetByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, balance_before, balance_after, status, description, reference, related_wallet_id, created_at, updated_at
		FROM transactions WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.WalletID,
			&tx.Type,
			&tx.Amount,
			&tx.BalanceBefore,
			&tx.BalanceAfter,
			&tx.Status,
			&tx.Description,
			&tx.Reference,
			&tx.RelatedWalletID,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, tx *pgxpool.Tx, id uuid.UUID, status models.TransactionStatus) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := (*tx).Exec(ctx, query, status, id)
	return err
}