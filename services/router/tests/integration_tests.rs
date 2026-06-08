// Integration tests for the Memzent Semantic Router gRPC surface.
// Requires Qdrant + router from docker-compose.test.yml:
//   docker compose -f docker-compose.test.yml up -d router
//   ROUTER_GRPC_ADDR=http://127.0.0.1:50051 cargo test --test integration_tests -- --ignored

pub mod router_proto {
    tonic::include_proto!("router");
}

use router_proto::semantic_router_client::SemanticRouterClient;
use router_proto::ToolRequest;
use tonic::transport::Channel;

#[tokio::test]
#[ignore = "requires Qdrant + router (docker-compose.test.yml)"]
async fn grpc_select_tools_returns_compressed_prompt() {
    let addr = std::env::var("ROUTER_GRPC_ADDR").unwrap_or_else(|_| "http://127.0.0.1:50051".to_string());
    let channel = Channel::from_shared(addr)
        .expect("valid router address")
        .connect()
        .await
        .expect("router gRPC connect");

    let mut client = SemanticRouterClient::new(channel);
    let response = client
        .select_tools(ToolRequest {
            prompt: "find customer tickets in CRM".to_string(),
            user_id: "test-org".to_string(),
            org_id: "test-org".to_string(),
            allowed_tool_ids: vec![],
            score_threshold_override: 0.0,
        })
        .await
        .expect("SelectTools RPC")
        .into_inner();

    assert!(
        !response.compressed_prompt.is_empty(),
        "router should return a compressed prompt"
    );
}
