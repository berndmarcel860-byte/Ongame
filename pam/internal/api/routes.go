package api

import (
	"github.com/berndmarcel860-byte/ongame/pam/internal/auth"
	"github.com/berndmarcel860-byte/ongame/pam/internal/service"
	"github.com/gin-gonic/gin"
)

// RouterConfig holds all dependencies for the HTTP router.
type RouterConfig struct {
	UserService    *service.UserService
	BalanceService *service.BalanceService
	GameService    *service.GameService
	JWTManager     *auth.JWTManager
	PAMAPIToken    string
}

// Router wraps gin.Engine with its dependencies.
type Router struct {
	engine     *gin.Engine
	userSvc    *service.UserService
	balanceSvc *service.BalanceService
	gameSvc    *service.GameService
}

// NewRouter constructs and returns a configured gin.Engine.
func NewRouter(cfg RouterConfig) *gin.Engine {
	r := &Router{
		engine:     gin.New(),
		userSvc:    cfg.UserService,
		balanceSvc: cfg.BalanceService,
		gameSvc:    cfg.GameService,
	}
	r.engine.Use(loggingMiddleware(), recoveryMiddleware())
	r.registerRoutes(cfg)
	return r.engine
}

func (r *Router) registerRoutes(cfg RouterConfig) {
	e := r.engine

	// Public auth routes
	authGroup := e.Group("/auth")
	{
		authGroup.POST("/register", r.handleRegister)
		authGroup.POST("/login", r.handleLogin)
	}

	// Authenticated player routes
	userGroup := e.Group("/users", jwtAuthMiddleware(cfg.JWTManager))
	{
		userGroup.GET("/me", r.handleGetMe)
		userGroup.GET("/me/transactions", r.handleGetMyTransactions)
	}

	// PAM endpoints (Valkyrie-to-PAM) - protected by API token
	pamGroup := e.Group("/players", pamAuthMiddleware(cfg.PAMAPIToken))
	{
		pamGroup.GET("/:userId/balance", r.handleGetPlayerBalance)
		pamGroup.POST("/:userId/transactions", r.handlePlayerTransaction)
		pamGroup.POST("/:userId/sessions", r.handleCreateSession)
		pamGroup.DELETE("/:userId/sessions/:sessionId", r.handleEndSession)
	}

	// Admin routes - JWT required (role check can be added later)
	adminGroup := e.Group("/admin", jwtAuthMiddleware(cfg.JWTManager))
	{
		adminGroup.GET("/users", r.handleAdminListUsers)
		adminGroup.GET("/users/:userId", r.handleAdminGetUser)
		adminGroup.PUT("/users/:userId/balance", r.handleAdminAdjustBalance)
		adminGroup.POST("/users/:userId/ban", r.handleAdminBanUser)
	}

	// Health check
	e.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
