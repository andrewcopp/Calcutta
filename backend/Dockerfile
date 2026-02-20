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

# Run as non-root user (distroless provides nonroot:nonroot)
USER nonroot:nonroot

# Run the built binary
CMD ["/app/bin/api"]

FROM golang:1.24 AS worker-builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/workers ./cmd/workers \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM python:3.12-slim AS worker-prod

WORKDIR /app

COPY --from=worker-builder /out/workers /app/bin/workers
COPY --from=worker-builder /out/migrate /app/bin/migrate
COPY --from=worker-builder /app/migrations /app/migrations

COPY data-science/requirements.txt /app/data-science/requirements.txt
RUN pip install --no-cache-dir -r /app/data-science/requirements.txt

COPY data-science/ /app/data-science/

RUN useradd --create-home --shell /bin/bash appuser
USER appuser

CMD ["/app/bin/workers"]

FROM golang:1.24 AS dev

WORKDIR /app

# Install Python for running data-science scripts
RUN apt-get update && apt-get install -y python3 python3-pip python3-venv && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/api"]