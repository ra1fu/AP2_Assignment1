# Order & Payment Microservices Platform

## Overview

This project implements a two-service microservices platform in Go following Clean Architecture principles. The platform consists of:

1. **Order Service** (port 8080) - Manages customer orders and their lifecycle
2. **Payment Service** (port 8081) - Processes and authorizes payments

## Architecture Decisions

### Clean Architecture Implementation

Both services follow a strict layered architecture:

```
service/
├── cmd/service-name/          # Entry point (Composition Root)
├── internal/
│   ├── domain/                # Domain models and interfaces (Ports)
│   ├── usecase/               # Business logic layer
│   ├── repository/            # Data access and external clients
│   ├── transport/http/        # HTTP handlers (Delivery layer)
│   └── app/                   # Application setup and DI configuration
├── migrations/                # SQL migration scripts
├── go.mod                     # Module definition
└── README.md
```

### Key Design Principles

1. **Dependency Inversion**: All dependencies flow inward. Use cases depend on interfaces (Ports), not concrete implementations.

2. **Separation of Concerns**:
   - **Domain Layer**: Pure business rules, no framework dependencies
   - **Use Case Layer**: Business logic and state transitions
   - **Repository Layer**: Data persistence and external service calls
   - **Transport Layer**: HTTP request/response handling

3. **Microservices Architecture**:
   - **Database per Service**: Each service has its own PostgreSQL database
   - **No Shared Code**: Each service has independent domain models
   - **Bounded Contexts**: Clear ownership boundaries
   - **Synchronous Communication**: REST-based inter-service communication with HTTP

4. **Resilience**:
   - HTTP client timeout: 2 seconds (max) for Payment Service calls
   - Proper error handling and status code mapping
   - Service unavailability returns 503 Service Unavailable

5. **Financial Accuracy**:
   - Amount stored as `int64` (cents), never `float64`
   - Example: 1000 = $10.00; 100000 = $1,000.00

## Microservices Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (API Consumer)                     │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │   Order Service (8080)      │
        ├────────────────────────────┤
        │ - HTTP Handler             │
        │ - Order Use Case           │
        │ - Order Repository (DB)    │
        │ - Payment Client (HTTP)    │
        └────────────┬───────────────┘
                     │
                     │ REST Call
                     │ (with 2sec timeout)
                     ▼
        ┌────────────────────────────┐
        │   Payment Service (8081)    │
        ├────────────────────────────┤
        │ - HTTP Handler             │
        │ - Payment Use Case         │
        │ - Payment Repository (DB)  │
        └────────────────────────────┘
```

## Business Rules

### Order Service

1. **Create Order**:
   - Amount must be > 0
   - Order created with "Pending" status
   - Payment authorization is synchronous
   - Status updated to "Paid" or "Failed" based on response

2. **Cancel Order**:
   - Only "Pending" orders can be cancelled
   - "Paid" orders cannot be cancelled

3. **Service Resilience**:
   - If Payment Service times out (>2 seconds), order marked as "Failed"
   - Returns 503 Service Unavailable to client

### Payment Service

1. **Payment Limits**:
   - Amount ≤ 100,000 cents → "Authorized"
   - Amount > 100,000 cents → "Declined"

2. **Idempotency**:
   - Each payment has a unique transaction ID
   - Duplicate requests prevented via unique constraint

## Database Schema

### Order Service Database

```sql
-- orders table
CREATE TABLE orders (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);

-- idempotency_keys table (for bonus feature)
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

### Payment Service Database

```sql
-- payments table
CREATE TABLE payments (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_payments_order_id ON payments(order_id);
```

## API Endpoints

### Order Service

#### 1. Create Order
```http
POST /orders
Content-Type: application/json
Idempotency-Key: <optional-key>

{
  "customer_id": "CUST-123",
  "item_name": "Laptop",
  "amount": 150000
}

Response (201 Created):
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-123",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Paid" | "Failed" | "Pending",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 2. Get Order
```http
GET /orders/{id}

Response (200 OK):
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-123",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Paid",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 3. Cancel Order
```http
PATCH /orders/{id}/cancel

Response (200 OK):
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-123",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Cancelled",
  "created_at": "2024-03-31T12:30:00Z"
}
```

### Payment Service

#### 1. Authorize Payment
```http
POST /payments
Content-Type: application/json

{
  "order_id": "ORD-1711234567890",
  "amount": 150000
}

Response (200 OK):
{
  "id": "PAY-1711234567890",
  "order_id": "ORD-1711234567890",
  "transaction_id": "TXN-1711234567890",
  "amount": 150000,
  "status": "Authorized" | "Declined",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 2. Get Payment Status
```http
GET /payments/{order_id}

Response (200 OK):
{
  "id": "PAY-1711234567890",
  "order_id": "ORD-1711234567890",
  "transaction_id": "TXN-1711234567890",
  "amount": 150000,
  "status": "Authorized",
  "created_at": "2024-03-31T12:30:00Z"
}
```

## Setup Instructions

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Environment variables configuration

### Environment Variables

#### Order Service (.env)
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=order_db
PORT=:8080
PAYMENT_SERVICE_URL=http://localhost:8081
```

#### Payment Service (.env)
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_db
PORT=:8081
```

### Database Setup

```bash
# Order Service Database
psql -U postgres -d order_db -f order-service/migrations/001_create_orders_table.sql

# Payment Service Database
psql -U postgres -d payment_db -f payment-service/migrations/001_create_payments_table.sql
```

### Running the Services

```bash
# Terminal 1: Order Service
cd order-service/cmd/order-service
go run main.go

# Terminal 2: Payment Service
cd payment-service/cmd/payment-service
go run main.go
```

## Failure Handling & Resilience

### Scenario: Payment Service Unavailable

1. **Order Service** calls Payment Service with 2-second timeout
2. If timeout occurs:
   - Connection refused → Order marked as "Failed"
   - Response: 503 Service Unavailable
3. Order remains in database for retry/investigation
4. Client receives clear error response

### Error Codes

- **201 Created**: Order successfully created
- **200 OK**: Successful GET/PATCH
- **400 Bad Request**: Invalid input or business rule violation
- **404 Not Found**: Resource not found
- **503 Service Unavailable**: Payment Service unreachable/timeout

## Bonus Features

### Idempotency

The Order Service supports idempotency via `Idempotency-Key` header:

```http
POST /orders
Idempotency-Key: abc123-unique-request-id

{
  "customer_id": "CUST-123",
  "item_name": "Laptop",
  "amount": 150000
}
```

- Same key ensures same request doesn't create duplicate orders
- Implementation uses `idempotency_keys` table
- If key exists, returns the previously created order

## Testing

### Using curl

```bash
# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"CUST-001","item_name":"Laptop","amount":150000}'

# Get an order
curl http://localhost:8080/orders/ORD-1711234567890

# Cancel an order
curl -X PATCH http://localhost:8080/orders/ORD-1711234567890/cancel

# Get payment status
curl http://localhost:8081/payments/ORD-1711234567890
```

### Using Postman

See `api-examples.md` for Postman collection links and examples.

## Project Structure Summary

```
Assignment1/
├── order-service/
│   ├── cmd/order-service/main.go
│   ├── internal/
│   │   ├── domain/order.go
│   │   ├── usecase/order_usecase.go
│   │   ├── repository/postgres.go
│   │   ├── transport/http/handler.go
│   │   └── app/app.go
│   ├── migrations/001_create_orders_table.sql
│   ├── go.mod
│   └── README.md
├── payment-service/
│   ├── cmd/payment-service/main.go
│   ├── internal/
│   │   ├── domain/payment.go
│   │   ├── usecase/payment_usecase.go
│   │   ├── repository/postgres.go
│   │   ├── transport/http/handler.go
│   │   └── app/app.go
│   ├── migrations/001_create_payments_table.sql
│   ├── go.mod
│   └── README.md
└── README.md (this file)
```

## Key Implementation Details

1. **No Shared Code**: Each service has its own domain models and repositories
2. **Manual DI**: Composition Root in each service's `main.go`
3. **HTTP Timeout**: Order Service → Payment Service has 2-second timeout
4. **Financial Accuracy**: All amounts stored as `int64` (cents)
5. **Clean Layers**: Domain models have zero framework dependencies
6. **Interface-Based**: Dependencies use Go interfaces (Ports pattern)

