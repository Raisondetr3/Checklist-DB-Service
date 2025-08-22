.PHONY: run build test clean proto-gen proto-clean

run:
	go run cmd/api/main.go

build:
	go build -o bin/api-service cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ logs/

deps:
	go mod download
	go mod tidy

proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       internal/proto/task.proto

proto-clean:
	rm -f pb/*.pb.go