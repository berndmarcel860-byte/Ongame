# Ongame — Online Gambling Platform

A full-stack online gambling platform built with Go, PostgreSQL, Redis, and Docker Compose.

## Architecture

```
┌──────────────┐     X-Api-Token     ┌──────────────┐
│   Valkyrie   │ ──────────────────► │  PAM Service │
│  (port 8090) │                     │  (port 8080) │
└──────┬───────┘                     └──────┬───────┘
       │                                    │
       │ game RPC                    ┌──────┴───────┐
       ▼                             │  PostgreSQL  │
┌──────────────┐                     │  + Redis     │
│  Slotopol    │                     └──────────────┘
│  (port 8100) │
└──────┬───────┘
       │
  ┌────┴────┐
  │  MySQL  │
  └─────────┘
```

| Service    | Description                                          | Port  |
|------------|------------------------------------------------------|-------|
| PAM        | Player Account Management — custom Go service        | 8080  |
| Valkyrie   | Open-source game aggregator (lobby + session proxy)  | 8090  |
| Slotopol   | Slot games server                                    | 8100  |
| PostgreSQL | Primary datastore for PAM                            | 5432  |
| Redis      | Session cache / idempotency                          | 6379  |
| MySQL      | Slotopol datastore                                   | 3306  |

## Quick Start

### Prerequisites
- Docker ≥ 24
- Docker Compose v2

### 1. Clone & configure
```bash
git clone https://github.com/berndmarcel860-byte/ongame.git
cd ongame
cp .env.example .env   # edit secrets
```

### 2. Start everything
```bash
docker compose up --build -d
```

### 3. Check health
```bash
curl http://localhost:8080/health
```

## Environment Variables

| Variable            | Default                     | Description                        |
|---------------------|-----------------------------|------------------------------------|
| `DATABASE_URL`      | `postgres://…@postgres/ongame` | PostgreSQL DSN                  |
| `REDIS_URL`         | `redis://redis:6379`        | Redis URL                          |
| `JWT_SECRET`        | *(must set)*                | HS256 signing secret               |
| `PAM_API_TOKEN`     | *(must set)*                | Static token Valkyrie sends to PAM |
| `POSTGRES_PASSWORD` | `postgres`                  | PostgreSQL root password           |
| `MYSQL_PASSWORD`    | `slotopol`                  | MySQL password for Slotopol        |
| `SLOTOPOL_API_KEY`  | *(optional)*                | Slotopol API key                   |
| `PORT`              | `8080`                      | PAM HTTP listen port               |

## PAM API Reference

### Authentication
All Valkyrie→PAM requests require `X-Api-Token: <PAM_API_TOKEN>` header.  
Player requests use `Authorization: Bearer <jwt>`.

### Player Balance
```
GET /players/{userId}/balance
→ { cash, bonus, locked, currency }
```

### Player Transaction
```
POST /players/{userId}/transactions
Body: { type, amount, currency, providerTransactionId, gameId, roundId }
→ { transactionId, cash, bonus, locked, currency }
```
- `type`: `WITHDRAW` or `DEPOSIT`
- Idempotent: repeated calls with same `providerTransactionId` return same result.

### Game Sessions
```
POST   /players/{userId}/sessions        → { sessionToken, sessionId }
DELETE /players/{userId}/sessions/{id}   → 204
```

### User Auth
```
POST /auth/register  { email, username, password, currency }
POST /auth/login     { email, password } → { token, user }
```

### User Profile
```
GET /users/me
GET /users/me/transactions?limit=20&offset=0
```

### Admin
```
GET  /admin/users
GET  /admin/users/{userId}
PUT  /admin/users/{userId}/balance   { amount, reason }
POST /admin/users/{userId}/ban       { reason }
```

## Development

```bash
# Run PAM locally (needs Postgres + Redis running)
cd pam
go run ./cmd/server

# Or use dev compose (mounts source for live reload)
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

## Database Schema

Migrations are applied automatically on startup from `pam/migrations/`.

| Table               | Purpose                              |
|---------------------|--------------------------------------|
| `users`             | Player accounts                      |
| `balances`          | One row per player, optimistic lock  |
| `transactions`      | Immutable audit ledger               |
| `game_sessions`     | Active / historical game sessions    |
| `compliance_records`| RG limits, bans, self-exclusions     |

## Security Notes

- Passwords hashed with bcrypt (cost 10)
- JWT signed with HS256, 24 h TTL
- Balances updated with optimistic locking (`version` column) to prevent race conditions
- Financial amounts stored as `DECIMAL(20,8)` — no floating-point rounding
- Negative balances blocked at DB constraint level
