package billing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripeHandler_CreateCheckoutSession_BadMethod(t *testing.T) {
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	req := httptest.NewRequest(http.MethodGet, "/checkout", nil)
	rr := httptest.NewRecorder()

	handler.CreateCheckoutSession(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestStripeHandler_CreateCheckoutSession_BadPayload(t *testing.T) {
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBuffer([]byte("invalid json")))
	rr := httptest.NewRecorder()

	handler.CreateCheckoutSession(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStripeHandler_CreateCheckoutSession_MinimumTopup(t *testing.T) {
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	payload := map[string]interface{}{
		"amount": 3.5, // Less than minimum $5.00
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.CreateCheckoutSession(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStripeHandler_CreateCheckoutSession_MissingPriceID(t *testing.T) {
	// Don't set STRIPE_PRO_PRICE_ID
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	payload := map[string]interface{}{
		"tier": "pro",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.CreateCheckoutSession(rr, req)

	// Since STRIPE_PRO_PRICE_ID is empty, it returns 503
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestStripeHandler_HandleWebhook_InvalidBody(t *testing.T) {
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	req := httptest.NewRequest(http.MethodPost, "/webhook", errReader(0))
	rr := httptest.NewRecorder()

	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestStripeHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	handler := NewStripeHandler(nil, nil, "secret", "pro", "biz", nil)
	
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Stripe-Signature", "invalid-signature")
	rr := httptest.NewRecorder()

	handler.HandleWebhook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}
