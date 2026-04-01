# Order Service

The Order Service manages customer orders and their lifecycle. It communicates with the Payment Service via REST API to authorize payments.

## Architecture

This service follows Clean Architecture with strict layering:

- **Domain Layer** (`internal/domain/`): Order entity, repository interfaces, use case interfaces
- **Use Case Layer** (`internal/usecase/`): Business logic for creating, retrieving, and canceling orders
- **Repository Layer** (`internal/repository/`): PostgreSQL order repository + HTTP client for Payment Service
- **Delivery Layer** (`internal/transport/http/`): Gin-based HTTP handlers
- **Application Layer** (`internal/app/`): Dependency injection and app initialization

## API Endpoints

### POST /orders
Create a new order and authorize payment.

**Request**:
```json
{
  "customer_id": "CUST-001",
  "item_name": "Laptop",
  "amount": 150000
}
```

**Response** (201 Created):
```json
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-001",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Paid",
  "created_at": "2024-03-31T12:30:00Z"
}
```

### GET /orders/{id}
Get order details.

### PATCH /orders/{id}/cancel
Cancel a pending order.

## Business Rules

1. Amount must be > 0
2. Only "Pending" orders can be canceled
3. "Paid" orders cannot be canceled
4. Order status updated based on payment response
5. If Payment Service times out (>2sec), order marked as "Failed"
6. On timeout, returns 503 Service Unavailable

## Environment Variables

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=order_db
PORT=:8080
PAYMENT_SERVICE_URL=http://localhost:8081
```

## Running the Service

```bash
cd cmd/order-service
go run main.go
```

Service starts on `http://localhost:8080`

## Database

PostgreSQL database: `order_db`

Tables:
- `orders` - Order entities
- `idempotency_keys` - Idempotency support (bonus feature)

See `migrations/` for schema.
