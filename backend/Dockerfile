FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code (needed for internal packages)
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/workers ./cmd/workers \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/base-debian12 AS prod

WORKDIR /app

COPY --from=builder /out/api /app/bin/api
COPY --from=builder /out/workers /app/bin/workers
COPY --from=builder /out/migrate /app/bin/migrate
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

# Run the built binary
CMD ["/app/bin/api"]

FROM golang:1.24 AS dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/api"]