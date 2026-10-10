# 多阶段构建
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/api .
COPY --from=builder /build/.env.example ./.env
COPY --from=builder /build/sql ./sql
COPY --from=builder /build/internal/web/static ./internal/web/static

EXPOSE 8182
CMD ["./api"]
