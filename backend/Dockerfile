FROM golang:1.22

WORKDIR /app

# Copy go mod files
COPY go.mod ./
# COPY go.sum ./ # Uncomment this when you have a go.sum file

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main cmd/server/main.go

# Run the binary
CMD ["/app/main"] 