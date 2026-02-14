# --- Stage 1: Builder ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Optimization: Cache Layering
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Binary
RUN CGO_ENABLED=0 GOOS=linux go build -o quickcart main.go

# --- Stage 2: Production ---

# Distroless Static
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy binary hasil build dari stage builder 
COPY --from=builder /app/quickcart /quickcart

# Copy .env sample
COPY --from=builder /app/.env.example .env

# Default Go 8080
EXPOSE 8080

# Security Best Practice
USER nonroot:nonroot

ENTRYPOINT ["/quickcart"]