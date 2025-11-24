FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o opgl-data main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/opgl-data .

EXPOSE 8081

CMD ["./opgl-data"]
