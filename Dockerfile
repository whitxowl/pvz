# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/pvz ./cmd/app/main.go
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.21 AS final

WORKDIR /app

COPY --from=builder /app/bin/pvz ./pvz
COPY --from=builder /go/bin/goose ./goose
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 3000 9000

CMD ["./pvz"]