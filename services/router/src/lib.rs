use compression_prompt::compressor::{Compressor, CompressorConfig};
use sha2::{Digest, Sha256};
use uuid::Uuid;

/// Default tool-matching score threshold when the client sends no override.
pub const DEFAULT_SCORE_THRESHOLD: f32 = 0.65;

/// Semantic prompt cache hit threshold (strict greater-than).
pub const SEMANTIC_CACHE_THRESHOLD: f32 = 0.88;

/// SHA-256 hex digest of text — used for cache keys and Qdrant point IDs.
pub fn calculate_hash(text: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(text);
    format!("{:x}", hasher.finalize())
}

/// Stop-word removal + compression-prompt pass for token-efficient routing.
pub fn compress_text(prompt: &str) -> String {
    let words = stop_words::get(stop_words::LANGUAGE::English);
    let filtered: Vec<&str> = prompt
        .split_whitespace()
        .filter(|w| !words.contains(&w.to_lowercase()))
        .collect();
    let intermediate = filtered.join(" ");

    let config = CompressorConfig::default();
    let compressor = Compressor::new(config);

    compressor
        .compress(&intermediate)
        .map(|res| res.compressed)
        .unwrap_or(intermediate)
}

pub fn resolve_score_threshold(override_val: f32) -> f32 {
    if override_val > 0.0 {
        override_val
    } else {
        DEFAULT_SCORE_THRESHOLD
    }
}

pub fn should_semantic_cache_hit(score: f32) -> bool {
    score > SEMANTIC_CACHE_THRESHOLD
}

pub fn estimate_tokens_saved(original_len: usize, compressed_len: usize) -> i32 {
    let raw = original_len as i32 - compressed_len as i32;
    raw / 4
}

/// Deterministic UUID v5 for tool registration deduplication.
pub fn tool_uuid(tool_id: &str) -> String {
    Uuid::new_v5(&Uuid::NAMESPACE_OID, tool_id.as_bytes()).to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hash_and_threshold_smoke() {
        assert_eq!(calculate_hash("x").len(), 64);
        assert_eq!(resolve_score_threshold(0.0), DEFAULT_SCORE_THRESHOLD);
        assert!(should_semantic_cache_hit(0.89));
    }
}
