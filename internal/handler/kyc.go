package handler

import (
	"net/http"

	"github.com/agri-finance/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KYCHandler struct {
	kycSvc *service.KYCService
}

func NewKYCHandler(kycSvc *service.KYCService) *KYCHandler {
	return &KYCHandler{
		kycSvc: kycSvc,
	}
}

// POST /kyc/submit
type SubmitKYCRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	Address      string `json:"address" binding:"required"`
	DocumentType string `json:"document_type" binding:"required"`
	DocumentID   string `json:"document_id" binding:"required"`
}

func (h *KYCHandler) SubmitKYC(c *gin.Context) {
	var req SubmitKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	kycRecord, err := h.kycSvc.SubmitKYC(c.Request.Context(), userID, service.KYCSubmission{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
		DocumentType: req.DocumentType,
		DocumentID:   req.DocumentID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, kycRecord)
}

// GET /kyc/:user_id
func (h *KYCHandler) GetKYC(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	kycRecord, err := h.kycSvc.GetKYC(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "KYC record not found"})
		return
	}

	c.JSON(http.StatusOK, kycRecord)
}

// POST /kyc/verify/:user_id
func (h *KYCHandler) VerifyKYC(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	kycRecord, err := h.kycSvc.VerifyKYC(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, kycRecord)
}
