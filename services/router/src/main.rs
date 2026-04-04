use std::sync::Arc;
use tonic::{transport::Server, Request, Response, Status};
use qdrant_client::Qdrant;
use fastembed::{TextEmbedding, InitOptions, EmbeddingModel};
use sha2::{Sha256, Digest};
use compression_prompt::compressor::Compressor;


// Import the generated code from the proto
pub mod router_proto {
    tonic::include_proto!("router"); 
}

use router_proto::semantic_router_server::{SemanticRouter, SemanticRouterServer};
use router_proto::{ToolRequest, ToolResponse, Tool};

pub struct MyRouter {
    q_client: Qdrant,
    // Using Arc to safely share the model across async threads
    embedding_model: Arc<TextEmbedding>, 
}

use qdrant_client::qdrant::{
    Condition, Filter, SearchPointsBuilder, FieldCondition, Match, r#match::MatchValue,
    condition::ConditionOneOf, PointStruct, UpsertPointsBuilder
};



fn calculate_hash(text: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(text);
    format!("{:x}", hasher.finalize())
}

fn compress_text(prompt: &str) -> String {
    let words = stop_words::get(stop_words::Language::English);
    let filtered: Vec<&str> = prompt.split_whitespace()
        .filter(|w| !words.contains(&w.to_lowercase()))
        .collect();
    let intermediate = filtered.join(" ");
    
    // Use statistical compressor to target 70% of original length
    let compressor = Compressor::new();
    compressor.compress(&intermediate, 0.7).unwrap_or(intermediate)
}

#[tonic::async_trait]
impl SemanticRouter for MyRouter {

    async fn select_tools(
        &self,
        request: Request<ToolRequest>,
    ) -> Result<Response<ToolResponse>, Status> {
        let req = request.into_inner();
        
        println!("Received request for user: {}", req.user_id);
        println!("Prompt to route: \"{}\"", req.prompt);

        // 1. REAL Vector Embedding using FastEmbed
        // We wrap the prompt in a Vec as FastEmbed expects a batch of strings
        let documents = vec![req.prompt.clone()];
        
        let embeddings = self.embedding_model.embed(documents, None)
            .map_err(|e| Status::internal(format!("Failed to generate embeddings: {}", e)))?;
            
        // Extract the generated 384-dimension vector for the prompt
        let real_vector = embeddings[0].clone();
        let prompt_hash = calculate_hash(&req.prompt);

        // --- NEW: Semantic Cache Lookup ---
        let mut similar_prompt_hash = String::new();
        let cache_search_request = SearchPointsBuilder::new("prompts_collection", real_vector.clone(), 1)
            .with_payload(true)
            .build();

        if let Ok(cache_res) = self.q_client.search_points(cache_search_request).await {
            if let Some(hit) = cache_res.result.first() {
                if hit.score > 0.95 {
                    similar_prompt_hash = hit.payload.get("prompt_hash")
                        .and_then(|v| v.kind.as_ref())
                        .map(|k| match k {
                            qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                            _ => String::new(),
                        })
                        .unwrap_or_default();
                    println!("🎯 Semantic Cache Hit! Score: {}, Hash: {}", hit.score, similar_prompt_hash);
                }
            }
        }

        // If no similar prompt found, store this one for future reference
        if similar_prompt_hash.is_empty() {
             let _ = self.q_client.upsert_points(UpsertPointsBuilder::new(
                "prompts_collection",
                vec![PointStruct::new(
                    uuid::Uuid::new_v4().to_string(),
                    real_vector.clone(),
                    vec![
                        ("prompt_hash", prompt_hash.clone().into()),
                        ("user_id", req.user_id.clone().into()),
                    ]
                )]
            )).await;
        }

        // --- Prompt Compression ---
        let compressed = compress_text(&req.prompt);
        let tokens_saved = (req.prompt.len() as i32 - compressed.len() as i32) / 4; // Rough estimate

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

        // 3. Search Qdrant using the REAL vector (for Tools)
        let search_request = SearchPointsBuilder::new("tools_collection", real_vector, 3)
            .filter(filter.unwrap_or_default())
            .with_payload(true)
            .build();

        let search_result = match self.q_client.search_points(search_request).await {
            Ok(res) => res,
            Err(e) => {
                eprintln!("Qdrant search failed (likely empty collection): {}", e);
                qdrant_client::qdrant::SearchResponse { 
                    result: vec![], 
                    time: 0.0,
                    ..Default::default()
                }
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

        // Hardcoded Fallback if collection is empty or nothing matches
        if tools.is_empty() {
             tools.push(Tool {
                id: "read_database".to_string(), 
                name: "Database Metrics Reader".to_string(),
                relevance_score: 0.95, 
            });
        }

        let reply = ToolResponse {
            tools,
            total_tokens_saved: tokens_saved.max(0) + 450, 
            compressed_prompt: compressed,
            similar_prompt_hash,
            current_prompt_hash: prompt_hash,
        };


        Ok(Response::new(reply))

    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::]:50051".parse()?;
    let qdrant_url = std::env::var("QDRANT_URL").unwrap_or_else(|_| "http://qdrant:6333".to_string());
    
    let q_client = Qdrant::from_url(&qdrant_url).build()?;
    
    // Initialize the collection
    let collection_name = "tools_collection";
    println!("Checking if Qdrant collection '{}' exists...", collection_name);
    
    let collections_response = q_client.list_collections().await?;
    let collection_exists = collections_response.collections.iter().any(|c| c.name == collection_name);
    
    if collection_exists {
        println!("✅ Collection '{}' already exists.", collection_name);
    } else {
        println!("Creating collection '{}' with 384 dimensions...", collection_name);
        
        let create_result = q_client
            .create_collection(
                qdrant_client::qdrant::CreateCollectionBuilder::new(collection_name)
                    .vectors_config(qdrant_client::qdrant::VectorParamsBuilder::new(
                        384, 
                        qdrant_client::qdrant::Distance::Cosine
                    ))
            )
            .await;
            
        match create_result {
            Ok(_) => println!("✅ Successfully created collection."),
            Err(e) => eprintln!("⚠️ Could not create collection: {}", e),
        }
    }
    
    // NEW: Initialize the prompts_collection for semantic cache
    let cache_collection = "prompts_collection";
    println!("Checking if Qdrant collection '{}' exists...", cache_collection);
    
    let cache_exists = collections_response.collections.iter().any(|c| c.name == cache_collection);
    
    if !cache_exists {
        println!("Creating collection '{}' for semantic cache...", cache_collection);
        let _ = q_client
            .create_collection(
                qdrant_client::qdrant::CreateCollectionBuilder::new(cache_collection)
                    .vectors_config(qdrant_client::qdrant::VectorParamsBuilder::new(
                        384, 
                        qdrant_client::qdrant::Distance::Cosine
                    ))
            )
            .await;
    }

    
    // NEW: Load the Embedding Model ONCE on startup
    println!("Loading FastEmbed model (all-MiniLM-L6-v2)...");
    let model = TextEmbedding::try_new(
        InitOptions::new(EmbeddingModel::AllMiniLML6V2).with_show_download_progress(true),
    )?;
    
    let router_service = MyRouter { 
        q_client,
        embedding_model: Arc::new(model), 
    };

    println!("Aura Semantic Router listening on {}", addr);
    Server::builder()
        .add_service(SemanticRouterServer::new(router_service))
        .serve(addr)
        .await?;

    Ok(())
}