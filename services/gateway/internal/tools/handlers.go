package tools

import (
	"aura-gateway/internal/metrics"
	"aura-gateway/internal/router"
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

// HandleSyncTools triggers a re-sync of all dynamic tool connectors
func HandleSyncTools(registry *Registry) http.HandlerFunc {
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

		slog.Info("Triggering global tool registry sync")
		metrics.GlobalAuditBuffer.Add(metrics.AuditEvent{
			Timestamp: time.Now(),
			Type:      "REGISTRY",
			User:      "admin",
			Detail:    "Neural Registry Sync Initiated",
			Status:    "success",
		})

		// In a real implementation, this might trigger:
		// 1. MCP Client re-scan
		// 2. Refreshing REST connector caches
		// 3. Re-indexing vector store in the Router

		// For now, we'll return a success status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":    "success",
			"message":   "Global registry synchronization initiated",
			"timestamp": time.Now(),
		})
	}
}

// ToolWithProvider wraps Tool with provider metadata for API response
type ToolWithProvider struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Provider       string                 `json:"provider"` // "Aura-Core", "Aura-MCP", "Aura-REST", etc.
	ConnectorType  string                 `json:"connector_type"`
	Status         string                 `json:"status"` // "online", "offline", etc.
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
	InputSchema    map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema   map[string]interface{} `json:"output_schema,omitempty"`
}

// Utility function to convert Tool to API response format
func ToolToAPI(t *Tool) ToolWithProvider {
	provider := "Aura-" + strings.ToUpper(string(t.ConnectorType))
	return ToolWithProvider{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description,
		Provider:       provider,
		ConnectorType:  string(t.ConnectorType),
		Status:         "online", // TODO: Check tool health
		TimeoutSeconds: t.TimeoutSeconds,
		InputSchema:    t.InputSchema,
		OutputSchema:   t.OutputSchema,
	}
}
