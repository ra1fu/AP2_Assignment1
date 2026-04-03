FROM golang:1.22-alpine AS builder

WORKDIR /app

# Загружаем зависимости
COPY go.mod ./
# Если будет go.sum: COPY go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN GOOS=linux go build -o payment-service ./cmd/payment-service/main.go

# Финальный легковесный образ
FROM alpine:latest  
WORKDIR /app

COPY --from=builder /app/payment-service .
# Копируем SQL-скрипты миграций
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081

CMD ["./payment-service"]
