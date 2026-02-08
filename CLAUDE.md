# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go run .           # Run the API server
go build ./...     # Build all packages
go test ./...      # Run all tests
```

## Environment Variables

Required environment variables (see `t.sh` for local setup):
- `DB_CONNECTION_STRING` - MongoDB connection string
- `JWT_SECRET` - JWT token signing secret
- `SALT` - Password hashing salt
- `SMTP_PASSWORD` - Email sending password
- `AUTH_TOKEN` - WhatsApp API auth token
- `PHONE_NUMBER_ID` - WhatsApp phone number ID
- `GOOGLE_API_KEY`, `GEOLOC_URL`, `WEATHER_URL` - Google Maps/Weather APIs
- `HUB_TOKEN` - WhatsApp webhook verification token

## Architecture

This is a Go API backend for cattle ranch management (ZapManejo). It receives WhatsApp messages via webhooks, parses them into structured data, and stores in MongoDB. Also provides REST API for the dashboard.

### Directory Structure

- `main.go` - HTTP server setup and routing (Gorilla Mux)
- `handlers.go` - REST API handlers for CRUD operations
- `auth.go` - Authentication handlers (register, login, verify, password reset)
- `internal/` - Business logic organized by domain:
  - `internal/chat/` - WhatsApp webhook processing
  - `internal/birth/`, `internal/death/`, etc. - Message parsers for each data type
  - `internal/db/` - MongoDB connection and queries
  - `internal/user/`, `internal/account/`, `internal/team/` - User/account management
  - `internal/breed/` - Breed parsing with nickname matching
- `db/schema/` - Seed data JSON files

### API Routes

**Public:**
- `POST /api/auth/register`, `/api/auth/login`, `/api/auth/verify-email`, `/api/auth/forgot-password`
- `GET/POST /chat/message` - WhatsApp webhook

**Protected (JWT required):**
- `GET/POST/PUT/PATCH/DELETE /api/data/{collection}` - CRUD for any collection
- `POST /api/upload/{collection}` - CSV upload
- `POST /api/upload/{collection}/json` - JSON bulk upload
- `GET /api/download/{collection}` - CSV export

### Key Patterns

**Message Parser Interface**: Each data type (birth, death, rain, etc.) implements:
```go
type Parser interface {
    GetCollection() string
    Parse(string) bool        // Detect if message matches
    Text(string) string       // Generate response
    Insert(*BaseMessageValues) error
}
```

**Multi-tenancy**: All data is scoped by `account` field. Queries always filter by account.

**Global vs Account Data**: System-wide records (default breeds, areas) use account `"000000000000000000000000"`. The API returns both global and account-specific records using `$in` filter.

**Authentication Middleware**: JWT token extracted and user_id injected into request context:
```go
userID := r.Context().Value("user_id").(string)
user, _ := user.Read(userID)
```

**Database Queries**: Use `db.ReadUnordered()` with optional filters:
```go
filters := map[string]string{"breed": "nelore"}
data, err := db.ReadUnordered("births", user.Account, filters)
```

### Database

MongoDB database: `possohelp`

Key collections:
- `users` - User accounts with hashed passwords
- `teams` - WhatsApp phone → account mappings
- `breeds`, `areas` - Reference data (supports global + per-account)
- `birth`, `death`, `rain`, `temperature` - Event data from WhatsApp

### WhatsApp Message Flow

1. Webhook receives message at `/chat/message`
2. `chat.ProcessEntries()` routes to appropriate parser
3. Parser extracts data (tag, breed, sex, area, etc.) from message text
4. Data inserted into MongoDB with account from team lookup
5. Response sent back via WhatsApp API
