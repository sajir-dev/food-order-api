# --- Build stage ---
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Install git for go mod and build tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/food-api ./cmd/api

# --- Run stage ---
FROM alpine:3.19

WORKDIR /app

# Install CA certificates and bash
RUN apk add --no-cache ca-certificates bash

# Copy the built binary and config
COPY --from=builder /app/food-api /app/
COPY config.yaml /app/
# Add wait-for-it script
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/food-api /app/wait-for-it.sh

# Expose the API port
EXPOSE 8080

# Run the binary
CMD ["/app/food-api"] 