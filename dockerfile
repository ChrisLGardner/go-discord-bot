FROM  alpine:latest

WORKDIR /root/

COPY . /root/

CMD ["./main"]
