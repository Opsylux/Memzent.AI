package billing

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLedger_HasSufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	ledger := NewLedger(db)
	ctx := context.Background()

	// Default/Empty
	has, err := ledger.HasSufficientBalance(ctx, "default")
	if err != nil || !has {
		t.Errorf("default org should have sufficient balance")
	}

	// Sufficient balance
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org1").
		WillReturnRows(sqlmock.NewRows([]string{"token_balance"}).AddRow(10.5))
	
	has, err = ledger.HasSufficientBalance(ctx, "org1")
	if err != nil || !has {
		t.Errorf("org1 should have sufficient balance")
	}

	// Insufficient balance
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org2").
		WillReturnRows(sqlmock.NewRows([]string{"token_balance"}).AddRow(0))
	
	has, err = ledger.HasSufficientBalance(ctx, "org2")
	if err != nil || has {
		t.Errorf("org2 should not have sufficient balance")
	}

	// Unknown org
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org3").
		WillReturnError(sql.ErrNoRows)
	
	_, err = ledger.HasSufficientBalance(ctx, "org3")
	if err == nil {
		t.Errorf("Expected error for unknown org")
	}

	// Query error
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org4").
		WillReturnError(fmt.Errorf("db error"))
	
	_, err = ledger.HasSufficientBalance(ctx, "org4")
	if err == nil {
		t.Errorf("Expected error for db error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestLedger_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	ledger := NewLedger(db)
	ctx := context.Background()

	// Default/Empty
	bal, err := ledger.GetBalance(ctx, "default")
	if err != nil || bal != 0 {
		t.Errorf("default org should have 0 balance")
	}

	// Normal
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org1").
		WillReturnRows(sqlmock.NewRows([]string{"token_balance"}).AddRow(15.2))
	
	bal, err = ledger.GetBalance(ctx, "org1")
	if err != nil || bal != 15.2 {
		t.Errorf("org1 should have 15.2 balance, got %v", bal)
	}

	// ErrNoRows
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org2").
		WillReturnError(sql.ErrNoRows)
	
	bal, err = ledger.GetBalance(ctx, "org2")
	if err != nil || bal != 0 {
		t.Errorf("org2 should have 0 balance on ErrNoRows, got %v", bal)
	}

	// Query error
	mock.ExpectQuery("SELECT COALESCE\\(token_balance, 0\\) FROM organizations").
		WithArgs("org3").
		WillReturnError(fmt.Errorf("db error"))
	
	_, err = ledger.GetBalance(ctx, "org3")
	if err == nil {
		t.Errorf("Expected db error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestLedger_Deduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	ledger := NewLedger(db)
	ctx := context.Background()

	// Default/Empty
	err = ledger.Deduct(ctx, "default", 10.0, "llm_query", "test")
	if err != nil {
		t.Errorf("default org should skip deduct")
	}
	err = ledger.Deduct(ctx, "org1", 0, "llm_query", "test")
	if err != nil {
		t.Errorf("0 amount should skip deduct")
	}

	// Normal deduction
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance = COALESCE\\(token_balance, 0\\) - \\$1").
		WithArgs(5.5, "org1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO billing_ledger").
		WithArgs("org1", -5.5, "llm_query", "desc").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = ledger.Deduct(ctx, "org1", 5.5, "llm_query", "desc")
	if err != nil {
		t.Errorf("unexpected error on deduct: %v", err)
	}

	// Begin error
	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))
	err = ledger.Deduct(ctx, "org1", 5.5, "llm_query", "desc")
	if err == nil {
		t.Errorf("expected begin error")
	}

	// Update error
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance").
		WithArgs(5.5, "org1").
		WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err = ledger.Deduct(ctx, "org1", 5.5, "llm_query", "desc")
	if err == nil {
		t.Errorf("expected update error")
	}

	// Insert error
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance").
		WithArgs(5.5, "org1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO billing_ledger").
		WithArgs("org1", -5.5, "llm_query", "desc").
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err = ledger.Deduct(ctx, "org1", 5.5, "llm_query", "desc")
	if err == nil {
		t.Errorf("expected insert error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestLedger_TopUp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	ledger := NewLedger(db)
	ctx := context.Background()

	// Default/Empty
	err = ledger.TopUp(ctx, "default", 10.0, "test")
	if err != nil {
		t.Errorf("default org should skip topup")
	}

	// Normal topup
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance = COALESCE\\(token_balance, 0\\) \\+ \\$1").
		WithArgs(50.0, "org1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO billing_ledger").
		WithArgs("org1", 50.0, "desc").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = ledger.TopUp(ctx, "org1", 50.0, "desc")
	if err != nil {
		t.Errorf("unexpected error on topup: %v", err)
	}

	// Begin error
	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))
	err = ledger.TopUp(ctx, "org1", 50.0, "desc")
	if err == nil {
		t.Errorf("expected begin error")
	}

	// Update error
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance").
		WithArgs(50.0, "org1").
		WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err = ledger.TopUp(ctx, "org1", 50.0, "desc")
	if err == nil {
		t.Errorf("expected update error")
	}

	// Insert error
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE organizations SET token_balance").
		WithArgs(50.0, "org1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO billing_ledger").
		WithArgs("org1", 50.0, "desc").
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err = ledger.TopUp(ctx, "org1", 50.0, "desc")
	if err == nil {
		t.Errorf("expected insert error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
