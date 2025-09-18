FROM golang:1.25.1-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /data-injector .

FROM alpine:3.22

WORKDIR /

COPY --from=builder /data-injector .

ENTRYPOINT ["/data-injector"]