# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

# Install dependencies (use apk instead of apt-get)
RUN apk add --no-cache ffmpeg

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build the Go application
RUN go build -o main .

# ---- Stage 2: Run (Smaller Image) ----
FROM alpine:latest

# Install only runtime dependencies (minimal)
RUN apk add --no-cache ffmpeg

# Set working directory
WORKDIR /app

# Copy compiled binary from builder stage
COPY --from=builder /app/main .

# Ensure binary has execution permissions
RUN chmod +x main

# Expose the correct port for Cloud Run
EXPOSE 8080

# Set environment variable (Cloud Run requires this)
ENV PORT=8080

# Run the application
CMD ["/app/main"]
