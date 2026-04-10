package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// GetKey returns the public key for a given kid (Key ID)
func (p *JWKSProvider) GetKey(kid string) (interface{}, error) {
	p.mu.RLock()
	key, ok := p.keys[kid]
	p.mu.RUnlock()

	if ok {
		return key, nil
	}

	// Not found or cache expired, fetch new ones
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
	defer p.mu.Unlock()

	// Rate limit refreshes (max once every 5 seconds)
	if time.Since(p.lastMod) < 5*time.Second {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", p.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	// Supabase Cloud requires apikey header even for discovery
	if p.apiKey != "" {
		req.Header.Set("apikey", p.apiKey)
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS, status: %d from %s", resp.StatusCode, p.url)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS from %s: %w", p.url, err)
	}

	for _, jwk := range jwks.Keys {
		var key interface{}
		var err error

		switch jwk.Kty {
		case "RSA":
			key, err = parseRSAKey(jwk)
		case "EC":
			key, err = parseECKey(jwk)
		default:
			continue // Skip unknown key types
		}

		if err == nil {
			p.keys[jwk.Kid] = key
		}
	}

	p.lastMod = time.Now()
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
