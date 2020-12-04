clean:
	rm -rf pb/*.go

ready:
	protoc \
    		--go_out=pb \
    		--go_opt=paths=source_relative \
    		--go-grpc_out=pb \
    		--go-grpc_opt=paths=source_relative \
    		--grpc-gateway_out=pb \
    		--proto_path=proto \
    		proto/*.proto
	go mod tidy

run:
	go install ./cs/server/
	server