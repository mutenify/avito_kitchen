# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/mock-restaurant ./cmd/mock-restaurant

# --- final stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/mock-restaurant /usr/local/bin/mock-restaurant

EXPOSE 8081

HEALTHCHECK --interval=3s --timeout=3s --retries=10 \
    CMD wget -q -O- http://localhost:8081/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/mock-restaurant"]
