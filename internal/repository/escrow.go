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
	ErrEscrowNotFound = errors.New("escrow not found")
)

type EscrowRepository struct {
	db *pgxpool.Pool
}

func NewEscrowRepository(db *pgxpool.Pool) *EscrowRepository {
	return &EscrowRepository{db: db}
}

func (r *EscrowRepository) Create(ctx context.Context, tx *pgxpool.Tx, escrow *models.EscrowTransaction) error {
	query := `
		INSERT INTO escrow_transactions (id, order_id, buyer_id, seller_id, amount, status, held_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := (*tx).Exec(ctx, query,
		escrow.ID,
		escrow.OrderID,
		escrow.BuyerID,
		escrow.SellerID,
		escrow.Amount,
		escrow.Status,
		escrow.HeldAt,
		escrow.CreatedAt,
		escrow.UpdatedAt,
	)
	return err
}

func (r *EscrowRepository) GetByOrderID(ctx context.Context, orderID string) (*models.EscrowTransaction, error) {
	query := `
		SELECT id, order_id, buyer_id, seller_id, amount, status, held_at, released_at, created_at, updated_at
		FROM escrow_transactions WHERE order_id = $1
	`
	var escrow models.EscrowTransaction
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&escrow.ID,
		&escrow.OrderID,
		&escrow.BuyerID,
		&escrow.SellerID,
		&escrow.Amount,
		&escrow.Status,
		&escrow.HeldAt,
		&escrow.ReleasedAt,
		&escrow.CreatedAt,
		&escrow.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEscrowNotFound
	}
	return &escrow, err
}

func (r *EscrowRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EscrowTransaction, error) {
	query := `
		SELECT id, order_id, buyer_id, seller_id, amount, status, held_at, released_at, created_at, updated_at
		FROM escrow_transactions WHERE id = $1
	`
	var escrow models.EscrowTransaction
	err := r.db.QueryRow(ctx, query, id).Scan(
		&escrow.ID,
		&escrow.OrderID,
		&escrow.BuyerID,
		&escrow.SellerID,
		&escrow.Amount,
		&escrow.Status,
		&escrow.HeldAt,
		&escrow.ReleasedAt,
		&escrow.CreatedAt,
		&escrow.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEscrowNotFound
	}
	return &escrow, err
}

func (r *EscrowRepository) UpdateStatus(ctx context.Context, tx *pgxpool.Tx, id uuid.UUID, status models.EscrowStatus, releasedAt *interface{}) error {
	query := `UPDATE escrow_transactions SET status = $1, released_at = $2, updated_at = NOW() WHERE id = $3`
	_, err := (*tx).Exec(ctx, query, status, releasedAt, id)
	return err
}

func (r *EscrowRepository) GetAll(ctx context.Context, limit, offset int) ([]models.EscrowTransaction, error) {
	query := `
		SELECT id, order_id, buyer_id, seller_id, amount, status, held_at, released_at, created_at, updated_at
		FROM escrow_transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var escrows []models.EscrowTransaction
	for rows.Next() {
		var e models.EscrowTransaction
		err := rows.Scan(
			&e.ID,
			&e.OrderID,
			&e.BuyerID,
			&e.SellerID,
			&e.Amount,
			&e.Status,
			&e.HeldAt,
			&e.ReleasedAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		escrows = append(escrows, e)
	}
	return escrows, nil
}
