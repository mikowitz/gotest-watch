# Use Go 1.25 on Ubuntu
FROM golang:1.25-bookworm

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Run tests
CMD ["go", "test", "-race", "-v", "./..."]
