package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	TxTypeWithdraw   = "WITHDRAW"
	TxTypeDeposit    = "DEPOSIT"
	TxTypeAdjustment = "ADJUSTMENT"

	TxStatusCompleted = "completed"
	TxStatusFailed    = "failed"
)

type Transaction struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	UserID                uuid.UUID `json:"userId" db:"user_id"`
	Type                  string    `json:"type" db:"type"`
	Amount                float64   `json:"amount" db:"amount"`
	Currency              string    `json:"currency" db:"currency"`
	ProviderTransactionID *string   `json:"providerTransactionId,omitempty" db:"provider_transaction_id"`
	GameID                *string   `json:"gameId,omitempty" db:"game_id"`
	RoundID               *string   `json:"roundId,omitempty" db:"round_id"`
	CashBefore            float64   `json:"cashBefore" db:"cash_before"`
	CashAfter             float64   `json:"cashAfter" db:"cash_after"`
	BonusBefore           float64   `json:"bonusBefore" db:"bonus_before"`
	BonusAfter            float64   `json:"bonusAfter" db:"bonus_after"`
	Status                string    `json:"status" db:"status"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
}

type TransactionRequest struct {
	Type                  string  `json:"type" binding:"required,oneof=WITHDRAW DEPOSIT"`
	Amount                float64 `json:"amount" binding:"required,gt=0"`
	Currency              string  `json:"currency" binding:"required"`
	ProviderTransactionID string  `json:"providerTransactionId" binding:"required"`
	GameID                string  `json:"gameId"`
	RoundID               string  `json:"roundId"`
}

type TransactionResponse struct {
	TransactionID string  `json:"transactionId"`
	Cash          float64 `json:"cash"`
	Bonus         float64 `json:"bonus"`
	Locked        float64 `json:"locked"`
	Currency      string  `json:"currency"`
}

type AdminDemoWinRateResponse struct {
	TotalPlayers   int64   `json:"totalPlayers"`
	BetRounds      int64   `json:"betRounds"`
	WinRounds      int64   `json:"winRounds"`
	WinRatePercent float64 `json:"winRatePercent"`
	PayoutPercent  float64 `json:"payoutPercent"`
}
