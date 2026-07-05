# Build Stage
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /make-backend ./cmd/make-backend

# Run Stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /make-backend .

EXPOSE 8080

CMD ["./make-backend"]