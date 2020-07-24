FROM golang:1 AS builder

COPY . /app/src

WORKDIR /app/src

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM  alpine:latest

WORKDIR /root/

COPY --from=builder /app/src/main .

CMD ["./main"]
