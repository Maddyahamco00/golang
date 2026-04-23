package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/agri-finance/platform/internal/config"
	"github.com/agri-finance/platform/internal/models"
	"github.com/agri-finance/platform/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrKYCNotFound           = errors.New("kyc record not found")
	ErrKYCVerificationFailed = errors.New("kyc verification failed")
)

type KYCService struct {
	kycRepo    *repository.KYCRepository
	walletRepo *repository.WalletRepository
	cfg        *config.KYCConfig
	httpClient *http.Client
}

func NewKYCService(
	kycRepo *repository.KYCRepository,
	walletRepo *repository.WalletRepository,
	cfg *config.KYCConfig,
) *KYCService {
	return &KYCService{
		kycRepo:    kycRepo,
		walletRepo: walletRepo,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type SubmitKYCRequest struct {
	UserID       uuid.UUID
	DocumentType string // BVN, NIN
	DocumentID   string
}

type KYCVerificationResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
}

func (s *KYCService) SubmitKYC(ctx context.Context, req SubmitKYCRequest) (*models.KYCRecord, error) {
	// Check if KYC already exists
	existing, err := s.kycRepo.GetByUserID(ctx, req.UserID)
	if err == nil && existing != nil {
		// If already verified, return existing
		if existing.Status == models.KYCStatusVerified {
			return existing, nil
		}
		// If pending, allow re-submission
	}

	// Call external KYC API (simulated)
	resp, err := s.verifyDocument(ctx, req.DocumentType, req.DocumentID)
	if err != nil {
		return nil, err
	}

	// Determine status based on verification
	status := models.KYCStatusPending
	if resp.Status == "verified" || resp.Status == "success" {
		status = models.KYCStatusVerified
	} else if resp.Status == "rejected" || resp.Status == "failed" {
		status = models.KYCStatusRejected
	}

	// Store KYC record
	kyc := &models.KYCRecord{
		ID:           uuid.New(),
		UserID:       req.UserID,
		DocumentType: req.DocumentType,
		DocumentID:   req.DocumentID,
		Status:       status,
		ResponseData: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Marshal response
	respJSON, _ := json.Marshal(resp)
	kyc.ResponseData = string(respJSON)

	if status == models.KYCStatusVerified {
		now := time.Now()
		kyc.VerifiedAt = &now
	}

	if existing != nil {
		// Update existing
		kyc.ID = existing.ID
		if err := s.kycRepo.Update(ctx, kyc); err != nil {
			return nil, err
		}
	} else {
		if err := s.kycRepo.Create(ctx, kyc); err != nil {
			return nil, err
		}
	}

	// Update wallet tier if verified
	if status == models.KYCStatusVerified {
		s.updateWalletTier(ctx, req.UserID, 2)
	}

	return kyc, nil
}

func (s *KYCService) verifyDocument(ctx context.Context, docType, docID string) (*KYCVerificationResponse, error) {
	// In production, this would call the actual KYC provider API
	// For MVP, we simulate the response
	// This is where you'd integrate with BVN/NIN providers

	// Simulated response
	return &KYCVerificationResponse{
		Status:  "verified",
		Message: "Document verified successfully",
	}, nil

	// Production code would look like:
	// url := fmt.Sprintf("%s/verify/%s", s.cfg.BaseURL, docType)
	// req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	// req.Header.Add("Authorization", "Bearer "+s.cfg.APIKey)
	// req.Header.Add("X-Document-ID", docID)
	// resp, err := s.httpClient.Do(req)
	// ...
}

func (s *KYCService) GetKYC(ctx context.Context, userID uuid.UUID) (*models.KYCRecord, error) {
	return s.kycRepo.GetByUserID(ctx, userID)
}

func (s *KYCService) GetAllKYC(ctx context.Context, limit, offset int) ([]models.KYCRecord, error) {
	return s.kycRepo.GetAll(ctx, limit, offset)
}

func (s *KYCService) updateWalletTier(ctx context.Context, userID uuid.UUID, tier int) error {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Update tier - in production, use a proper update method
	wallet.Tier = tier
	return nil
}

// GetTierLimit returns the daily transaction limit for a given tier
func (s *KYCService) GetTierLimit(tier int, limits *config.LimitsConfig) int64 {
	if tier >= 2 {
		return limits.Tier2DailyLimit
	}
	return limits.Tier1DailyLimit
}
