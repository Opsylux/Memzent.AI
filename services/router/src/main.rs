use tonic::{transport::Server, Request, Response, Status};
use qdrant_client::Qdrant;

// Import the generated code from the proto
pub mod router_proto {
    tonic::include_proto!("router"); 
}

use router_proto::semantic_router_server::{SemanticRouter, SemanticRouterServer};
use router_proto::{ToolRequest, ToolResponse, Tool};

pub struct MyRouter {
    q_client: Qdrant,
}

use qdrant_client::qdrant::{
    Condition, Filter, SearchPointsBuilder, FieldCondition, Match, r#match::MatchValue,
    condition::ConditionOneOf
};

#[tonic::async_trait]
impl SemanticRouter for MyRouter {
    async fn select_tools(
        &self,
        request: Request<ToolRequest>,
    ) -> Result<Response<ToolResponse>, Status> {
        let req = request.into_inner();
        
        println!("Received request for user: {}", req.user_id);

        // 1. Mock Vector Embedding (In production, call an embedding model API here)
        // Using a 384-dimension vector typical for open-source models (e.g., all-MiniLM-L6-v2)
        let mock_vector: Vec<f32> = vec![0.1; 384];

        // 2. Build Payload Filter for RBAC (allowed_tool_ids)
        let mut filter = None;
        if !req.allowed_tool_ids.is_empty() {
            let should_conditions: Vec<Condition> = req.allowed_tool_ids.iter().map(|id| {
                Condition {
                    condition_one_of: Some(ConditionOneOf::Field(FieldCondition {
                        key: "tool_id".to_string(),
                        r#match: Some(Match {
                            match_value: Some(MatchValue::Keyword(id.clone())),
                        }),
                        ..Default::default()
                    })),
                }
            }).collect();

            filter = Some(Filter {
                should: should_conditions,
                ..Default::default()
            });
        }

        // 3. Search Qdrant
        let search_request = SearchPointsBuilder::new("tools_collection", mock_vector, 3)
            .filter(filter.unwrap_or_default())
            .with_payload(true)
            .build();

        let search_result = match self.q_client.search_points(search_request).await {
            Ok(res) => res,
            Err(e) => {
                eprintln!("Qdrant search failed: {}", e);
                return Err(Status::internal("Vector search failed"));
            }
        };

        // 4. Map Results to ToolResponse
        let mut tools = Vec::new();
        for scored_point in search_result.result {
            let payload = scored_point.payload;
            let tool_id = payload.get("tool_id")
                .and_then(|v| v.kind.as_ref())
                .map(|k| match k {
                    qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                    _ => "unknown".to_string(),
                })
                .unwrap_or_else(|| "unknown".to_string());
                
            let tool_name = payload.get("tool_name")
                .and_then(|v| v.kind.as_ref())
                .map(|k| match k {
                    qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                    _ => tool_id.clone(),
                })
                .unwrap_or_else(|| tool_id.clone());

            tools.push(Tool {
                id: tool_id,
                name: tool_name,
                relevance_score: scored_point.score,
            });
        }

        if tools.is_empty() {
             tools.push(Tool {
                id: "tool_fallback".to_string(),
                name: "fallback_tool".to_string(),
                relevance_score: 0.1,
            });
        }

        let reply = ToolResponse {
            tools,
            total_tokens_saved: 450,
        };

        Ok(Response::new(reply))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Docker: Bind to [::]:50051 instead of [::1]:50051
    let addr = "[::]:50051".parse()?;
    
    // Use environment variable for Qdrant URL
    let qdrant_url = std::env::var("QDRANT_URL").unwrap_or_else(|_| "http://qdrant:6333".to_string());
    
    let q_client = Qdrant::from_url(&qdrant_url).build()?;
    
    let router_service = MyRouter { q_client };

    println!("Aura Semantic Router listening on {}", addr);
    println!("Connecting to Qdrant at: {}", qdrant_url);

    Server::builder()
        .add_service(SemanticRouterServer::new(router_service))
        .serve(addr)
        .await?;

    Ok(())
}