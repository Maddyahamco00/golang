package handler

import (
	"net/http"

	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletSvc *service.WalletService
}

func NewWalletHandler(walletSvc *service.WalletService) *WalletHandler {
	return &WalletHandler{walletSvc: walletSvc}
}

// POST /wallet/create
type CreateWalletRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Currency string `json:"currency" binding:"required"`
	Tier     int    `json:"tier" binding:"required,oneof=1 2"`
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	wallet, err := h.walletSvc.CreateWallet(c.Request.Context(), service.CreateWalletRequest{
		UserID:   userID,
		Currency: req.Currency,
		Tier:     req.Tier,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// GET /wallet/:id
func (h *WalletHandler) GetWallet(c *gin.Context) {
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

// GET /wallet/:id/transactions
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet_id"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		limit = ParseIntDefault(l, 20)
	}
	if o := c.Query("offset"); o != "" {
		offset = ParseIntDefault(o, 0)
	}

	transactions, err := h.walletSvc.GetTransactions(c.Request.Context(), walletID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    transactions,
		"limit":   limit,
		"offset":  offset,
	})
}

