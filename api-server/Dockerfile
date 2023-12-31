# api-server/Dockerfile
FROM golang:1.21

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the service's specific files
COPY api-server/ .

# Build the application
RUN go build -o ./app ./app.go

# Run the binary
CMD ["./app"]
