ready:
	go mod tidy
	go install \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc

gen:
	protoc \
		-I$(HOME)/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=pb \
		--go_opt=paths=source_relative \
		--go-grpc_out=pb \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb \
		--plugin=protoc-gen-grpc-gateway=$(HOME)/go/bin/protoc-gen-grpc-gateway \
		--proto_path=proto \
		proto/*.proto

clean:
	rm -rf pb/*.go

run:
	go install cs/server/*.go
	auth_server