FROM  alpine:latest

WORKDIR /root/

COPY ./artifacts/ /root/

CMD ["./main"]
