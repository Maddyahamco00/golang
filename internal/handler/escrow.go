package handler

import (
	"net/http"

	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EscrowHandler struct {
	escrowSvc *service.EscrowService
	walletSvc *service.WalletService
}

func NewEscrowHandler(escrowSvc *service.EscrowService, walletSvc *service.WalletService) *EscrowHandler {
	return &EscrowHandler{
		escrowSvc: escrowSvc,
		walletSvc: walletSvc,
	}
}

// POST /escrow/create
type CreateEscrowRequest struct {
	SellerID       string `json:"seller_id" binding:"required"`
	BuyerID        string `json:"buyer_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	ExternalRef    string `json:"external_ref" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *EscrowHandler) CreateEscrow(c *gin.Context) {
	var req CreateEscrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sellerID, err := uuid.Parse(req.SellerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seller_id"})
		return
	}

	buyerID, err := uuid.Parse(req.BuyerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid buyer_id"})
		return
	}

	escrow, err := h.escrowSvc.CreateEscrow(c.Request.Context(), sellerID, buyerID, req.Amount, req.ExternalRef, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, escrow)
}

// POST /escrow/release
type ReleaseEscrowRequest struct {
	EscrowID       string `json:"escrow_id" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *EscrowHandler) ReleaseEscrow(c *gin.Context) {
	var req ReleaseEscrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	escrowID, err := uuid.Parse(req.EscrowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid escrow_id"})
		return
	}

	escrow, err := h.escrowSvc.ReleaseEscrow(c.Request.Context(), escrowID, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, escrow)
}

// GET /escrow/:id
func (h *EscrowHandler) GetEscrow(c *gin.Context) {
	escrowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid escrow_id"})
		return
	}

	escrow, err := h.escrowSvc.GetEscrow(c.Request.Context(), escrowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "escrow not found"})
		return
	}

	c.JSON(http.StatusOK, escrow)
}
