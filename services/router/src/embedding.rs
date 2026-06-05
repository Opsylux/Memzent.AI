use std::sync::Arc;
use dashmap::DashMap;
use fastembed::TextEmbedding;
use sha2::{Sha256, Digest};
use tonic::Status;
use compression_prompt::compressor::Compressor;

pub fn calculate_hash(text: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(text);
    format!("{:x}", hasher.finalize())
}

pub fn compress_text(prompt: &str) -> String {
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

/// In-process embedding cache using lock-free concurrent DashMap.
/// FastEmbed / all-MiniLM-L6-v2 is fully deterministic — identical text always
/// produces identical vectors.
pub struct EmbeddingCache {
    cache: DashMap<String, Vec<f32>>,
    max_size: usize,
}

impl EmbeddingCache {
    pub fn new(max_size: usize) -> Self {
        Self {
            cache: DashMap::with_capacity(max_size),
            max_size,
        }
    }

    pub fn get(&self, text: &str) -> Option<Vec<f32>> {
        let key = calculate_hash(text);
        self.cache.get(&key).map(|v| v.value().clone())
    }

    pub fn set(&self, text: &str, vector: Vec<f32>) {
        let key = calculate_hash(text);
        if self.cache.len() >= self.max_size {
            if let Some(entry) = self.cache.iter().next() {
                let evict_key = entry.key().clone();
                drop(entry);
                self.cache.remove(&evict_key);
            }
        }
        self.cache.insert(key, vector);
    }
}

/// Wraps the embedding model + cache into a single helper for reuse across handlers.
pub struct Embedder {
    pub model: Arc<TextEmbedding>,
    pub cache: Arc<EmbeddingCache>,
}

impl Embedder {
    pub fn new(model: Arc<TextEmbedding>, cache: Arc<EmbeddingCache>) -> Self {
        Self { model, cache }
    }

    /// Returns a cached embedding or computes one. Lock-free on cache hit.
    pub fn embed(&self, text: &str) -> Result<Vec<f32>, Status> {
        if let Some(cached) = self.cache.get(text) {
            return Ok(cached);
        }
        let embeddings = self.model
            .embed(vec![text.to_string()], None)
            .map_err(|e| Status::internal(format!("Failed to generate embeddings: {}", e)))?;
        let vector = embeddings[0].clone();
        self.cache.set(text, vector.clone());
        Ok(vector)
    }
}
