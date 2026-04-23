package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Balance      int64      `json:"balance" db:"balance"` // in kobo/kents
	Currency     string     `json:"currency" db:"currency"`
	Tier         int        `json:"tier" db:"tier"` // 1 or 2
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type TransactionType string

const (
	TransactionTypeTransfer    TransactionType = "transfer"
	TransactionTypeDeposit     TransactionType = "deposit"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeEscrowHold  TransactionType = "escrow_hold"
	TransactionTypeEscrowRelease TransactionType = "escrow_release"
	TransactionTypeLoanDisbursement TransactionType = "loan_disbursement"
	TransactionTypeLoanRepayment TransactionType = "loan_repayment"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed   TransactionStatus = "failed"
	TransactionStatusReversed TransactionStatus = "reversed"
)

type Transaction struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	WalletID        uuid.UUID       `json:"wallet_id" db:"wallet_id"`
	Type            TransactionType `json:"type" db:"type"`
	Amount          int64           `json:"amount" db:"amount"`
	BalanceBefore   int64           `json:"balance_before" db:"balance_before"`
	BalanceAfter    int64           `json:"balance_after" db:"balance_after"`
	Status          TransactionStatus `json:"status" db:"status"`
	Description     string          `json:"description" db:"description"`
	Reference       string          `json:"reference" db:"reference"` // idempotency key
	RelatedWalletID *uuid.UUID      `json:"related_wallet_id,omitempty" db:"related_wallet_id"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type LedgerEntryType string

const (
	LedgerEntryTypeDebit  LedgerEntryType = "debit"
	LedgerEntryTypeCredit LedgerEntryType = "credit"
)

type LedgerEntry struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TransactionID uuid.UUID       `json:"transaction_id" db:"transaction_id"`
	WalletID      uuid.UUID       `json:"wallet_id" db:"wallet_id"`
	EntryType     LedgerEntryType `json:"entry_type" db:"entry_type"`
	Amount        int64           `json:"amount" db:"amount"`
	BalanceBefore int64           `json:"balance_before" db:"balance_before"`
	BalanceAfter  int64           `json:"balance_after" db:"balance_after"`
	Description   string          `json:"description" db:"description"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusVerified KYCStatus = "verified"
	KYCStatusRejected KYCStatus = "rejected"
)

type KYCRecord struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	DocumentType string     `json:"document_type" db:"document_type"` // BVN, NIN
	DocumentID   string     `json:"document_id" db:"document_id"`
	Status       KYCStatus  `json:"status" db:"status"`
	ResponseData string     `json:"response_data" db:"response_data"` // JSON response from provider
	VerifiedAt   *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type LoanStatus string

const (
	LoanStatusPending   LoanStatus = "pending"
	LoanStatusApproved  LoanStatus = "approved"
	LoanStatusRejected  LoanStatus = "rejected"
	LoanStatusActive    LoanStatus = "active"
	LoanStatusPaidOff   LoanStatus = "paid_off"
	LoanStatusDefaulted LoanStatus = "defaulted"
)

type Loan struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	UserID          uuid.UUID   `json:"user_id" db:"user_id"`
	Amount          int64       `json:"amount" db:"amount"`
	InterestRate    float64     `json:"interest_rate" db:"interest_rate"` // e.g., 0.05 = 5%
	TotalRepayment  int64       `json:"total_repayment" db:"total_repayment"`
	AmountPaid      int64       `json:"amount_paid" db:"amount_paid"`
	Status          LoanStatus  `json:"status" db:"status"`
	ApprovedAt      *time.Time  `json:"approved_at,omitempty" db:"approved_at"`
	DisbursedAt     *time.Time  `json:"disbursed_at,omitempty" db:"disbursed_at"`
	DueDate         time.Time   `json:"due_date" db:"due_date"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

type EscrowTransaction struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	OrderID     string               `json:"order_id" db:"order_id"`
	BuyerID     uuid.UUID            `json:"buyer_id" db:"buyer_id"`
	SellerID    uuid.UUID            `json:"seller_id" db:"seller_id"`
	Amount      int64                `json:"amount" db:"amount"`
	Status      EscrowStatus         `json:"status" db:"status"`
	HeldAt      time.Time            `json:"held_at" db:"held_at"`
	ReleasedAt  *time.Time           `json:"released_at,omitempty" db:"released_at"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

type EscrowStatus string

const (
	EscrowStatusHeld    EscrowStatus = "held"
	EscrowStatusReleased EscrowStatus = "released"
	EscrowStatusRefunded EscrowStatus = "refunded"
)