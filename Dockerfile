FROM golang:1.14.14-stretch
LABEL maintainer avvero

ADD . /app
WORKDIR /app
RUN go get github.com/stretchr/stew
RUN go get github.com/stretchr/signature
RUN go get github.com/fatih/pool
RUN go build -o main .
CMD ["/app/main", "-httpPort=8080", "-infoUpdateInterval=30", "-heraldEndpoint=https://avvero.pw/bot/kid/api/message"]
