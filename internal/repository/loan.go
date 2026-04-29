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
	ErrLoanNotFound = errors.New("loan not found")
)

type LoanRepository struct {
	db *pgxpool.Pool
}

func NewLoanRepository(db *pgxpool.Pool) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, tx pgx.Tx, loan *models.Loan) error {
	query := `
		INSERT INTO loans (id, user_id, amount, interest_rate, total_repayment, amount_paid, status, approved_at, disbursed_at, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := tx.Exec(ctx, query,
		loan.ID,
		loan.UserID,
		loan.Amount,
		loan.InterestRate,
		loan.TotalRepayment,
		loan.AmountPaid,
		loan.Status,
		loan.ApprovedAt,
		loan.DisbursedAt,
		loan.DueDate,
		loan.CreatedAt,
		loan.UpdatedAt,
	)
	return err
}

func (r *LoanRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Loan, error) {
	query := `
		SELECT id, user_id, amount, interest_rate, total_repayment, amount_paid, status, approved_at, disbursed_at, due_date, created_at, updated_at
		FROM loans WHERE id = $1
	`
	var loan models.Loan
	err := r.db.QueryRow(ctx, query, id).Scan(
		&loan.ID,
		&loan.UserID,
		&loan.Amount,
		&loan.InterestRate,
		&loan.TotalRepayment,
		&loan.AmountPaid,
		&loan.Status,
		&loan.ApprovedAt,
		&loan.DisbursedAt,
		&loan.DueDate,
		&loan.CreatedAt,
		&loan.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLoanNotFound
	}
	return &loan, err
}

func (r *LoanRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Loan, error) {
	query := `
		SELECT id, user_id, amount, interest_rate, total_repayment, amount_paid, status, approved_at, disbursed_at, due_date, created_at, updated_at
		FROM loans WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var loan models.Loan
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Amount,
			&loan.InterestRate,
			&loan.TotalRepayment,
			&loan.AmountPaid,
			&loan.Status,
			&loan.ApprovedAt,
			&loan.DisbursedAt,
			&loan.DueDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		loans = append(loans, loan)
	}
	return loans, nil
}

func (r *LoanRepository) Update(ctx context.Context, tx pgx.Tx, loan *models.Loan) error {
	query := `
		UPDATE loans
		SET amount_paid = $1, status = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := tx.Exec(ctx, query, loan.AmountPaid, loan.Status, loan.ID)
	return err
}

func (r *LoanRepository) GetAll(ctx context.Context, limit, offset int) ([]models.Loan, error) {
	query := `
		SELECT id, user_id, amount, interest_rate, total_repayment, amount_paid, status, approved_at, disbursed_at, due_date, created_at, updated_at
		FROM loans
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var loan models.Loan
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Amount,
			&loan.InterestRate,
			&loan.TotalRepayment,
			&loan.AmountPaid,
			&loan.Status,
			&loan.ApprovedAt,
			&loan.DisbursedAt,
			&loan.DueDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		loans = append(loans, loan)
	}
	return loans, nil
}

func (r *LoanRepository) GetActiveLoans(ctx context.Context) ([]models.Loan, error) {
	query := `
		SELECT id, user_id, amount, interest_rate, total_repayment, amount_paid, status, approved_at, disbursed_at, due_date, created_at, updated_at
		FROM loans WHERE status = 'active'
		ORDER BY due_date ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var loan models.Loan
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Amount,
			&loan.InterestRate,
			&loan.TotalRepayment,
			&loan.AmountPaid,
			&loan.Status,
			&loan.ApprovedAt,
			&loan.DisbursedAt,
			&loan.DueDate,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		loans = append(loans, loan)
	}
	return loans, nil
}
