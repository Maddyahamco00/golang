package handler

import (
	"net/http"

	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoanHandler struct {
	loanSvc *service.LoanService
}

func NewLoanHandler(loanSvc *service.LoanService) *LoanHandler {
	return &LoanHandler{
		loanSvc: loanSvc,
	}
}

// POST /loan/eligibility
type CheckEligibilityRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (h *LoanHandler) CheckEligibility(c *gin.Context) {
	var req CheckEligibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	eligibility, err := h.loanSvc.CheckEligibility(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, eligibility)
}

// POST /loan/apply
type ApplyLoanRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	Duration       int    `json:"duration" binding:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *LoanHandler) ApplyForLoan(c *gin.Context) {
	var req ApplyLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	loan, err := h.loanSvc.ApplyForLoan(c.Request.Context(), userID, req.Amount, req.Duration, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, loan)
}

// POST /loan/disburse/:id
func (h *LoanHandler) DisburseLoan(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan_id"})
		return
	}

	loan, err := h.loanSvc.DisburseLoan(c.Request.Context(), loanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// POST /loan/repay
type RepayLoanRequest struct {
	LoanID         string `json:"loan_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *LoanHandler) RepayLoan(c *gin.Context) {
	var req RepayLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loanID, err := uuid.Parse(req.LoanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan_id"})
		return
	}

	loan, err := h.loanSvc.RepayLoan(c.Request.Context(), loanID, req.Amount, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// GET /loan/:id
func (h *LoanHandler) GetLoan(c *gin.Context) {
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

// GET /loan/user/:user_id
func (h *LoanHandler) GetUserLoans(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	loans, err := h.loanSvc.GetUserLoans(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": loans})
}
