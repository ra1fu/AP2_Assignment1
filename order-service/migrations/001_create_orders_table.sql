-- Create orders table for Order Service
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for customer_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);

-- Create idempotency_keys table for idempotency feature (bonus)
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for order_id in idempotency_keys
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_order_id ON idempotency_keys(order_id);
