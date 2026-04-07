package billing

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeHandler struct {
	db            *sql.DB
	webhookSecret string
	proProductID  string
	bizProductID  string
}

func NewStripeHandler(db *sql.DB, secret, proID, bizID string) *StripeHandler {
	return &StripeHandler{
		db:            db,
		webhookSecret: secret,
		proProductID:  proID,
		bizProductID:  bizID,
	}
}

func (h *StripeHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading Stripe webhook body", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		slog.Error("Error verifying Stripe webhook signature", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			slog.Error("Error unmarshaling Stripe session", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.handleCheckoutCompleted(&session)

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &sub)
		if err != nil {
			slog.Error("Error unmarshaling Stripe subscription", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.handleSubscriptionDeleted(&sub)

	default:
		slog.Info("Unhandled Stripe event type", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *StripeHandler) handleCheckoutCompleted(session *stripe.CheckoutSession) {
	customerID := session.Customer.ID
	// In a real app, you'd map the Stripe Customer ID to an Organization ID
	// We'll update the organization's tier based on the product purchased.

	slog.Info("Stripe Checkout Completed", "customer_id", customerID, "status", session.Status)

	// Mock Logic: Update DB
	_, err := h.db.Exec("UPDATE organizations SET subscription_tier = 'pro' WHERE stripe_customer_id = $1", customerID)
	if err != nil {
		slog.Error("Failed to update organization tier on checkout", "customer_id", customerID, "error", err)
	}
}

func (h *StripeHandler) handleSubscriptionDeleted(sub *stripe.Subscription) {
	customerID := sub.Customer.ID
	slog.Info("Stripe Subscription Deleted", "customer_id", customerID)

	// Downgrade to free
	_, err := h.db.Exec("UPDATE organizations SET subscription_tier = 'free' WHERE stripe_customer_id = $1", customerID)
	if err != nil {
		slog.Error("Failed to downgrade organization tier on cancellation", "customer_id", customerID, "error", err)
	}
}
