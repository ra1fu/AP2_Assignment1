# API Examples

## curl Examples

### Order Service

#### 1. Create an Order (Amount < Payment Limit)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-001",
    "item_name": "Laptop",
    "amount": 150000
  }'

# Expected Response (201 Created):
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-001",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Paid",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 2. Create an Order (Amount > Payment Limit - 100000)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-002",
    "item_name": "Server",
    "amount": 150000
  }'

# Expected Response (Payment Declined):
{
  "id": "ORD-1711234567891",
  "customer_id": "CUST-002",
  "item_name": "Server",
  "amount": 150000,
  "status": "Failed",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 3. Get Order Details
```bash
curl -X GET http://localhost:8080/orders/ORD-1711234567890

# Expected Response (200 OK):
{
  "id": "ORD-1711234567890",
  "customer_id": "CUST-001",
  "item_name": "Laptop",
  "amount": 150000,
  "status": "Paid",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 4. Cancel Order (Pending Status)
```bash
# First, create a pending order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-003",
    "item_name": "Mouse",
    "amount": 50000
  }'

# Then cancel it
curl -X PATCH http://localhost:8080/orders/ORD-1711234567892/cancel

# Expected Response (200 OK):
{
  "id": "ORD-1711234567892",
  "customer_id": "CUST-003",
  "item_name": "Mouse",
  "amount": 50000,
  "status": "Cancelled",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 5. Try to Cancel Paid Order (Should Fail)
```bash
curl -X PATCH http://localhost:8080/orders/ORD-1711234567890/cancel

# Expected Response (400 Bad Request):
{
  "error": "cannot cancel order with status 'Paid': only 'Pending' orders can be cancelled"
}
```

#### 6. Create Order with Idempotency Key (Bonus Feature)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-12345" \
  -d '{
    "customer_id": "CUST-004",
    "item_name": "Keyboard",
    "amount": 30000
  }'

# Send same request again - should return same order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-12345" \
  -d '{
    "customer_id": "CUST-004",
    "item_name": "Keyboard",
    "amount": 30000
  }'

# Both responses return the same order ID
```

### Payment Service

#### 1. Authorize Payment (Amount Within Limit)
```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-1711234567890",
    "amount": 50000
  }'

# Expected Response (200 OK):
{
  "id": "PAY-1711234567890",
  "order_id": "ORD-1711234567890",
  "transaction_id": "TXN-1711234567890",
  "amount": 50000,
  "status": "Authorized",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 2. Authorize Payment (Amount Exceeds Limit)
```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-1711234567891",
    "amount": 150000
  }'

# Expected Response (200 OK - but Declined):
{
  "id": "PAY-1711234567891",
  "order_id": "ORD-1711234567891",
  "transaction_id": "TXN-1711234567891",
  "amount": 150000,
  "status": "Declined",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 3. Get Payment Status
```bash
curl -X GET http://localhost:8081/payments/ORD-1711234567890

# Expected Response (200 OK):
{
  "id": "PAY-1711234567890",
  "order_id": "ORD-1711234567890",
  "transaction_id": "TXN-1711234567890",
  "amount": 50000,
  "status": "Authorized",
  "created_at": "2024-03-31T12:30:00Z"
}
```

### Failure Scenarios

#### 1. Service Unavailability (2-second Timeout)
```bash
# Stop the Payment Service
# Then try to create an order

curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-005",
    "item_name": "Monitor",
    "amount": 80000
  }'

# Expected Response (503 Service Unavailable):
{
  "error": "payment service is unavailable"
}

# The order will be created with status "Failed"
curl http://localhost:8080/orders/ORD-[created-id]

# Order Response:
{
  "id": "ORD-[created-id]",
  "customer_id": "CUST-005",
  "item_name": "Monitor",
  "amount": 80000,
  "status": "Failed",
  "created_at": "2024-03-31T12:30:00Z"
}
```

#### 2. Invalid Input (Missing Required Field)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-006"
  }'

# Expected Response (400 Bad Request):
{
  "error": "json: key \"item_name\" is required; key \"amount\" is required"
}
```

#### 3. Invalid Amount (Negative or Zero)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "CUST-007",
    "item_name": "Product",
    "amount": -1000
  }'

# Expected Response (400 Bad Request):
{
  "error": "amount must be greater than 0"
}
```

#### 4. Order Not Found
```bash
curl -X GET http://localhost:8080/orders/ORD-NONEXISTENT

# Expected Response (404 Not Found):
{
  "error": "order not found"
}
```

## Postman Collection

You can import the following into Postman as a collection:

```json
{
  "info": {
    "name": "Order & Payment Microservices",
    "description": "API collection for Order and Payment services",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Order Service",
      "item": [
        {
          "name": "Create Order",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"customer_id\": \"CUST-001\", \"item_name\": \"Laptop\", \"amount\": 150000}"
            },
            "url": {
              "raw": "http://localhost:8080/orders",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["orders"]
            }
          }
        },
        {
          "name": "Get Order",
          "request": {
            "method": "GET",
            "url": {
              "raw": "http://localhost:8080/orders/{{order_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["orders", "{{order_id}}"]
            }
          }
        },
        {
          "name": "Cancel Order",
          "request": {
            "method": "PATCH",
            "url": {
              "raw": "http://localhost:8080/orders/{{order_id}}/cancel",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["orders", "{{order_id}}", "cancel"]
            }
          }
        }
      ]
    },
    {
      "name": "Payment Service",
      "item": [
        {
          "name": "Authorize Payment",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"order_id\": \"ORD-1711234567890\", \"amount\": 150000}"
            },
            "url": {
              "raw": "http://localhost:8081/payments",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["payments"]
            }
          }
        },
        {
          "name": "Get Payment Status",
          "request": {
            "method": "GET",
            "url": {
              "raw": "http://localhost:8081/payments/{{order_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["payments", "{{order_id}}"]
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "order_id",
      "value": "ORD-1711234567890"
    }
  ]
}
```

## Testing Checklist

- [ ] Create order with valid amount → Status: Paid
- [ ] Create order with amount > 100000 → Status: Failed
- [ ] Get order by ID → Returns correct details
- [ ] Cancel pending order → Status: Cancelled
- [ ] Try to cancel paid order → Error message
- [ ] Payment service timeout → 503 Service Unavailable
- [ ] Invalid input validation → 400 Bad Request
- [ ] Order not found → 404 Not Found
- [ ] Idempotent requests → Same response for same key
