# --- Build Stage ---
FROM golang:1.22 as builder
WORKDIR /app

# Copy go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy all backend files
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# --- Deploy Stage ---
FROM alpine:latest
WORKDIR /root/

# Copy built binary
COPY --from=builder /app/app .

# .env is optional (remove jika tidak ingin upload env)
# COPY --from=builder /app/.env .

# Alpine dependencies (fix Railway crash)
RUN apk add --no-cache tzdata

EXPOSE 8080

CMD ["./app"]
