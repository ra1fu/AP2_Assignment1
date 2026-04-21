package repository

import (
	"database/sql"
	"errors"
	"time"

	"payment-service/internal/domain"
)

// PostgresPaymentRepository implements domain.PaymentRepository.
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository creates a new PostgreSQL payment repository.
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Save stores a payment in the database.
func (r *PostgresPaymentRepository) Save(payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := r.db.Exec(
		query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
		payment.CreatedAt,
	)

	return err
}

// GetByOrderID retrieves a payment by order ID.
func (r *PostgresPaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments
		WHERE order_id = $1
		LIMIT 1
	`

	payment := &Payment{}
	err := r.db.QueryRow(query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
		&payment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("payment not found")
	}
	if err != nil {
		return nil, err
	}

	return payment.ToDomainPayment(), nil
}

// GetByID retrieves a payment by ID.
func (r *PostgresPaymentRepository) GetByID(id string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments
		WHERE id = $1
	`

	payment := &Payment{}
	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
		&payment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("payment not found")
	}
	if err != nil {
		return nil, err
	}

	return payment.ToDomainPayment(), nil
}

// FindByAmountRange retrieves payments within a specific amount range.
func (r *PostgresPaymentRepository) FindByAmountRange(min, max int64) ([]*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments
		WHERE ($1 = 0 OR amount >= $1) AND ($2 = 0 OR amount <= $2)
	`
	rows, err := r.db.Query(query, min, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p.ToDomainPayment())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

// Payment is a repository model (internal to repository, not part of domain).
type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}

// ToDomainPayment converts repository model to domain model.
func (p *Payment) ToDomainPayment() *domain.Payment {
	return &domain.Payment{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
	}
}

// FromDomainPayment converts domain model to repository model.
func FromDomainPayment(p *domain.Payment) *Payment {
	return &Payment{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
	}
}
