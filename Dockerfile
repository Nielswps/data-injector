FROM golang:1.25.1-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -trimpath" -o /go-redis-tool .

FROM alpine:3.19

WORKDIR /

COPY --from=builder /go-redis-tool .

ENTRYPOINT ["/data-injector"]