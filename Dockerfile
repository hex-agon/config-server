FROM golang:1.18-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN GOOS=linux GOARCH=amd64 go build -o config-server

FROM alpine
WORKDIR /app
COPY --from=builder /app/config-server ./
CMD ["./config-server"]