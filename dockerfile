FROM  alpine:latest

WORKDIR /root/

COPY ./main /root/

CMD ["./main"]
