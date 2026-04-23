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
	ErrEscrowNotFound     = errors.New("escrow not found")
	ErrEscrowAlreadyHeld  = errors.New("escrow already held")
	ErrEscrowNotHeld      = errors.New("escrow not in held status")
	ErrEscrowAlreadyReleased = errors.New("escrow already released")
)

type EscrowService struct {
	escrowRepo *repository.EscrowRepository
	walletSvc  *WalletService
	ledgerSvc  *ledger.LedgerService
	db         *pgxpool.Pool
}

func NewEscrowService(
	escrowRepo *repository.EscrowRepository,
	walletSvc *WalletService,
	ledgerSvc *ledger.LedgerService,
	db *pgxpool.Pool,
) *EscrowService {
	return &EscrowService{
		escrowRepo: escrowRepo,
		walletSvc:  walletSvc,
		ledgerSvc:  ledgerSvc,
		db:         db,
	}
}

type CreateEscrowRequest struct {
	OrderID   string
	BuyerID   uuid.UUID
	SellerID  uuid.UUID
	Amount    int64
	Reference string
}

func (s *EscrowService) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*models.EscrowTransaction, error) {
	// Check if order already exists
	existing, err := s.escrowRepo.GetByOrderID(ctx, req.OrderID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Get buyer wallet
	buyerWallet, err := s.walletSvc.GetWalletByUserID(ctx, req.BuyerID)
	if err != nil {
		return nil, errors.New("buyer wallet not found")
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Transfer from buyer to escrow (internal escrow wallet)
	// In production, this would go to a dedicated escrow account
	if err := s.ledgerSvc.ExecuteTransfer(ctx, &tx, buyerWallet.ID, buyerWallet.ID, req.Amount, req.Reference, "escrow_hold:"+req.OrderID); err != nil {
		return nil, err
	}

	// Create escrow record
	escrow := &models.EscrowTransaction{
		ID:       uuid.New(),
		OrderID:  req.OrderID,
		BuyerID:  req.BuyerID,
		SellerID: req.SellerID,
		Amount:   req.Amount,
		Status:   models.EscrowStatusHeld,
		HeldAt:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.escrowRepo.Create(ctx, &tx, escrow); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return escrow, nil
}

type ReleaseEscrowRequest struct {
	OrderID   string
	Reference string
}

func (s *EscrowService) ReleaseEscrow(ctx context.Context, req ReleaseEscrowRequest) (*models.EscrowTransaction, error) {
	// Get escrow
	escrow, err := s.escrowRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, ErrEscrowNotFound
	}

	if escrow.Status != models.EscrowStatusHeld {
		return nil, ErrEscrowNotHeld
	}

	// Get seller wallet
	sellerWallet, err := s.walletSvc.GetWalletByUserID(ctx, escrow.SellerID)
	if err != nil {
		return nil, errors.New("seller wallet not found")
	}

	// Get buyer wallet
	buyerWallet, err := s.walletSvc.GetWalletByUserID(ctx, escrow.BuyerID)
	if err != nil {
		return nil, errors.New("buyer wallet not found")
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Release funds from escrow to seller
	// First debit from buyer (reverse the hold)
	if err := s.ledgerSvc.ExecuteWithdrawal(ctx, &tx, buyerWallet.ID, escrow.Amount, req.Reference+"_debit", "escrow_release:"+req.OrderID); err != nil {
		return nil, err
	}

	// Then credit to seller
	if err := s.ledgerSvc.ExecuteDeposit(ctx, &tx, sellerWallet.ID, escrow.Amount, models.TransactionTypeEscrowRelease, req.Reference+"_credit", "escrow_release:"+req.OrderID); err != nil {
		return nil, err
	}

	// Update escrow status
	now := time.Now()
	if err := s.escrowRepo.UpdateStatus(ctx, &tx, escrow.ID, models.EscrowStatusReleased, &now); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.escrowRepo.GetByOrderID(ctx, req.OrderID)
}

func (s *EscrowService) GetEscrow(ctx context.Context, orderID string) (*models.EscrowTransaction, error) {
	return s.escrowRepo.GetByOrderID(ctx, orderID)
}

func (s *EscrowService) GetAllEscrows(ctx context.Context, limit, offset int) ([]models.EscrowTransaction, error) {
	return s.escrowRepo.GetAll(ctx, limit, offset)
}