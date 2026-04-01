# Payment Service

The Payment Service processes and authorizes payments. It validates transaction limits and returns payment status to the Order Service.

## Architecture

This service follows Clean Architecture with strict layering:

- **Domain Layer** (`internal/domain/`): Payment entity, repository interface, use case interface
- **Use Case Layer** (`internal/usecase/`): Business logic for payment authorization
- **Repository Layer** (`internal/repository/`): PostgreSQL payment repository
- **Delivery Layer** (`internal/transport/http/`): Gin-based HTTP handlers
- **Application Layer** (`internal/app/`): Dependency injection and app initialization

## API Endpoints

### POST /payments
Authorize a payment.

**Request**:
```json
{
  "order_id": "ORD-1711234567890",
  "amount": 150000
}
```

**Response** (200 OK):
```json
{
  "id": "PAY-1711234567890",
  "order_id": "ORD-1711234567890",
  "transaction_id": "TXN-1711234567890",
  "amount": 150000,
  "status": "Authorized",
  "created_at": "2024-03-31T12:30:00Z"
}
```

**Status Values**:
- `"Authorized"` - Amount ≤ 100,000 cents
- `"Declined"` - Amount > 100,000 cents

### GET /payments/{order_id}
Get payment status for an order.

## Business Rules

1. Payment amount must be > 0
2. Amount limit: 100,000 cents ($1,000)
   - Amount ≤ 100,000 → "Authorized"
   - Amount > 100,000 → "Declined"
3. Each payment gets unique transaction ID
4. Financial accuracy: `int64` only, never `float64`

## Environment Variables

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_db
PORT=:8081
```

## Running the Service

```bash
cd cmd/payment-service
go run main.go
```

Service starts on `http://localhost:8081`

## Database

PostgreSQL database: `payment_db`

Table:
- `payments` - Payment entities with order_id, transaction_id, status

See `migrations/` for schema.
