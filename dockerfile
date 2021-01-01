FROM  alpine:latest

RUN apk --no-cache add tzdata

WORKDIR /root/

COPY ./artifacts/ /root/

CMD ["./go-discord-bot"]
