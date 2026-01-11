# Payment Processing System

Multi-currency payment processing system handling payments between internal users and to external accounts in USD, EUR, and GBP.

## Architecture

**Layered Architecture:**

- HTTP API (Echo) → Service Layer → SQLC Queries
- No repository layer - services use SQLC directly
- Interfaces defined in same file as implementations

**Key Design Decisions:**

- **Two-Step Payment Flow**: Initiate → Confirm with PIN for security
- **PIN Validation**: 4-digit numeric PIN stored as bcrypt hash in users table
- **Transaction Expiration**: Initiated transactions expire after 10 minutes
- **Double-Entry Ledger**: Immutable ledger entries are the source of truth for balances
- **Wallet Locking**: `SELECT FOR UPDATE` for concurrent transaction safety
- **Idempotency**: Required `Idempotency-Key` header for mutation endpoints
- **Provider Abstraction**: Clean interface with multiple provider support (CurrencyCloud, dLocal)
- **Async Processing**: Redis queue for external payouts and webhook processing
- **No Foreign Keys**: App-level referential integrity for better control and performance

## Setup

```bash
# Install dependencies
go mod download

# Generate SQLC code (required first)
make sqlc-generate

# Start services
make dev-up
# or
docker compose up --build

# Run locally
make run
```

## API Endpoints

All endpoints require `X-User-ID` header and `Idempotency-Key` header (for mutations).

### Payments

- `POST /api/payments/internal` - Internal transfer between users

  - Same user (different currency): Processes immediately
  - Different user: Creates initiated transaction, requires PIN confirmation
  - Supports multi-currency transfers with exchange rate locking at initiate
  - Uses account number and bank code to identify recipient (like name enquiry)

  ```json
  {
  	"from_currency": "USD",
  	"to_account_number": "1234567890",
  	"to_bank_code": "044",
  	"amount": {
  		"amount": 10000,
  		"currency": "EUR"
  	}
  }
  ```

- `POST /api/payments/external` - External transfer

  - Always creates initiated transaction, requires PIN confirmation
  - Supports multi-currency transfers with exchange rate locking at initiate

  ```json
  {
  	"bank_account_id": "text_id",
  	"from_currency": "USD",
  	"amount": {
  		"amount": 5000,
  		"currency": "EUR"
  	}
  }
  ```

- `POST /api/payments/:id/confirm` - Confirm transaction with PIN

  - Validates PIN and processes transaction
  - Transaction must be in `initiated` status and not expired (10 minutes)

  ```json
  {
  	"pin": "1234"
  }
  ```

- `GET /api/payments/:id` - Get transaction status

- `GET /api/transactions?cursor=&limit=20` - Get transaction history (cursor-based pagination)

  Query parameters:

  - `cursor` (optional): Base64-encoded cursor from previous response for pagination
  - `limit` (optional): Number of transactions per page (default: 20, max: 100)

  Response:

  ```json
  {
  	"transactions": [
  		{
  			"id": "tx-id",
  			"type": "internal",
  			"amount": 10000,
  			"currency": "USD",
  			"status": "completed",
  			"created_at": "2024-01-01T00:00:00Z"
  		}
  	],
  	"next_cursor": "base64-encoded-cursor"
  }
  ```

  Returns transactions in descending order (newest first). Use `next_cursor` from response to fetch next page.

### Webhooks

- `POST /api/webhooks/:provider` - Receive provider callbacks

### Name Enquiry

- `POST /api/name-enquiry` - Enquire account name and verify if internal/external

  ```json
  {
  	"account_number": "0123456789",
  	"bank_code": "044"
  }
  ```

  Response:

  ```json
  {
  	"account_name": "John Doe",
  	"is_internal": true,
  	"currency": "USD"
  }
  ```

### Wallets

- `GET /api/wallets` - Get all wallets for the authenticated user

  Returns wallets with bank account details and cached balance (optimized, not ledger sum).

  Response:

  ```json
  [
  	{
  		"id": "wallet-id",
  		"currency": "USD",
  		"balance": 50000,
  		"account_number": "1234567890",
  		"bank_name": "Test Bank",
  		"account_name": "John Doe",
  		"provider": "CurrencyCloud",
  		"created_at": "2024-01-01T00:00:00Z",
  		"updated_at": "2024-01-01T00:00:00Z"
  	}
  ]
  ```

### Exchange Rates

- `GET /api/exchange-rate?from=USD&to=EUR` - Get exchange rate between currencies

  Response:

  ```json
  {
  	"from_currency": "USD",
  	"to_currency": "EUR",
  	"rate": 0.85
  }
  ```

### Utilities

- `GET /health` - Health check

- `GET /docs` - API documentation (ReDoc)

## Data Model

### Core Tables

- **users**: Account holders (internal/external)
- **wallets**: One per user per currency (USD, EUR, GBP)
- **transactions**: Transaction intent and state
- **ledger_entries**: Immutable source of truth for all balance changes (double-entry)
- **idempotency_keys**: Ensures at-least-once processing safety
- **bank_accounts**: External payout destinations
- **webhook_events**: Provider callback events

### Key Principles

- Balances stored in wallets table (denormalized for performance)
- Balance calculated from ledger entries (source of truth)
- Every transaction creates debit/credit ledger entries
- Ledger entries are never updated or deleted
- Transactions use idempotency keys to prevent duplicates

## Money Handling

- All amounts stored as integers in minor units (cents/pence)
- No floating point arithmetic
- Strong Money type with Currency validation
- Supports USD, EUR, GBP

## Providers

- **CurrencyCloud**: Supports USD, EUR
- **dLocal**: Supports USD, EUR, GBP

Provider selection based on currency support. External payouts processed asynchronously via Redis queue.

## Development

```bash
make run              # Run locally
make test             # Run tests
make lint             # Lint code
make sqlc-generate    # Generate SQLC code
make dev-up           # Start docker compose
make dev-down         # Stop docker compose
```

## Environment Variables

| Variable            | Default                  | Description          |
| ------------------- | ------------------------ | -------------------- |
| `PORT`              | `8080`                   | Server port          |
| `DATABASE_HOST`     | `localhost`              | PostgreSQL host      |
| `DATABASE_PORT`     | `5432`                   | PostgreSQL port      |
| `DATABASE_NAME`     | `payment_service`        | Database name        |
| `DATABASE_USERNAME` | `postgres`               | Database user        |
| `DATABASE_PASSWORD` | `password`               | Database password    |
| `REDIS_URL`         | `redis://localhost:6379` | Redis connection URL |

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-verbose
```
