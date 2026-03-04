FROM golang:1.25.1-alpine3.21 AS builder

RUN apk --no-cache add ca-certificates && \
    apk upgrade --no-cache

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Prepare runtime dirs for distroless stage
RUN mkdir -p /app/uploads

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/server

# Final stage - distroless runtime
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main /app/main

# Copy migrations directory
COPY --from=builder /app/pkg/db/migrations /app/pkg/db/migrations

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/uploads /app/uploads

EXPOSE 8080

CMD ["/app/main"]
