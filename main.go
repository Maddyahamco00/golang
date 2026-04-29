package main

import (
	"fmt"
	"log"

	"github.com/agri-finance/platform/internal/config"
	"github.com/agri-finance/platform/internal/db"
	"github.com/agri-finance/platform/internal/handler"
	"github.com/agri-finance/platform/internal/repository"
	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
)

/*
=============================================================================
WEEK 1: PROJECT SETUP & FOUNDATION (ACTIVE)
- Database connection
- Wallet creation API
- Basic wallet operations (Get, List)

WEEK 2: TRANSFER ENGINE (COMMENTED - Enable after Week 1 testing)
- Atomic transactions with double-entry ledger
- Idempotency middleware
- Transfer, Deposit, Withdraw endpoints

WEEK 3: ESCROW, KYC, LOANS (COMMENTED - Enable after Week 2 testing)
- Escrow system for secure payments
- KYC integration with external provider
- Loan system with eligibility checks

WEEK 4: ADMIN & SECURITY (COMMENTED - Enable after Week 3 testing)
- Admin APIs for all entities
- JWT authentication
- HMAC request signing
- Rate limiting
- Audit logging
=============================================================================
*/

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize repositories (Week 1 - only wallet and transaction needed)
	walletRepo := repository.NewWalletRepository(database.Pool)
	txRepo := repository.NewTransactionRepository(database.Pool)

	/* WEEK 2: Uncomment after Week 1 testing
	ledgerRepo := repository.NewLedgerRepository(database.Pool)
	escrowRepo := repository.NewEscrowRepository(database.Pool)
	kycRepo := repository.NewKYCRepository(database.Pool)
	loanRepo := repository.NewLoanRepository(database.Pool)
	auditRepo := repository.NewAuditRepository(database.Pool)
	*/

	// Initialize wallet service (Week 1 - basic version without ledger)
	walletSvc := service.NewWalletService(walletRepo, txRepo, nil, database.Pool)

	/* WEEK 2: Uncomment after Week 1 testing
	// Initialize ledger service
	ledgerSvc := ledger.NewLedgerService(ledgerRepo, walletRepo, txRepo)

	// Initialize escrow service
	escrowSvc := service.NewEscrowService(escrowRepo, walletRepo, txRepo, ledgerSvc, database.Pool)

	// Initialize KYC service
	kycSvc := service.NewKYCService(kycRepo, cfg.KYC)

	// Initialize loan service
	loanSvc := service.NewLoanService(loanRepo, walletRepo, txRepo, ledgerSvc, database.Pool)
	*/

	// Initialize handlers (Week 1 - only wallet handler)
	walletHandler := handler.NewWalletHandler(walletSvc)

	/* WEEK 2-4: Uncomment after Week 1 testing
	escrowHandler := handler.NewEscrowHandler(escrowSvc, walletSvc)
	kycHandler := handler.NewKYCHandler(kycSvc)
	loanHandler := handler.NewLoanHandler(loanSvc)
	adminHandler := handler.NewAdminHandler(walletSvc, escrowSvc, loanSvc, kycSvc)

	// Initialize middleware (Week 2-4)
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
	})
	defer redisClient.Close()

	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(redisClient, 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(cfg.App.JWTSecret)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.Limits.RequestsPerMinute)
	*/

	// Setup router
	r := gin.Default()

	/* WEEK 2: Add rate limiting after testing
	r.Use(rateLimiter)
	*/

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Week 1: Wallet routes (basic - no auth for now)
	wallet := r.Group("/wallet")
	{
		wallet.POST("/create", walletHandler.CreateWallet)
		wallet.GET("/:id", walletHandler.GetWallet)
		wallet.GET("/:id/transactions", walletHandler.GetTransactions)
	}

	/* WEEK 2: Uncomment after Week 1 testing
	// Wallet routes with transfer capabilities
	wallet.POST("/transfer", idempotencyMiddleware.Handle, walletHandler.Transfer)
	wallet.POST("/deposit", idempotencyMiddleware.Handle, walletHandler.Deposit)
	wallet.POST("/withdraw", idempotencyMiddleware.Handle, walletHandler.Withdraw)

	// Escrow routes (Week 3)
	escrow := r.Group("/escrow")
	escrow.Use(authMiddleware.Authenticate)
	{
		escrow.POST("/create", idempotencyMiddleware.Handle, escrowHandler.CreateEscrow)
		escrow.POST("/release", idempotencyMiddleware.Handle, escrowHandler.ReleaseEscrow)
		escrow.GET("/:id", escrowHandler.GetEscrow)
	}

	// KYC routes (Week 3)
	kyc := r.Group("/kyc")
	{
		kyc.POST("/submit", kycHandler.SubmitKYC)
		kyc.GET("/:user_id", kycHandler.GetKYC)
		kyc.POST("/verify/:user_id", kycHandler.VerifyKYC)
	}

	// Loan routes (Week 3)
	loan := r.Group("/loan")
	loan.Use(authMiddleware.Authenticate)
	{
		loan.POST("/eligibility", loanHandler.CheckEligibility)
		loan.POST("/apply", idempotencyMiddleware.Handle, loanHandler.ApplyForLoan)
		loan.POST("/disburse/:id", loanHandler.DisburseLoan)
		loan.POST("/repay", idempotencyMiddleware.Handle, loanHandler.RepayLoan)
		loan.GET("/:id", loanHandler.GetLoan)
		loan.GET("/user/:user_id", loanHandler.GetUserLoans)
	}

	// Admin routes (Week 4)
	admin := r.Group("/admin")
	admin.Use(authMiddleware.AdminOnly)
	{
		admin.GET("/wallets", adminHandler.GetAllWallets)
		admin.GET("/wallets/:id", adminHandler.GetWallet)
		admin.GET("/transactions", adminHandler.GetAllTransactions)
		admin.GET("/transactions/:id", adminHandler.GetTransaction)
		admin.GET("/escrow", adminHandler.GetAllEscrows)
		admin.GET("/escrow/:id", adminHandler.GetEscrow)
		admin.GET("/kyc", adminHandler.GetAllKYC)
		admin.GET("/kyc/:user_id", adminHandler.GetKYC)
		admin.GET("/loans", adminHandler.GetAllLoans)
		admin.GET("/loans/:id", adminHandler.GetLoan)
		admin.GET("/audit", adminHandler.GetAuditLogs)
	}
	*/

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Week 1 mode: Basic wallet operations only")
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
