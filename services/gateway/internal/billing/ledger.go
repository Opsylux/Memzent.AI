package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Ledger struct {
	db *sql.DB
}

func NewLedger(db *sql.DB) *Ledger {
	return &Ledger{db: db}
}

type OrgSettings struct {
	TokenBalance    float64
	DefaultProvider string
	DefaultModel    string
}

type Transaction struct {
	ID              string    `json:"id"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetRecentTransactions retrieves the latest billing ledger logs for an organization
func (l *Ledger) GetRecentTransactions(ctx context.Context, orgID string, limit int) ([]Transaction, error) {
	if orgID == "default" || orgID == "" {
		return nil, nil
	}
	rows, err := l.db.QueryContext(ctx, `
		SELECT id, amount, transaction_type, COALESCE(description, ''), created_at
		FROM billing_ledger
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.ID, &tx.Amount, &tx.TransactionType, &tx.Description, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// GetOrgSettings retrieves organization preferences and balance details
func (l *Ledger) GetOrgSettings(ctx context.Context, orgID string) (*OrgSettings, error) {
	if orgID == "default" || orgID == "" {
		return &OrgSettings{TokenBalance: 999999.0}, nil
	}

	var balance float64
	var defaultProvider sql.NullString
	var defaultModel sql.NullString

	query := `
		SELECT COALESCE(token_balance, 0), default_provider, default_model
		FROM organizations
		WHERE id = $1
	`
	err := l.db.QueryRowContext(ctx, query, orgID).Scan(&balance, &defaultProvider, &defaultModel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", orgID)
		}
		return nil, err
	}

	return &OrgSettings{
		TokenBalance:    balance,
		DefaultProvider: defaultProvider.String,
		DefaultModel:    defaultModel.String,
	}, nil
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

// SpendLimitStatus represents whether org has hit daily/monthly caps
type SpendLimitStatus struct {
	DailySpend     float64  `json:"daily_spend"`
	MonthlySpend   float64  `json:"monthly_spend"`
	DailyLimit     *float64 `json:"daily_limit,omitempty"`
	MonthlyLimit   *float64 `json:"monthly_limit,omitempty"`
	DailyExceeded  bool     `json:"daily_exceeded"`
	MonthlyExceeded bool    `json:"monthly_exceeded"`
}

// CheckSpendLimits verifies if the org is within configured spend limits.
// Returns nil if no limits are set or limits not exceeded.
func (l *Ledger) CheckSpendLimits(ctx context.Context, orgID string) (*SpendLimitStatus, error) {
	if orgID == "default" || orgID == "" {
		return nil, nil
	}

	var dailyLimit, monthlyLimit sql.NullFloat64
	err := l.db.QueryRowContext(ctx,
		"SELECT daily_spend_limit, monthly_spend_limit FROM organizations WHERE id = $1", orgID,
	).Scan(&dailyLimit, &monthlyLimit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// If no limits configured, skip
	if !dailyLimit.Valid && !monthlyLimit.Valid {
		return nil, nil
	}

	status := &SpendLimitStatus{}

	// Get today's spend
	_ = l.db.QueryRowContext(ctx,
		"SELECT COALESCE(ABS(SUM(amount)), 0) FROM billing_ledger WHERE org_id = $1 AND amount < 0 AND created_at > CURRENT_DATE",
		orgID,
	).Scan(&status.DailySpend)

	// Get this month's spend
	_ = l.db.QueryRowContext(ctx,
		"SELECT COALESCE(ABS(SUM(amount)), 0) FROM billing_ledger WHERE org_id = $1 AND amount < 0 AND created_at > DATE_TRUNC('month', CURRENT_DATE)",
		orgID,
	).Scan(&status.MonthlySpend)

	if dailyLimit.Valid {
		status.DailyLimit = &dailyLimit.Float64
		status.DailyExceeded = status.DailySpend >= dailyLimit.Float64
	}
	if monthlyLimit.Valid {
		status.MonthlyLimit = &monthlyLimit.Float64
		status.MonthlyExceeded = status.MonthlySpend >= monthlyLimit.Float64
	}

	return status, nil
}

// SetSpendLimits updates the org's daily/monthly spend caps
func (l *Ledger) SetSpendLimits(ctx context.Context, orgID string, dailyLimit, monthlyLimit *float64) error {
	_, err := l.db.ExecContext(ctx,
		"UPDATE organizations SET daily_spend_limit = $2, monthly_spend_limit = $3 WHERE id = $1",
		orgID, dailyLimit, monthlyLimit,
	)
	return err
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

// SpendSummary represents aggregated spend data for a time period
type SpendSummary struct {
	Period       string  `json:"period"`        // "24h", "7d", "30d"
	TotalSpend   float64 `json:"total_spend"`
	RequestCount int     `json:"request_count"`
	CacheHits    int     `json:"cache_hits"`
	CacheSavings float64 `json:"cache_savings"`
}

// ProviderSpend breaks down spending by LLM provider
type ProviderSpend struct {
	Provider     string  `json:"provider"`
	TotalSpend   float64 `json:"total_spend"`
	RequestCount int     `json:"request_count"`
}

// BudgetStatus represents the current budget health for planning tools
type BudgetStatus struct {
	OrgID           string          `json:"org_id"`
	CurrentBalance  float64         `json:"current_balance"`
	Tier            string          `json:"tier"`
	SpendSummaries  []SpendSummary  `json:"spend_summaries"`
	ProviderBreakdown []ProviderSpend `json:"provider_breakdown"`
	DailyAvgSpend   float64         `json:"daily_avg_spend"`
	ProjectedDaysRemaining float64  `json:"projected_days_remaining"`
	BurnRate        float64         `json:"burn_rate_per_hour"`
	SpendLimit      *float64        `json:"spend_limit,omitempty"`
}

// GetBudgetStatus computes a comprehensive budget/spend report for planning tools
func (l *Ledger) GetBudgetStatus(ctx context.Context, orgID string) (*BudgetStatus, error) {
	if orgID == "default" || orgID == "" {
		return &BudgetStatus{OrgID: orgID, CurrentBalance: 999999}, nil
	}

	status := &BudgetStatus{OrgID: orgID}

	// 1. Current balance + tier
	var tier sql.NullString
	err := l.db.QueryRowContext(ctx,
		"SELECT COALESCE(token_balance, 0), COALESCE(tier, 'free') FROM organizations WHERE id = $1", orgID,
	).Scan(&status.CurrentBalance, &tier)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}
	status.Tier = tier.String

	// 2. Spend summaries for 24h, 7d, 30d
	periods := []struct {
		label    string
		interval string
	}{
		{"24h", "1 day"},
		{"7d", "7 days"},
		{"30d", "30 days"},
	}

	for _, p := range periods {
		var spend float64
		var reqCount, cacheCount int
		var savings float64

		// Total spend (negative amounts in ledger = spending)
		err := l.db.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT COALESCE(ABS(SUM(amount)), 0), COUNT(*)
			FROM billing_ledger
			WHERE org_id = $1 AND amount < 0 AND created_at > NOW() - INTERVAL '%s'
		`, p.interval), orgID).Scan(&spend, &reqCount)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		// Cache hits count + savings
		_ = l.db.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT COUNT(*), COALESCE(ABS(SUM(amount)), 0)
			FROM billing_ledger
			WHERE org_id = $1 AND transaction_type = 'cache_hit' AND created_at > NOW() - INTERVAL '%s'
		`, p.interval), orgID).Scan(&cacheCount, &savings)

		status.SpendSummaries = append(status.SpendSummaries, SpendSummary{
			Period:       p.label,
			TotalSpend:   spend,
			RequestCount: reqCount,
			CacheHits:    cacheCount,
			CacheSavings: savings,
		})
	}

	// 3. Provider breakdown (last 30 days)
	rows, err := l.db.QueryContext(ctx, `
		SELECT 
			COALESCE(SPLIT_PART(description, ' ', 1), 'unknown') as provider,
			COALESCE(ABS(SUM(amount)), 0) as total,
			COUNT(*) as cnt
		FROM billing_ledger
		WHERE org_id = $1 AND amount < 0 AND transaction_type != 'cache_hit'
		  AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY SPLIT_PART(description, ' ', 1)
		ORDER BY total DESC
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ps ProviderSpend
			_ = rows.Scan(&ps.Provider, &ps.TotalSpend, &ps.RequestCount)
			status.ProviderBreakdown = append(status.ProviderBreakdown, ps)
		}
	}
	if status.ProviderBreakdown == nil {
		status.ProviderBreakdown = make([]ProviderSpend, 0)
	}

	// 4. Compute burn rate and projections
	if len(status.SpendSummaries) >= 2 {
		// Use 7-day spend for projection (more stable than 24h)
		weeklySpend := status.SpendSummaries[1].TotalSpend
		status.DailyAvgSpend = weeklySpend / 7.0
		if status.DailyAvgSpend > 0 {
			status.BurnRate = status.DailyAvgSpend / 24.0
			status.ProjectedDaysRemaining = status.CurrentBalance / status.DailyAvgSpend
		}
	}

	return status, nil
}

// GetSpendTimeseries returns daily spend for the last N days (for charts)
func (l *Ledger) GetSpendTimeseries(ctx context.Context, orgID string, days int) ([]map[string]any, error) {
	if days <= 0 {
		days = 30
	}

	rows, err := l.db.QueryContext(ctx, `
		SELECT DATE(created_at) as day, COALESCE(ABS(SUM(amount)), 0) as spend, COUNT(*) as requests
		FROM billing_ledger
		WHERE org_id = $1 AND amount < 0 AND created_at > NOW() - MAKE_INTERVAL(days => $2)
		GROUP BY DATE(created_at)
		ORDER BY day ASC
	`, orgID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		var day time.Time
		var spend float64
		var requests int
		if err := rows.Scan(&day, &spend, &requests); err != nil {
			continue
		}
		result = append(result, map[string]any{
			"date":     day.Format("2006-01-02"),
			"spend":    spend,
			"requests": requests,
		})
	}
	return result, nil
}
