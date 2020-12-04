FROM golang:1.15

MAINTAINER Mohammad Jafari <m.jafari9877@gmail.com>

ADD . /go/src/github.com/mjafari98/go-auth/

WORKDIR /go/src/github.com/mjafari98/go-auth/
RUN go mod tidy
RUN go install github.com/mjafari98/go-auth/cs/server/

ENTRYPOINT ["/go/bin/server"]
