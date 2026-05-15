package models

import (
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	CashBalance    float64   `json:"cash" db:"cash_balance"`
	BonusBalance   float64   `json:"bonus" db:"bonus_balance"`
	LockedBalance  float64   `json:"locked" db:"locked_balance"`
	Currency       string    `json:"currency" db:"currency"`
	Version        int       `json:"-" db:"version"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// BalanceResponse is returned to Valkyrie PAM client.
type BalanceResponse struct {
	Cash     float64 `json:"cash"`
	Bonus    float64 `json:"bonus"`
	Locked   float64 `json:"locked"`
	Currency string  `json:"currency"`
}

type AdjustBalanceRequest struct {
	Amount float64 `json:"amount" binding:"required"`
	Reason string  `json:"reason" binding:"required"`
}
