package repository

import (
	"context"

	"github.com/agri-finance/platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerRepository struct {
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) CreateEntry(ctx context.Context, tx pgx.Tx, entry *models.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (id, transaction_id, wallet_id, entry_type, amount, balance_before, balance_after, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := tx.Exec(ctx, query,
		entry.ID,
		entry.TransactionID,
		entry.WalletID,
		entry.EntryType,
		entry.Amount,
		entry.BalanceBefore,
		entry.BalanceAfter,
		entry.Description,
		entry.CreatedAt,
	)
	return err
}

func (r *LedgerRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]models.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, wallet_id, entry_type, amount, balance_before, balance_after, description, created_at
		FROM ledger_entries WHERE transaction_id = $1
	`
	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LedgerEntry
	for rows.Next() {
		var entry models.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.WalletID,
			&entry.EntryType,
			&entry.Amount,
			&entry.BalanceBefore,
			&entry.BalanceAfter,
			&entry.Description,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *LedgerRepository) GetByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, wallet_id, entry_type, amount, balance_before, balance_after, description, created_at
		FROM ledger_entries WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LedgerEntry
	for rows.Next() {
		var entry models.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.WalletID,
			&entry.EntryType,
			&entry.Amount,
			&entry.BalanceBefore,
			&entry.BalanceAfter,
			&entry.Description,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
