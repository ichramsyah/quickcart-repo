# QuickCart Sample Application

A simple Go web application.

## Features

- Single binary that serves both frontend (HTML) and backend (JSON API)
- PostgreSQL database connection
- Health check endpoint
- Auto-creates tables and sample data on startup

## Requirements

- Go 1.21+
- PostgreSQL (optional - app runs without it)

## Quick Start

### Without Database

```bash
go run main.go
```

The app will start on http://localhost:8080 but products won't be displayed.

### With Database

1. Start PostgreSQL:
```bash
# Using Docker
docker run -d \
  --name quickcart-db \
  -e POSTGRES_USER=quickcart \
  -e POSTGRES_PASSWORD=quickcart123 \
  -e POSTGRES_DB=quickcart \
  -p 5432:5432 \
  postgres:15-alpine
```

2. Create `.env` file (optional - uses defaults if not provided):
```bash
cp .env.example .env
# Edit .env if needed
```

3. Run the application:
```bash
go run main.go
```

4. Open http://localhost:8080

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | quickcart | Database user |
| DB_PASSWORD | quickcart123 | Database password |
| DB_NAME | quickcart | Database name |

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Web UI - displays products |
| `/health` | GET | Health check (JSON) |
| `/api/products` | GET | List all products (JSON) |
| `/api/error` | GET | Returns 500 error (for testing monitoring) |
| `/api/slow` | GET | Slow response (for testing latency alerts) |

## Testing Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```
```json
{
  "status": "healthy",
  "database": "connected",
  "time": "2024-01-15T10:30:00Z"
}
```

### Error Endpoint (for monitoring testing)
```bash
curl http://localhost:8080/api/error
```
Returns HTTP 500 with error message. Use this to test error rate alerts.

### Slow Endpoint (for latency testing)
```bash
# Default 3 second delay
curl http://localhost:8080/api/slow

# Custom delay (in milliseconds, max 30000)
curl http://localhost:8080/api/slow?delay=5000
```
Use this to test latency/response time alerts.
