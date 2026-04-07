package tools

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
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

// HandleRegisterTool registers a new tool (admin-only)
func HandleRegisterTool(registry *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if user is admin (could also check scopes from JWT)
		userRole, ok := r.Context().Value("user_role").(string)
		if !ok || userRole != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Tool registered", "id", tool.ID, "name", tool.Name, "connector_type", tool.ConnectorType)

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

		// Check admin role
		userRole, ok := r.Context().Value("user_role").(string)
		if !ok || userRole != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		slog.Info("Tool disabled", "id", toolID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "disabled", "tool_id": toolID})
	}
}

// ToolWithProvider wraps Tool with provider metadata for API response
type ToolWithProvider struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Provider        string                 `json:"provider"` // "Aura-Core", "Aura-MCP", "Aura-REST", etc.
	ConnectorType   string                 `json:"connector_type"`
	Status          string                 `json:"status"` // "online", "offline", etc.
	TimeoutSeconds  int                    `json:"timeout_seconds,omitempty"`
	InputSchema     map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema    map[string]interface{} `json:"output_schema,omitempty"`
}

// Utility function to convert Tool to API response format
func ToolToAPI(t *Tool) ToolWithProvider {
	provider := "Aura-" + strings.ToUpper(string(t.ConnectorType))
	return ToolWithProvider{
		ID:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		Provider:      provider,
		ConnectorType: string(t.ConnectorType),
		Status:        "online", // TODO: Check tool health
		TimeoutSeconds: t.TimeoutSeconds,
		InputSchema:   t.InputSchema,
		OutputSchema:  t.OutputSchema,
	}
}
