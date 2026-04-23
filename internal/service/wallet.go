package service

import (
	"context"
	"errors"
	"time"

	"github.com/agri-finance/platform/internal/ledger"
	"github.com/agri-finance/platform/internal/models"
	"github.com/agri-finance/platform/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrWalletNotActive     = errors.New("wallet is not active")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrSelfTransfer        = errors.New("cannot transfer to same wallet")
)

type WalletService struct {
	walletRepo *repository.WalletRepository
	txRepo     *repository.TransactionRepository
	ledgerSvc  *ledger.LedgerService
	db         *pgxpool.Pool
}

func NewWalletService(
	walletRepo *repository.WalletRepository,
	txRepo *repository.TransactionRepository,
	ledgerSvc *ledger.LedgerService,
	db *pgxpool.Pool,
) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
		txRepo:     txRepo,
		ledgerSvc:  ledgerSvc,
		db:         db,
	}
}

type CreateWalletRequest struct {
	UserID   uuid.UUID
	Currency string
	Tier     int
}

func (s *WalletService) CreateWallet(ctx context.Context, req CreateWalletRequest) (*models.Wallet, error) {
	// Check if wallet already exists
	existing, err := s.walletRepo.GetByUserID(ctx, req.UserID)
	if err == nil && existing != nil {
		return existing, nil
	}

	wallet := &models.Wallet{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Balance:   0,
		Currency:  req.Currency,
		Tier:      req.Tier,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	return s.walletRepo.GetByID(ctx, walletID)
}

func (s *WalletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

type TransferRequest struct {
	FromWalletID uuid.UUID
	ToWalletID   uuid.UUID
	Amount       int64
	Reference    string // Idempotency key
	Description  string
}

func (s *WalletService) Transfer(ctx context.Context, req TransferRequest) (*models.Transaction, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Prevent self-transfer
	if req.FromWalletID == req.ToWalletID {
		return nil, ErrSelfTransfer
	}

	// Check idempotency - if reference already exists, return existing transaction
	existingTx, err := s.txRepo.GetByReference(ctx, req.Reference)
	if err == nil && existingTx != nil {
		return existingTx, nil
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Execute transfer using ledger service
	if err := s.ledgerSvc.ExecuteTransfer(ctx, &tx, req.FromWalletID, req.ToWalletID, req.Amount, req.Reference, req.Description); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Get the created transaction
	return s.txRepo.GetByReference(ctx, req.Reference)
}

type DepositRequest struct {
	WalletID    uuid.UUID
	Amount      int64
	Reference   string
	Type        models.TransactionType
	Description string
}

func (s *WalletService) Deposit(ctx context.Context, req DepositRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Check idempotency
	existingTx, err := s.txRepo.GetByReference(ctx, req.Reference)
	if err == nil && existingTx != nil {
		return existingTx, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := s.ledgerSvc.ExecuteDeposit(ctx, &tx, req.WalletID, req.Amount, req.Type, req.Reference, req.Description); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.txRepo.GetByReference(ctx, req.Reference)
}

type WithdrawRequest struct {
	WalletID    uuid.UUID
	Amount      int64
	Reference   string
	Description string
}

func (s *WalletService) Withdraw(ctx context.Context, req WithdrawRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Check idempotency
	existingTx, err := s.txRepo.GetByReference(ctx, req.Reference)
	if err == nil && existingTx != nil {
		return existingTx, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := s.ledgerSvc.ExecuteWithdrawal(ctx, &tx, req.WalletID, req.Amount, req.Reference, req.Description); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.txRepo.GetByReference(ctx, req.Reference)
}

func (s *WalletService) GetTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.txRepo.GetByWalletID(ctx, walletID, limit, offset)
}
