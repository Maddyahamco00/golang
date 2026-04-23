package repository

import (
	"context"

	"github.com/agri-finance/platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.OldValues,
		log.NewValues,
		log.IPAddress,
		log.CreatedAt,
	)
	return err
}

func (r *AuditRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at
		FROM audit_logs WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *AuditRepository) GetByEntityID(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at
		FROM audit_logs WHERE entity_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}