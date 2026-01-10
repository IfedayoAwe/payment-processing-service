# Payment Processing System

Multi-currency payment processing system for handling payments between internal users and to external accounts in USD, EUR, and GBP.

## Requirements

- Process payments between internal users (USD, EUR, GBP)
- Process payments to external accounts (USD, EUR, GBP)
- Maintain proper financial records

## Tech Stack

- Golang
- PostgreSQL
- Redis
- Docker Compose
- SQLC

## Setup

```bash
# Install dependencies
go mod download

# Start services
make dev-up
# or
docker compose up --build
```

Services:

- API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Development

```bash
make run          # Run locally
make test         # Run tests
make lint         # Lint code
make sqlc-generate # Generate SQLC code
```

## API Documentation

- ReDoc: http://localhost:8080/docs
- Health: http://localhost:8080/health

## Authentication

All requests require:

```
X-User-ID: user_1
```

## Environment Variables

| Variable            | Default                  |
| ------------------- | ------------------------ |
| `PORT`              | `8080`                   |
| `DATABASE_HOST`     | `localhost`              |
| `DATABASE_PORT`     | `5432`                   |
| `DATABASE_NAME`     | `payment_service`        |
| `DATABASE_USERNAME` | `postgres`               |
| `DATABASE_PASSWORD` | `password`               |
| `REDIS_URL`         | `redis://localhost:6379` |

## Migrations

```bash
make create-migration name=add_payments_table
```
