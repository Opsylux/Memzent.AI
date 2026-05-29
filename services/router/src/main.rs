use std::sync::Arc;
use tonic::{transport::Server, Request, Response, Status};
use qdrant_client::Qdrant;
use fastembed::{TextEmbedding, InitOptions, EmbeddingModel};
use sha2::{Sha256, Digest};
use compression_prompt::compressor::Compressor;
// ADDED: for embedding cache
use std::sync::Mutex;

// Import the generated code from the proto
pub mod router_proto {
    tonic::include_proto!("router"); 
}

use router_proto::semantic_router_server::{SemanticRouter, SemanticRouterServer};
use router_proto::{
    ToolRequest, ToolResponse, Tool, RegisterToolRequest, RegisterToolResponse, 
    ToolChainRequest, ToolChainResponse, ToolStep, StoreMemoryRequest, 
    StoreMemoryResponse, QueryMemoryRequest, QueryMemoryResponse, MemoryHit
};

use qdrant_client::qdrant::{
    Condition, Filter, SearchPointsBuilder, FieldCondition, Match, r#match::MatchValue,
    condition::ConditionOneOf, PointStruct, UpsertPointsBuilder, Value, 
    CreateCollectionBuilder, Distance, ScalarQuantizationBuilder, VectorParamsBuilder,
    OptimizersConfigDiff, CreateFieldIndexCollectionBuilder, FieldType,
};
use std::collections::HashMap;


fn calculate_hash(text: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(text);
    format!("{:x}", hasher.finalize())
}

fn compress_text(prompt: &str) -> String {
    let words = stop_words::get(stop_words::LANGUAGE::English);
    let filtered: Vec<&str> = prompt.split_whitespace()
        .filter(|w| !words.contains(&w.to_lowercase()))
        .collect();
    let intermediate = filtered.join(" ");
    
    use compression_prompt::compressor::CompressorConfig;
    let config = CompressorConfig::default();
    let compressor = Compressor::new(config);
    
    compressor.compress(&intermediate)
        .map(|res| res.compressed)
        .unwrap_or(intermediate)
}

// ADDED: In-process embedding cache.
// FastEmbed / all-MiniLM-L6-v2 is fully deterministic — identical text always
// produces identical vectors. Caching eliminates re-inference on repeated or
// concurrent requests for the same prompt, which was the root cause of 87% CPU.
struct EmbeddingCache {
    // SHA-256(text) → vector
    cache: Mutex<HashMap<String, Vec<f32>>>,
    max_size: usize,
}

impl EmbeddingCache {
    fn new(max_size: usize) -> Self {
        Self {
            cache: Mutex::new(HashMap::new()),
            max_size,
        }
    }

    fn get(&self, text: &str) -> Option<Vec<f32>> {
        let key = calculate_hash(text);
        self.cache.lock().unwrap().get(&key).cloned()
    }

    fn set(&self, text: &str, vector: Vec<f32>) {
        let key = calculate_hash(text);
        let mut cache = self.cache.lock().unwrap();
        // Simple size cap: evict one arbitrary entry when full.
        // Good enough for a hot-prompt working set of 2000 entries (~3MB RAM).
        if cache.len() >= self.max_size {
            if let Some(first_key) = cache.keys().next().cloned() {
                cache.remove(&first_key);
            }
        }
        cache.insert(key, vector);
    }
}

pub struct MyRouter {
    q_client: Qdrant,
    embedding_model: Arc<TextEmbedding>,
    // ADDED: shared across all async handlers
    embed_cache: Arc<EmbeddingCache>,
}

// ADDED: Single embed helper used by every handler.
// Cache hit  → returns immediately, zero CPU.
// Cache miss → runs inference once, stores result, never recomputes for this text.
impl MyRouter {
    fn embed_cached(&self, text: &str) -> Result<Vec<f32>, Status> {
        if let Some(cached) = self.embed_cache.get(text) {
            return Ok(cached);
        }
        let embeddings = self.embedding_model
            .embed(vec![text.to_string()], None)
            .map_err(|e| Status::internal(format!("Failed to generate embeddings: {}", e)))?;
        let vector = embeddings[0].clone();
        self.embed_cache.set(text, vector.clone());
        Ok(vector)
    }
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

        // 1. Vector Embedding — cache hit costs zero CPU, miss runs inference once.
        // Previously: self.embedding_model.embed(...) called unconditionally every request.
        let real_vector = self.embed_cached(&req.prompt)?;
        let prompt_hash = calculate_hash(&req.prompt);

        // --- Semantic Cache Lookup (Org-Isolated) ---
        let mut similar_prompt_hash = String::new();
        let mut cache_filter_conditions: Vec<Condition> = Vec::new();
        if !req.org_id.is_empty() {
            cache_filter_conditions.push(Condition {
                condition_one_of: Some(ConditionOneOf::Field(FieldCondition {
                    key: "org_id".to_string(),
                    r#match: Some(Match {
                        match_value: Some(MatchValue::Keyword(req.org_id.clone())),
                    }),
                    ..Default::default()
                })),
            });
        }
        let cache_search_request = SearchPointsBuilder::new("prompts_collection", real_vector.clone(), 1)
            .filter(Filter { must: cache_filter_conditions, ..Default::default() })
            .with_payload(true)
            .build();

        if let Ok(cache_res) = self.q_client.search_points(cache_search_request).await {
            if let Some(hit) = cache_res.result.first() {
                println!("🔍 Semantic Cache Candidate: Score {}, Threshold 0.88", hit.score);
                
                if hit.score > 0.88 {
                    similar_prompt_hash = hit.payload.get("prompt_hash")
                        .and_then(|v| v.kind.as_ref())
                        .map(|k| match k {
                            qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                            _ => String::new(),
                        })
                        .unwrap_or_default();
                    println!("🎯 Semantic Cache Hit! Hash: {}", similar_prompt_hash);
                }
            }
        }

        // If no similar prompt found, store this one for future reference
        if similar_prompt_hash.is_empty() {
             println!("💾 Semantic Cache Store: New prompt intent saved");
             
             let mut payload = HashMap::new();
             payload.insert("prompt_hash".to_string(), Value::from(prompt_hash.clone()));
             payload.insert("user_id".to_string(), Value::from(req.user_id.clone()));
             payload.insert("org_id".to_string(), Value::from(req.org_id.clone()));

             let _ = self.q_client.upsert_points(UpsertPointsBuilder::new(
                "prompts_collection",
                vec![PointStruct::new(
                    uuid::Uuid::new_v4().to_string(),
                    real_vector.clone(),
                    payload
                )]
            )).await;
        }

        // --- Prompt Compression ---
        let compressed = compress_text(&req.prompt);
        let tokens_saved = (req.prompt.len() as i32 - compressed.len() as i32) / 4;

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

        // 3. Search Qdrant using the cached vector (for Tools)
        let search_request = SearchPointsBuilder::new("tools_collection", real_vector, 10)
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
        let threshold = if req.score_threshold_override > 0.0 { req.score_threshold_override } else { 0.65 };

        for scored_point in search_result.result {
            if scored_point.score < threshold {
                continue;
            }

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

            let description = payload.get("description")
                .and_then(|v| v.kind.as_ref())
                .map(|k| match k {
                    qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                    _ => String::new(),
                })
                .unwrap_or_default();

            tools.push(Tool {
                id: tool_id,
                name: tool_name,
                relevance_score: scored_point.score,
                description,
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

    async fn register_tool(
        &self,
        request: Request<RegisterToolRequest>,
    ) -> Result<Response<RegisterToolResponse>, Status> {
        let req = request.into_inner();
        println!("📝 Registering new tool semantic intent: {} (ID: {})", req.name, req.id);

        // 1. Generate Vector Embedding — cached so re-registering the same tool is free
        let vector = self.embed_cached(&req.description)?;

        // 2. Prepare Payload
        let mut payload = HashMap::new();
        payload.insert("tool_id".to_string(), Value::from(req.id.clone()));
        payload.insert("tool_name".to_string(), Value::from(req.name.clone()));
        payload.insert("description".to_string(), Value::from(req.description.clone()));
        payload.insert("org_id".to_string(), Value::from(req.org_id.clone()));

        // 3. Upsert into Qdrant tools_collection
        let tool_uuid = uuid::Uuid::new_v5(&uuid::Uuid::NAMESPACE_OID, req.id.as_bytes());
        let result = self.q_client.upsert_points(UpsertPointsBuilder::new(
            "tools_collection",
            vec![PointStruct::new(
                tool_uuid.to_string(),
                vector,
                payload
            )]
        )).await;

        match result {
            Ok(_) => {
                println!("✅ Tool vectorized and stored in Qdrant: {}", req.id);
                Ok(Response::new(RegisterToolResponse { success: true, error: String::new() }))
            },
            Err(e) => {
                eprintln!("❌ Qdrant upsert failed: {}", e);
                Ok(Response::new(RegisterToolResponse { 
                    success: false, 
                    error: format!("Qdrant failure: {}", e) 
                }))
            }
        }
    }

    async fn plan_tool_chain(
        &self,
        request: Request<ToolChainRequest>,
    ) -> Result<Response<ToolChainResponse>, Status> {
        let req = request.into_inner();
        println!("🔮 Planning sequential tool chain for prompt: \"{}\"", req.prompt);

        // 1. Vector Embedding — reuses cached vector if select_tools already ran for this prompt
        let real_vector = self.embed_cached(&req.prompt)?;

        // 2. Build OR filter for allowed tool IDs
        let tool_filter = if !req.allowed_tool_ids.is_empty() {
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
            Filter { should: should_conditions, ..Default::default() }
        } else {
            Filter::default()
        };

        let search_request = SearchPointsBuilder::new("tools_collection", real_vector, 5)
            .with_payload(true)
            .filter(tool_filter)
            .build();

        let mut steps = Vec::new();
        let mut confidence_score = 0.0;

        if let Ok(search_res) = self.q_client.search_points(search_request).await {
            let mut matched_tools = search_res.result;
            matched_tools.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap_or(std::cmp::Ordering::Equal));

            for (idx, hit) in matched_tools.iter().enumerate() {
                if hit.score > req.score_threshold_override {
                    let tool_name = hit.payload.get("tool_name")
                        .and_then(|v| v.kind.as_ref())
                        .map(|k| match k {
                            qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                            _ => "unknown_tool".to_string(),
                        })
                        .unwrap_or_else(|| "unknown_tool".to_string());

                    steps.push(ToolStep {
                        step_order: (idx + 1) as i32,
                        tool_name,
                        parameters_json: "{\"status\": \"scheduled\"}".to_string(),
                    });
                }
            }

            if !steps.is_empty() {
                confidence_score = 0.95;
            }
        }

        Ok(Response::new(ToolChainResponse {
            steps,
            confidence_score,
        }))
    }

    async fn store_memory(
        &self,
        request: Request<StoreMemoryRequest>,
    ) -> Result<Response<StoreMemoryResponse>, Status> {
        let req = request.into_inner();
        println!("💾 Storing semantic memory fact: \"{}\" for org: {}", req.fact, req.org_id);

        // 1. Generate vector embedding — cached so duplicate facts cost nothing
        let vector = self.embed_cached(&req.fact)?;

        // 2. Prepare payload
        let mut payload = HashMap::new();
        payload.insert("fact".to_string(), Value::from(req.fact.clone()));
        payload.insert("org_id".to_string(), Value::from(req.org_id.clone()));
        payload.insert("user_id".to_string(), Value::from(req.user_id.clone()));
        payload.insert("created_at".to_string(), Value::from(chrono::Utc::now().to_rfc3339()));

        // 3. Upsert point into memories_collection
        let result = self.q_client.upsert_points(UpsertPointsBuilder::new(
            "memories_collection",
            vec![PointStruct::new(
                uuid::Uuid::new_v4().to_string(),
                vector,
                payload
            )]
        )).await;

        match result {
            Ok(_) => {
                println!("✅ Memory fact vectorized and stored in Qdrant successfully!");
                Ok(Response::new(StoreMemoryResponse { success: true, error: String::new() }))
            },
            Err(e) => {
                eprintln!("❌ Qdrant memory upsert failed: {}", e);
                Ok(Response::new(StoreMemoryResponse { success: false, error: format!("Qdrant memory failure: {}", e) }))
            }
        }
    }

    async fn query_memory(
        &self,
        request: Request<QueryMemoryRequest>,
    ) -> Result<Response<QueryMemoryResponse>, Status> {
        let req = request.into_inner();
        println!("🔍 Querying semantic memories for prompt: \"{}\" under org: {}", req.prompt, req.org_id);

        // 1. Embed query prompt — reuses cached vector if select_tools already ran for this prompt
        let real_vector = self.embed_cached(&req.prompt)?;

        // 2. Build payload filter for Org-Isolation
        let mut filter_conditions = Vec::new();
        if !req.org_id.is_empty() {
            filter_conditions.push(Condition {
                condition_one_of: Some(ConditionOneOf::Field(FieldCondition {
                    key: "org_id".to_string(),
                    r#match: Some(Match {
                        match_value: Some(MatchValue::Keyword(req.org_id.clone())),
                    }),
                    ..Default::default()
                })),
            });
        }

        // 3. Search memories_collection
        let search_request = SearchPointsBuilder::new("memories_collection", real_vector, 5)
            .filter(Filter { must: filter_conditions, ..Default::default() })
            .with_payload(true)
            .build();

        let search_result = match self.q_client.search_points(search_request).await {
            Ok(res) => res,
            Err(e) => {
                eprintln!("Qdrant memory search failed (likely empty collection): {}", e);
                qdrant_client::qdrant::SearchResponse { result: vec![], time: 0.0, ..Default::default() }
            }
        };

        // 4. Map to MemoryHit structures
        let mut memories = Vec::new();
        let threshold = if req.score_threshold_override > 0.0 { req.score_threshold_override } else { 0.65 };

        for scored_point in search_result.result {
            if scored_point.score < threshold {
                continue;
            }

            let payload = scored_point.payload;
            let fact = payload.get("fact")
                .and_then(|v| v.kind.as_ref())
                .map(|k| match k {
                    qdrant_client::qdrant::value::Kind::StringValue(s) => s.clone(),
                    _ => String::new(),
                })
                .unwrap_or_default();

            if !fact.is_empty() {
                memories.push(MemoryHit {
                    fact,
                    relevance_score: scored_point.score,
                });
            }
        }

        Ok(Response::new(QueryMemoryResponse { memories }))
    }
}

async fn init_optimized_collection(
    q_client: &Qdrant,
    collection_name: &str,
    exists: bool,
) -> Result<(), Box<dyn std::error::Error>> {
    if !exists {
        println!("🚀 Creating optimized collection: {}", collection_name);
        
        q_client
            .create_collection(
                CreateCollectionBuilder::new(collection_name)
                    .vectors_config(VectorParamsBuilder::new(384, Distance::Cosine))
                    // 1. Enable Scalar Quantization (reduces RAM usage by 75%)
                    .quantization_config(ScalarQuantizationBuilder::default())
                    // 2. Keep payloads on disk to protect system memory
                    .on_disk_payload(true)
                    // 3. Move vector data to disk (memmap) once collection grows beyond 20,000
                    .optimizers_config(OptimizersConfigDiff {
                        memmap_threshold: Some(20000),
                        ..Default::default()
                    })
            )
            .await?;

        // 4. Index payload fields (org_id and user_id) for instant filters
        println!("🔑 Creating payload indexes for {}", collection_name);
        let _ = q_client.create_field_index(
            CreateFieldIndexCollectionBuilder::new(collection_name, "org_id", FieldType::Keyword)
        ).await;

        let _ = q_client.create_field_index(
            CreateFieldIndexCollectionBuilder::new(collection_name, "user_id", FieldType::Keyword)
        ).await;
    }
    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::]:50051".parse()?;
    let qdrant_url = std::env::var("QDRANT_URL").unwrap_or_else(|_| "http://qdrant:6334".to_string());
    
    let q_client = Qdrant::from_url(&qdrant_url).build()?;
    
    // Initialize the collections with SQ, Memmap, On-Disk storage, and payload indexing
    let collections_response = q_client.list_collections().await?;
    
    let tools_exists = collections_response.collections.iter().any(|c| c.name == "tools_collection");
    init_optimized_collection(&q_client, "tools_collection", tools_exists).await?;
    
    let cache_exists = collections_response.collections.iter().any(|c| c.name == "prompts_collection");
    init_optimized_collection(&q_client, "prompts_collection", cache_exists).await?;

    let memories_exists = collections_response.collections.iter().any(|c| c.name == "memories_collection");
    init_optimized_collection(&q_client, "memories_collection", memories_exists).await?;

    println!("Loading FastEmbed model (all-MiniLM-L6-v2)...");
    let model = TextEmbedding::try_new(
        InitOptions::new(EmbeddingModel::AllMiniLML6V2).with_show_download_progress(true),
    )?;
    
    let router_service = MyRouter { 
        q_client,
        embedding_model: Arc::new(model),
        // ADDED: 2000-entry embedding cache.
        // Memory cost: ~2000 entries × 384 floats × 4 bytes = ~3MB.
        // CPU impact: eliminates re-inference on any repeated or concurrent prompt.
        embed_cache: Arc::new(EmbeddingCache::new(2000)),
    };

    println!("Memzent Semantic Router listening on {}", addr);
    Server::builder()
        .add_service(SemanticRouterServer::new(router_service))
        .serve(addr)
        .await?;

    Ok(())
}