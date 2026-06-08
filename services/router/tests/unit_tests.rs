// Pure unit tests for the Memzent Semantic Router library.

use memzent_router::{
    calculate_hash, compress_text, estimate_tokens_saved, resolve_score_threshold,
    should_semantic_cache_hit, tool_uuid, DEFAULT_SCORE_THRESHOLD, SEMANTIC_CACHE_THRESHOLD,
};

#[test]
fn test_hash_is_deterministic() {
    let h1 = calculate_hash("hello world");
    let h2 = calculate_hash("hello world");
    assert_eq!(h1, h2);
}

#[test]
fn test_hash_length_is_64_chars() {
    assert_eq!(calculate_hash("any input").len(), 64);
}

#[test]
fn test_compress_text_retains_semantic_terms() {
    let result = compress_text("find customer lookup CRM database error");
    assert!(result.contains("customer") || result.contains("CRM") || result.contains("database"));
}

#[test]
fn test_resolve_score_threshold_defaults() {
    assert_eq!(resolve_score_threshold(0.0), DEFAULT_SCORE_THRESHOLD);
    assert_eq!(resolve_score_threshold(-0.1), DEFAULT_SCORE_THRESHOLD);
    assert_eq!(resolve_score_threshold(0.90), 0.90);
}

#[test]
fn test_semantic_cache_threshold() {
    assert!(should_semantic_cache_hit(0.95));
    assert!(!should_semantic_cache_hit(SEMANTIC_CACHE_THRESHOLD));
    assert!(!should_semantic_cache_hit(0.75));
}

#[test]
fn test_estimate_tokens_saved() {
    assert_eq!(estimate_tokens_saved(100, 100), 0);
    assert_eq!(estimate_tokens_saved(100, 50), 12);
    assert_eq!(estimate_tokens_saved(2000, 600), 350);
}

#[test]
fn test_tool_uuid_is_stable() {
    let u1 = tool_uuid("customer_lookup");
    let u2 = tool_uuid("customer_lookup");
    assert_eq!(u1, u2);
    assert_ne!(tool_uuid("customer_lookup"), tool_uuid("order_search"));
}
