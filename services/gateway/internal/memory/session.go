// Package memory implements short-term conversation persistence and long-term semantic memory logic.
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"memzent-gateway/internal/llm"
)

// ChatSession represents an active conversation thread
type ChatSession struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionManager orchestrates chat session persistence in PostgreSQL
type SessionManager struct {
	db *sql.DB
}

// NewSessionManager instantiates a session manager
func NewSessionManager(db *sql.DB) *SessionManager {
	return &SessionManager{db: db}
}

// CreateSession creates a new chat session for an organization
func (sm *SessionManager) CreateSession(ctx context.Context, orgID, userID, title string) (string, error) {
	if sm.db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	if title == "" {
		title = "New Conversation"
	}

	query := `
		INSERT INTO chat_sessions (org_id, user_id, title)
		VALUES ($1::uuid, $2, $3)
		RETURNING id
	`
	var id string
	err := sm.db.QueryRowContext(ctx, query, orgID, userID, title).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create chat session: %w", err)
	}

	slog.Info("Chat session created", "session_id", id, "org_id", orgID)
	return id, nil
}

// ListSessions retrieves all sessions for an organization ordered by last update
func (sm *SessionManager) ListSessions(ctx context.Context, orgID string) ([]*ChatSession, error) {
	if sm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, org_id, user_id, title, created_at, updated_at
		FROM chat_sessions
		WHERE org_id = $1::uuid
		ORDER BY updated_at DESC
	`
	rows, err := sm.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*ChatSession
	for rows.Next() {
		var s ChatSession
		var userID sql.NullString
		err := rows.Scan(&s.ID, &s.OrgID, &userID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat session: %w", err)
		}
		s.UserID = userID.String
		sessions = append(sessions, &s)
	}

	return sessions, nil
}

// DeleteSession prunes a chat session and all cascading messages
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	if sm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := "DELETE FROM chat_sessions WHERE id = $1::uuid"
	_, err := sm.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	slog.Info("Chat session deleted", "session_id", sessionID)
	return nil
}

// AppendMessage inserts a user or assistant response into the conversation history
func (sm *SessionManager) AppendMessage(ctx context.Context, sessionID, role, content string) error {
	if sm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert message
	queryMsg := `
		INSERT INTO chat_messages (session_id, role, content)
		VALUES ($1::uuid, $2, $3)
	`
	_, err = tx.ExecContext(ctx, queryMsg, sessionID, role, content)
	if err != nil {
		return fmt.Errorf("failed to insert chat message: %w", err)
	}

	// 2. Touch the session's updated_at timestamp
	querySession := "UPDATE chat_sessions SET updated_at = NOW() WHERE id = $1::uuid"
	_, err = tx.ExecContext(ctx, querySession, sessionID)
	if err != nil {
		return fmt.Errorf("failed to touch chat session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit message append transaction: %w", err)
	}

	return nil
}

// GetSessionMessages fetches the last N messages inside a conversation session
func (sm *SessionManager) GetSessionMessages(ctx context.Context, sessionID string, limit int) ([]llm.Message, error) {
	if sm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT role, content
		FROM (
			SELECT role, content, created_at
			FROM chat_messages
			WHERE session_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2
		) sub
		ORDER BY created_at ASC
	`
	rows, err := sm.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []llm.Message
	for rows.Next() {
		var m llm.Message
		err := rows.Scan(&m.Role, &m.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}
