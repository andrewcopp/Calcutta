FROM golang:1.24

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code (needed for internal packages)
COPY . .

# Build the application
RUN go build -o /app/bin/api ./cmd/api \
    && go build -o /app/bin/workers ./cmd/workers \
    && go build -o /app/bin/migrate ./cmd/migrate

EXPOSE 8080

# Run the built binary
CMD ["/app/bin/api"]