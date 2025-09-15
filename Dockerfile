FROM golang:1.22.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -trimpath" -o /go-redis-tool .

FROM alpine:3.19

WORKDIR /

COPY --from=builder /go-redis-tool .

ENTRYPOINT ["/data-injector"]