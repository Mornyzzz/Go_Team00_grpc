
k := $(filter-out $@,$(MAKECMDGOALS))

generate:
	protoc \
	--go_out=test \
	--go_opt=paths=source_relative \
	--go-grpc_out=test \
    --go-grpc_opt=paths=source_relative \
	scheme.proto

server:
	@cd grpc_server && go run server.go

client:
	@cd grpc_client && go run client.go -k $(k)


.PHONY: all