package usecase

import (
	"fmt"
	"time"

	"payment-service/internal/domain"
)

// PaymentUseCase implements the business logic for payments.
type PaymentUseCase struct {
	repo domain.PaymentRepository
}

// NewPaymentUseCase creates a new payment use case.
func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

// AuthorizePayment authorizes a payment.
// Returns "Authorized" if amount <= 100000 (1000 units), otherwise "Declined".
func (uc *PaymentUseCase) AuthorizePayment(orderID string, amount int64) (*domain.Payment, error) {
	// Business rule: Payment limit is 100000 cents (1000 units)
	const maxAmount = 100000

	status := "Authorized"
	if amount > maxAmount {
		status = "Declined"
	}

	transactionID := generateTransactionID()

	payment := &domain.Payment{
		ID:            generatePaymentID(),
		OrderID:       orderID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
		CreatedAt:     time.Now(),
	}

	err := uc.repo.Save(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return payment, nil
}

// GetPaymentStatus retrieves the status of a payment by order ID.
func (uc *PaymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return payment, nil
}

// generatePaymentID generates a unique payment ID.
func generatePaymentID() string {
	return fmt.Sprintf("PAY-%d", time.Now().UnixNano())
}

// generateTransactionID generates a unique transaction ID.
func generateTransactionID() string {
	return fmt.Sprintf("TXN-%d", time.Now().UnixNano())
}
