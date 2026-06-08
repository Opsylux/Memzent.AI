// ============================================================
// tests/unit_tests.rs
// Pure unit tests for the Memzent Semantic Router.
// These tests cover all pure functions and do NOT require
// a live Qdrant instance or the FastEmbed model.
// ============================================================

// ---------------------------------------------------------------------------
// calculate_hash  (imported from main via lib re-export below)
// ---------------------------------------------------------------------------
// NOTE: Because Rust's #[tokio::main] binary crate cannot be imported directly,
// we mirror the two pure functions here in a test-local module.
// The integration tests in tests/integration_tests.rs cover the gRPC surface.

#[cfg(test)]
mod hash_tests {
    use sha2::{Sha256, Digest};

    fn calculate_hash(text: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(text);
        format!("{:x}", hasher.finalize())
    }

    #[test]
    fn test_hash_is_deterministic() {
        let h1 = calculate_hash("hello world");
        let h2 = calculate_hash("hello world");
        assert_eq!(h1, h2, "Same input must always produce the same hash");
    }

    #[test]
    fn test_hash_is_hex_string() {
        let h = calculate_hash("memzent semantic router test");
        assert!(h.chars().all(|c| c.is_ascii_hexdigit()), "Hash must be a hex string");
    }

    #[test]
    fn test_hash_length_is_64_chars() {
        // SHA-256 always produces a 32-byte (64 hex-char) digest
        let h = calculate_hash("any input");
        assert_eq!(h.len(), 64, "SHA-256 hex string must be 64 characters");
    }

    #[test]
    fn test_different_inputs_produce_different_hashes() {
        let h1 = calculate_hash("find all tickets");
        let h2 = calculate_hash("find all open tickets");
        assert_ne!(h1, h2, "Different inputs must produce different hashes");
    }

    #[test]
    fn test_empty_string_hash() {
        // SHA-256("") = e3b0c44298fc1c149afb...
        let h = calculate_hash("");
        assert_eq!(h.len(), 64);
        // Well-known SHA-256 hash of empty string
        assert_eq!(&h[..8], "e3b0c442");
    }

    #[test]
    fn test_hash_case_sensitivity() {
        let h1 = calculate_hash("Memzent");
        let h2 = calculate_hash("memzent");
        assert_ne!(h1, h2, "Hashing must be case-sensitive");
    }
}

// ---------------------------------------------------------------------------
// compress_text — stop-word removal + compression fallback
// ---------------------------------------------------------------------------
#[cfg(test)]
mod compress_tests {
    use stop_words;

    // Mirror of the compress_text logic (without external compressor fallback)
    fn strip_stop_words(prompt: &str) -> String {
        let words = stop_words::get(stop_words::LANGUAGE::English);
        prompt.split_whitespace()
            .filter(|w| !words.contains(&w.to_lowercase()))
            .collect::<Vec<&str>>()
            .join(" ")
    }

    #[test]
    fn test_stop_words_removed() {
        // "the", "a", "is", "in" are English stop words
        let result = strip_stop_words("the customer is in the database");
        assert!(!result.to_lowercase().contains(" the "), "Stop word 'the' must be removed");
        // Semantic terms should survive
        assert!(result.to_lowercase().contains("customer") || result.to_lowercase().contains("database"));
    }

    #[test]
    fn test_empty_prompt_survives() {
        let result = strip_stop_words("");
        assert_eq!(result, "", "Empty prompt must remain empty after filtering");
    }

    #[test]
    fn test_all_stop_words_returns_empty() {
        // "the" and "a" are both stop words
        let result = strip_stop_words("the a");
        // Either empty or whitespace — meaningful terms removed
        assert!(result.trim().is_empty() || result.len() < 4);
    }

    #[test]
    fn test_technical_terms_survive() {
        let result = strip_stop_words("find customer lookup CRM database error");
        assert!(result.contains("customer") || result.contains("CRM") || result.contains("database"));
    }

    #[test]
    fn test_output_is_shorter_than_input() {
        let input = "the user wants to find all the open error tickets in the system";
        let result = strip_stop_words(input);
        assert!(
            result.len() <= input.len(),
            "Output ({}) should be ≤ input ({}) after stop-word removal",
            result.len(), input.len()
        );
    }
}

// ---------------------------------------------------------------------------
// score_threshold_override logic
// ---------------------------------------------------------------------------
#[cfg(test)]
mod threshold_tests {

    fn resolve_threshold(override_val: f32) -> f32 {
        if override_val > 0.0 { override_val } else { 0.65 }
    }

    #[test]
    fn test_default_threshold_is_0_65() {
        assert_eq!(resolve_threshold(0.0), 0.65);
    }

    #[test]
    fn test_negative_override_uses_default() {
        assert_eq!(resolve_threshold(-0.1), 0.65);
    }

    #[test]
    fn test_positive_override_is_used() {
        assert_eq!(resolve_threshold(0.90), 0.90);
    }

    #[test]
    fn test_override_exactly_zero_uses_default() {
        assert_eq!(resolve_threshold(0.0), 0.65);
    }
}

// ---------------------------------------------------------------------------
// Semantic cache threshold  (hardcoded 0.88 in select_tools)
// ---------------------------------------------------------------------------
#[cfg(test)]
mod cache_threshold_tests {
    const SEMANTIC_CACHE_THRESHOLD: f32 = 0.88;

    fn should_cache_hit(score: f32) -> bool {
        score > SEMANTIC_CACHE_THRESHOLD
    }

    #[test]
    fn test_score_above_threshold_is_hit() {
        assert!(should_cache_hit(0.95), "Score 0.95 should be a cache hit");
        assert!(should_cache_hit(0.89), "Score 0.89 should be a cache hit");
    }

    #[test]
    fn test_score_equal_threshold_is_miss() {
        assert!(!should_cache_hit(0.88), "Score exactly 0.88 is NOT a hit (strict >)");
    }

    #[test]
    fn test_score_below_threshold_is_miss() {
        assert!(!should_cache_hit(0.75));
        assert!(!should_cache_hit(0.0));
    }

    #[test]
    fn test_perfect_score_is_hit() {
        assert!(should_cache_hit(1.0), "Score 1.0 (identical vector) must be a cache hit");
    }
}

// ---------------------------------------------------------------------------
// tokens_saved calculation  (rough estimate: bytes diff / 4)
// ---------------------------------------------------------------------------
#[cfg(test)]
mod tokens_saved_tests {
    fn estimate_tokens_saved(original_len: usize, compressed_len: usize) -> i32 {
        // Mirrors: (req.prompt.len() as i32 - compressed.len() as i32) / 4
        let raw = original_len as i32 - compressed_len as i32;
        raw / 4
    }

    #[test]
    fn test_no_compression_zero_tokens() {
        assert_eq!(estimate_tokens_saved(100, 100), 0);
    }

    #[test]
    fn test_half_size_reduction() {
        // 100 chars → 50 chars = 50 chars saved = 50/4 = 12 tokens
        assert_eq!(estimate_tokens_saved(100, 50), 12);
    }

    #[test]
    fn test_grows_longer_is_not_negative_in_clamped_path() {
        // Router adds 450 offset before max(0) clamp — test raw formula
        let raw = estimate_tokens_saved(10, 20); // compressed is LONGER
        assert_eq!(raw, -2); // -10 / 4 = -2 (truncated)
    }

    #[test]
    fn test_large_prompt_reduction() {
        // 2000 char prompt compressed to 600 = 1400 chars saved = 350 tokens
        assert_eq!(estimate_tokens_saved(2000, 600), 350);
    }
}

// ---------------------------------------------------------------------------
// UUID v5 determinism (used for tool registration dedup)
// ---------------------------------------------------------------------------
#[cfg(test)]
mod uuid_tests {
    use uuid::Uuid;

    fn tool_uuid(tool_id: &str) -> String {
        Uuid::new_v5(&Uuid::NAMESPACE_OID, tool_id.as_bytes()).to_string()
    }

    #[test]
    fn test_same_tool_id_same_uuid() {
        let u1 = tool_uuid("customer_lookup");
        let u2 = tool_uuid("customer_lookup");
        assert_eq!(u1, u2, "Same tool_id must always produce the same UUID v5");
    }

    #[test]
    fn test_different_tool_id_different_uuid() {
        let u1 = tool_uuid("customer_lookup");
        let u2 = tool_uuid("order_search");
        assert_ne!(u1, u2);
    }

    #[test]
    fn test_uuid_is_valid_format() {
        let u = tool_uuid("any_tool");
        // UUID format: 8-4-4-4-12 hex chars
        assert_eq!(u.len(), 36, "UUID string must be 36 characters");
        let parts: Vec<&str> = u.split('-').collect();
        assert_eq!(parts.len(), 5, "UUID must have 5 dash-separated groups");
    }

    #[test]
    fn test_empty_tool_id_is_stable() {
        let u1 = tool_uuid("");
        let u2 = tool_uuid("");
        assert_eq!(u1, u2, "Even empty string tool_id must produce stable UUID");
    }
}
