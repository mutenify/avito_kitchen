# syntax=docker/dockerfile:1

# --- build stage ---
# Фиксируем minor-версию тега образа под go.mod (go 1.26.2) — иначе на CI/на
# чужой машине может подтянуться более старый golang:alpine с несовместимым
# тулчейном, и сборка упадёт с непонятной ошибкой про версию Go. Патч-версию
# не фиксируем: такого точного тега на Docker Hub может не быть, а 1.26-alpine
# всегда указывает на актуальный патч ветки 1.26.x (>= 1.26.2).
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Кэш-слой с зависимостями: пока go.mod/go.sum не поменялись, docker переиспользует
# этот слой и не скачивает модули заново при каждой правке кода.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 — статический бинарник без зависимости от libc контейнера,
# поэтому в финальный образ можно брать даже минимальный alpine без сборочных тулов.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api

# --- final stage ---
FROM alpine:3.20

# ca-certificates — на будущее, если Core API начнёт стучаться наружу по HTTPS;
# wget — только для HEALTHCHECK ниже (сам бинарник им не пользуется).
RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/api /usr/local/bin/api

EXPOSE 8080

HEALTHCHECK --interval=3s --timeout=3s --retries=10 \
    CMD wget -q -O- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
