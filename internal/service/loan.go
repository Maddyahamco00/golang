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
	ErrLoanNotFound      = errors.New("loan not found")
	ErrLoanNotEligible   = errors.New("not eligible for loan")
	ErrLoanAlreadyActive = errors.New("loan already active")
)

type LoanService struct {
	loanRepo   *repository.LoanRepository
	walletRepo *repository.WalletRepository
	txRepo     *repository.TransactionRepository
	ledgerSvc  *ledger.LedgerService
	db         *pgxpool.Pool
}

func NewLoanService(
	loanRepo *repository.LoanRepository,
	walletRepo *repository.WalletRepository,
	txRepo *repository.TransactionRepository,
	ledgerSvc *ledger.LedgerService,
	db *pgxpool.Pool,
) *LoanService {
	return &LoanService{
		loanRepo:   loanRepo,
		walletRepo: walletRepo,
		txRepo:     txRepo,
		ledgerSvc:  ledgerSvc,
		db:         db,
	}
}

type LoanScore struct {
	TransactionVolume int64   // Total transaction volume
	ActiveDays        int     // Days since first transaction
	LoanCount         int     // Previous loans
	RepaymentRate     float64 // Percentage of loans repaid on time
}

type ApplyLoanRequest struct {
	UserID   uuid.UUID
	Amount   int64
	Duration int // in days
}

// CalculateLoanEligibility calculates if user is eligible for a loan
func (s *LoanService) CalculateLoanEligibility(ctx context.Context, userID uuid.UUID) (*LoanScore, error) {
	// Get user's wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user's transactions
	transactions, err := s.txRepo.GetByWalletID(ctx, wallet.ID, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Calculate transaction volume
	var volume int64
	var firstTxTime time.Time
	for _, tx := range transactions {
		volume += tx.Amount
		if firstTxTime.IsZero() || tx.CreatedAt.Before(firstTxTime) {
			firstTxTime = tx.CreatedAt
		}
	}

	// Get user's loans
	loans, err := s.loanRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate repayment rate
	var repaidOnTime int
	var totalLoans int
	for _, loan := range loans {
		if loan.Status == models.LoanStatusPaidOff {
			totalLoans++
			if loan.DisbursedAt != nil {
				dueDate := loan.DueDate
				// Check if paid before due date
				if loan.UpdatedAt.Before(dueDate) {
					repaidOnTime++
				}
			}
		}
		if loan.Status == models.LoanStatusActive {
			totalLoans++
		}
	}

	repaymentRate := 0.0
	if totalLoans > 0 {
		repaymentRate = float64(repaidOnTime) / float64(totalLoans)
	}

	// Calculate active days
	activeDays := 0
	if !firstTxTime.IsZero() {
		activeDays = int(time.Since(firstTxTime).Hours() / 24)
	}

	return &LoanScore{
		TransactionVolume: volume,
		ActiveDays:        activeDays,
		LoanCount:         totalLoans,
		RepaymentRate:     repaymentRate,
	}, nil
}

// DetermineLoanAmount determines the maximum loan amount based on score
func (s *LoanService) DetermineLoanAmount(score *LoanScore) int64 {
	// Base amount from transaction volume (can borrow up to 50% of 3-month volume)
	baseAmount := score.TransactionVolume * 50 / 100

	// Cap based on activity
	if score.ActiveDays < 30 {
		baseAmount = baseAmount * 30 / 100 // Only 30% for new users
	} else if score.ActiveDays < 90 {
		baseAmount = baseAmount * 70 / 100 // 70% for established users
	}

	// Bonus for good repayment history
	if score.RepaymentRate >= 0.9 {
		baseAmount = baseAmount * 120 / 100 // 20% bonus
	}

	// Cap at reasonable maximum
	if baseAmount > 1000000 { // 10,000 NGN max for MVP
		baseAmount = 1000000
	}

	return baseAmount
}

func (s *LoanService) ApplyForLoan(ctx context.Context, req ApplyLoanRequest) (*models.Loan, error) {
	// Check for existing active loans
	loans, err := s.loanRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	for _, loan := range loans {
		if loan.Status == models.LoanStatusActive {
			return nil, ErrLoanAlreadyActive
		}
	}

	// Calculate eligibility
	score, err := s.CalculateLoanEligibility(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Determine max eligible amount
	maxAmount := s.DetermineLoanAmount(score)
	if req.Amount > maxAmount {
		return nil, ErrLoanNotEligible
	}

	// Basic eligibility checks
	if score.ActiveDays < 7 {
		return nil, ErrLoanNotEligible
	}

	// Calculate interest (5% for MVP)
	interestRate := 0.05
	totalRepayment := req.Amount + int64(float64(req.Amount)*interestRate)

	// Calculate due date
	duration := req.Duration
	if duration == 0 {
		duration = 30 // Default 30 days
	}
	dueDate := time.Now().AddDate(0, 0, duration)

	// Create loan record
	loan := &models.Loan{
		ID:             uuid.New(),
		UserID:         req.UserID,
		Amount:         req.Amount,
		InterestRate:   interestRate,
		TotalRepayment: totalRepayment,
		AmountPaid:     0,
		Status:         models.LoanStatusPending,
		DueDate:        dueDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// For MVP, auto-approve (in production, this would go through review)
	now := time.Now()
	loan.Status = models.LoanStatusApproved
	loan.ApprovedAt = &now

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := s.loanRepo.Create(ctx, tx, loan); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return loan, nil
}

type DisburseLoanRequest struct {
	LoanID    uuid.UUID
	Reference string
}

func (s *LoanService) DisburseLoan(ctx context.Context, req DisburseLoanRequest) (*models.Loan, error) {
	// Get loan
	loan, err := s.loanRepo.GetByID(ctx, req.LoanID)
	if err != nil {
		return nil, ErrLoanNotFound
	}

	if loan.Status != models.LoanStatusApproved {
		return nil, errors.New("loan not approved")
	}

	// Get user wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, loan.UserID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Disburse to wallet
	if err := s.ledgerSvc.ExecuteDeposit(ctx, tx, wallet.ID, loan.Amount, models.TransactionTypeLoanDisbursement, req.Reference, "loan_disbursement"); err != nil {
		return nil, err
	}

	// Update loan status
	now := time.Now()
	loan.Status = models.LoanStatusActive
	loan.DisbursedAt = &now

	if err := s.loanRepo.Update(ctx, tx, loan); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.loanRepo.GetByID(ctx, req.LoanID)
}

type RepayLoanRequest struct {
	LoanID    uuid.UUID
	Amount    int64
	Reference string
}

func (s *LoanService) RepayLoan(ctx context.Context, req RepayLoanRequest) (*models.Loan, error) {
	// Get loan
	loan, err := s.loanRepo.GetByID(ctx, req.LoanID)
	if err != nil {
		return nil, ErrLoanNotFound
	}

	if loan.Status != models.LoanStatusActive {
		return nil, errors.New("loan not active")
	}

	// Get user wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, loan.UserID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Process repayment (withdraw from wallet)
	if err := s.ledgerSvc.ExecuteWithdrawal(ctx, tx, wallet.ID, req.Amount, req.Reference, "loan_repayment"); err != nil {
		return nil, err
	}

	// Update loan
	loan.AmountPaid += req.Amount
	if loan.AmountPaid >= loan.TotalRepayment {
		loan.Status = models.LoanStatusPaidOff
	}

	if err := s.loanRepo.Update(ctx, tx, loan); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.loanRepo.GetByID(ctx, req.LoanID)
}

func (s *LoanService) GetLoan(ctx context.Context, loanID uuid.UUID) (*models.Loan, error) {
	return s.loanRepo.GetByID(ctx, loanID)
}

func (s *LoanService) GetUserLoans(ctx context.Context, userID uuid.UUID) ([]models.Loan, error) {
	return s.loanRepo.GetByUserID(ctx, userID)
}

func (s *LoanService) GetAllLoans(ctx context.Context, limit, offset int) ([]models.Loan, error) {
	return s.loanRepo.GetAll(ctx, limit, offset)
}
