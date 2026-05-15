package service

import (
	"context"
	"fmt"
	"time"

	"github.com/berndmarcel860-byte/ongame/pam/internal/cache"
	"github.com/berndmarcel860-byte/ongame/pam/internal/db"
	"github.com/berndmarcel860-byte/ongame/pam/internal/models"
	"github.com/google/uuid"
)

type GameService struct {
	db    *db.Postgres
	cache *cache.Redis
}

func NewGameService(database *db.Postgres, redisClient *cache.Redis) *GameService {
	return &GameService{db: database, cache: redisClient}
}

// CreateSession creates a game session and returns a session token.
func (s *GameService) CreateSession(ctx context.Context, userID uuid.UUID, req *models.CreateSessionRequest) (*models.SessionResponse, error) {
	var sessionID uuid.UUID
	err := s.db.DB.QueryRowContext(ctx, `
		INSERT INTO game_sessions (user_id, game_id, provider)
		VALUES ($1, $2, $3)
		RETURNING id`,
		userID, req.GameID, req.Provider,
	).Scan(&sessionID)
	if err != nil {
		return nil, fmt.Errorf("create game session: %w", err)
	}

	sessionToken := uuid.New().String()
	// Cache: token -> userID for fast PAM auth lookups.
	_ = s.cache.Set(ctx, gameSessionKey(sessionToken), userID.String(), 12*time.Hour)
	// Cache: token -> sessionID
	_ = s.cache.Set(ctx, gameSessionIDKey(sessionToken), sessionID.String(), 12*time.Hour)

	return &models.SessionResponse{
		SessionToken: sessionToken,
		SessionID:    sessionID.String(),
	}, nil
}

// EndSession marks a game session as ended.
func (s *GameService) EndSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	res, err := s.db.DB.ExecContext(ctx, `
		UPDATE game_sessions SET ended_at = NOW()
		WHERE id = $1 AND user_id = $2 AND ended_at IS NULL`,
		sessionID, userID)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session not found or already ended")
	}

	// Invalidate session token cache entries for this session.
	_ = s.cache.Del(ctx, gameSessionIDKey(sessionID.String()))
	return nil
}

// ResolvePlayerToken looks up the player UUID from a session token.
func (s *GameService) ResolvePlayerToken(ctx context.Context, sessionToken string) (uuid.UUID, error) {
	userIDStr, err := s.cache.Get(ctx, gameSessionKey(sessionToken))
	if err != nil {
		// Fallback to DB
		return s.resolveFromDB(ctx, sessionToken)
	}
	return uuid.Parse(userIDStr)
}

func (s *GameService) resolveFromDB(ctx context.Context, sessionToken string) (uuid.UUID, error) {
	// Session token is stored only in Redis; if not found, session is invalid/expired.
	return uuid.Nil, fmt.Errorf("session token not found or expired")
}

func gameSessionKey(token string) string {
	return "game_session:user:" + token
}

func gameSessionIDKey(token string) string {
	return "game_session:id:" + token
}
