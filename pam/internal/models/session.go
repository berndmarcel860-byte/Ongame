package models

import (
	"time"

	"github.com/google/uuid"
)

type GameSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	GameID       string     `json:"gameId" db:"game_id"`
	Provider     string     `json:"provider" db:"provider"`
	StartedAt    time.Time  `json:"startedAt" db:"started_at"`
	EndedAt      *time.Time `json:"endedAt,omitempty" db:"ended_at"`
	TotalWagered float64    `json:"totalWagered" db:"total_wagered"`
	TotalWon     float64    `json:"totalWon" db:"total_won"`
}

type CreateSessionRequest struct {
	GameID   string `json:"gameId" binding:"required"`
	Provider string `json:"provider"`
}

type SessionResponse struct {
	SessionToken string `json:"sessionToken"`
	SessionID    string `json:"sessionId"`
}
