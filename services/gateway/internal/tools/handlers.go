package tools

import (
	"memzent-gateway/internal/metrics"
	"memzent-gateway/internal/router"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RegisterRequest is the payload for registering a new tool
type RegisterRequest struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	ConnectorType  string                 `json:"connector_type"` // "mcp", "rest", "sql", "graphql", etc.
	Endpoint       string                 `json:"endpoint"`       // URL, connection string, or tool name
	InputSchema    map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema   map[string]interface{} `json:"output_schema,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
	RequiresAuth   bool                   `json:"requires_auth,omitempty"`
}

// HandleRegisterTool registers a new tool (admin-only) and notifies the semantic router.
func HandleRegisterTool(registry *Registry, routerClient *router.RouterClient, auditLogger *metrics.PersistentAuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Allow verified 'admin' (from DB) or global 'platform_staff'
		userRole, ok := r.Context().Value("user_role").(string)
		isAdmin := ok && (userRole == "admin" || userRole == "platform_staff")
		if !isAdmin {
			http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validation
		if req.ID == "" || req.Name == "" || req.Endpoint == "" {
			http.Error(w, "Missing required fields: id, name, endpoint", http.StatusBadRequest)
			return
		}

		if req.ConnectorType == "" {
			req.ConnectorType = "mcp" // Default to MCP for backward compatibility
		}

		if req.TimeoutSeconds == 0 {
			req.TimeoutSeconds = 15 // Default timeout
		}

		tool := &Tool{
			ID:             req.ID,
			Name:           req.Name,
			Description:    req.Description,
			ConnectorType:  ToolConnectorType(req.ConnectorType),
			Endpoint:       req.Endpoint,
			InputSchema:    req.InputSchema,
			OutputSchema:   req.OutputSchema,
			TimeoutSeconds: req.TimeoutSeconds,
			Enabled:        true,
			RequiresAuth:   req.RequiresAuth,
		}

		if err := registry.RegisterTool(r.Context(), tool); err != nil {
			slog.Error("Failed to register tool", "error", err, "tool_id", req.ID)
			if auditLogger != nil {
				auditLogger.Log(r.Context(), metrics.AuditEvent{
					Timestamp: time.Now(),
					OrgID:     "system", // Generic org if failed
					Type:      "ERROR",
					Detail:    fmt.Sprintf("Tool Reg Fail: %s", req.ID),
					Status:    "error",
				}, map[string]interface{}{"error": err.Error(), "tool_id": req.ID})
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 2. Notify Semantic Router (Vectorization)
		if routerClient != nil {
			orgID, _ := r.Context().Value("org_id").(string)
			go func() {
				_, err := routerClient.RegisterTool(context.Background(), tool.ID, tool.Name, tool.Description, orgID)
				if err != nil {
					slog.Error("Failed to vectorize tool in router", "tool_id", tool.ID, "error", err)
				}
			}()
		}

		slog.Info("Tool registered", "id", tool.ID, "name", tool.Name, "connector_type", tool.ConnectorType)
		if auditLogger != nil {
			auditLogger.Log(r.Context(), metrics.AuditEvent{
				Timestamp: time.Now(),
				OrgID:     "system", // Placeholder for auditing the registry action
				Type:      "REGISTRY",
				User:      userRole,
				Detail:    fmt.Sprintf("New Node Integrated: %s", tool.Name),
				Status:    "success",
			}, map[string]interface{}{"tool_id": tool.ID, "name": tool.Name})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tool)
	}
}

// HandleUpdateTool updates an existing tool's configuration (admin-only)
func HandleUpdateTool(registry *Registry, routerClient *router.RouterClient, auditLogger *metrics.PersistentAuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userRole, ok := r.Context().Value("user_role").(string)
		isAdmin := ok && (userRole == "admin" || userRole == "platform_staff")
		if !isAdmin {
			http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
			return
		}

		// Extract tool ID from URL path /v1/tools/{toolId}
		toolID := strings.TrimPrefix(r.URL.Path, "/v1/tools/")
		if toolID == "" {
			http.Error(w, "Bad Request: tool ID required", http.StatusBadRequest)
			return
		}

		// Get existing tool first
		existing, err := registry.GetTool(r.Context(), toolID)
		if err != nil || existing == nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		// Decode partial update (only overwrite fields that are present)
		var req struct {
			Name           *string                `json:"name"`
			Description    *string                `json:"description"`
			ConnectorType  *string                `json:"connector_type"`
			Endpoint       *string                `json:"endpoint"`
			Config         map[string]interface{} `json:"config"`
			InputSchema    map[string]interface{} `json:"input_schema"`
			OutputSchema   map[string]interface{} `json:"output_schema"`
			TimeoutSeconds *int                   `json:"timeout_seconds"`
			Enabled        *bool                  `json:"enabled"`
			RequiresAuth   *bool                  `json:"requires_auth"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Apply partial updates
		if req.Name != nil {
			existing.Name = *req.Name
		}
		if req.Description != nil {
			existing.Description = *req.Description
		}
		if req.ConnectorType != nil {
			existing.ConnectorType = ToolConnectorType(*req.ConnectorType)
		}
		if req.Endpoint != nil {
			existing.Endpoint = *req.Endpoint
		}
		if req.Config != nil {
			existing.Config = req.Config
		}
		if req.InputSchema != nil {
			existing.InputSchema = req.InputSchema
		}
		if req.OutputSchema != nil {
			existing.OutputSchema = req.OutputSchema
		}
		if req.TimeoutSeconds != nil {
			existing.TimeoutSeconds = *req.TimeoutSeconds
		}
		if req.Enabled != nil {
			existing.Enabled = *req.Enabled
		}
		if req.RequiresAuth != nil {
			existing.RequiresAuth = *req.RequiresAuth
		}

		if err := registry.UpdateTool(r.Context(), existing); err != nil {
			slog.Error("Failed to update tool", "error", err, "tool_id", toolID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Re-vectorize if description changed
		if req.Description != nil && routerClient != nil {
			orgID, _ := r.Context().Value("org_id").(string)
			go func() {
				_, err := routerClient.RegisterTool(context.Background(), existing.ID, existing.Name, existing.Description, orgID)
				if err != nil {
					slog.Error("Failed to re-vectorize tool after update", "tool_id", existing.ID, "error", err)
				}
			}()
		}

		slog.Info("Tool updated", "id", toolID, "name", existing.Name)
		if auditLogger != nil {
			auditLogger.Log(r.Context(), metrics.AuditEvent{
				Timestamp: time.Now(),
				OrgID:     "system",
				Type:      "REGISTRY",
				User:      userRole,
				Detail:    fmt.Sprintf("Tool Updated: %s", existing.Name),
				Status:    "success",
			}, map[string]interface{}{"tool_id": toolID})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
	}
}

// HandleGetTool returns a single tool by ID
func HandleGetTool(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		toolID := strings.TrimPrefix(r.URL.Path, "/v1/tools/")
		if toolID == "" {
			http.Error(w, "Bad Request: tool ID required", http.StatusBadRequest)
			return
		}

		tool, err := registry.GetTool(r.Context(), toolID)
		if err != nil {
			slog.Error("Failed to get tool", "error", err, "tool_id", toolID)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if tool == nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tool)
	}
}

// HandleDisableTool soft-deletes a tool (admin-only)
func HandleDisableTool(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Allow verified 'admin' (from DB) or global 'platform_staff'
		userRole, ok := r.Context().Value("user_role").(string)
		isAdmin := ok && (userRole == "admin" || userRole == "platform_staff")
		if !isAdmin {
			http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
			return
		}

		// Extract tool ID from URL path /v1/tools/{toolId}
		toolID := strings.TrimPrefix(r.URL.Path, "/v1/tools/")

		if toolID == "" || toolID == "/v1/tools" {
			http.Error(w, "Bad Request: tool ID required", http.StatusBadRequest)
			return
		}

		if err := registry.DisableTool(r.Context(), toolID); err != nil {
			slog.Error("Failed to disable tool", "error", err, "tool_id", toolID)
			metrics.GlobalAuditBuffer.Add(metrics.AuditEvent{
				Timestamp: time.Now(),
				Type:      "ERROR",
				Detail:    fmt.Sprintf("Dissolution Fail: %s", toolID),
				Status:    "error",
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Tool disabled", "id", toolID)
		metrics.GlobalAuditBuffer.Add(metrics.AuditEvent{
			Timestamp: time.Now(),
			Type:      "REGISTRY",
			User:      "admin",
			Detail:    fmt.Sprintf("Infrastructure Node Dissolved: %s", toolID),
			Status:    "warning",
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "tool_id": toolID})
	}
}

// HandleSyncTools triggers a real re-sync: polls Postgres for drifted tools and
// pushes each one to the Rust Router for vectorization in Qdrant.
// Admin-only. Returns a summary of what was synced.
func HandleSyncTools(registry *Registry, routerClient *router.RouterClient, auditLogger *metrics.PersistentAuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		userRole, ok := r.Context().Value("user_role").(string)
		isAdmin := ok && (userRole == "admin" || userRole == "platform_staff")
		if !isAdmin {
			http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
			return
		}

		if registry == nil {
			http.Error(w, "Tool registry unavailable", http.StatusServiceUnavailable)
			return
		}

		var syncedIDs []string
		var syncErrors []string

		// Build the sync callback that pushes each drifted tool to Qdrant via gRPC
		onSync := func(ctx context.Context, tools []*Tool) {
			for _, t := range tools {
				orgID := ""
				if t.OrgID != nil {
					orgID = *t.OrgID
				}
				if routerClient != nil {
					ok, err := routerClient.RegisterTool(ctx, t.ID, t.Name, t.Description, orgID)
					if err != nil || !ok {
						errMsg := fmt.Sprintf("%s: gRPC error", t.ID)
						if err != nil {
							errMsg = fmt.Sprintf("%s: %s", t.ID, err.Error())
						}
						syncErrors = append(syncErrors, errMsg)
						slog.Error("[ToolSync] Vectorization failed", "tool_id", t.ID, "error", err)
						continue
					}
				}
				syncedIDs = append(syncedIDs, t.ID)
				slog.Info("[ToolSync] Tool vectorized", "tool_id", t.ID, "name", t.Name)
			}
		}

		n, err := registry.Refresh(r.Context(), onSync)
		if err != nil {
			slog.Error("[ToolSync] Registry refresh failed", "error", err)
			http.Error(w, "Registry refresh failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("[ToolSync] Manual sync complete", "synced", n)

		if auditLogger != nil {
			auditLogger.Log(r.Context(), metrics.AuditEvent{
				Timestamp: time.Now(),
				Type:      "REGISTRY",
				User:      userRole,
				Detail:    fmt.Sprintf("Manual Qdrant Sync: %d tools vectorized", n),
				Status:    "success",
			}, map[string]interface{}{"synced_count": n, "synced_ids": syncedIDs})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":         "success",
			"tools_synced":   n,
			"synced_ids":     syncedIDs,
			"errors":         syncErrors,
			"last_refresh":   registry.LastRefreshTime(),
			"timestamp":      time.Now(),
		})
	}
}

// HandleRegistryStatus returns the current state of the Tool Registry sync.
func HandleRegistryStatus(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if registry == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":       "healthy",
			"last_refresh": registry.LastRefreshTime(),
		})
	}
}

// ToolWithProvider wraps Tool with provider metadata for API response
type ToolWithProvider struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Provider       string                 `json:"provider"` // "Memzent-Core", "Memzent-MCP", "Memzent-REST", etc.
	ConnectorType  string                 `json:"connector_type"`
	Status         string                 `json:"status"` // "online", "offline", etc.
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
	InputSchema    map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema   map[string]interface{} `json:"output_schema,omitempty"`
}

// Utility function to convert Tool to API response format
func ToolToAPI(t *Tool) ToolWithProvider {
	provider := "Memzent-" + strings.ToUpper(string(t.ConnectorType))
	return ToolWithProvider{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description,
		Provider:       provider,
		ConnectorType:  string(t.ConnectorType),
		Status:         "online", // Assumed online; real health probing deferred to tool registry refresh cycle
		TimeoutSeconds: t.TimeoutSeconds,
		InputSchema:    t.InputSchema,
		OutputSchema:   t.OutputSchema,
	}
}
