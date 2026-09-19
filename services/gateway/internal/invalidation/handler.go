// services/gateway/internal/invalidation/handler.go
package invalidation

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// HandleInvalidate returns an HTTP handler for POST /v1/cache/invalidate.
//
// It accepts an InvalidationEvent. The org is always taken from the
// authenticated request context (never trusted from the body) to preserve
// tenant isolation. Body example:
//
//	{ "change_type": "tool_data_changed", "tool_ids": ["github-repo"], "reason": "push" }
func HandleInvalidate(inv *Invalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		orgID, _ := r.Context().Value("org_id").(string)
		if orgID == "" {
			http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
			return
		}
		if inv == nil {
			http.Error(w, `{"error":"invalidation not enabled"}`, http.StatusServiceUnavailable)
			return
		}

		var ev InvalidationEvent
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		ev.OrgID = orgID // enforce tenant scope from auth context
		if ev.Timestamp.IsZero() {
			ev.Timestamp = time.Now()
		}
		if ev.ChangeType == ChangeToolData && len(ev.ToolIDs) == 0 {
			http.Error(w, `{"error":"tool_ids required for tool_data_changed"}`, http.StatusBadRequest)
			return
		}

		res, err := inv.HandleEvent(r.Context(), ev)
		if err != nil {
			slog.Warn("cache invalidation failed", "org_id", orgID, "change_type", ev.ChangeType, "error", err)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		slog.Info("🧹 Cache invalidation event",
			"org_id", orgID, "change_type", ev.ChangeType,
			"keys_deleted", res.KeysDeleted, "version_bumped", res.VersionBumped)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":        true,
			"change_type":    res.ChangeType,
			"keys_deleted":   res.KeysDeleted,
			"version_bumped": res.VersionBumped,
			"new_version":    res.NewVersion,
			"org_id":         orgID,
		})
	}
}
