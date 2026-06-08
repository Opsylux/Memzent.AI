package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Webhook represents a registered webhook subscription
type Webhook struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"` // Only returned on creation
	Events      []string  `json:"events"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeliveryLog represents a webhook delivery attempt
type DeliveryLog struct {
	ID            string     `json:"id"`
	WebhookID     string     `json:"webhook_id"`
	EventType     string     `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	ResponseCode  *int       `json:"response_code,omitempty"`
	Error         string     `json:"error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Registry manages webhook subscriptions in Postgres
type Registry struct {
	db *sql.DB
}

// NewRegistry creates a webhook registry backed by Postgres
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{db: db}
}

// Create registers a new webhook for an org
func (r *Registry) Create(ctx context.Context, wh *Webhook) error {
	query := `
		INSERT INTO webhooks (org_id, url, secret, events, enabled, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		wh.OrgID, wh.URL, wh.Secret, pq.Array(wh.Events), wh.Enabled, wh.Description,
	).Scan(&wh.ID, &wh.CreatedAt, &wh.UpdatedAt)
}

// List returns all webhooks for an org
func (r *Registry) List(ctx context.Context, orgID string) ([]Webhook, error) {
	query := `
		SELECT id, org_id, url, events, enabled, description, created_at, updated_at
		FROM webhooks WHERE org_id = $1::uuid
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	webhooks := make([]Webhook, 0)
	for rows.Next() {
		var wh Webhook
		err := rows.Scan(&wh.ID, &wh.OrgID, &wh.URL, pq.Array(&wh.Events),
			&wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, rows.Err()
}

// Get retrieves a single webhook by ID (scoped to org)
func (r *Registry) Get(ctx context.Context, orgID, webhookID string) (*Webhook, error) {
	query := `
		SELECT id, org_id, url, events, enabled, description, created_at, updated_at
		FROM webhooks WHERE id = $1::uuid AND org_id = $2::uuid
	`
	var wh Webhook
	err := r.db.QueryRowContext(ctx, query, webhookID, orgID).Scan(
		&wh.ID, &wh.OrgID, &wh.URL, pq.Array(&wh.Events),
		&wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &wh, err
}

// Update modifies a webhook's mutable fields
func (r *Registry) Update(ctx context.Context, wh *Webhook) error {
	query := `
		UPDATE webhooks SET url = $3, events = $4, enabled = $5, description = $6, updated_at = NOW()
		WHERE id = $1::uuid AND org_id = $2::uuid
	`
	res, err := r.db.ExecContext(ctx, query, wh.ID, wh.OrgID, wh.URL,
		pq.Array(wh.Events), wh.Enabled, wh.Description)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// Delete removes a webhook
func (r *Registry) Delete(ctx context.Context, orgID, webhookID string) error {
	query := `DELETE FROM webhooks WHERE id = $1::uuid AND org_id = $2::uuid`
	res, err := r.db.ExecContext(ctx, query, webhookID, orgID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// GetSubscribers returns all enabled webhooks subscribed to a given event type for an org
func (r *Registry) GetSubscribers(ctx context.Context, orgID, eventType string) ([]Webhook, error) {
	query := `
		SELECT id, org_id, url, secret, events, enabled, description, created_at, updated_at
		FROM webhooks
		WHERE org_id = $1::uuid AND enabled = true AND $2 = ANY(events)
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var wh Webhook
		err := rows.Scan(&wh.ID, &wh.OrgID, &wh.URL, &wh.Secret, pq.Array(&wh.Events),
			&wh.Enabled, &wh.Description, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, rows.Err()
}

// LogDelivery records a delivery attempt
func (r *Registry) LogDelivery(ctx context.Context, log *DeliveryLog) error {
	query := `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status, attempts, last_attempt_at, response_code, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		log.WebhookID, log.EventType, log.Payload, log.Status,
		log.Attempts, log.LastAttemptAt, log.ResponseCode, log.Error,
	).Scan(&log.ID, &log.CreatedAt)
}

// GetDeliveryLogs returns recent delivery logs for a webhook
func (r *Registry) GetDeliveryLogs(ctx context.Context, orgID, webhookID string, limit int) ([]DeliveryLog, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT d.id, d.webhook_id, d.event_type, d.payload, d.status, d.attempts,
		       d.last_attempt_at, d.response_code, d.error, d.created_at
		FROM webhook_deliveries d
		JOIN webhooks w ON w.id = d.webhook_id
		WHERE d.webhook_id = $1::uuid AND w.org_id = $2::uuid
		ORDER BY d.created_at DESC LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, webhookID, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]DeliveryLog, 0)
	for rows.Next() {
		var l DeliveryLog
		err := rows.Scan(&l.ID, &l.WebhookID, &l.EventType, &l.Payload, &l.Status,
			&l.Attempts, &l.LastAttemptAt, &l.ResponseCode, &l.Error, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
