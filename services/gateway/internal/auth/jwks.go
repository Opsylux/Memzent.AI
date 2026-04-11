package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// JWKS holds the JSON Web Key Set retrieved from an IdP
type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"` // RSA modulus
	E   string   `json:"e"` // RSA exponent
	X   string   `json:"x"` // EC x coordinate
	Y   string   `json:"y"` // EC y coordinate
	Crv string   `json:"crv"` // EC curve
	X5c []string `json:"x5c"` // X.509 certificate chain
}

// JWKSProvider fetches and caches JWKS keys
type JWKSProvider struct {
	url     string
	apiKey  string
	mu      sync.RWMutex
	keys    map[string]interface{}
	lastMod time.Time
}

func NewJWKSProvider(url, apiKey string) *JWKSProvider {
	return &JWKSProvider{
		url:    url,
		apiKey: apiKey,
		keys:   make(map[string]interface{}),
	}
}

// SeedKey pre-loads a known public key into the cache.
// This is the primary fallback when the JWKS endpoint is unreachable (e.g.
// returns 401 due to network policy).  Call this from main.go with the key
// material obtained from the Supabase dashboard.
func (p *JWKSProvider) SeedKey(kid string, key interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.keys[kid] = key
	slog.Info("JWKS: static key seeded", "kid", kid)
}

// GetKey returns the public key for a given kid (Key ID)
func (p *JWKSProvider) GetKey(kid string) (interface{}, error) {
	p.mu.RLock()
	key, ok := p.keys[kid]
	p.mu.RUnlock()

	if ok {
		return key, nil
	}

	// Not found in cache – try a live fetch
	if err := p.Refresh(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	key, ok = p.keys[kid]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key with kid %s not found in JWKS from %s", kid, p.url)
	}

	return key, nil
}

// Refresh fetches the latest keys from the JWKS URL
func (p *JWKSProvider) Refresh() error {
	p.mu.Lock()
	// Rate limit refreshes (max once every 5 seconds)
	if time.Since(p.lastMod) < 5*time.Second {
		p.mu.Unlock()
		return nil
	}
	// Set lastMod immediately to prevent concurrent refresh attempts
	// but we don't return an error here so others can still use the cache
	p.lastMod = time.Now()
	url := p.url
	apiKey := p.apiKey
	p.mu.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("apikey", apiKey)
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("JWKS: remote fetch failed, using cached/seeded keys",
			"status", resp.StatusCode, "url", url)
		return nil
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	newKeys := make(map[string]interface{})
	for _, jwk := range jwks.Keys {
		var key interface{}
		var parseErr error

		switch jwk.Kty {
		case "RSA":
			key, parseErr = parseRSAKey(jwk)
		case "EC":
			key, parseErr = parseECKey(jwk)
		default:
			continue
		}

		if parseErr == nil {
			newKeys[jwk.Kid] = key
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for kid, key := range newKeys {
		if _, exists := p.keys[kid]; !exists {
			slog.Info("JWKS: new key loaded", "kid", kid)
		}
		p.keys[kid] = key
	}

	return nil
}

func parseRSAKey(jwk JSONWebKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	var e int
	for _, b := range eBytes {
		e <<= 8
		e |= int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func parseECKey(jwk JSONWebKey) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, err
	}

	var curve elliptic.Curve
	switch jwk.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", jwk.Crv)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

// ParseECJWKLiteral parses an inline JWK JSON object into an *ecdsa.PublicKey.
// Use this to seed a known key at startup without needing the JWKS endpoint.
func ParseECJWKLiteral(jwkJSON string) (string, *ecdsa.PublicKey, error) {
	var jwk JSONWebKey
	if err := json.Unmarshal([]byte(jwkJSON), &jwk); err != nil {
		return "", nil, fmt.Errorf("failed to parse JWK JSON: %w", err)
	}
	if jwk.Kty != "EC" {
		return "", nil, fmt.Errorf("expected EC key, got %s", jwk.Kty)
	}
	key, err := parseECKey(jwk)
	if err != nil {
		return "", nil, err
	}
	return jwk.Kid, key, nil
}
