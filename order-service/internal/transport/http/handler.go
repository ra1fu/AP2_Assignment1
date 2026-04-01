package http

import (
	"net/http"

	"order-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	service domain.OrderUseCase
}

// NewOrderHandler creates a new order handler.
func NewOrderHandler(service domain.OrderUseCase) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrderRequest is the request body for POST /orders.
type CreateOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required,gt=0"`
}

// OrderResponse is the response body for order endpoints.
type OrderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(req.CustomerID, req.ItemName, req.Amount)
	if err != nil {
		// Check if this is a timeout/service unavailable error
		if isServiceUnavailableError(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service is unavailable"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	order, err := h.service.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	response := OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// CancelOrder handles PATCH /orders/{id}/cancel.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	err := h.service.CancelOrder(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, _ := h.service.GetOrder(orderID)
	response := OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// SetupRoutes configures the order routes.
func SetupRoutes(router *gin.Engine, handler *OrderHandler) {
	router.POST("/orders", handler.CreateOrder)
	router.GET("/orders/:id", handler.GetOrder)
	router.PATCH("/orders/:id/cancel", handler.CancelOrder)
}

// isServiceUnavailableError checks if the error is related to service unavailability.
func isServiceUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return errMsg == "payment authorization failed: failed to call payment service: context deadline exceeded" ||
		errMsg == "payment authorization failed: failed to call payment service: EOF" ||
		errMsg == "payment authorization failed: failed to call payment service: connection refused" ||
		errMsg == "payment authorization failed: failed to call payment service: dial tcp: lookup: no such host"
}
