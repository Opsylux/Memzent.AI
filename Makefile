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

# Run semantic cache correctness tests (requires MEMZENT_API_KEY env var)
test-cache:
	cd services/gateway && go run scripts/test_semantic_cache/main.go

# Run agent memory & session knowledge tests (requires MEMZENT_API_KEY env var)
test-memory:
	cd services/gateway && go run scripts/test_agent_memory/main.go

# Run entity extraction cache guard tests (requires MEMZENT_API_KEY env var)
test-entity:
	cd services/gateway && go run scripts/test_entity_extraction/main.go

# Run full Evolution Pipeline integration test E1-E5 (requires MEMZENT_API_KEY + running gateway)
test-evolution:
	cd services/gateway && go run scripts/test_evolution/main.go

.PHONY: gen-proto up down test-flow test-cache test-memory test-entity test-evolution