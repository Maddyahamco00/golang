package handler

import (
	"net/http"

	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	walletSvc *service.WalletService
	escrowSvc *service.EscrowService
	loanSvc   *service.LoanService
	kycSvc    *service.KYCService
}

func NewAdminHandler(
	walletSvc *service.WalletService,
	escrowSvc *service.EscrowService,
	loanSvc *service.LoanService,
	kycSvc *service.KYCService,
) *AdminHandler {
	return &AdminHandler{
		walletSvc: walletSvc,
		escrowSvc: escrowSvc,
		loanSvc:   loanSvc,
		kycSvc:    kycSvc,
	}
}

// GET /admin/wallets
func (h *AdminHandler) GetAllWallets(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	// For now, return empty - would need a List method in wallet service
	c.JSON(http.StatusOK, gin.H{
		"data":   []interface{}{},
		"limit":  limit,
		"offset": offset,
	})
}

// GET /admin/wallets/:id
func (h *AdminHandler) GetWallet(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet_id"})
		return
	}

	wallet, err := h.walletSvc.GetWallet(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GET /admin/transactions
func (h *AdminHandler) GetAllTransactions(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	// Would need a List method in wallet service
	c.JSON(http.StatusOK, gin.H{
		"data":   []interface{}{},
		"limit":  limit,
		"offset": offset,
	})
}

// GET /admin/loans
func (h *AdminHandler) GetAllLoans(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	loans, err := h.loanSvc.GetAllLoans(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   loans,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /admin/loans/:id
func (h *AdminHandler) GetLoan(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan_id"})
		return
	}

	loan, err := h.loanSvc.GetLoan(c.Request.Context(), loanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// POST /admin/loans/:id/disburse
func (h *AdminHandler) DisburseLoan(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan_id"})
		return
	}

	loan, err := h.loanSvc.DisburseLoan(c.Request.Context(), service.DisburseLoanRequest{
		LoanID:    loanID,
		Reference: "admin_disburse_" + loanID.String(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// GET /admin/escrows
func (h *AdminHandler) GetAllEscrows(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	escrows, err := h.escrowSvc.GetAllEscrows(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   escrows,
		"limit":  limit,
		"offset": offset,
	})
}

// POST /admin/escrows/:orderId/release
func (h *AdminHandler) ReleaseEscrow(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order_id"})
		return
	}

	escrow, err := h.escrowSvc.ReleaseEscrow(c.Request.Context(), service.ReleaseEscrowRequest{
		OrderID:   orderID,
		Reference: "admin_release_" + orderID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, escrow)
}

// GET /admin/kyc
func (h *AdminHandler) GetAllKYC(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	records, err := h.kycSvc.GetAllKYC(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   records,
		"limit":  limit,
		"offset": offset,
	})
}