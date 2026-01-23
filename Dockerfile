FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download

# Build Proxy
RUN CGO_ENABLED=0 go build -o /bin/proxy cmd/proxy/main.go

# Build Metrics App
RUN CGO_ENABLED=0 go build -o /bin/metrics_app cmd/metrics_app/main.go

# --- Proxy Image ---
FROM alpine:latest AS proxy
WORKDIR /app
# Copy the static binary directly from host
COPY proxy-linux-amd64 /app/proxy
RUN chmod +x /app/proxy
ENTRYPOINT ["/app/proxy"]

# --- Metrics App Image ---
FROM alpine:latest AS metrics-app
WORKDIR /app
COPY --from=builder /bin/metrics_app /app/metrics_app
ENTRYPOINT ["/app/metrics_app"]
