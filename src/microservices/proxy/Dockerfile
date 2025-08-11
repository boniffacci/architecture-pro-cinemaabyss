FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o proxy-service .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/proxy-service .
EXPOSE 8000
CMD ["./proxy-service"]