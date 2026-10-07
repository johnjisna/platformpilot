FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o platformpilot ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/platformpilot .

EXPOSE 8080

CMD ["./platformpilot"]
