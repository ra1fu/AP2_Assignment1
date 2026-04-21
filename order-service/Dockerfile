FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY ./order-service/go.mod ./order-service/go.sum* ./
RUN go mod download
COPY ./order-service .
RUN go mod tidy
RUN GOOS=linux go build -o order-service ./cmd/order-service/main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080 50051
CMD ["./order-service"]
