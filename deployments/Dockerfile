# Stage 1: Build the binary
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy and download dependencies (leveraging Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary (CGO_ENABLED=0 for static linking)
RUN CGO_ENABLED=0 GOOS=linux go build -o sr_plus_app ./cmd/api

# Stage 2: Final minimal image
FROM alpine:latest

# Add CA certificates (needed for HTTPS calls to Turtlogs)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/sr_plus_app .

# Expose the API port
EXPOSE 8080

CMD ["./sr_plus_app"]
