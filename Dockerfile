FROM golang:1.25.6-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o csv-processor ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/csv-processor .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./csv-processor"]
