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
	ErrKYCNotFound = errors.New("kyc record not found")
)

type KYCRepository struct {
	db *pgxpool.Pool
}

func NewKYCRepository(db *pgxpool.Pool) *KYCRepository {
	return &KYCRepository{db: db}
}

func (r *KYCRepository) Create(ctx context.Context, kyc *models.KYCRecord) error {
	query := `
		INSERT INTO kyc_records (id, user_id, document_type, document_id, status, response_data, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		kyc.ID,
		kyc.UserID,
		kyc.DocumentType,
		kyc.DocumentID,
		kyc.Status,
		kyc.ResponseData,
		kyc.VerifiedAt,
		kyc.CreatedAt,
		kyc.UpdatedAt,
	)
	return err
}

func (r *KYCRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.KYCRecord, error) {
	query := `
		SELECT id, user_id, document_type, document_id, status, response_data, verified_at, created_at, updated_at
		FROM kyc_records WHERE user_id = $1
	`
	var kyc models.KYCRecord
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&kyc.ID,
		&kyc.UserID,
		&kyc.DocumentType,
		&kyc.DocumentID,
		&kyc.Status,
		&kyc.ResponseData,
		&kyc.VerifiedAt,
		&kyc.CreatedAt,
		&kyc.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrKYCNotFound
	}
	return &kyc, err
}

func (r *KYCRepository) Update(ctx context.Context, kyc *models.KYCRecord) error {
	query := `
		UPDATE kyc_records 
		SET status = $1, response_data = $2, verified_at = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, kyc.Status, kyc.ResponseData, kyc.VerifiedAt, kyc.ID)
	return err
}

func (r *KYCRepository) GetAll(ctx context.Context, limit, offset int) ([]models.KYCRecord, error) {
	query := `
		SELECT id, user_id, document_type, document_id, status, response_data, verified_at, created_at, updated_at
		FROM kyc_records
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.KYCRecord
	for rows.Next() {
		var kyc models.KYCRecord
		err := rows.Scan(
			&kyc.ID,
			&kyc.UserID,
			&kyc.DocumentType,
			&kyc.DocumentID,
			&kyc.Status,
			&kyc.ResponseData,
			&kyc.VerifiedAt,
			&kyc.CreatedAt,
			&kyc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, kyc)
	}
	return records, nil
}
