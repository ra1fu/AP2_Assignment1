# Order & Payment Microservices Platform

## Overview

This project implements a two-service microservices platform in Go following Clean Architecture principles. For Assignment 2, we migrated inter-service communication from REST to **gRPC** using a **Contract-First approach**:

1. **Contract Repository (Repo A)**: Contains Protocol Buffers (`.proto`) and GitHub Actions for automated code generation.
2. **Generated Repository (Repo B)**: Stores the generated Go code (the `*.pb.go` and `*_grpc.pb.go` files).
3. **Order Service** (port 8080 for REST, 50051 for gRPC Streaming) - Manages customer orders and streams updates via gRPC. 
4. **Payment Service** (port 50052 for gRPC Unary) - Processes and authorizes payments using gRPC. Includes a logging interceptor.

## Architecture Decisions

### Clean Architecture Implementation

Both services follow a strict layered architecture:

```
service/
├── cmd/service-name/          # Entry point (Composition Root)
├── internal/
│   ├── domain/                # Domain models and interfaces (Ports)
│   ├── usecase/               # Business logic layer
│   ├── repository/            # Data access and external clients (gRPC Client)
│   ├── transport/grpc/        # gRPC Servers (Delivery layer)
│   ├── transport/http/        # HTTP handlers (Delivery layer for public API)
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
   - **Transport Layer**: gRPC and HTTP request/response handling

3. **Microservices Architecture**:
   - **Database per Service**: Each service has its own PostgreSQL database
   - **No Shared Code**: Services share contracts via the external generated Repo B, no shared domain models
   - **Bounded Contexts**: Clear ownership boundaries
   - **Synchronous Communication**: gRPC-based inter-service communication 

4. **Resilience**:
   - gRPC context timeout: 2 seconds (max) for Payment Service calls
   - Proper error handling and `google.golang.org/grpc/status` mapping

5. **Financial Accuracy**:
   - Amount stored as `int64` (cents) in domain, passed as `double` in pb
   - Example: 1000 = $10.00; 100000 = $1,000.00

## Microservices Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (API Consumer)                   │
└────────────────────┬────────────────────────────────────────────┘
                     │ REST (Create Order)
                     │ gRPC (SubscribeToOrderUpdates)
                     ▼
        ┌────────────────────────────┐
        │   Order Service            │
        ├────────────────────────────┤
        │ - HTTP Handler (:8080)     │
        │ - gRPC Streaming (:50051)  │
        │ - Order Use Case           │
        │ - Order Repository (DB)    │
        │ - Payment Client (gRPC)    │
        └────────────┬───────────────┘
                     │
                     │ gRPC Call (ProcessPayment)
                     │ (with 2sec timeout)
                     ▼
        ┌────────────────────────────┐
        │   Payment Service          │
        ├────────────────────────────┤
        │ - gRPC Server (:50052)     │
        │ - Logging Interceptor      │
        │ - Payment Use Case         │
        │ - Payment Repository (DB)  │
        └────────────────────────────┘
```

## Submitting Evidence (For Reviewer)

To run the project:
1. Ensure your `repo-b` (generated repository) is properly linked in the `go.mod` files of `order-service` and `payment-service`. (Use `go get github.com/youruser/repo-b` once the remote repository is populated by the `ap2-contracts` GitHub Action, or substitute it with a local workspace replacement).
2. Run `docker-compose up --build`
3. Try hitting the REST endpoint `POST localhost:8080/orders` to see standard REST functionality.
4. Try connecting with a gRPC Client (e.g. `grpcurl` or Postman) to `localhost:50051` via `orderv1.OrderService.SubscribeToOrderUpdates` to watch the real-time server-side streaming endpoint pulling status updates!

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

## Defense Notes

- **Bounded Contexts**: Order domain and Payment domain are completely separate
- **Service Decomposition**: Each service manages its own lifecycle and data
- **Financial Rules**: Amount validation and payment limits enforced strictly
- **Resilience**: Proper timeout handling prevents hanging requests
- **Status Codes**: HTTP status codes correctly map to business outcomes
- **Idempotency**: Duplicate requests handled safely with Idempotency-Key

---

**Deadline**: 01.04.2026 23:59  
**Author**: Advanced Programming 2 - Assignment 1
