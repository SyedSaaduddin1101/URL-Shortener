.PHONY: proto build run clean

# Generate gRPC code from protobuf
proto:
	protoc --go_out=. --go-grpc_out=. proto/kv.proto

# Build the binary
build: proto
	go build -o bin/server ./cmd/server

# Run the server locally (without Docker)
run: build
	./bin/server

# Clean up
clean:
	rm -rf bin/
	go clean