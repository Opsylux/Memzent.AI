package billing

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

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
		slog.Info("Stripe Checkout Completed", "customer_id", session.Customer.ID, "status", session.Status)

	case "customer.subscription.created", "customer.subscription.updated":
		var sub stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &sub)
		if err != nil {
			slog.Error("Error unmarshaling Stripe subscription", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.handleSubscriptionChanged(&sub)

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

func (h *StripeHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tier string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	priceID := ""
	switch req.Tier {
	case "pro":
		priceID = os.Getenv("STRIPE_PRO_PRICE_ID")
	case "business":
		priceID = os.Getenv("STRIPE_BIZ_PRICE_ID")
	default:
		http.Error(w, "Invalid Tier", http.StatusBadRequest)
		return
	}

	orgID := r.Header.Get("X-Org-ID")
	// In a real app, you'd create a Checkout Session here using the Stripe SDK
	// For this demo, we'll return a mock URL that Stripe would provide
	
	slog.Info("Creating Stripe Checkout Session", "tier", req.Tier, "org_id", orgID, "price_id", priceID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": "https://checkout.stripe.com/pay/mock_session_" + req.Tier,
	})
}

func (h *StripeHandler) handleSubscriptionChanged(sub *stripe.Subscription) {
	customerID := sub.Customer.ID
	if len(sub.Items.Data) == 0 {
		return
	}

	// Resolve the Product ID from the subscription item
	productID := sub.Items.Data[0].Price.Product.ID
	tier := "free"

	if productID == h.proProductID {
		tier = "pro"
	} else if productID == h.bizProductID {
		tier = "business"
	}

	slog.Info("Stripe Subscription Created/Updated", "customer_id", customerID, "product_id", productID, "mapped_tier", tier)

	_, err := h.db.Exec("UPDATE organizations SET subscription_tier = $1 WHERE stripe_customer_id = $2", tier, customerID)
	if err != nil {
		slog.Error("Failed to update organization tier on subscription change", "customer_id", customerID, "tier", tier, "error", err)
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
