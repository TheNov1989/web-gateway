FROM golang

ENV GO111MODULE=on

ADD . /go/src/github.com/TheNov1989/web-gateway/

RUN go get github.com/TheNov1989/web-gateway/server@latest
RUN go install github.com/TheNov1989/web-gateway/server@latest
RUN go install github.com/TheNov1989/web-gateway/setuptool@latest
RUN go install github.com/TheNov1989/web-gateway/connectiontest@latest

ENTRYPOINT ["/go/bin/server"]

EXPOSE 443
