FROM golang:1.24.3 AS builder
LABEL authors="Radosław Głasek <01180779@pw.edu.pl>"

WORKDIR /app

COPY go.mod server.go ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /token-transfer-api ./server.go

FROM alpine:latest AS runner

WORKDIR /root/

COPY --from=builder /token-transfer-api .

EXPOSE 8080

CMD ["./token-transfer-api"]
