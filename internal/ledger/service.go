package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/agri-finance/platform/internal/models"
	"github.com/agri-finance/platform/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerService struct {
	ledgerRepo *repository.LedgerRepository
	walletRepo *repository.WalletRepository
	txRepo     *repository.TransactionRepository
}

func NewLedgerService(ledgerRepo *repository.LedgerRepository, walletRepo *repository.WalletRepository, txRepo *repository.TransactionRepository) *LedgerService {
	return &LedgerService{
		ledgerRepo: ledgerRepo,
		walletRepo: walletRepo,
		txRepo:     txRepo,
	}
}

// ExecuteTransfer performs an atomic transfer between two wallets using double-entry bookkeeping
// This is the CORE financial operation that ensures data integrity
func (s *LedgerService) ExecuteTransfer(ctx context.Context, tx *pgxpool.Tx, fromWalletID, toWalletID uuid.UUID, amount int64, reference, description string) error {
	// 1. Lock both wallets in a consistent order (by ID) to prevent deadlocks
	walletIDs := []uuid.UUID{fromWalletID, toWalletID}
	if fromWalletID.String() > toWalletID.String() {
		walletIDs = []uuid.UUID{toWalletID, fromWalletID}
	}

	wallets := make(map[uuid.UUID]*models.Wallet)
	for _, id := range walletIDs {
		wallet, err := s.walletRepo.GetForUpdate(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("failed to lock wallet %s: %w", id, err)
		}
		wallets[id] = wallet
	}

	fromWallet := wallets[fromWalletID]
	toWallet := wallets[toWalletID]

	// 2. Validate sufficient balance
	if fromWallet.Balance < amount {
		return fmt.Errorf("insufficient balance: have %d, need %d", fromWallet.Balance, amount)
	}

	// 3. Create the transaction record
	transaction := &models.Transaction{
		ID:              uuid.New(),
		WalletID:        fromWalletID,
		Type:            models.TransactionTypeTransfer,
		Amount:          amount,
		BalanceBefore:   fromWallet.Balance,
		BalanceAfter:    fromWallet.Balance - amount,
		Status:          models.TransactionStatusCompleted,
		Description:     description,
		Reference:       reference,
		RelatedWalletID: &toWalletID,
		CreatedAt:       getTime(),
		UpdatedAt:       getTime(),
	}

	if err := s.txRepo.Create(ctx, tx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 4. Create DEBIT ledger entry (money leaving fromWallet)
	debitEntry := &models.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transaction.ID,
		WalletID:      fromWalletID,
		EntryType:     models.LedgerEntryTypeDebit,
		Amount:        amount,
		BalanceBefore: fromWallet.Balance,
		BalanceAfter:  fromWallet.Balance - amount,
		Description:   fmt.Sprintf("Transfer to %s: %s", toWalletID.String()[:8], description),
		CreatedAt:     getTime(),
	}

	if err := s.ledgerRepo.CreateEntry(ctx, tx, debitEntry); err != nil {
		return fmt.Errorf("failed to create debit entry: %w", err)
	}

	// 5. Update fromWallet balance
	newFromBalance := fromWallet.Balance - amount
	if err := s.walletRepo.UpdateBalance(ctx, tx, fromWalletID, newFromBalance); err != nil {
		return fmt.Errorf("failed to update from wallet balance: %w", err)
	}

	// 6. Create CREDIT ledger entry (money entering toWallet)
	creditEntry := &models.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transaction.ID,
		WalletID:      toWalletID,
		EntryType:     models.LedgerEntryTypeCredit,
		Amount:        amount,
		BalanceBefore: toWallet.Balance,
		BalanceAfter:  toWallet.Balance + amount,
		Description:   fmt.Sprintf("Transfer from %s: %s", fromWalletID.String()[:8], description),
		CreatedAt:     getTime(),
	}

	if err := s.ledgerRepo.CreateEntry(ctx, tx, creditEntry); err != nil {
		return fmt.Errorf("failed to create credit entry: %w", err)
	}

	// 7. Update toWallet balance
	newToBalance := toWallet.Balance + amount
	if err := s.walletRepo.UpdateBalance(ctx, tx, toWalletID, newToBalance); err != nil {
		return fmt.Errorf("failed to update to wallet balance: %w", err)
	}

	return nil
}

// ExecuteDeposit handles money coming INTO a wallet (e.g., funding, loan disbursement)
func (s *LedgerService) ExecuteDeposit(ctx context.Context, tx *pgxpool.Tx, walletID uuid.UUID, amount int64, txType models.TransactionType, reference, description string) error {
	// Lock wallet
	wallet, err := s.walletRepo.GetForUpdate(ctx, tx, walletID)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	// Create transaction
	transaction := &models.Transaction{
		ID:            uuid.New(),
		WalletID:      walletID,
		Type:          txType,
		Amount:        amount,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  wallet.Balance + amount,
		Status:        models.TransactionStatusCompleted,
		Description:   description,
		Reference:     reference,
		CreatedAt:     getTime(),
		UpdatedAt:     getTime(),
	}

	if err := s.txRepo.Create(ctx, tx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Create credit entry (money coming in)
	creditEntry := &models.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transaction.ID,
		WalletID:      walletID,
		EntryType:     models.LedgerEntryTypeCredit,
		Amount:        amount,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  wallet.Balance + amount,
		Description:   description,
		CreatedAt:     getTime(),
	}

	if err := s.ledgerRepo.CreateEntry(ctx, tx, creditEntry); err != nil {
		return fmt.Errorf("failed to create credit entry: %w", err)
	}

	// Update balance
	newBalance := wallet.Balance + amount
	if err := s.walletRepo.UpdateBalance(ctx, tx, walletID, newBalance); err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

// ExecuteWithdrawal handles money leaving a wallet
func (s *LedgerService) ExecuteWithdrawal(ctx context.Context, tx *pgxpool.Tx, walletID uuid.UUID, amount int64, reference, description string) error {
	// Lock wallet
	wallet, err := s.walletRepo.GetForUpdate(ctx, tx, walletID)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	// Validate sufficient balance
	if wallet.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Create transaction
	transaction := &models.Transaction{
		ID:            uuid.New(),
		WalletID:      walletID,
		Type:          models.TransactionTypeWithdrawal,
		Amount:        amount,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  wallet.Balance - amount,
		Status:        models.TransactionStatusCompleted,
		Description:   description,
		Reference:     reference,
		CreatedAt:     getTime(),
		UpdatedAt:     getTime(),
	}

	if err := s.txRepo.Create(ctx, tx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Create debit entry (money going out)
	debitEntry := &models.LedgerEntry{
		ID:            uuid.New(),
		TransactionID: transaction.ID,
		WalletID:      walletID,
		EntryType:     models.LedgerEntryTypeDebit,
		Amount:        amount,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  wallet.Balance - amount,
		Description:   description,
		CreatedAt:     getTime(),
	}

	if err := s.ledgerRepo.CreateEntry(ctx, tx, debitEntry); err != nil {
		return fmt.Errorf("failed to create debit entry: %w", err)
	}

	// Update balance
	newBalance := wallet.Balance - amount
	if err := s.walletRepo.UpdateBalance(ctx, tx, walletID, newBalance); err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

// now() helper for testability
func getTime() time.Time {
	return time.Now()
}