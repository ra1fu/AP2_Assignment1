# Architecture Diagram

## System Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           External Client                                 │
│                      (Postman / curl / Browser)                           │
└─────────────────────────────────┬──────────────────────────────────────┘
                                  │
                 ┌────────────────┴────────────────┐
                 │                                 │
                 ▼                                 ▼
    ┌────────────────────────────┐    ┌────────────────────────────┐
    │   ORDER SERVICE (8080)     │    │  PAYMENT SERVICE (8081)    │
    ├────────────────────────────┤    ├────────────────────────────┤
    │                            │    │                            │
    │  ┌──────────────────────┐  │    │  ┌──────────────────────┐  │
    │  │ HTTP Handler Layer   │  │    │  │ HTTP Handler Layer   │  │
    │  │ (Delivery/Transport) │  │    │  │ (Delivery/Transport) │  │
    │  │                      │  │    │  │                      │  │
    │  │ POST /orders         │  │    │  │ POST /payments       │  │
    │  │ GET /orders/{id}     │  │    │  │ GET /payments/{id}   │  │
    │  │ PATCH /orders/{id}.. │  │    │  └──────────────────────┘  │
    │  └──────────┬───────────┘  │    │         ▲                   │
    │             │              │    │         │                   │
    │             ▼              │    │         │                   │
    │  ┌──────────────────────┐  │    │  ┌──────┴──────────────┐   │
    │  │ Use Case Layer       │  │    │  │ Use Case Layer      │   │
    │  │ (Business Logic)     │  │    │  │ (Business Logic)    │   │
    │  │                      │  │    │  │                     │   │
    │  │ CreateOrder()        │  │    │  │ AuthorizePayment()  │   │
    │  │ GetOrder()           │  │    │  │ GetPaymentStatus()  │   │
    │  │ CancelOrder()        │  │    │  └──────────────────────┘  │
    │  └──────────┬───────────┘  │    │         ▲                   │
    │             │              │    │         │                   │
    │             ▼              │    │         │                   │
    │  ┌──────────────────────┐  │    │  ┌──────┴──────────────┐   │
    │  │ Repository Layer     │  │    │  │ Repository Layer    │   │
    │  │ (Data Access)        │  │    │  │ (Data Access)       │   │
    │  │                      │  │    │  │                     │   │
    │  │ OrderRepository      │  │    │  │ PaymentRepository   │   │
    │  │ (PostgreSQL)         │  │    │  │ (PostgreSQL)        │   │
    │  │                      │  │    │  └──────────────────────┘  │
    │  │ PaymentClient        │  │    │         ▲                   │
    │  │ (HTTP to Payment)    │  │    │         │                   │
    │  └──────────┬───────────┘  │    │         │                   │
    │             │              │    └─────────┤────────────────┘  │
    │             ▼              │              │                    │
    │  ┌──────────────────────┐  │              │                    │
    │  │ Domain Layer         │  │              │                    │
    │  │                      │  │              │                    │
    │  │ Order Entity         │  │              │                    │
    │  │ OrderRepository (I)  │  │              │                    │
    │  │ PaymentClient (I)    │  │              │                    │
    │  │ OrderUseCase (I)     │  │              │                    │
    │  │                      │  │              │                    │
    │  └──────────────────────┘  │              │                    │
    │                            │              │                    │
    └────────────────────────────┘              │                    │
             │                                  │                    │
             ▼                                  ▼                    │
    ┌─────────────────────┐            ┌─────────────────────┐      │
    │                     │            │                     │      │
    │  order_db           │            │  payment_db         │      │
    │  (PostgreSQL)       │            │  (PostgreSQL)       │      │
    │                     │            │                     │      │
    │  ┌───────────────┐  │            │  ┌───────────────┐  │      │
    │  │ orders table  │  │            │  │ payments tbl  │  │      │
    │  └───────────────┘  │            │  └───────────────┘  │      │
    │  ┌───────────────┐  │            └─────────────────────┘      │
    │  │ idempotency.. │  │                                         │
    │  └───────────────┘  │                                         │
    └─────────────────────┘                                         │
                                                                     │
                           REST Call (2sec timeout)                 │
                           ──────────────────────────┐              │
                                                     └──────────────┘
```

## Detailed Sequence Diagram: Create Order

```
Client                Order Service              Payment Service           Database
  │                         │                           │                      │
  │ POST /orders            │                           │                      │
  ├────────────────────────►│                           │                      │
  │                         │                           │                      │
  │                         │ 1. Create Order (Pending) │                      │
  │                         │ ─────────────────────────────────────────────────►│
  │                         │                           │                      │ Insert
  │                         │                           │                      │ (orders)
  │                         │                           │                      │
  │                         │ 2. POST /payments         │                      │
  │                         │ (with 2-second timeout)   │                      │
  │                         ├──────────────────────────►│                      │
  │                         │                           │                      │
  │                         │                           │ 3. Check Limit        │
  │                         │                           │ (amount ≤ 100000)    │
  │                         │                           │                      │
  │                         │                           │ 4. Create Payment    │
  │                         │                           │ ──────────────────────►│
  │                         │                           │                      │ Insert
  │                         │                           │                      │ (payments)
  │                         │                           │                      │
  │                         │ 5. Return Payment Response │                      │
  │                         │◄──────────────────────────┤                      │
  │                         │                           │                      │
  │                         │ 6. Update Order Status    │                      │
  │                         │    ("Paid" or "Failed")   │                      │
  │                         │ ─────────────────────────────────────────────────►│
  │                         │                           │                      │ Update
  │                         │                           │                      │ (orders)
  │                         │                           │                      │
  │ 200 Created / 503       │                           │                      │
  │◄────────────────────────┤                           │                      │
  │                         │                           │                      │
```

## Timeout Scenario

```
Client              Order Service          Payment Service         Network
  │                       │                       │                   │
  │ POST /orders          │                       │                   │
  ├──────────────────────►│                       │                   │
  │                       │                       │                   │
  │                       │ 1. POST /payments (Start timer: 2sec)    │
  │                       ├──────────────────────►│                   │
  │                       │                       │ (Service Down)    │
  │                       │                       │ (No Response)     │
  │                       │                       │                   │
  │                       │ [WAITING...]          │                   │
  │                       │ [1 sec elapsed]       │                   │
  │                       │ [2 sec elapsed]       │                   │
  │                       │ TIMEOUT! ◄────────────┘                   │
  │                       │                       │                   │
  │                       │ 2. Create Order       │                   │
  │                       │    with status        │                   │
  │                       │    "Failed"           │                   │
  │                       │ ──┐                   │                   │
  │                       │   │ (DB Update)       │                   │
  │                       │ ◄─┘                   │                   │
  │                       │                       │                   │
  │ 503 Service           │                       │                   │
  │ Unavailable ◄─────────┤                       │                   │
  │                       │                       │                   │
```

## Layered Architecture Within a Service

### Example: Order Service

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Request                       │
└──────────────────────────┬──────────────────────────┘
                           │
                           ▼
        ┌──────────────────────────────────────┐
        │    DELIVERY/TRANSPORT LAYER          │
        │  (Framework-Specific: Gin)           │
        │  ┌──────────────────────────────────┐│
        │  │ CreateOrderHandler()              ││
        │  │ - Parse JSON                     ││
        │  │ - Validate input format          ││
        │  │ - Call use case                  ││
        │  │ - Return HTTP response           ││
        │  └──────────────────────────────────┘│
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │    USE CASE / BUSINESS LOGIC LAYER   │
        │  (Pure logic, framework-agnostic)    │
        │  ┌──────────────────────────────────┐│
        │  │ CreateOrder()                    ││
        │  │ - Validate business rules        ││
        │  │   (amount > 0)                   ││
        │  │ - Create order                   ││
        │  │ - Call payment client            ││
        │  │ - Update order status            ││
        │  │ - Return domain object           ││
        │  └──────────────────────────────────┘│
        └──────────────┬───────────────────────┘
                       │
       ┌───────────────┴──────────────┐
       │                              │
       ▼                              ▼
   ┌────────────────────┐      ┌──────────────────────┐
   │ REPOSITORY LAYER   │      │ REPOSITORY LAYER     │
   │ (Data Access)      │      │ (External Services)  │
   │ ┌──────────────────┤      │ ┌──────────────────────┤
   │ │ OrderRepository  │      │ │ PaymentClient        │
   │ │ - Save()         │      │ │ - AuthorizePayment() │
   │ │ - GetByID()      │      │ │ - GetPaymentStatus() │
   │ │ - Update()       │      │ └──────────────────────┤
   │ │                  │      │ (HTTP impl, 2sec TO)   │
   │ └──────────────────┤      └──────────────────────┘
   │ (PostgreSQL impl)  │
   └────────┬───────────┘
            │
            ▼
       ┌────────────────┐
       │   PostgreSQL   │
       │   Database     │
       └────────────────┘

        ┌──────────────────────────────────────┐
        │      DOMAIN LAYER                    │
        │  (Pure business models)              │
        │  ┌──────────────────────────────────┐│
        │  │ type Order struct                 ││
        │  │ type OrderRepository interface    ││
        │  │ type PaymentClient interface      ││
        │  │ type OrderUseCase interface       ││
        │  │                                   ││
        │  │ (Zero framework dependencies!)    ││
        │  └──────────────────────────────────┘│
        └──────────────────────────────────────┘
```

## Dependency Flow (Inward - Clean Architecture)

```
                    DOMAIN
                      ▲
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        │             │             │
    USE CASE    USE CASE    USE CASE
        │             │             │
        └─────────────┼─────────────┘
                      │
                      ▲
        ┌─────────────┼─────────────┐
        │             │             │
    REPOSITORY   REPOSITORY   REPOSITORY
    (DB)         (Http)       (Cache)
        │             │             │
        └─────────────┼─────────────┘
                      │
                      ▲
                 DELIVERY/HTTP
```

**Key Rule**: Outer layers can depend on inner layers, but inner layers
can NEVER depend on outer layers. All dependencies point inward.

## Database Schema Relationships

```
ORDER SERVICE (order_db)
┌─────────────────────────────────────────┐
│ orders                                  │
├─────────────────────────────────────────┤
│ id (VARCHAR, PK)                        │
│ customer_id (VARCHAR)                   │◄──┐
│ item_name (VARCHAR)                    │    │
│ amount (BIGINT)                        │    │
│ status (VARCHAR)                        │    │
│ created_at (TIMESTAMP)                  │    │
└─────────────────────────────────────────┘    │
                                               │
┌─────────────────────────────────────────┐    │
│ idempotency_keys                        │    │
├─────────────────────────────────────────┤    │
│ key (VARCHAR, PK)                       │    │
│ order_id (VARCHAR, FK) ─────────────────┼────┘
│ created_at (TIMESTAMP)                  │
└─────────────────────────────────────────┘


PAYMENT SERVICE (payment_db)
┌─────────────────────────────────────────┐
│ payments                                │
├─────────────────────────────────────────┤
│ id (VARCHAR, PK)                        │
│ order_id (VARCHAR)  ──────────────┐     │
│ transaction_id (VARCHAR, UNIQUE)  │     │
│ amount (BIGINT)                   │     │
│ status (VARCHAR)                  │     │
│ created_at (TIMESTAMP)            │     │
└─────────────────────────────────────────┘
          ▲                          │
          │                          │
          └──────────────────────────┘
        (Logical reference, NOT FK)
    (No actual foreign key because
     databases are independent)
```

## Deployment View

```
┌────────────────────────────────────────────────────┐
│              Docker/VM Environment                 │
├────────────────────────────────────────────────────┤
│                                                    │
│ ┌──────────────────┐    ┌──────────────────┐    │
│ │ Order Service    │    │ Payment Service  │    │
│ │  (Go binary)     │    │  (Go binary)     │    │
│ │  Port: 8080      │    │  Port: 8081      │    │
│ └────────┬─────────┘    └────────┬─────────┘    │
│          │                       │               │
│          ▼                       ▼               │
│ ┌──────────────────┐    ┌──────────────────┐   │
│ │  PostgreSQL      │    │  PostgreSQL      │   │
│ │  order_db        │    │  payment_db      │   │
│ │  Port: 5432      │    │  Port: 5432      │   │
│ └──────────────────┘    └──────────────────┘   │
│                                                 │
└────────────────────────────────────────────────────┘
         ▲                         ▲
         │                         │
         │  REST (HTTP)            │
         │  (2sec timeout)         │
         └─────────────────────────┘
             (Inter-service)
```

---

This architecture ensures:
- ✓ Clean separation of concerns
- ✓ Independent microservices
- ✓ Testable design
- ✓ Framework independence in domain logic
- ✓ Resilient communication with timeouts
- ✓ Financial accuracy (int64 for money)
