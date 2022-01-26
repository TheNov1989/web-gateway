FROM golang

ENV GO111MODULE=on

ADD . /go/src/github.com/google/web-api-gateway

RUN go get github.com/google/web-gatewayserver@latest
RUN go install github.com/google/web-gatewayserver@latest
RUN go install github.com/google/web-gatewaysetuptool@latest
RUN go install github.com/google/web-gatewayconnectiontest@latest

ENTRYPOINT ["/go/bin/server"]

EXPOSE 443


