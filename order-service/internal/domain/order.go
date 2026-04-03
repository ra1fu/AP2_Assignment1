package domain

import "time"

// Order represents an order entity in the Order Service domain.
type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64  // Amount in cents (e.g., 1000 = $10.00)
	Status     string // "Pending", "Paid", "Failed", "Cancelled"
	CreatedAt  time.Time
}

// OrderRepository is the interface for order persistence.
type OrderRepository interface {
	Save(order *Order) error
	GetByID(id string) (*Order, error)
	GetRecentOrders(limit int) ([]*Order, error)
	Update(order *Order) error
}

// PaymentClient is the interface for calling the Payment Service.
type PaymentClient interface {
	AuthorizePayment(orderID string, amount int64) (transactionID string, status string, err error)
	GetPaymentStatus(orderID string) (status string, err error)
}

// OrderUseCase defines the business logic for orders.
type OrderUseCase interface {
	CreateOrder(customerID, itemName string, amount int64) (*Order, error)
	GetOrder(id string) (*Order, error)
	GetRecentOrders(limit int) ([]*Order, error)
	CancelOrder(id string) error
}
