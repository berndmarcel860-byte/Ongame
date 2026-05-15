package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/berndmarcel860-byte/ongame/pam/internal/auth"
	"github.com/berndmarcel860-byte/ongame/pam/internal/cache"
	"github.com/berndmarcel860-byte/ongame/pam/internal/db"
	"github.com/berndmarcel860-byte/ongame/pam/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBanned         = errors.New("user is banned")
	ErrUserInactive       = errors.New("user account is inactive")
)

type UserService struct {
	db         *db.Postgres
	cache      *cache.Redis
	jwtManager *auth.JWTManager
}

func NewUserService(database *db.Postgres, redisClient *cache.Redis, jwtManager *auth.JWTManager) *UserService {
	return &UserService{db: database, cache: redisClient, jwtManager: jwtManager}
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var user models.User
	err = s.db.DB.QueryRowContext(ctx, `
		INSERT INTO users (email, username, password_hash, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, password_hash, kyc_status, is_active, is_banned, is_self_excluded, currency, created_at, updated_at`,
		req.Email, req.Username, string(hash), req.Currency,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.KYCStatus, &user.IsActive, &user.IsBanned, &user.IsSelfExcluded,
		&user.Currency, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return nil, ErrEmailTaken
		}
		if isUniqueViolation(err, "users_username_key") {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	// Create initial balance record
	_, err = s.db.DB.ExecContext(ctx, `
		INSERT INTO balances (user_id, currency) VALUES ($1, $2)`,
		user.ID, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("create balance: %w", err)
	}

	return &user, nil
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.IsBanned {
		return nil, ErrUserBanned
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	token, err := s.jwtManager.Generate(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// Cache session
	_ = s.cache.Set(ctx, sessionKey(token), user.ID.String(), 24*time.Hour)

	return &models.LoginResponse{Token: token, User: user}, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.DB.QueryRowContext(ctx, `
		SELECT id, email, username, password_hash, kyc_status, is_active, is_banned, is_self_excluded, currency, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.KYCStatus, &user.IsActive, &user.IsBanned, &user.IsSelfExcluded,
		&user.Currency, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.DB.QueryRowContext(ctx, `
		SELECT id, email, username, password_hash, kyc_status, is_active, is_banned, is_self_excluded, currency, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.KYCStatus, &user.IsActive, &user.IsBanned, &user.IsSelfExcluded,
		&user.Currency, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	rows, err := s.db.DB.QueryContext(ctx, `
		SELECT id, email, username, password_hash, kyc_status, is_active, is_banned, is_self_excluded, currency, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash,
			&u.KYCStatus, &u.IsActive, &u.IsBanned, &u.IsSelfExcluded,
			&u.Currency, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (s *UserService) BanUser(ctx context.Context, userID uuid.UUID) error {
	res, err := s.db.DB.ExecContext(ctx, `UPDATE users SET is_banned = true, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("ban user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserService) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	rows, err := s.db.DB.QueryContext(ctx, `
		SELECT id, user_id, type, amount, currency, provider_transaction_id, game_id, round_id,
		       cash_before, cash_after, bonus_before, bonus_after, status, created_at
		FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var txns []*models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Currency,
			&t.ProviderTransactionID, &t.GameID, &t.RoundID,
			&t.CashBefore, &t.CashAfter, &t.BonusBefore, &t.BonusAfter,
			&t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, &t)
	}
	return txns, rows.Err()
}

func sessionKey(token string) string {
	return "session:" + token
}

// isUniqueViolation checks if a pq error is a unique constraint violation for the given constraint.
func isUniqueViolation(err error, constraint string) bool {
	return err != nil && (fmt.Sprintf("%v", err) == fmt.Sprintf("ERROR: duplicate key value violates unique constraint \"%s\" (SQLSTATE 23505)", constraint) ||
		containsStr(err.Error(), constraint))
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
