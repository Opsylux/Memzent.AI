package billing

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
	"aura-gateway/internal/metrics"
	"context"
	"fmt"
	"time"
)
type StripeHandler struct {
	db            *sql.DB
	webhookSecret string
	proProductID  string
	bizProductID  string
	auditLogger   *metrics.PersistentAuditLogger
}

func NewStripeHandler(db *sql.DB, secret, proID, bizID string, audit *metrics.PersistentAuditLogger) *StripeHandler {
	// Configure Stripe key globally for this handler
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	
	return &StripeHandler{
		db:            db,
		webhookSecret: secret,
		proProductID:  proID,
		bizProductID:  bizID,
		auditLogger:   audit,
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
	slog.Info("Creating real Stripe Checkout Session", "tier", req.Tier, "org_id", orgID, "price_id", priceID)

	successURL := os.Getenv("STRIPE_SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:3000/dashboard/billing?session_id={CHECKOUT_SESSION_ID}"
	}
	cancelURL := os.Getenv("STRIPE_CANCEL_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:3000/dashboard/billing?status=cancel"
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"org_id": orgID,
			"tier":   req.Tier,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		slog.Error("Stripe Checkout Session creation failed", "error", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": sess.URL,
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

	// TODO: Persist tier change to database using customerID mapping

	if h.auditLogger != nil {
		h.auditLogger.Log(context.Background(), metrics.AuditEvent{
			Timestamp: time.Now(),
			OrgID:     "system", // customerID mapping needed for better org scoping
			Type:      "BILLING",
			Detail:    fmt.Sprintf("Subscription Updated: %s", tier),
			Status:    "success",
		}, map[string]interface{}{"tier": tier, "customer_id": customerID})
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
