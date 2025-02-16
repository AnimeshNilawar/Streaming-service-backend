# Use an official Golang image
FROM golang:1.22

# Install dependencies
RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy the application source code
COPY . .

# Build the Go application
RUN go mod tidy
RUN go build -o app

# Expose the port your app runs on
EXPOSE 8080

# Run the application
CMD ["./app"]
