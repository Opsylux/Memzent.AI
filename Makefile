# Root Makefile

gen-proto:
	# Generate Go code for the Gateway directly into the target package
	protoc --go_out=services/gateway/internal/router --go_opt=paths=source_relative \
	       --go-grpc_out=services/gateway/internal/router --go-grpc_opt=paths=source_relative \
	       -I proto proto/router.proto
	# Generate Rust code for the Router
	# (Assumes you have tonic-build configured in the Rust service)
	cd services/router && cargo build

up:
	docker-compose up -d --build

down:
	docker-compose down
# View live logs from both Go and Rust
logs:
	docker-compose logs -f gateway router

# Run the end-to-end neural integration test client
test-flow:
	cd services/gateway && go run scripts/test_flow.go

.PHONY: gen-proto up down test-flow