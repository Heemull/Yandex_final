# Этап сборки
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/web ./web

EXPOSE 7540

# Переменные окружения по умолчанию (можно переопределить при запуске)
ENV TODO_PORT=7540
ENV WEB_DIR=./web
ENV TODO_DBFILE=scheduler.db
ENV TODO_PASSWORD=12345

# Запуск приложения
CMD ["./main"]