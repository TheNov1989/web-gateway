FROM golang

ENV GO111MODULE=on

ADD . /go/src/github.com/google/web-api-gateway

RUN go get github.com/TheNov1989/web-gatewayserver@latest
RUN go install github.com/TheNov1989/web-gatewayserver@latest
RUN go install github.com/TheNov1989/web-gatewaysetuptool@latest
RUN go install github.com/TheNov1989/web-gatewayconnectiontest@latest

ENTRYPOINT ["/go/bin/server"]

EXPOSE 443


