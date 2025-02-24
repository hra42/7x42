# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server

# Final stage
FROM alpine:3.21

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
COPY web/ web/

# Create non-root user
RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./main"]