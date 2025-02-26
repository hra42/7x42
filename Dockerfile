# Build stage for Go
FROM golang:1.24-alpine AS go-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server

# Build stage for frontend assets
FROM node:20-alpine AS js-builder

WORKDIR /app

# Copy package.json and install dependencies
COPY package.json package-lock.json* ./
RUN npm ci

# Copy frontend source files
COPY tailwind.config.js ./

# Create directory structure and copy source files
COPY web ./web

# Build frontend assets
RUN npm run build

# Final stage
FROM alpine:3.21

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/main .

# Copy the entire web directory with built assets
COPY --from=js-builder /app/web ./web

# Create non-root user
RUN adduser -D appuser
RUN chown -R appuser:appuser ./web
USER appuser

EXPOSE 8080

CMD ["./main"]