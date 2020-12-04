FROM golang:1.15

MAINTAINER Mohammad Jafari <m.jafari9877@gmail.com>

ADD . /go/src/github.com/mjafari98/go-auth/
WORKDIR /go/src/github.com/mjafari98/go-auth

RUN apt-get update
RUN apt-get install -y protobuf-compiler
RUN go install \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc

RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
