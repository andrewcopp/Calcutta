FROM golang:1.22

WORKDIR /app

# Copy go mod files
COPY go.mod ./
# COPY go.sum ./ # Uncomment this when you have a go.sum file

# Download dependencies
RUN go mod download

# The source code will be mounted as a volume in development
# COPY . .

EXPOSE 8080

# The command will be overridden by docker-compose.yml in development
CMD ["go", "run", "cmd/server/main.go"] 