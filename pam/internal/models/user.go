package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Username       string    `json:"username" db:"username"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	KYCStatus      string    `json:"kycStatus" db:"kyc_status"`
	IsActive       bool      `json:"isActive" db:"is_active"`
	IsBanned       bool      `json:"isBanned" db:"is_banned"`
	IsSelfExcluded bool      `json:"isSelfExcluded" db:"is_self_excluded"`
	Currency       string    `json:"currency" db:"currency"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8"`
	Currency string `json:"currency" binding:"required,len=3"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type BanRequest struct {
	Reason string `json:"reason" binding:"required"`
}
