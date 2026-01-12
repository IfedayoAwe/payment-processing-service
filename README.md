# Payment Processing Service

Multi-currency payment processing system handling internal transfers between users and external transfers to bank accounts. Supports USD, EUR, and GBP currencies with exchange rate locking and double-entry ledger accounting.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Application                      │
└────────────────────────────┬────────────────────────────────┘
                              │
                              │ HTTP/REST
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                         API Server                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Handlers   │──│  Middleware  │──│    Routes   │      │
│  └──────┬───────┘  └──────────────┘  └──────────────┘      │
└─────────┼───────────────────────────────────────────────────┘
          │
          │ Service Layer
          │
┌─────────▼───────────────────────────────────────────────────┐
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Payment    │  │   Wallet     │  │   Ledger    │      │
│  │   Service    │  │   Service    │  │   Service   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐     │
│  │   External   │  │   Name       │  │   Webhook    │     │
│  │   Transfer   │  │   Enquiry    │  │   Service    │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
┌─────────▼──────────┐ ┌─────▼──────┐ ┌───────▼──────────┐
│   PostgreSQL       │ │  RabbitMQ  │ │   Providers     │
│   (SQLC)           │ │   (Queue)  │ │ (CurrencyCloud  │
│                    │ │            │ │    dLocal)      │
└────────────────────┘ └────────────┘ └─────────────────┘
          │
          │
┌─────────▼───────────────────────────────────────────────────┐
│              Background Workers                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Outbox     │  │   Payout     │  │   Webhook    │      │
│  │   Worker     │  │   Worker     │  │   Worker     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Quick Start

```bash
# Clone the repository
git clone git@github.com:IfedayoAwe/payment-processing-service.git
cd payment-processing-service

# Start all services (PostgreSQL, RabbitMQ, API)
docker-compose up --build

# The service will automatically:
# - Run database migrations
# - Seed test data (2 users with wallets and bank accounts)
# - Start the API server on http://localhost:8080
# - Start background workers for async processing
```

The API will be available at `http://localhost:8080`.

### Test Users

Two test users are automatically seeded:

- **user_1** (John Doe) - PIN: `12345`
- **user_2** (Jane Doe) - PIN: `12345`

Both users have wallets in USD, EUR, and GBP with initial balances. Use the test endpoint to retrieve full details:

```bash
GET /api/test/users
```

This endpoint returns user IDs, PINs, wallet details, bank account information, and balances for easy testing.

## Request Flow

### Internal Transfer Request Flow

```
Client
  │
  ├─► POST /api/payments/internal
  │   Headers: X-User-ID, Idempotency-Key
  │   Body: {from_currency, to_account_number, to_bank_code, amount}
  │
  ├─► Handler: PaymentHandler.CreateInternalTransfer
  │   │
  │   ├─► Validate Request
  │   ├─► Check Idempotency Key
  │   └─► Service: PaymentService.CreateInternalTransfer
  │       │
  │       ├─► Get Exchange Rate (if different currency)
  │       ├─► Find Recipient Wallet
  │       ├─► Create Transaction (status: initiated)
  │       │
  │       └─► Process Based on Recipient
  │           │
  │           ├─► Same User, Different Currency
  │           │   └─► Process Immediately
  │           │       ├─► Lock Wallets
  │           │       ├─► Create Ledger Entries
  │           │       ├─► Update Balances
  │           │       └─► Return (status: completed)
  │           │
  │           └─► Different User
  │               └─► Return Transaction ID (requires PIN)
  │
  └─► POST /api/payments/:id/confirm (if different user)
      Headers: X-User-ID
      Body: {pin}
      │
      ├─► Handler: PaymentHandler.ConfirmTransaction
      └─► Service: PaymentService.ConfirmInternalTransfer
          │
          ├─► Validate PIN
          ├─► Check Expiration (10 min)
          ├─► Lock Wallets
          ├─► Create Ledger Entries
          ├─► Update Balances
          └─► Return (status: completed)
```

### External Transfer Request Flow

```
Client
  │
  ├─► POST /api/payments/external
  │   Headers: X-User-ID, Idempotency-Key
  │   Body: {from_currency, to_account_number, to_bank_code, amount}
  │
  ├─► Handler: PaymentHandler.CreateExternalTransfer
  │   │
  │   ├─► Validate Request
  │   ├─► Check Idempotency Key
  │   └─► Service: ExternalTransferService.CreateExternalTransfer
  │       │
  │       ├─► Get Exchange Rate
  │       ├─► Create Transaction (status: initiated)
  │       └─► Store Recipient Details
  │
  └─► POST /api/payments/:id/confirm
      Headers: X-User-ID
      Body: {pin}
      │
      ├─► Handler: PaymentHandler.ConfirmTransaction
      └─► Service: ExternalTransferService.ConfirmExternalTransfer
          │
          ├─► Validate PIN
          ├─► Check Expiration (10 min)
          ├─► Begin DB Transaction
          │   │
          │   ├─► Lock Wallet
          │   ├─► Check Funds
          │   ├─► Create Debit Entry
          │   ├─► Create Credit Entry (external system)
          │   ├─► Update Balance
          │   ├─► Update Transaction (status: pending)
          │   └─► Create Outbox Entry (atomic)
          │
          └─► Commit Transaction
              │
              └─► Background Processing
                  │
                  ├─► Outbox Worker (every 2s)
                  │   ├─► Get Unprocessed Entries
                  │   ├─► Enqueue to RabbitMQ
                  │   └─► Mark Processed
                  │
                  └─► Payout Worker
                      ├─► Dequeue Job
                      ├─► Call Provider API
                      └─► Update Transaction (status: completed)
```

## API Usage

### Authentication

All API endpoints (except `/api/test/users`) require the `X-User-ID` header:

```
X-User-ID: user_1
```

### Idempotency

All mutation endpoints require the `Idempotency-Key` header to prevent duplicate processing:

```
Idempotency-Key: unique-key-per-request
```

### Endpoints

#### 1. Get Test Users

```
GET /api/test/users
```

Returns test user data including user IDs, PINs, wallets, bank accounts, and balances. No authentication required.

**Response:**

```json
{
	"data": {
		"users": [
			{
				"user_id": "user_1",
				"name": "John Doe",
				"pin": "12345",
				"wallets": [
					{
						"id": "wallet_user1_usd",
						"currency": "USD",
						"balance": 100.5,
						"account_number": "1000000001",
						"bank_name": "Test Bank",
						"bank_code": "044",
						"account_name": "John Doe",
						"provider": "currencycloud"
					}
				]
			}
		]
	}
}
```

#### 2. Name Enquiry

```
POST /api/name-enquiry
Headers: X-User-ID
```

Check if an account number and bank code belong to an internal user or external account.

**Request:**

```json
{
	"account_number": "1000000001",
	"bank_code": "044"
}
```

**Response:**

```json
{
	"data": {
		"account_name": "John Doe",
		"is_internal": true,
		"currency": "USD"
	}
}
```

#### 3. Get Exchange Rate

```
GET /api/exchange-rate?from=USD&to=EUR
Headers: X-User-ID
```

Get current exchange rate between two currencies.

**Response:**

```json
{
	"data": {
		"from_currency": "USD",
		"to_currency": "EUR",
		"rate": 0.85
	}
}
```

#### 4. Get User Wallets

```
GET /api/wallets
Headers: X-User-ID
```

Get all wallets for the authenticated user with bank account details and cached balances.

**Response:**

```json
{
	"data": [
		{
			"id": "wallet_user1_usd",
			"currency": "USD",
			"balance": 100.5,
			"account_number": "1000000001",
			"bank_name": "Test Bank",
			"bank_code": "044",
			"account_name": "John Doe",
			"provider": "currencycloud",
			"created_at": "2026-01-11T00:00:00Z",
			"updated_at": "2026-01-11T00:00:00Z"
		}
	]
}
```

#### 5. Create Internal Transfer

```
POST /api/payments/internal
Headers: X-User-ID, Idempotency-Key
```

Initiate an internal transfer between users. Uses account number and bank code to identify recipient.

**Flow:**

- Same user, different currency: Processes immediately (no PIN required)
- Different user: Creates initiated transaction, requires PIN confirmation

**Request:**

```json
{
	"from_currency": "USD",
	"to_account_number": "2000000001",
	"to_bank_code": "044",
	"amount": {
		"amount": 100.5,
		"currency": "EUR"
	}
}
```

**Response (Immediate):**

```json
{
	"data": {
		"id": "tx-id",
		"status": "completed",
		"amount": 100.5,
		"currency": "EUR"
	},
	"message": "transfer completed successfully"
}
```

**Response (Requires Confirmation):**

```json
{
	"data": {
		"id": "tx-id",
		"status": "initiated",
		"amount": 100.5,
		"currency": "EUR",
		"exchange_rate": 0.85
	},
	"message": "transfer initiated, please confirm with PIN"
}
```

#### 6. Create External Transfer

```
POST /api/payments/external
Headers: X-User-ID, Idempotency-Key
```

Initiate an external transfer to a bank account outside the system. Always requires PIN confirmation.

**Request:**

```json
{
	"from_currency": "USD",
	"to_account_number": "9999999999",
	"to_bank_code": "044",
	"amount": {
		"amount": 50.0,
		"currency": "GBP"
	}
}
```

**Response:**

```json
{
	"data": {
		"id": "tx-id",
		"status": "initiated",
		"amount": 50.0,
		"currency": "GBP",
		"exchange_rate": 0.75
	},
	"message": "external transfer initiated, please confirm with PIN"
}
```

#### 7. Confirm Transaction

```
POST /api/payments/:id/confirm
Headers: X-User-ID
```

Confirm an initiated transaction with PIN. Transaction must be in `initiated` status and not expired (10 minutes from creation).

**Request:**

```json
{
	"pin": "12345"
}
```

**Response (Internal - Immediate):**

```json
{
	"data": {
		"id": "tx-id",
		"status": "completed",
		"amount": 100.5,
		"currency": "EUR"
	},
	"message": "transaction confirmed and completed successfully"
}
```

**Response (External - Queued):**

```json
{
	"data": {
		"id": "tx-id",
		"status": "pending",
		"amount": 50.0,
		"currency": "GBP",
		"provider_reference": "{\"account_number\":\"9999999999\",\"bank_code\":\"044\"}"
	},
	"message": "transaction confirmed and queued for processing"
}
```

Note: External transfers return `pending` status after confirmation and are processed asynchronously by a worker. The transaction status will change to `completed` once the payout worker successfully processes the transfer. Check transaction status later to see final provider details.

#### 8. Get Transaction

```
GET /api/payments/:id
Headers: X-User-ID
```

Get transaction details by ID.

**Response:**

```json
{
	"data": {
		"id": "tx-id",
		"idempotency_key": "key-123",
		"from_wallet_id": "wallet_user1_usd",
		"to_wallet_id": "wallet_user2_eur",
		"type": "internal",
		"amount": 100.5,
		"currency": "EUR",
		"status": "completed",
		"exchange_rate": 0.85,
		"created_at": "2026-01-11T00:00:00Z",
		"updated_at": "2026-01-11T00:00:00Z"
	}
}
```

#### 9. Get Transaction History

```
GET /api/transactions?cursor=&limit=20
Headers: X-User-ID
```

Get paginated transaction history using cursor-based pagination.

**Query Parameters:**

- `cursor` (optional): Base64-encoded cursor from previous response
- `limit` (optional): Number of transactions per page (default: 20, max: 100)

**Response:**

```json
{
	"data": {
		"transactions": [
			{
				"id": "tx-id",
				"type": "internal",
				"amount": 100.5,
				"currency": "EUR",
				"status": "completed",
				"created_at": "2026-01-11T00:00:00Z"
			}
		],
		"next_cursor": "base64-encoded-cursor"
	}
}
```

#### 10. Receive Webhook

```
POST /api/webhooks/:provider?reference=provider-ref
Headers: X-User-ID
```

Receive webhook events from payment providers. Events are queued for asynchronous processing.

**Path Parameters:**

- `provider`: Provider name (e.g., `currencycloud`, `dlocal`)

**Query Parameters:**

- `reference`: Provider reference ID (optional, can be in body)

**Request:**

```json
{
	"event_type": "payout.completed",
	"reference": "TXN-12345",
	"transaction_id": "tx-id",
	"payload": {}
}
```

**Response:**

```json
{
	"data": {
		"status": "received"
	},
	"message": "webhook received successfully"
}
```

## Money Handling

All monetary amounts in API requests and responses use major units (dollars, euros, pounds) as floating-point numbers. Internally, amounts are stored as integers in the smallest currency unit (cents/pence) to avoid floating-point precision issues.

**Example:**

- API Request: `{"amount": 100.50, "currency": "USD"}`
- Internal Storage: `10050` (cents)

## Transaction States

- **initiated**: Transaction created, awaiting PIN confirmation
- **pending**: Transaction confirmed, being processed (external transfers)
- **completed**: Transaction successfully processed
- **failed**: Transaction failed (insufficient funds, provider error, etc.)

## Transaction Expiration

Initiated transactions expire after 10 minutes. Expired transactions cannot be confirmed and must be re-initiated.

## Error Responses

All errors follow a consistent format:

```json
{
	"error": "error_code",
	"message": "Human-readable error message"
}
```

Common error codes:

- `400`: Bad Request (validation errors, invalid parameters)
- `401`: Unauthorized (missing or invalid X-User-ID)
- `404`: Not Found (transaction, wallet, or account not found)
- `409`: Conflict (duplicate idempotency key)
- `500`: Internal Server Error

## API Documentation

Interactive API documentation is available at:

- ReDoc UI: `http://localhost:8080/docs`
- OpenAPI JSON: `http://localhost:8080/docs/openapi.json`

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Generate SQLC code
make sqlc-generate

# Run tests
make test

# Run linter
make lint

# Run locally (requires PostgreSQL and RabbitMQ)
make run
```

### Environment Variables

| Variable            | Default                              | Description             |
| ------------------- | ------------------------------------ | ----------------------- |
| `PORT`              | `8080`                               | Server port             |
| `DATABASE_HOST`     | `localhost`                          | PostgreSQL host         |
| `DATABASE_PORT`     | `5432`                               | PostgreSQL port         |
| `DATABASE_NAME`     | `payment_service`                    | Database name           |
| `DATABASE_USERNAME` | `postgres`                           | Database user           |
| `DATABASE_PASSWORD` | `password`                           | Database password       |
| `RABBITMQ_URL`      | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |

### Database Migrations

Migrations run automatically on startup. To manually run:

```bash
# Using golang-migrate
migrate -path migrations -database "postgres://postgres:password@localhost:5432/payment_service?sslmode=disable" up
```

### Seed Data

Seed data is applied automatically on startup. The seed script is idempotent and can be run multiple times safely.

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed design decisions, trade-offs, and improvements.
