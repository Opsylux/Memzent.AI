package billing

import (
	"context"
	"database/sql"
	"fmt"
)

type Ledger struct {
	db *sql.DB
}

func NewLedger(db *sql.DB) *Ledger {
	return &Ledger{db: db}
}

// HasSufficientBalance checks if the org has a token balance > 0
func (l *Ledger) HasSufficientBalance(ctx context.Context, orgID string) (bool, error) {
	if orgID == "default" || orgID == "" {
		return true, nil // Skip check for system/default org
	}
	var balance float64
	err := l.db.QueryRowContext(ctx, "SELECT COALESCE(token_balance, 0) FROM organizations WHERE id = $1", orgID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("organization not found: %s", orgID)
		}
		return false, err
	}
	return balance > 0, nil
}

// GetBalance retrieves the current token balance for the organization
func (l *Ledger) GetBalance(ctx context.Context, orgID string) (float64, error) {
	if orgID == "default" || orgID == "" {
		return 0, nil
	}
	var balance float64
	err := l.db.QueryRowContext(ctx, "SELECT COALESCE(token_balance, 0) FROM organizations WHERE id = $1", orgID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Return 0 for unknown org
		}
		return 0, err
	}
	return balance, nil
}
// Deduct subtracts the amount from the org's balance and records the transaction
func (l *Ledger) Deduct(ctx context.Context, orgID string, amount float64, txType, desc string) error {
	if amount == 0 || orgID == "default" || orgID == "" {
		return nil
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE organizations SET token_balance = COALESCE(token_balance, 0) - $1 WHERE id = $2", amount, orgID)
	if err != nil {
		return err
	}

	// Record ledger entry (negative amount for deduction)
	_, err = tx.ExecContext(ctx,
		"INSERT INTO billing_ledger (org_id, amount, transaction_type, description) VALUES ($1, $2, $3, $4)",
		orgID, -amount, txType, desc,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// TopUp adds the amount to the org's balance and records the transaction
func (l *Ledger) TopUp(ctx context.Context, orgID string, amount float64, desc string) error {
	if orgID == "default" || orgID == "" {
		return nil
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE organizations SET token_balance = COALESCE(token_balance, 0) + $1 WHERE id = $2", amount, orgID)
	if err != nil {
		return err
	}

	// Positive amount for topup
	_, err = tx.ExecContext(ctx,
		"INSERT INTO billing_ledger (org_id, amount, transaction_type, description) VALUES ($1, $2, 'stripe_topup', $3)",
		orgID, amount, desc,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
