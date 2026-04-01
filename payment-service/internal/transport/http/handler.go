package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-service/internal/domain"
)

// PaymentHandler handles HTTP requests for payments.
type PaymentHandler struct {
	service domain.PaymentService
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(service domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// AuthorizePaymentRequest is the request body for POST /payments.
type AuthorizePaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required,gt=0"`
}

// PaymentResponse is the response body for payment endpoints.
type PaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// AuthorizePayment handles POST /payments.
func (h *PaymentHandler) AuthorizePayment(c *gin.Context) {
	var req AuthorizePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.service.AuthorizePayment(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
		return
	}

	response := PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPaymentStatus handles GET /payments/{order_id}.
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
		return
	}

	payment, err := h.service.GetPaymentStatus(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	response := PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// SetupRoutes configures the payment routes.
func SetupRoutes(router *gin.Engine, handler *PaymentHandler) {
	router.POST("/payments", handler.AuthorizePayment)
	router.GET("/payments/:order_id", handler.GetPaymentStatus)
}
