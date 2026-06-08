mod embedding;
mod collections;
mod handlers;

pub mod router_proto {
    tonic::include_proto!("router");
}

use std::sync::Arc;
use tonic::transport::Server;
use qdrant_client::Qdrant;
use fastembed::{TextEmbedding, InitOptions, EmbeddingModel};

use router_proto::semantic_router_server::SemanticRouterServer;
use embedding::{EmbeddingCache, Embedder};
use collections::init_optimized_collection;
use handlers::MyRouter;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::]:50051".parse()?;
    let qdrant_url = std::env::var("QDRANT_URL").unwrap_or_else(|_| "http://qdrant:6334".to_string());

    let q_client = Qdrant::from_url(&qdrant_url).build()?;

    // Initialize collections with SQ, Memmap, On-Disk storage, and payload indexing
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

    let embedder = Embedder::new(
        Arc::new(model),
        Arc::new(EmbeddingCache::new(2000)),
    );

    let router_service = MyRouter { q_client, embedder };

    println!("Memzent Semantic Router listening on {}", addr);
    Server::builder()
        .add_service(SemanticRouterServer::new(router_service))
        .serve(addr)
        .await?;

    Ok(())
}
