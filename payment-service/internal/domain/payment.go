package domain

import "time"

// Payment represents a payment entity in the Payment Service domain.
type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64  // Amount in cents (e.g., 1000 = $10.00)
	Status        string // "Authorized", "Declined"
	CreatedAt     time.Time
}

// PaymentRepository is the interface for payment persistence.
type PaymentRepository interface {
	Save(payment *Payment) error
	GetByOrderID(orderID string) (*Payment, error)
	GetByID(id string) (*Payment, error)
	FindByAmountRange(min, max int64) ([]*Payment, error)
}

// PaymentService defines the business logic for payments.
type PaymentService interface {
	AuthorizePayment(orderID string, amount int64) (*Payment, error)
	GetPaymentStatus(orderID string) (*Payment, error)
}
