package usecase

import (
	"fmt"
	"time"

	"order-service/internal/domain"
)

// OrderUseCase implements domain.OrderUseCase.
type OrderUseCase struct {
	orderRepo     domain.OrderRepository
	paymentClient domain.PaymentClient
}

// NewOrderUseCase creates a new order use case.
func NewOrderUseCase(orderRepo domain.OrderRepository, paymentClient domain.PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		paymentClient: paymentClient,
	}
}

// CreateOrder creates a new order and attempts to authorize payment.
// Flow:
// 1. Create an Order with status "Pending" in the DB.
// 2. Call Payment Service POST /payments.
// 3. Update Order status to "Paid" or "Failed" in the DB based on the response.
func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64) (*domain.Order, error) {
	// Validate business rules
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	orderID := generateOrderID()

	// Create order with Pending status
	order := &domain.Order{
		ID:         orderID,
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
		CreatedAt:  time.Now(),
	}

	err := uc.orderRepo.Save(order)
	if err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Call Payment Service to authorize payment
	_, paymentStatus, err := uc.paymentClient.AuthorizePayment(orderID, amount)
	if err != nil {
		// Payment service is unavailable or timeout
		// Mark order as Failed and return error
		order.Status = "Failed"
		_ = uc.orderRepo.Update(order)
		return order, fmt.Errorf("payment authorization failed: %w", err)
	}

	// Update order status based on payment response
	if paymentStatus == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	err = uc.orderRepo.Update(order)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return order, nil
}

// GetOrder retrieves an order by ID.
func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

// GetRecentOrders retrieves the most recent orders up to the limit.
func (uc *OrderUseCase) GetRecentOrders(limit int) ([]*domain.Order, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	orders, err := uc.orderRepo.GetRecentOrders(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}

	return orders, nil
}

// CancelOrder cancels an order.
// Business rule: Only "Pending" orders can be cancelled. Once a payment is successful ("Paid"),
// cancellation is prohibited.
func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Enforce business rule: only Pending orders can be cancelled
	if order.Status != "Pending" {
		return fmt.Errorf("cannot cancel order with status '%s': only 'Pending' orders can be cancelled", order.Status)
	}

	order.Status = "Cancelled"
	err = uc.orderRepo.Update(order)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// generateOrderID generates a unique order ID.
func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
