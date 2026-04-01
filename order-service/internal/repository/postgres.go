package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"order-service/internal/domain"
)

// PostgresOrderRepository implements domain.OrderRepository.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Save stores an order in the database.
func (r *PostgresOrderRepository) Save(order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)

	return err
}

// GetByID retrieves an order by ID.
func (r *PostgresOrderRepository) GetByID(id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders
		WHERE id = $1
	`

	order := &domain.Order{}
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}

// Update updates an existing order in the database.
func (r *PostgresOrderRepository) Update(order *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

// CheckIdempotency checks if a request with the same idempotency key already exists.
func (r *PostgresOrderRepository) CheckIdempotency(idempotencyKey string) (string, error) {
	query := `
		SELECT order_id FROM idempotency_keys WHERE key = $1
	`

	var orderID string
	err := r.db.QueryRow(query, idempotencyKey).Scan(&orderID)

	if err == sql.ErrNoRows {
		return "", nil // No existing request
	}
	if err != nil {
		return "", fmt.Errorf("failed to check idempotency: %w", err)
	}

	return orderID, nil
}

// SaveIdempotencyKey saves an idempotency key for a request.
func (r *PostgresOrderRepository) SaveIdempotencyKey(idempotencyKey, orderID string) error {
	query := `
		INSERT INTO idempotency_keys (key, order_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO NOTHING
	`

	_, err := r.db.Exec(query, idempotencyKey, orderID, time.Now())
	return err
}
