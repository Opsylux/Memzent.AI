package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSessionManager_CreateSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sm := NewSessionManager(db)
	ctx := context.Background()

	t.Run("DB Not Initialized", func(t *testing.T) {
		nilMgr := NewSessionManager(nil)
		_, err := nilMgr.CreateSession(ctx, "org1", "user1", "")
		if err == nil {
			t.Errorf("expected error when DB is nil, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO chat_sessions").
			WithArgs("org1", "user1", "New Conversation").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("session-uuid-001"))

		id, err := sm.CreateSession(ctx, "org1", "user1", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != "session-uuid-001" {
			t.Errorf("expected session-uuid-001, got %q", id)
		}
	})
}

func TestSessionManager_ListSessions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sm := NewSessionManager(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("DB Not Initialized", func(t *testing.T) {
		nilMgr := NewSessionManager(nil)
		_, err := nilMgr.ListSessions(ctx, "org1")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, org_id, user_id, title, created_at, updated_at FROM chat_sessions").
			WithArgs("org1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "org_id", "user_id", "title", "created_at", "updated_at"}).
				AddRow("session-01", "org1", "user1", "Title 1", now, now).
				AddRow("session-02", "org1", nil, "Title 2", now, now))

		sessions, err := sm.ListSessions(ctx, "org1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sessions) != 2 {
			t.Fatalf("expected 2 sessions, got %d", len(sessions))
		}
		if sessions[0].UserID != "user1" || sessions[1].UserID != "" {
			t.Errorf("unexpected scanned session values: %+v, %+v", sessions[0], sessions[1])
		}
	})
}

func TestSessionManager_DeleteSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sm := NewSessionManager(db)
	ctx := context.Background()

	t.Run("DB Not Initialized", func(t *testing.T) {
		nilMgr := NewSessionManager(nil)
		err := nilMgr.DeleteSession(ctx, "session-01")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM chat_sessions WHERE id = \\$1").
			WithArgs("session-01").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := sm.DeleteSession(ctx, "session-01")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestSessionManager_AppendMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sm := NewSessionManager(db)
	ctx := context.Background()

	t.Run("DB Not Initialized", func(t *testing.T) {
		nilMgr := NewSessionManager(nil)
		err := nilMgr.AppendMessage(ctx, "session-01", "user", "hi")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO chat_messages").
			WithArgs("session-01", "user", "hi").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE chat_sessions SET updated_at = NOW\\(\\)").
			WithArgs("session-01").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := sm.AppendMessage(ctx, "session-01", "user", "hi")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("Transaction Fail", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

		err := sm.AppendMessage(ctx, "session-01", "user", "hi")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestSessionManager_GetSessionMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	sm := NewSessionManager(db)
	ctx := context.Background()

	t.Run("DB Not Initialized", func(t *testing.T) {
		nilMgr := NewSessionManager(nil)
		_, err := nilMgr.GetSessionMessages(ctx, "session-01", 10)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery("SELECT role, content FROM").
			WithArgs("session-01", 10).
			WillReturnRows(sqlmock.NewRows([]string{"role", "content"}).
				AddRow("user", "hello").
				AddRow("assistant", "world"))

		messages, err := sm.GetSessionMessages(ctx, "session-01", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}
		if messages[0].Role != "user" || messages[1].Content != "world" {
			t.Errorf("unexpected messages output: %+v", messages)
		}
	})

	t.Run("Success Default Limit", func(t *testing.T) {
		mock.ExpectQuery("SELECT role, content FROM").
			WithArgs("session-01", 20).
			WillReturnRows(sqlmock.NewRows([]string{"role", "content"}))

		_, err := sm.GetSessionMessages(ctx, "session-01", 0)
		if err != nil {
			t.Errorf("unexpected error on default limit: %v", err)
		}
	})
}
