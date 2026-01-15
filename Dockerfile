# ---- Build stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/api ./cmd/api

# ---- Run stage ----
FROM alpine:3.20

WORKDIR /app

# Needed for HTTPS calls to Snowflake
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api /app/api

ENV PORT=8080
EXPOSE 8080

CMD ["/app/api"]
