package api

import (
	"net/http"
	"strings"

	"github.com/berndmarcel860-byte/ongame/pam/internal/auth"
	"github.com/gin-gonic/gin"
)

// pamAuthMiddleware validates the static X-Api-Token header used by Valkyrie.
func pamAuthMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.GetHeader("X-Api-Token")
		if t == "" || t != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API token"})
			return
		}
		c.Next()
	}
}

// jwtAuthMiddleware validates Bearer JWT tokens for player/admin endpoints.
func jwtAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}
		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// loggingMiddleware logs method, path and status.
func loggingMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// recoveryMiddleware recovers from panics.
func recoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}
