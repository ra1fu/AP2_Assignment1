-- Create payments table for Payment Service
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for order_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
