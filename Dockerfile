FROM golang:latest
LABEL maintainer avvero

ADD . /app
WORKDIR /app
RUN go get github.com/stretchr/stew
RUN go get github.com/stretchr/signature
RUN go build -o main .
CMD ["/app/main", "-httpPort=8080", "-infoUpdateInterval=5", "-heraldEndpoint=https://f2g.site/bot/herald/api/message", "-metricsPullInterval=5", "-graphiteUrl=172.16.81.155:2003", "-graphiteDashboard=http://172.16.81.155:8181"]