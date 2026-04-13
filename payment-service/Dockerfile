FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY ./payment-service/go.mod ./payment-service/go.sum* ./
COPY ./mock-repo-b /mock-repo-b
RUN go mod download
COPY ./payment-service .
RUN go mod tidy
RUN GOOS=linux go build -o payment-service ./cmd/payment-service/main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8081 50052
CMD ["./payment-service"]
