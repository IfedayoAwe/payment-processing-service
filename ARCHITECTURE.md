# Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Handlers   │  │  Middleware  │  │    Routes    │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                    Service Layer                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Payment    │  │   Wallet     │  │   Ledger    │          │
│  │   Service    │  │   Service    │  │   Service   │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                  │                  │                  │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐         │
│  │   External   │  │   Name       │  │   Webhook    │         │
│  │   Transfer   │  │   Enquiry    │  │   Service    │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                    Data & Infrastructure Layer                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  PostgreSQL  │  │    Redis     │  │  Providers   │          │
│  │   (SQLC)     │  │   (Queue)    │  │ (CurrencyCloud│          │
│  │              │  │              │  │    dLocal)   │          │
│  └──────────────┘  └──────┬───────┘  └──────────────┘          │
└────────────────────────────┼────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │  Background       │
                    │  Workers          │
                    │  - Outbox         │
                    │  - Payout         │
                    │  - Webhook        │
                    └───────────────────┘
```

## Payment Flow Diagrams

### Internal Transfer Flow

```
User Request
    │
    ├─► POST /api/payments/internal
    │       │
    │       ├─► Validate Request
    │       ├─► Check Idempotency
    │       └─► PaymentService.CreateInternalTransfer
    │               │
    │               ├─► Get Exchange Rate (if different currency)
    │               ├─► Create Transaction (status: initiated)
    │               │
    │               └─► Check Recipient
    │                       │
    │                       ├─► Same User, Different Currency
    │                       │       └─► Process Immediately
    │                       │               ├─► Lock Wallets (SELECT FOR UPDATE)
    │                       │               ├─► Create Ledger Entries (debit/credit)
    │                       │               ├─► Update Wallet Balances
    │                       │               └─► Update Transaction (status: completed)
    │                       │
    │                       └─► Different User
    │                               └─► Return Transaction ID (requires PIN)
    │
    └─► POST /api/payments/:id/confirm (if different user)
            │
            ├─► Validate PIN
            ├─► Check Transaction Expiration (10 min)
            └─► PaymentService.ConfirmInternalTransfer
                    │
                    ├─► Lock Wallets (SELECT FOR UPDATE)
                    ├─► Create Ledger Entries (debit/credit)
                    ├─► Update Wallet Balances
                    └─► Update Transaction (status: completed)
```

### External Transfer Flow

```
User Request
    │
    ├─► POST /api/payments/external
    │       │
    │       ├─► Validate Request
    │       ├─► Check Idempotency
    │       └─► ExternalTransferService.CreateExternalTransfer
    │               │
    │               ├─► Get Exchange Rate
    │               ├─► Create Transaction (status: initiated)
    │               └─► Store Recipient Details
    │
    └─► POST /api/payments/:id/confirm
            │
            ├─► Validate PIN
            ├─► Check Transaction Expiration (10 min)
            └─► ExternalTransferService.ConfirmExternalTransfer
                    │
                    ├─► Begin DB Transaction
                    │       │
                    │       ├─► Lock Wallet (SELECT FOR UPDATE)
                    │       ├─► Check Sufficient Funds
                    │       ├─► Create Debit Entry (user wallet)
                    │       ├─► Create Credit Entry (external system)
                    │       ├─► Update Wallet Balance
                    │       ├─► Update Transaction (status: pending)
                    │       └─► Create Outbox Entry (atomic)
                    │
                    └─► Commit Transaction
                            │
                            └─► Outbox Worker (async)
                                    │
                                    ├─► Get Unprocessed Entries (FOR UPDATE SKIP LOCKED)
                                    ├─► Enqueue to Redis Queue
                                    └─► Mark Outbox Entry Processed
                                            │
                                            └─► Payout Worker (async)
                                                    │
                                                    ├─► Get Job from Queue
                                                    ├─► Call External Provider
                                                    ├─► Update Transaction (status: completed)
                                                    └─► Store Provider Reference
```

## Outbox Pattern Flow

```
┌─────────────────────────────────────────────────────────────┐
│  External Transfer Confirmation                              │
│                                                              │
│  1. Begin Transaction                                        │
│  2. Update Transaction Status                                │
│  3. Create Ledger Entries                                    │
│  4. Create Outbox Entry (atomic)                            │
│  5. Commit Transaction                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Outbox Table                                                │
│  ┌─────────────┬──────────┬──────────┬──────────┐          │
│  │ id         │ job_type │ payload  │ processed│          │
│  ├─────────────┼──────────┼──────────┼──────────┤          │
│  │ outbox_1   │ payout   │ {...}    │ false    │          │
│  └─────────────┴──────────┴──────────┴──────────┘          │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Outbox Worker (every 2 seconds)                             │
│                                                              │
│  1. Begin Transaction                                        │
│  2. SELECT * FROM outbox                                     │
│     WHERE processed = false                                  │
│     FOR UPDATE SKIP LOCKED                                   │
│     LIMIT 10                                                 │
│  3. For each entry:                                          │
│     - Enqueue to Redis Queue                                │
│     - Mark as processed                                      │
│  4. Commit Transaction                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Redis Queue                                                 │
│  ┌────────────────────────────────────────────┐            │
│  │ queue:payout                                │            │
│  │ [{"transaction_id": "...", ...}]           │            │
│  └────────────────────────────────────────────┘            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Payout Worker                                               │
│                                                              │
│  1. Dequeue Job                                              │
│  2. Process Payout                                           │
│  3. Update Transaction                                       │
└─────────────────────────────────────────────────────────────┘
```

## Database Schema

```
┌──────────────┐
│    users     │
├──────────────┤
│ id (TEXT)    │
│ pin_hash     │
│ created_at   │
└──────┬───────┘
       │
       │ 1:N
       │
┌──────▼───────┐      ┌──────────────┐
│ bank_accounts│      │   wallets    │
├──────────────┤      ├──────────────┤
│ id (TEXT)    │◄─────┤ id (TEXT)    │
│ user_id      │      │ user_id      │
│ bank_code    │      │ currency     │
│ account_num  │      │ balance      │
│ provider     │      │ bank_acc_id  │
└──────────────┘      └──────┬───────┘
                             │
                             │ 1:N
                             │
                    ┌────────▼────────┐
                    │ ledger_entries  │
                    ├─────────────────┤
                    │ id (TEXT)       │
                    │ wallet_id       │
                    │ transaction_id  │
                    │ amount          │
                    │ currency        │
                    │ account_type    │
                    │ balance_before  │
                    │ balance_after   │
                    └────────┬────────┘
                             │
                             │ N:1
                             │
                    ┌────────▼────────┐
                    │  transactions   │
                    ├─────────────────┤
                    │ id (TEXT)       │
                    │ idempotency_key │
                    │ trace_id        │
                    │ from_wallet_id  │
                    │ to_wallet_id    │
                    │ type            │
                    │ amount          │
                    │ currency        │
                    │ status          │
                    │ exchange_rate   │
                    │ provider_ref    │
                    └─────────────────┘
                             │
                             │ 1:1
                             │
                    ┌────────▼────────┐
                    │ idempotency_    │
                    │     keys        │
                    ├─────────────────┤
                    │ key (TEXT)      │
                    │ transaction_id │
                    └─────────────────┘

┌──────────────┐
│   outbox     │
├──────────────┤
│ id (TEXT)    │
│ job_type     │
│ payload      │
│ processed    │
│ retry_count  │
└──────────────┘

┌──────────────┐
│ webhook_     │
│   events     │
├──────────────┤
│ id (TEXT)    │
│ provider_name│
│ event_type   │
│ provider_ref │
│ transaction_id│
│ payload      │
└──────────────┘
```

## Design Decisions

### 1. Ledger-First Design

Balances are derived from immutable ledger entries, ensuring auditability and correctness under failure. The `wallets.balance` column is denormalized for performance but the ledger remains the source of truth.

**Why:** Financial systems require complete audit trails. Immutable ledger entries provide a complete history of all balance changes, enabling reconciliation and debugging.

### 2. Idempotency as a First-Class Concept

Duplicate requests are common in payment systems; enforcing idempotency at the API layer prevents double-spend scenarios. All mutation endpoints require an `Idempotency-Key` header.

**Why:** Network retries, client bugs, and user behavior can cause duplicate requests. Idempotency keys ensure that retrying a request with the same key returns the same result without side effects.

### 3. Explicit State Machines

Clear transaction states (`initiated`, `pending`, `completed`, `failed`) make the system predictable, debuggable, and safe for async flows.

**Why:** State machines prevent invalid state transitions and make the system's behavior explicit. This is critical for financial systems where incorrect state transitions can lead to money loss.

### 4. Typed Core Models

Strong typing with dedicated `Money` and `Currency` types prevents category errors, improves readability, and enforces correctness across providers.

**Why:** Type safety catches errors at compile time. The `Money` type ensures amounts are always paired with currencies and prevents mixing different currency amounts.

### 5. Async External Flows

External money movement is unreliable by nature; async processing isolates failures and improves resilience. External transfers are queued and processed by background workers.

**Why:** External payment providers can be slow or fail. Synchronous processing would block API responses and create poor user experience. Async processing allows the API to respond immediately while processing happens in the background.

### 6. Outbox Pattern

Outbox pattern ensures reliable message delivery. Job payloads are written to the database within the same transaction as business logic, then processed asynchronously by workers.

**Why:** Prevents data inconsistency where a transaction commits but the queue enqueue fails. Ensures atomicity between business logic and message publishing.

### 7. Simple but Realistic Provider Abstraction

Implements extensibility without overengineering. Provider interfaces are segregated by action (PayoutProvider, NameEnquiryProvider, ExchangeRateProvider) following Interface Segregation Principle.

**Why:** Multiple payment providers are needed for different currencies and regions. Clean interfaces allow adding new providers without changing core business logic.

### 8. Redis for Async Work

Chosen for simplicity and speed in a take-home, while clearly stating production alternatives (SQS, RabbitMQ, Kafka).

**Why:** Redis is simple to set up and provides the necessary queue semantics. In production, a dedicated message queue would provide better durability and features.

### 9. No Float Math

Prevents rounding drift and financial inconsistencies at scale. All amounts stored as integers in smallest currency unit (cents/pence).

**Why:** Floating-point arithmetic introduces rounding errors that accumulate over time. Integer arithmetic is exact and prevents financial discrepancies.

### 10. DB Locking Over Caching

Correctness > performance. Balance reads use row-level locking (`SELECT FOR UPDATE`) instead of risky caching.

**Why:** Concurrent transactions can cause race conditions. Row-level locking ensures serializable isolation and prevents double-spending. Caching balances would require complex invalidation logic and risk inconsistencies.

### 11. Two-Step Payment Flow

Initiate transaction (returns ID) → Confirm with PIN. Separates transaction creation from execution, allowing rate locking and validation before commitment.

**Why:** Provides security (PIN confirmation) and allows exchange rates to be locked at initiation time, preventing rate changes between initiation and confirmation.

### 12. Double-Entry Ledger

Every transaction creates both debit and credit entries. Internal transfers debit sender and credit receiver. External transfers debit user wallet and credit external system (ledger entries with NULL wallet_id represent balancing with the outside world).

**Why:** Double-entry bookkeeping is the standard for financial systems. It ensures the accounting equation always balances and provides complete auditability. External system entries (NULL wallet_id) represent money leaving/entering the platform, balancing with external forces rather than an internal company wallet.

### 13. Wallet-Bank Account Linkage

Wallets reference bank accounts. This enables multi-provider support (same user can have USD wallet with CurrencyCloud and another USD wallet with dLocal).

**Why:** Real-world users have multiple bank accounts with different providers. This design supports that use case while maintaining clear relationships.

### 14. Cursor-Based Pagination

Transaction history uses cursor-based pagination instead of offset-based to avoid performance degradation with large datasets.

**Why:** Offset-based pagination becomes slow as the offset increases. Cursor-based pagination maintains consistent performance regardless of dataset size.

### 15. Structured Logging

Uses `zerolog` for structured logging with context propagation throughout the application.

**Why:** Structured logs are easier to query and analyze. Context propagation enables tracing requests across service boundaries.

### 16. Distributed Tracing

Trace IDs are generated per request and propagated through logs, database, queues, and provider calls.

**Why:** Enables end-to-end request tracing for observability, debugging, and auditing.

## Trade-offs

### 1. Denormalized Balance Column

**Trade-off:** Balance stored in `wallets` table for performance, but ledger is source of truth.

**Rationale:** Reading balance from ledger requires SUM aggregation which is slow. Cached balance provides fast reads while ledger ensures correctness. Balance is recalculated from ledger after each transaction.

### 2. No Foreign Keys

**Trade-off:** App-level referential integrity instead of database foreign keys.

**Rationale:** Provides more control over deletion behavior and can improve write performance. Requires careful application logic to maintain integrity.

### 3. TEXT IDs Instead of UUIDs

**Trade-off:** Using TEXT for IDs allows flexibility but loses UUID benefits (uniqueness guarantees, no collisions).

**Rationale:** Simplifies ID generation and allows custom ID formats. In production, would use UUIDs or ULIDs for better uniqueness guarantees.

### 4. Synchronous Internal Transfers

**Trade-off:** Internal transfers process immediately (same user) or require confirmation (different user), while external transfers are always async.

**Rationale:** Internal transfers are fast and reliable (same database). External transfers are slow and unreliable (external API calls). Different flows optimize for each case.

### 5. Exchange Rate Stored as String

**Trade-off:** Exchange rate stored as TEXT in database instead of DECIMAL.

**Rationale:** Simplifies code and avoids precision issues in JSON serialization. In production, would use DECIMAL(20,8) for better precision and queryability.

### 6. Single Redis Queue

**Trade-off:** Using Redis for both payout and webhook queues instead of separate queues.

**Rationale:** Simpler setup for take-home. In production, would use separate queues with different priorities and retry policies.

### 7. PIN Stored as Bcrypt Hash

**Trade-off:** Using bcrypt with default cost (10) instead of higher cost for better security.

**Rationale:** Default cost provides good security while maintaining acceptable performance. In production, would use cost 12+ and consider rate limiting.

## Improvements with More Time

### Database

- Use UUIDs or ULIDs for better uniqueness guarantees
- Add database foreign keys for referential integrity
- Use DECIMAL(20,8) for exchange rates
- Add database indexes for common query patterns
- Implement connection pooling optimization

### Testing

- Add integration tests with testcontainers
- Add end-to-end tests for complete flows
- Increase unit test coverage for edge cases
- Add load testing for concurrent transactions
- Add chaos engineering tests for failure scenarios

### Observability

- Add distributed tracing (OpenTelemetry)
- Add metrics (Prometheus) for transaction rates, latencies, errors
- Add alerting for critical failures
- Add structured logging with correlation IDs

### Security

- Implement rate limiting for API endpoints
- Add request signing/authentication beyond X-User-ID
- Implement PIN attempt limiting and account locking
- Add audit logging for all financial operations
- Encrypt sensitive data at rest

### Performance

- Implement read replicas for balance queries
- Add caching layer for exchange rates (with TTL)
- Optimize database queries with proper indexes
- Implement connection pooling
- Add horizontal scaling support

### Reliability

- Implement retry logic with exponential backoff for external calls
- Add circuit breakers for external provider calls
- Implement dead letter queues for failed jobs
- Add transaction reconciliation and recovery
- Implement idempotency key cleanup (TTL)

### Features

- Add transaction reversal/cancellation
- Implement multi-currency wallet aggregation
- Add transaction fees
- Implement scheduled/recurring payments
- Add payment limits and fraud detection

### Infrastructure

- Replace Redis with dedicated message queue (SQS, RabbitMQ, Kafka)
- Add database replication and failover
- Implement blue-green deployments
- Add canary releases
- Implement feature flags

### Code Quality

- Add more comprehensive error handling
- Implement request/response validation middleware
- Add API versioning
- Implement graceful shutdown
- Add health check endpoints for dependencies
