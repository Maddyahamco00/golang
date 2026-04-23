package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/agri-finance/platform/internal/config"
	"github.com/agri-finance/platform/internal/db"
	"github.com/agri-finance/platform/internal/handler"
	"github.com/agri-finance/platform/internal/ledger"
	"github.com/agri-finance/platform/internal/middleware"
	"github.com/agri-finance/platform/internal/repository"
	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

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

	// Initialize Redis for idempotency
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	// Initialize repositories
	walletRepo := repository.NewWalletRepository(database.Pool)
	txRepo := repository.NewTransactionRepository(database.Pool)
	ledgerRepo := repository.NewLedgerRepository(database.Pool)
	escrowRepo := repository.NewEscrowRepository(database.Pool)
	kycRepo := repository.NewKYCRepository(database.Pool)
	loanRepo := repository.NewLoanRepository(database.Pool)
	auditRepo := repository.NewAuditRepository(database.Pool)

	// Initialize ledger service
	ledgerSvc := ledger.NewLedgerService(ledgerRepo, walletRepo, txRepo)

	// Initialize wallet service
	walletSvc := service.NewWalletService(walletRepo, txRepo, ledgerSvc, database.Pool)

	// Initialize escrow service
	escrowSvc := service.NewEscrowService(escrowRepo, walletRepo, txRepo, ledgerSvc, database.Pool)

	// Initialize KYC service
	kycSvc := service.NewKYCService(kycRepo, cfg.KYC)

	// Initialize loan service
	loanSvc := service.NewLoanService(loanRepo, walletRepo, txRepo, ledgerSvc, database.Pool)

	// Initialize handlers
	walletHandler := handler.NewWalletHandler(walletSvc)
	escrowHandler := handler.NewEscrowHandler(escrowSvc, walletSvc)
	kycHandler := handler.NewKYCHandler(kycSvc)
	loanHandler := handler.NewLoanHandler(loanSvc)
	adminHandler := handler.NewAdminHandler(walletSvc, escrowSvc, loanSvc, kycSvc)

	// Initialize middleware
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(redisClient, 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(cfg.App.JWTSecret)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.Limits.RequestsPerMinute)

	// Setup router
	r := gin.Default()

	// Apply global middleware
	r.Use(rateLimiter)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Wallet routes (with idempotency)
	wallet := r.Group("/wallet")
	wallet.Use(authMiddleware.Authenticate)
	{
		wallet.POST("/create", idempotencyMiddleware.Handle, walletHandler.CreateWallet)
		wallet.GET("/:id", walletHandler.GetWallet)
		wallet.GET("/:id/transactions", walletHandler.GetTransactions)
		wallet.POST("/transfer", idempotencyMiddleware.Handle, walletHandler.Transfer)
		wallet.POST("/deposit", idempotencyMiddleware.Handle, walletHandler.Deposit)
		wallet.POST("/withdraw", idempotencyMiddleware.Handle, walletHandler.Withdraw)
	}

	// Escrow routes
	escrow := r.Group("/escrow")
	escrow.Use(authMiddleware.Authenticate)
	{
		escrow.POST("/create", idempotencyMiddleware.Handle, escrowHandler.CreateEscrow)
		escrow.POST("/release", idempotencyMiddleware.Handle, escrowHandler.ReleaseEscrow)
		escrow.GET("/:id", escrowHandler.GetEscrow)
	}

	// KYC routes
	kyc := r.Group("/kyc")
	{
		kyc.POST("/submit", kycHandler.SubmitKYC)
		kyc.GET("/:user_id", kycHandler.GetKYC)
		kyc.POST("/verify/:user_id", kycHandler.VerifyKYC)
	}

	// Loan routes
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

	// Admin routes
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

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
