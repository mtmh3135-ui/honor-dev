# --- Build Stage ---
FROM golang:1.22 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# --- Deploy Stage ---
FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./app"]
