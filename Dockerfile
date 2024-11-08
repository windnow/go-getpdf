# Первый этап - сборка приложения
FROM golang:1.23-alpine as builder

WORKDIR /build

RUN apk add --no-cache git build-base

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main main.go

# Второй этап - финальный образ с приложением
FROM alpine:latest

WORKDIR /srv/

# Установка необходимых пакетов, включая Chromium
RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl chromium nss freetype ttf-freefont harfbuzz dumb-init

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /build/main .

# Копируем данные и создаем необходимые директории
COPY template.html .
COPY logo.png .
RUN mkdir res
RUN touch res/config.toml

# Устанавливаем путь к браузеру Chromium в переменные окружения
ENV CHROME_BIN="/usr/bin/chromium-browser"
ENV CHROME_PATH="/usr/lib/chromium/"

# Порт, который будет прослушивать сервер
EXPOSE 8080

# Используем dumb-init для корректного управления процессами внутри контейнера
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Запускаем приложение
CMD ["./main"]
