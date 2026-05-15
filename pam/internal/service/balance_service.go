package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/berndmarcel860-byte/ongame/pam/internal/db"
	"github.com/berndmarcel860-byte/ongame/pam/internal/models"
	"github.com/google/uuid"
)

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
	ErrBalanceNotFound      = errors.New("balance not found")
)

type BalanceService struct {
	db *db.Postgres
}

func NewBalanceService(database *db.Postgres) *BalanceService {
	return &BalanceService{db: database}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*models.BalanceResponse, error) {
	var b models.Balance
	err := s.db.DB.QueryRowContext(ctx, `
SELECT id, user_id, cash_balance, bonus_balance, locked_balance, currency, version, updated_at
FROM balances WHERE user_id = $1`, userID,
	).Scan(&b.ID, &b.UserID, &b.CashBalance, &b.BonusBalance, &b.LockedBalance, &b.Currency, &b.Version, &b.UpdatedAt)
	if err != nil {
		return nil, ErrBalanceNotFound
	}
	return &models.BalanceResponse{
		Cash:     b.CashBalance,
		Bonus:    b.BonusBalance,
		Locked:   b.LockedBalance,
		Currency: b.Currency,
	}, nil
}

// ProcessTransaction handles WITHDRAW/DEPOSIT atomically with idempotency.
func (s *BalanceService) ProcessTransaction(ctx context.Context, userID uuid.UUID, req *models.TransactionRequest) (*models.TransactionResponse, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Idempotency: check if provider transaction already processed.
	if req.ProviderTransactionID != "" {
		existing, err := findExistingTx(ctx, tx, req.ProviderTransactionID)
		if err == nil && existing != nil {
			var bal models.Balance
			if berr := scanBalance(ctx, tx, userID, &bal); berr != nil {
				return nil, berr
			}
			return &models.TransactionResponse{
				TransactionID: existing.ID.String(),
				Cash:          bal.CashBalance,
				Bonus:         bal.BonusBalance,
				Locked:        bal.LockedBalance,
				Currency:      bal.Currency,
			}, nil
		}
	}

	// Lock balance row for atomic update.
	var bal models.Balance
	if err := lockBalance(ctx, tx, userID, &bal); err != nil {
		return nil, ErrBalanceNotFound
	}

	cashBefore := bal.CashBalance
	bonusBefore := bal.BonusBalance

	switch req.Type {
	case models.TxTypeWithdraw:
		if bal.CashBalance < req.Amount {
			return nil, ErrInsufficientFunds
		}
		bal.CashBalance -= req.Amount
	case models.TxTypeDeposit:
		bal.CashBalance += req.Amount
	}

	_, err = tx.ExecContext(ctx, `
UPDATE balances
SET cash_balance = $1, version = version + 1, updated_at = NOW()
WHERE user_id = $2 AND version = $3`,
		bal.CashBalance, userID, bal.Version)
	if err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	var txnID uuid.UUID
	err = tx.QueryRowContext(ctx, `
INSERT INTO transactions
  (user_id, type, amount, currency, provider_transaction_id, game_id, round_id,
   cash_before, cash_after, bonus_before, bonus_after, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id`,
		userID, req.Type, req.Amount, req.Currency,
		nullableStr(req.ProviderTransactionID), nullableStr(req.GameID), nullableStr(req.RoundID),
		cashBefore, bal.CashBalance, bonusBefore, bal.BonusBalance, models.TxStatusCompleted,
	).Scan(&txnID)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &models.TransactionResponse{
		TransactionID: txnID.String(),
		Cash:          bal.CashBalance,
		Bonus:         bal.BonusBalance,
		Locked:        bal.LockedBalance,
		Currency:      bal.Currency,
	}, nil
}

// AdjustBalance allows admins to manually credit/debit a balance.
func (s *BalanceService) AdjustBalance(ctx context.Context, userID uuid.UUID, amount float64, reason string) error {
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var bal models.Balance
	if err := lockBalance(ctx, tx, userID, &bal); err != nil {
		return ErrBalanceNotFound
	}

	cashBefore := bal.CashBalance
	bal.CashBalance += amount
	if bal.CashBalance < 0 {
		return ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx, `
UPDATE balances SET cash_balance = $1, version = version + 1, updated_at = NOW()
WHERE user_id = $2`, bal.CashBalance, userID)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	txType := models.TxTypeAdjustment
	txAmount := amount
	if amount < 0 {
		txAmount = -amount
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO transactions
  (user_id, type, amount, currency, cash_before, cash_after, bonus_before, bonus_after, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		userID, txType, txAmount, bal.Currency,
		cashBefore, bal.CashBalance, bal.BonusBalance, bal.BonusBalance, models.TxStatusCompleted)
	if err != nil {
		return fmt.Errorf("insert adjustment tx: %w", err)
	}

	return tx.Commit()
}

// helpers

func lockBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID, bal *models.Balance) error {
	return tx.QueryRowContext(ctx, `
SELECT id, user_id, cash_balance, bonus_balance, locked_balance, currency, version
FROM balances WHERE user_id = $1 FOR UPDATE`, userID,
	).Scan(&bal.ID, &bal.UserID, &bal.CashBalance, &bal.BonusBalance, &bal.LockedBalance, &bal.Currency, &bal.Version)
}

func scanBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID, bal *models.Balance) error {
	return tx.QueryRowContext(ctx, `
SELECT id, user_id, cash_balance, bonus_balance, locked_balance, currency, version
FROM balances WHERE user_id = $1`, userID,
	).Scan(&bal.ID, &bal.UserID, &bal.CashBalance, &bal.BonusBalance, &bal.LockedBalance, &bal.Currency, &bal.Version)
}

func findExistingTx(ctx context.Context, tx *sql.Tx, providerTxID string) (*models.Transaction, error) {
	var t models.Transaction
	err := tx.QueryRowContext(ctx, `
SELECT id, user_id, type, amount, currency, provider_transaction_id, game_id, round_id,
       cash_before, cash_after, bonus_before, bonus_after, status, created_at
FROM transactions WHERE provider_transaction_id = $1`, providerTxID,
	).Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Currency,
		&t.ProviderTransactionID, &t.GameID, &t.RoundID,
		&t.CashBefore, &t.CashAfter, &t.BonusBefore, &t.BonusAfter,
		&t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
