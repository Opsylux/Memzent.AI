package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

// SQLConnector executes tools via direct SQL queries
type SQLConnector struct {
	db         *sql.DB
	connString string
	readOnly   bool
}

// Dangerous SQL statements that modify data or schema
var dangerousStmtPattern = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE|EXEC|EXECUTE)\b`)

// Multiple statement detection (prevents chaining attacks)
var multiStmtPattern = regexp.MustCompile(`;\s*\S`)

// NewSQLConnector creates a SQL connector for a database
func NewSQLConnector(connString string) *SQLConnector {
	return &SQLConnector{
		connString: connString,
		readOnly:   true,
	}
}

// Connect establishes database connection (called once at startup)
func (c *SQLConnector) Connect(ctx context.Context) error {
	var err error
	connStr := c.connString
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		if strings.Contains(connStr, "?") {
			if !strings.Contains(connStr, "binary_parameters=") {
				connStr += "&binary_parameters=yes"
			}
		} else {
			connStr += "?binary_parameters=yes"
		}
	}

	c.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	slog.Info("SQL connector connected", "driver", "postgres")
	return nil
}

// Execute runs a SQL query and returns results
func (c *SQLConnector) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	start := time.Now()

	if c.db == nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    "database connection not initialized",
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Create context with timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	// Extract query from inputs
	query, ok := req.Inputs["query"].(string)
	if !ok || query == "" {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    "query field is required and must be a string",
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Guard: reject multiple statements (prevents injection chaining)
	if multiStmtPattern.MatchString(query) {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    "multiple SQL statements are not allowed",
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Guard: reject write/DDL operations when in read-only mode
	if c.readOnly && dangerousStmtPattern.MatchString(query) {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    "write operations are not permitted — SQL connector is read-only",
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	slog.Info("SQL connector executing query", "tool_id", req.ToolID, "query_length", len(query))

	// Execute query with row limit
	const maxRows = 1000
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ExecutionResponse{
				ToolID:   req.ToolID,
				Status:   "timeout",
				Error:    "SQL query exceeded timeout",
				Duration: int(time.Since(start).Milliseconds()),
			}, nil
		}
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("query execution failed: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}
	defer rows.Close()

	// Fetch column names
	cols, err := rows.Columns()
	if err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("failed to read columns: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	// Fetch all rows as maps (capped at maxRows)
	var result []map[string]interface{}
	for rows.Next() {
		if len(result) >= maxRows {
			slog.Warn("SQL query hit row limit", "tool_id", req.ToolID, "max_rows", maxRows)
			break
		}
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &ExecutionResponse{
				ToolID:   req.ToolID,
				Status:   "error",
				Error:    fmt.Sprintf("failed to scan row: %v", err),
				Duration: int(time.Since(start).Milliseconds()),
			}, nil
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return &ExecutionResponse{
			ToolID:   req.ToolID,
			Status:   "error",
			Error:    fmt.Sprintf("row iteration error: %v", err),
			Duration: int(time.Since(start).Milliseconds()),
		}, nil
	}

	return &ExecutionResponse{
		ToolID:   req.ToolID,
		Status:   "success",
		Data:     result,
		Duration: int(time.Since(start).Milliseconds()),
	}, nil
}

// Validate checks if the SQL request is valid
func (c *SQLConnector) Validate(req *ExecutionRequest) error {
	if req.ToolID == "" {
		return fmt.Errorf("tool_id is required")
	}
	if _, ok := req.Inputs["query"].(string); !ok {
		return fmt.Errorf("inputs must contain 'query' field (string)")
	}
	return nil
}

// HealthCheck verifies database connectivity
func (c *SQLConnector) HealthCheck(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database connection not initialized")
	}
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	slog.Info("SQL connector health check passed")
	return nil
}

// Type returns the connector type
func (c *SQLConnector) Type() ConnectorType {
	return TypeSQL
}

// Close closes database connection
func (c *SQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
