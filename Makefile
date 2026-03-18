# Root Makefile

gen-proto:
	# Generate Go code for the Gateway
	protoc --go_out=. --go-grpc_out=. proto/router.proto
	# Generate Rust code for the Router
	# (Assumes you have tonic-build configured in the Rust service)
	cd services/router && cargo build

up:
	docker-compose up -d

down:
	docker-compose down
# View live logs from both Go and Rust
logs:
	docker-compose logs -f gateway router

.PHONY: gen-proto up down