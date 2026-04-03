-- Создаем базу для payment_service
CREATE DATABASE payment_db;

-- Переключаемся на order_db и создаем таблицы
\c order_db;

CREATE TABLE orders (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);

CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Переключаемся на payment_db и создаем таблицы
\c payment_db;

CREATE TABLE payments (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_payments_order_id ON payments(order_id);
