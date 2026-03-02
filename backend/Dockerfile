FROM golang:1.23.5-alpine3.21 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final stage - use specific alpine version
FROM alpine:3.21

RUN apk --no-cache add ca-certificates && \
    apk upgrade --no-cache

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy migrations directory
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations

# Create uploads directory
RUN mkdir -p uploads

EXPOSE 8080

CMD ["./main"]