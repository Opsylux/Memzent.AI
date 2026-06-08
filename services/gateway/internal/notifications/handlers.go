package notifications

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// HandleWebhooks handles CRUD on /v1/webhooks (list + create)
func HandleWebhooks(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		if orgID == "" {
			http.Error(w, "Unauthorized: org context required", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			webhooks, err := registry.List(r.Context(), orgID)
			if err != nil {
				slog.Error("[Webhooks] List failed", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(webhooks)

		case http.MethodPost:
			var req struct {
				URL         string   `json:"url"`
				Events      []string `json:"events"`
				Description string   `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			if req.URL == "" {
				http.Error(w, "Missing required field: url", http.StatusBadRequest)
				return
			}
			if len(req.Events) == 0 {
				http.Error(w, "Missing required field: events (at least one event type)", http.StatusBadRequest)
				return
			}

			// Validate event types
			for _, ev := range req.Events {
				if !isValidEventType(ev) {
					http.Error(w, "Invalid event type: "+ev, http.StatusBadRequest)
					return
				}
			}

			// Generate signing secret
			secret := generateSecret()

			wh := &Webhook{
				OrgID:       orgID,
				URL:         req.URL,
				Secret:      secret,
				Events:      req.Events,
				Enabled:     true,
				Description: req.Description,
			}

			if err := registry.Create(r.Context(), wh); err != nil {
				slog.Error("[Webhooks] Create failed", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			slog.Info("[Webhooks] Created", "id", wh.ID, "org", orgID, "url", wh.URL, "events", wh.Events)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(wh) // Includes secret on creation only

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

// HandleWebhookByID handles GET/PUT/DELETE on /v1/webhooks/{id}
func HandleWebhookByID(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, _ := r.Context().Value("org_id").(string)
		if orgID == "" {
			http.Error(w, "Unauthorized: org context required", http.StatusUnauthorized)
			return
		}

		webhookID := strings.TrimPrefix(r.URL.Path, "/v1/webhooks/")
		if webhookID == "" || webhookID == "/" {
			http.Error(w, "Bad Request: webhook ID required", http.StatusBadRequest)
			return
		}

		// Strip /deliveries suffix for delivery logs sub-route
		if strings.HasSuffix(webhookID, "/deliveries") {
			webhookID = strings.TrimSuffix(webhookID, "/deliveries")
			handleDeliveryLogs(registry, orgID, webhookID, w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			wh, err := registry.Get(r.Context(), orgID, webhookID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if wh == nil {
				http.Error(w, "Webhook not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(wh)

		case http.MethodPut:
			var req struct {
				URL         *string  `json:"url"`
				Events      []string `json:"events"`
				Enabled     *bool    `json:"enabled"`
				Description *string  `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			existing, err := registry.Get(r.Context(), orgID, webhookID)
			if err != nil || existing == nil {
				http.Error(w, "Webhook not found", http.StatusNotFound)
				return
			}

			if req.URL != nil {
				existing.URL = *req.URL
			}
			if req.Events != nil {
				for _, ev := range req.Events {
					if !isValidEventType(ev) {
						http.Error(w, "Invalid event type: "+ev, http.StatusBadRequest)
						return
					}
				}
				existing.Events = req.Events
			}
			if req.Enabled != nil {
				existing.Enabled = *req.Enabled
			}
			if req.Description != nil {
				existing.Description = *req.Description
			}

			if err := registry.Update(r.Context(), existing); err != nil {
				slog.Error("[Webhooks] Update failed", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(existing)

		case http.MethodDelete:
			if err := registry.Delete(r.Context(), orgID, webhookID); err != nil {
				slog.Error("[Webhooks] Delete failed", "error", err)
				http.Error(w, "Webhook not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

// HandleEventTypes returns the list of subscribable event types
func HandleEventTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"event_types": AllEventTypes,
		})
	}
}

func handleDeliveryLogs(registry *Registry, orgID, webhookID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := registry.GetDeliveryLogs(r.Context(), orgID, webhookID, 50)
	if err != nil {
		slog.Error("[Webhooks] Get delivery logs failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func isValidEventType(t string) bool {
	for _, valid := range AllEventTypes {
		if t == valid {
			return true
		}
	}
	return false
}

func generateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "whsec_" + hex.EncodeToString(b)
}
