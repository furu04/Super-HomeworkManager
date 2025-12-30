# Builder stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git if needed for fetching dependencies (sometimes needed even with go modules)
# RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary since we are using pure Go SQLite driver (glebarez/sqlite)
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy web assets (templates, static files)
COPY --from=builder /app/web ./web

# Expose port (adjust if your app uses a different port)
EXPOSE 8080

# Run the application
ENTRYPOINT ["./server"]
